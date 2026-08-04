<script setup lang="ts">
import {
  LngLatBounds,
  Map,
  Marker,
  NavigationControl,
  type MapMouseEvent,
  type StyleSpecification,
} from 'maplibre-gl'
import 'maplibre-gl/dist/maplibre-gl.css'
import { NAlert, NButton, NInput, NSelect, NSpin, NTag, NTooltip } from 'naive-ui'
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import { useRouter } from 'vue-router'

import { listSysConfigs } from '@/services/sysconfig'

import type { FootprintPlace } from '@/services/footprints'
import type { SysConfigGroup, SysConfigItem } from '@/services/sysconfig'

type SelectedMapPlace = {
  cityName?: string
  regionName?: string
  countryName?: string
  countryCode?: string
  latitude: number
  longitude: number
}

type NominatimAddress = {
  borough?: string
  city?: string
  city_district?: string
  country?: string
  country_code?: string
  county?: string
  district?: string
  hamlet?: string
  municipality?: string
  province?: string
  region?: string
  state?: string
  suburb?: string
  town?: string
  village?: string
}

type NominatimResult = {
  address?: NominatimAddress
  boundingbox?: [string, string, string, string]
  display_name: string
  lat: string
  lon: string
  name?: string
  place_id: number
}

type MapProvider = 'osm' | 'tianditu'
type TiandituLayer = 'vector' | 'imagery'

type MapSettings = {
  provider: MapProvider
  tiandituKey: string
  tiandituLayer: TiandituLayer
}

const DEFAULT_CENTER: [number, number] = [104.1954, 35.8617]
const DEFAULT_ZOOM = 3
const OSM_TILE_URL =
  import.meta.env.VITE_OSM_TILE_URL?.trim() || 'https://tile.openstreetmap.org/{z}/{x}/{y}.png'
const NOMINATIM_BASE_URL = (
  import.meta.env.VITE_NOMINATIM_BASE_URL?.trim() || 'https://nominatim.openstreetmap.org'
).replace(/\/$/, '')
const REQUEST_INTERVAL_MS = 1000
const MAP_CONFIG_KEYS = ['map.provider', 'map.tianditu.key', 'map.tianditu.layer']

let nominatimQueue = Promise.resolve()
let lastNominatimRequestAt = 0
const nominatimCache = new globalThis.Map<string, unknown>()

const props = defineProps<{
  latitude: number
  longitude: number
  places: FootprintPlace[]
  selectedPlaceId: number | null
}>()

const router = useRouter()

const emit = defineEmits<{
  pick: [coordinates: { latitude: number; longitude: number }]
  resolvePlace: [place: SelectedMapPlace]
  selectPlace: [id: number]
}>()

const mapEl = ref<HTMLDivElement | null>(null)
const loading = ref(false)
const resolving = ref(false)
const searching = ref(false)
const errorMessage = ref('')
const searchErrorMessage = ref('')
const searchQuery = ref('')
const searchResults = ref<NominatimResult[]>([])
const selectedSearchResult = ref<string | null>(null)
const mapProvider = ref<MapProvider>('osm')

let map: Map | null = null
let currentMarker: Marker | null = null
let placeMarkers: Marker[] = []
let resolveRequestId = 0
let destroyed = false
let mapSettings: MapSettings = {
  provider: 'osm',
  tiandituKey: '',
  tiandituLayer: 'vector',
}

const hasCoordinate = computed(
  () => props.selectedPlaceId !== null || props.latitude !== 0 || props.longitude !== 0,
)

const coordinateLabel = computed(() => ({
  latitude: Number(props.latitude.toFixed(6)),
  longitude: Number(props.longitude.toFixed(6)),
}))

const searchOptions = computed(() =>
  searchResults.value.map((result) => ({
    label: result.display_name,
    value: String(result.place_id),
  })),
)

function collectConfigItems(groups: SysConfigGroup[]): SysConfigItem[] {
  return groups.flatMap((group) => [
    ...(group.items ?? []),
    ...collectConfigItems(group.children ?? []),
  ])
}

async function loadMapSettings() {
  const tree = await listSysConfigs(MAP_CONFIG_KEYS)
  const items = [...(tree.items ?? []), ...collectConfigItems(tree.groups ?? [])]
  const value = (key: string) => items.find((item) => item.key === key)?.value
  const provider = value('map.provider')
  const tiandituLayer = value('map.tianditu.layer')
  mapSettings = {
    provider: provider === 'tianditu' ? 'tianditu' : 'osm',
    tiandituKey:
      typeof value('map.tianditu.key') === 'string' ? String(value('map.tianditu.key')) : '',
    tiandituLayer: tiandituLayer === 'imagery' ? 'imagery' : 'vector',
  }
  mapProvider.value = mapSettings.provider
}

function tiandituTiles(layer: string, key: string) {
  const query =
    `SERVICE=WMTS&REQUEST=GetTile&VERSION=1.0.0&LAYER=${layer}` +
    `&STYLE=default&TILEMATRIXSET=w&FORMAT=tiles&TILEMATRIX={z}` +
    `&TILEROW={y}&TILECOL={x}&tk=${encodeURIComponent(key)}`
  return Array.from(
    { length: 8 },
    (_, index) => `https://t${index}.tianditu.gov.cn/${layer}_w/wmts?${query}`,
  )
}

function createMapStyle(settings: MapSettings): StyleSpecification {
  if (settings.provider === 'tianditu') {
    const baseLayer = settings.tiandituLayer === 'imagery' ? 'img' : 'vec'
    const labelLayer = settings.tiandituLayer === 'imagery' ? 'cia' : 'cva'
    return {
      version: 8,
      sources: {
        'tianditu-base': {
          type: 'raster',
          tiles: tiandituTiles(baseLayer, settings.tiandituKey),
          tileSize: 256,
          attribution: '&copy; <a href="https://www.tianditu.gov.cn/">天地图</a>',
        },
        'tianditu-label': {
          type: 'raster',
          tiles: tiandituTiles(labelLayer, settings.tiandituKey),
          tileSize: 256,
        },
      },
      layers: [
        { id: 'tianditu-base', type: 'raster', source: 'tianditu-base' },
        { id: 'tianditu-label', type: 'raster', source: 'tianditu-label' },
      ],
    }
  }

  return {
    version: 8,
    sources: {
      osm: {
        type: 'raster',
        tiles: [OSM_TILE_URL],
        tileSize: 256,
        attribution:
          '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap contributors</a>',
      },
    },
    layers: [{ id: 'osm', type: 'raster', source: 'osm' }],
  }
}

function sleep(duration: number) {
  return new Promise((resolve) => window.setTimeout(resolve, duration))
}

function nominatimRequest<T>(url: URL): Promise<T> {
  const cacheKey = url.toString()
  if (nominatimCache.has(cacheKey)) return Promise.resolve(nominatimCache.get(cacheKey) as T)

  const request = nominatimQueue.then(async () => {
    const wait = Math.max(0, REQUEST_INTERVAL_MS - (Date.now() - lastNominatimRequestAt))
    if (wait > 0) await sleep(wait)
    lastNominatimRequestAt = Date.now()
    const response = await fetch(url, { headers: { Accept: 'application/json' } })
    if (!response.ok) throw new Error(`Nominatim 请求失败 (${response.status})`)
    const data = (await response.json()) as T
    nominatimCache.set(cacheKey, data)
    return data
  })

  nominatimQueue = request.then(
    () => undefined,
    () => undefined,
  )
  return request
}

function normalizedCoordinates(longitude: number, latitude: number) {
  return {
    latitude: Number(latitude.toFixed(6)),
    longitude: Number(longitude.toFixed(6)),
  }
}

function placeDetails(result: NominatimResult, latitude: number, longitude: number) {
  const address = result.address ?? {}
  const locality =
    address.city ||
    address.town ||
    address.village ||
    address.municipality ||
    address.city_district ||
    address.district ||
    address.county ||
    address.borough ||
    address.suburb ||
    address.hamlet ||
    result.name ||
    result.display_name.split(',')[0]?.trim()
  const region = address.state || address.province || address.region
  return {
    cityName: locality,
    regionName: region === locality ? undefined : region,
    countryName: address.country,
    countryCode: address.country_code?.toUpperCase(),
    latitude,
    longitude,
  }
}

function emitCoordinates(longitude: number, latitude: number) {
  const coordinates = normalizedCoordinates(longitude, latitude)
  emit('pick', coordinates)
  return coordinates
}

async function resolveCoordinates(latitude: number, longitude: number, requestId: number) {
  resolving.value = true
  searchErrorMessage.value = ''
  try {
    const url = new URL(`${NOMINATIM_BASE_URL}/reverse`)
    url.searchParams.set('format', 'jsonv2')
    url.searchParams.set('addressdetails', '1')
    url.searchParams.set('accept-language', 'zh-CN')
    url.searchParams.set('lat', String(latitude))
    url.searchParams.set('lon', String(longitude))
    const result = await nominatimRequest<NominatimResult>(url)
    if (destroyed || requestId !== resolveRequestId) return
    emit('resolvePlace', placeDetails(result, latitude, longitude))
  } catch {
    if (requestId === resolveRequestId) {
      searchErrorMessage.value = '坐标已更新，但地点信息解析失败'
    }
  } finally {
    if (requestId === resolveRequestId) resolving.value = false
  }
}

function selectCoordinates(longitude: number, latitude: number) {
  const requestId = ++resolveRequestId
  const coordinates = emitCoordinates(longitude, latitude)
  void resolveCoordinates(coordinates.latitude, coordinates.longitude, requestId)
}

function clearMarkers() {
  currentMarker?.remove()
  currentMarker = null
  placeMarkers.forEach((marker) => marker.remove())
  placeMarkers = []
}

function renderMarkers() {
  if (!map) return
  clearMarkers()

  placeMarkers = props.places.map((place) => {
    const isSelected = place.id === props.selectedPlaceId
    const marker = new Marker({
      color: isSelected ? '#18a058' : '#737373',
      scale: isSelected ? 0.9 : 0.7,
    })
      .setLngLat([place.longitude, place.latitude])
      .addTo(map!)
    const element = marker.getElement()
    element.title = place.cityName
    element.addEventListener('click', (event) => {
      event.stopPropagation()
      emit('selectPlace', place.id)
    })
    return marker
  })

  if (!hasCoordinate.value || props.selectedPlaceId !== null) return
  currentMarker = new Marker({ color: '#2080f0', draggable: true, scale: 0.95 })
    .setLngLat([props.longitude, props.latitude])
    .addTo(map)
  currentMarker.getElement().title = '当前选点，可拖动调整'
  currentMarker.on('dragend', () => {
    const position = currentMarker?.getLngLat()
    if (position) selectCoordinates(position.lng, position.lat)
  })
}

async function searchLocation() {
  const query = searchQuery.value.trim()
  if (!query) return
  searching.value = true
  searchErrorMessage.value = ''
  selectedSearchResult.value = null
  try {
    const url = new URL(`${NOMINATIM_BASE_URL}/search`)
    url.searchParams.set('format', 'jsonv2')
    url.searchParams.set('addressdetails', '1')
    url.searchParams.set('accept-language', 'zh-CN')
    url.searchParams.set('limit', '5')
    url.searchParams.set('q', query)
    searchResults.value = await nominatimRequest<NominatimResult[]>(url)
    if (searchResults.value.length === 0) searchErrorMessage.value = '没有找到匹配地点'
  } catch (error) {
    searchErrorMessage.value = error instanceof Error ? error.message : '地点搜索失败'
  } finally {
    searching.value = false
  }
}

function selectSearchResult(value: string | null) {
  selectedSearchResult.value = value
  const result = searchResults.value.find((item) => String(item.place_id) === value)
  if (!result) return
  const longitude = Number(result.lon)
  const latitude = Number(result.lat)
  if (!Number.isFinite(longitude) || !Number.isFinite(latitude)) return

  resolveRequestId += 1
  const coordinates = emitCoordinates(longitude, latitude)
  emit('resolvePlace', placeDetails(result, coordinates.latitude, coordinates.longitude))
  if (!map) return

  if (result.boundingbox) {
    const [south, north, west, east] = result.boundingbox.map(Number)
    if ([south, north, west, east].every(Number.isFinite)) {
      map.fitBounds(
        [
          [west, south],
          [east, north],
        ],
        { padding: 48, maxZoom: 13 },
      )
      return
    }
  }
  map.flyTo({ center: [longitude, latitude], zoom: 11 })
}

async function initializeMap() {
  if (!mapEl.value) return
  loading.value = true
  errorMessage.value = ''
  try {
    await loadMapSettings()
    if (mapSettings.provider === 'tianditu' && !mapSettings.tiandituKey.trim()) {
      errorMessage.value = '已选择天地图，请先配置天地图 Key'
      loading.value = false
      return
    }
    map = new Map({
      container: mapEl.value,
      style: createMapStyle(mapSettings),
      center: hasCoordinate.value ? [props.longitude, props.latitude] : DEFAULT_CENTER,
      zoom: hasCoordinate.value ? 8 : DEFAULT_ZOOM,
    })
    map.addControl(new NavigationControl({ showCompass: false }), 'bottom-right')
    map.on('click', (event: MapMouseEvent) => {
      selectCoordinates(event.lngLat.lng, event.lngLat.lat)
    })
    map.on('load', () => {
      loading.value = false
      renderMarkers()
    })
    map.on('error', () => {
      searchErrorMessage.value = '部分地图瓦片加载失败，请检查网络或瓦片服务地址'
      loading.value = false
    })
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : '地图初始化失败'
    loading.value = false
  }
}

function openMapSettings() {
  void router.push({ name: 'settings', query: { tab: 'map' } })
}

function focusCurrent() {
  if (!map || !hasCoordinate.value) return
  map.flyTo({ center: [props.longitude, props.latitude], zoom: 12 })
}

function resetView() {
  if (!map) return
  const coordinates = props.places.map(
    (place) => [place.longitude, place.latitude] as [number, number],
  )
  if (hasCoordinate.value) coordinates.push([props.longitude, props.latitude])
  if (coordinates.length === 0) {
    map.flyTo({ center: DEFAULT_CENTER, zoom: DEFAULT_ZOOM })
    return
  }
  if (coordinates.length === 1) {
    map.flyTo({ center: coordinates[0], zoom: 8 })
    return
  }
  const bounds = coordinates.reduce(
    (result, coordinate) => result.extend(coordinate),
    new LngLatBounds(coordinates[0], coordinates[0]),
  )
  map.fitBounds(bounds, { padding: 48, maxZoom: 9 })
}

watch(
  () => [props.latitude, props.longitude, props.selectedPlaceId, props.places] as const,
  () => nextTick(renderMarkers),
  { deep: true },
)

onMounted(() => nextTick(initializeMap))

onUnmounted(() => {
  destroyed = true
  resolveRequestId += 1
  clearMarkers()
  map?.remove()
  map = null
})
</script>

<template>
  <div class="w-full overflow-hidden rounded border border-naive-border bg-naive-card">
    <div
      class="flex min-h-11 flex-wrap items-center justify-between gap-2 border-b border-naive-border px-3 py-2"
    >
      <div class="flex flex-wrap items-center gap-2">
        <NTag
          size="small"
          :bordered="false"
        >
          经度 {{ coordinateLabel.longitude }}
        </NTag>
        <NTag
          size="small"
          :bordered="false"
        >
          纬度 {{ coordinateLabel.latitude }}
        </NTag>
        <span class="text-xs opacity-45">{{ places.length }} 个已有城市</span>
        <NTag
          size="small"
          :bordered="false"
          type="info"
        >
          {{ mapProvider === 'tianditu' ? '天地图' : 'OpenStreetMap' }}
        </NTag>
        <span
          v-if="resolving"
          class="text-xs opacity-55"
          >正在解析地点...</span
        >
      </div>
      <div class="flex items-center gap-1">
        <NTooltip>
          <template #trigger>
            <NButton
              quaternary
              circle
              size="small"
              :disabled="!hasCoordinate"
              @click="focusCurrent"
            >
              <template #icon><div class="iconify ph--crosshair" /></template>
            </NButton>
          </template>
          定位当前选点
        </NTooltip>
        <NTooltip>
          <template #trigger>
            <NButton
              quaternary
              circle
              size="small"
              @click="resetView"
            >
              <template #icon><div class="iconify ph--arrows-out" /></template>
            </NButton>
          </template>
          显示全部足迹
        </NTooltip>
      </div>
    </div>
    <div class="space-y-2 border-b border-naive-border p-3">
      <div class="flex gap-2">
        <NInput
          v-model:value="searchQuery"
          clearable
          placeholder="搜索城市、景区或徒步地点"
          @keyup.enter="searchLocation"
        />
        <NButton
          type="primary"
          :loading="searching"
          :disabled="!searchQuery.trim()"
          @click="searchLocation"
        >
          <template #icon><div class="iconify ph--magnifying-glass" /></template>
          搜索
        </NButton>
      </div>
      <NSelect
        v-if="searchOptions.length"
        :value="selectedSearchResult"
        :options="searchOptions"
        placeholder="选择匹配地点"
        @update:value="selectSearchResult"
      />
      <NAlert
        v-if="searchErrorMessage"
        type="warning"
        :show-icon="false"
      >
        {{ searchErrorMessage }}
      </NAlert>
    </div>
    <NAlert
      v-if="errorMessage"
      class="m-3"
      type="error"
      title="地图不可用"
    >
      {{ errorMessage }}
      <div
        v-if="mapProvider === 'tianditu'"
        class="mt-2"
      >
        <NButton
          size="small"
          @click="openMapSettings"
          >打开地图设置</NButton
        >
      </div>
    </NAlert>
    <NSpin
      v-else
      :show="loading"
    >
      <div
        ref="mapEl"
        class="h-[400px] w-full"
      />
    </NSpin>
  </div>
</template>

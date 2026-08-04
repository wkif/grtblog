<script setup lang="ts">
import {
  NButton,
  NCard,
  NDatePicker,
  NForm,
  NFormItemGi,
  NGrid,
  NInput,
  NInputNumber,
  NSelect,
  NSpace,
  NSwitch,
} from 'naive-ui'
import { computed, reactive, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import { ScrollContainer } from '@/components'
import { useDiscreteApi } from '@/composables/useDiscreteApi'
import { listAlbums } from '@/services/albums'
import {
  createFootprint,
  getFootprint,
  listFootprintPlaces,
  updateFootprint,
} from '@/services/footprints'

import CoordinatePicker from './components/CoordinatePicker.vue'

import type { AlbumListItem } from '@/services/albums'
import type { FootprintJourneyPayload, FootprintPlace } from '@/services/footprints'

const route = useRoute()
const router = useRouter()
const { message } = useDiscreteApi()
const footprintId = computed(() => {
  const value = Number(route.params.id)
  return Number.isInteger(value) && value > 0 ? value : null
})
const isEdit = computed(() => footprintId.value !== null)
const loading = ref(false)
const places = ref<FootprintPlace[]>([])
const albums = ref<AlbumListItem[]>([])
const selectedPlaceId = ref<number | null>(null)
const generatedSlug = ref('')
const journeyDateValue = ref<number | null>(Date.now())
const endedAtValue = ref<number | null>(null)
const distanceKm = ref<number | null>(null)
const durationHours = ref<number | null>(null)
const durationMinutes = ref<number | null>(null)

const form = reactive<FootprintJourneyPayload>({
  place: {
    slug: '',
    cityName: '',
    regionName: null,
    countryName: '中国',
    countryCode: 'CN',
    latitude: 0,
    longitude: 0,
  },
  title: '',
  journeyDate: '',
  endedAt: null,
  summary: null,
  cover: null,
  distanceMeters: null,
  durationSeconds: null,
  trackUrl: null,
  albumIds: [],
  isPublished: true,
  sortOrder: 0,
})

const placeOptions = computed(() =>
  places.value.map((place) => ({
    label: `${place.cityName}${place.regionName ? ` · ${place.regionName}` : ''} · ${place.countryName}`,
    value: place.id,
  })),
)

const albumOptions = computed(() =>
  albums.value.map((album) => ({
    label: `${album.title}${album.isPublished ? '' : ' （未发布）'}`,
    value: album.id,
  })),
)

function dateISO(value: number | null) {
  if (value == null) return null
  const date = new Date(value)
  return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')}T00:00:00Z`
}

function resetForm() {
  Object.assign(form, {
    place: {
      slug: '',
      cityName: '',
      regionName: null,
      countryName: '中国',
      countryCode: 'CN',
      latitude: 0,
      longitude: 0,
    },
    title: '',
    journeyDate: '',
    endedAt: null,
    summary: null,
    cover: null,
    distanceMeters: null,
    durationSeconds: null,
    trackUrl: null,
    albumIds: [],
    isPublished: true,
    sortOrder: 0,
  })
  selectedPlaceId.value = null
  generatedSlug.value = ''
  journeyDateValue.value = Date.now()
  endedAtValue.value = null
  distanceKm.value = null
  durationHours.value = null
  durationMinutes.value = null
}

function placeSlug(cityName: string, latitude: number, longitude: number) {
  const city = cityName.trim().toLowerCase()
  if (!city) return ''
  const normalized = city
    .normalize('NFKD')
    .replace(/[\u0300-\u036f]/g, '')
    .replace(/[^\p{L}\p{N}]+/gu, '-')
    .replace(/^-+|-+$/g, '')
  if (normalized) return normalized
  const coordinate = `${latitude.toFixed(4)}-${longitude.toFixed(4)}`
  return `place-${coordinate.replace(/[^\d-]/g, '')}`
}

type PlaceDraft = Omit<FootprintPlace, 'id' | 'createdAt' | 'updatedAt'>

function applyPlace(place: PlaceDraft, options: { id?: number | null; autoSlug?: boolean } = {}) {
  const nextSlug = placeSlug(place.cityName, place.latitude, place.longitude)
  const shouldUpdateSlug =
    options.autoSlug !== false &&
    (!form.place.slug.trim() || form.place.slug === generatedSlug.value)

  Object.assign(form.place, {
    slug: place.slug,
    cityName: place.cityName,
    regionName: place.regionName ?? null,
    countryName: place.countryName,
    countryCode: place.countryCode ?? null,
    latitude: place.latitude,
    longitude: place.longitude,
  })
  if (shouldUpdateSlug) form.place.slug = place.slug || nextSlug
  generatedSlug.value = form.place.slug
  selectedPlaceId.value = options.id ?? null
}

function selectPlace(id: number | string | null) {
  const normalizedId = id == null || id === '' ? null : Number(id)
  const place = places.value.find((item) => item.id === normalizedId)
  if (!place) {
    selectedPlaceId.value = null
    return
  }
  applyPlace(place, { id: place.id, autoSlug: true })
}

function newPlace() {
  form.place.slug = ''
  form.place.cityName = ''
  form.place.regionName = null
  form.place.countryName = '中国'
  form.place.countryCode = 'CN'
  form.place.latitude = 0
  form.place.longitude = 0
  selectedPlaceId.value = null
  generatedSlug.value = ''
}

function updateCityName(value: string) {
  const shouldUpdateSlug = !form.place.slug.trim() || form.place.slug === generatedSlug.value
  form.place.cityName = value
  selectedPlaceId.value = null
  if (shouldUpdateSlug) {
    form.place.slug = placeSlug(value, form.place.latitude, form.place.longitude)
    generatedSlug.value = form.place.slug
  }
}

function updateSlug(value: string) {
  form.place.slug = value
  selectedPlaceId.value = null
  generatedSlug.value = ''
}

function detachPlace() {
  selectedPlaceId.value = null
}

function pickCoordinates(coordinates: { latitude: number; longitude: number }) {
  selectedPlaceId.value = null
  form.place.latitude = coordinates.latitude
  form.place.longitude = coordinates.longitude
}

function resolveMapPlace(place: {
  cityName?: string
  regionName?: string
  countryName?: string
  countryCode?: string
  latitude: number
  longitude: number
}) {
  const nextCityName = place.cityName?.trim() || form.place.cityName
  const nextRegionName = place.regionName?.trim() || form.place.regionName
  const nextCountryName = place.countryName?.trim() || form.place.countryName
  const nextCountryCode = place.countryCode?.trim() || form.place.countryCode
  const nextSlug = placeSlug(nextCityName, place.latitude, place.longitude)

  selectedPlaceId.value = null
  form.place.cityName = nextCityName
  form.place.regionName = nextRegionName
  form.place.countryName = nextCountryName
  form.place.countryCode = nextCountryCode
  form.place.latitude = place.latitude
  form.place.longitude = place.longitude
  if (!form.place.slug.trim() || form.place.slug === generatedSlug.value) {
    form.place.slug = nextSlug
  }
  generatedSlug.value = form.place.slug
}

let optionsPromise: Promise<void> | null = null
let loadRequestId = 0

function loadOptions() {
  if (optionsPromise) return optionsPromise
  optionsPromise = Promise.all([listFootprintPlaces(), listAlbums({ page: 1, pageSize: 100 })])
    .then(([placeItems, albumResult]) => {
      places.value = placeItems
      albums.value = albumResult.items
    })
    .catch((error) => {
      optionsPromise = null
      throw error
    })
  return optionsPromise
}

async function load() {
  const requestId = ++loadRequestId
  const id = footprintId.value
  loading.value = true
  try {
    resetForm()
    await loadOptions()
    if (requestId !== loadRequestId || !id) return
    const record = await getFootprint(id)
    if (requestId !== loadRequestId) return
    Object.assign(form, {
      title: record.title,
      journeyDate: record.journeyDate,
      endedAt: record.endedAt ?? null,
      summary: record.summary ?? null,
      cover: record.cover ?? null,
      distanceMeters: record.distanceMeters ?? null,
      durationSeconds: record.durationSeconds ?? null,
      trackUrl: record.trackUrl ?? null,
      albumIds: record.albums.map((album) => album.id),
      isPublished: record.isPublished,
      sortOrder: record.sortOrder,
    })
    applyPlace(record.place, { id: record.place.id, autoSlug: false })
    generatedSlug.value = record.place.slug
    journeyDateValue.value = Date.parse(record.journeyDate)
    endedAtValue.value = record.endedAt ? Date.parse(record.endedAt) : null
    distanceKm.value = record.distanceMeters == null ? null : record.distanceMeters / 1000
    if (record.durationSeconds == null) {
      durationHours.value = null
      durationMinutes.value = null
    } else {
      const totalMinutes = Math.round(record.durationSeconds / 60)
      durationHours.value = Math.floor(totalMinutes / 60)
      durationMinutes.value = totalMinutes % 60
    }
  } catch (error) {
    if (requestId === loadRequestId) {
      message.error(error instanceof Error ? error.message : '加载足迹失败')
    }
  } finally {
    if (requestId === loadRequestId) loading.value = false
  }
}

async function save() {
  if (!form.title.trim() || !form.place.cityName.trim() || !form.place.slug.trim()) {
    message.warning('请填写行程名称、城市和城市标识')
    return
  }
  if (!journeyDateValue.value) {
    message.warning('请选择行程日期')
    return
  }
  const duration = (durationHours.value ?? 0) * 3600 + (durationMinutes.value ?? 0) * 60
  form.journeyDate = dateISO(journeyDateValue.value)!
  form.endedAt = dateISO(endedAtValue.value)
  form.distanceMeters = distanceKm.value == null ? null : Math.round(distanceKm.value * 1000)
  form.durationSeconds = duration > 0 ? duration : null
  loading.value = true
  try {
    if (footprintId.value) await updateFootprint(footprintId.value, form)
    else await createFootprint(form)
    message.success('足迹记录已保存')
    router.push({ name: 'footprintList' })
  } catch (error) {
    message.error(error instanceof Error ? error.message : '保存失败')
  } finally {
    loading.value = false
  }
}

watch(footprintId, () => void load(), { immediate: true })
</script>

<template>
  <ScrollContainer wrapper-class="pb-8">
    <NCard
      :title="isEdit ? '编辑足迹' : '新增足迹'"
      segmented
    >
      <NForm
        label-placement="left"
        label-width="110"
        :disabled="loading"
      >
        <NGrid
          :cols="2"
          :x-gap="24"
          responsive="screen"
          item-responsive
        >
          <NFormItemGi
            :span="2"
            label="复用城市"
          >
            <NSpace
              class="w-full"
              align="center"
            >
              <NSelect
                class="min-w-80"
                filterable
                clearable
                :value="selectedPlaceId"
                :options="placeOptions"
                placeholder="选择已有城市"
                @update:value="selectPlace"
              />
              <NButton @click="newPlace">新建城市</NButton>
            </NSpace>
          </NFormItemGi>
          <NFormItemGi label="城市"
            ><NInput
              :value="form.place.cityName"
              @update:value="updateCityName"
          /></NFormItemGi>
          <NFormItemGi label="城市标识"
            ><NInput
              :value="form.place.slug"
              placeholder="alxa-tengger"
              @update:value="updateSlug"
          /></NFormItemGi>
          <NFormItemGi label="省/州"
            ><NInput
              v-model:value="form.place.regionName"
              @update:value="detachPlace"
          /></NFormItemGi>
          <NFormItemGi label="国家"
            ><NInput
              v-model:value="form.place.countryName"
              @update:value="detachPlace"
          /></NFormItemGi>
          <NFormItemGi label="国家代码"
            ><NInput
              v-model:value="form.place.countryCode"
              placeholder="CN"
              @update:value="detachPlace"
          /></NFormItemGi>
          <NFormItemGi label="纬度"
            ><NInputNumber
              v-model:value="form.place.latitude"
              :min="-90"
              :max="90"
              :precision="6"
              @update:value="detachPlace"
          /></NFormItemGi>
          <NFormItemGi label="经度"
            ><NInputNumber
              v-model:value="form.place.longitude"
              :min="-180"
              :max="180"
              :precision="6"
              @update:value="detachPlace"
          /></NFormItemGi>
          <NFormItemGi
            :span="2"
            label="地图选点"
          >
            <CoordinatePicker
              :latitude="form.place.latitude"
              :longitude="form.place.longitude"
              :places="places"
              :selected-place-id="selectedPlaceId"
              @pick="pickCoordinates"
              @resolve-place="resolveMapPlace"
              @select-place="selectPlace"
            />
          </NFormItemGi>
          <NFormItemGi
            :span="2"
            label="行程名称"
            ><NInput v-model:value="form.title"
          /></NFormItemGi>
          <NFormItemGi label="开始日期"
            ><NDatePicker
              v-model:value="journeyDateValue"
              type="date"
              clearable
          /></NFormItemGi>
          <NFormItemGi label="结束日期"
            ><NDatePicker
              v-model:value="endedAtValue"
              type="date"
              clearable
          /></NFormItemGi>
          <NFormItemGi label="徒步里程"
            ><NInputNumber
              v-model:value="distanceKm"
              :min="0"
              :precision="2"
              ><template #suffix>km</template></NInputNumber
            ></NFormItemGi
          >
          <NFormItemGi label="徒步时长">
            <NSpace>
              <NInputNumber
                v-model:value="durationHours"
                :min="0"
                :precision="0"
                ><template #suffix>小时</template></NInputNumber
              >
              <NInputNumber
                v-model:value="durationMinutes"
                :min="0"
                :max="59"
                :precision="0"
                ><template #suffix>分钟</template></NInputNumber
              >
            </NSpace>
          </NFormItemGi>
          <NFormItemGi
            :span="2"
            label="轨迹链接"
            ><NInput
              v-model:value="form.trackUrl"
              placeholder="https://..."
          /></NFormItemGi>
          <NFormItemGi
            :span="2"
            label="关联相册"
            ><NSelect
              v-model:value="form.albumIds"
              multiple
              filterable
              :options="albumOptions"
          /></NFormItemGi>
          <NFormItemGi
            :span="2"
            label="封面图"
            ><NInput
              v-model:value="form.cover"
              placeholder="不填时使用关联相册封面"
          /></NFormItemGi>
          <NFormItemGi
            :span="2"
            label="行程记录"
            ><NInput
              v-model:value="form.summary"
              type="textarea"
              :rows="5"
          /></NFormItemGi>
          <NFormItemGi label="排序"
            ><NInputNumber
              v-model:value="form.sortOrder"
              :precision="0"
          /></NFormItemGi>
          <NFormItemGi label="公开展示"><NSwitch v-model:value="form.isPublished" /></NFormItemGi>
        </NGrid>
        <NSpace justify="end"
          ><NButton @click="router.back()">取消</NButton
          ><NButton
            type="primary"
            :loading="loading"
            @click="save"
            >保存行程</NButton
          ></NSpace
        >
      </NForm>
    </NCard>
  </ScrollContainer>
</template>

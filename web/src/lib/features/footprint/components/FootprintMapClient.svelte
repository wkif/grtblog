<script lang="ts">
	import { onMount } from 'svelte';
	import { LocateFixed, RotateCcw } from 'lucide-svelte';
	import {
		LngLatBounds,
		Map as MapLibreMap,
		Marker,
		NavigationControl,
		setWorkerUrl,
		type StyleSpecification
	} from 'maplibre-gl';
	import 'maplibre-gl/dist/maplibre-gl.css';
	import maplibreWorkerUrl from 'maplibre-gl/dist/maplibre-gl-worker.mjs?worker&url';
	import type { FootprintMapSettings, FootprintPlaceGroup } from '../types';

	let {
		places,
		mapSettings,
		selectedSlug,
		onSelect
	}: {
		places: FootprintPlaceGroup[];
		mapSettings: FootprintMapSettings;
		selectedSlug: string | null;
		onSelect: (slug: string) => void;
	} = $props();

	const DEFAULT_CENTER: [number, number] = [104.1954, 35.8617];
	const DEFAULT_ZOOM = 2.5;
	const OSM_TILE_URL = 'https://tile.openstreetmap.org/{z}/{x}/{y}.png';

	let mapElement: HTMLDivElement;
	let map: MapLibreMap | null = null;
	let mapReady = $state(false);
	let loadError = $state('');
	const markers: Record<string, Marker> = {};

	function tiandituTiles(layer: string, key: string) {
		const query =
			`SERVICE=WMTS&REQUEST=GetTile&VERSION=1.0.0&LAYER=${layer}` +
			`&STYLE=default&TILEMATRIXSET=w&FORMAT=tiles&TILEMATRIX={z}` +
			`&TILEROW={y}&TILECOL={x}&tk=${encodeURIComponent(key)}`;
		return Array.from(
			{ length: 8 },
			(_, index) => `https://t${index}.tianditu.gov.cn/${layer}_w/wmts?${query}`
		);
	}

	function createMapStyle(settings: FootprintMapSettings): StyleSpecification {
		if (settings.provider === 'tianditu') {
			if (!settings.tiandituKey.trim()) {
				throw new Error('后台已选择天地图，请先在地图设置中配置 Key');
			}
			const baseLayer = settings.tiandituLayer === 'imagery' ? 'img' : 'vec';
			const labelLayer = settings.tiandituLayer === 'imagery' ? 'cia' : 'cva';
			return {
				version: 8,
				sources: {
					'tianditu-base': {
						type: 'raster',
						tiles: tiandituTiles(baseLayer, settings.tiandituKey),
						tileSize: 256,
						attribution: '&copy; 天地图'
					},
					'tianditu-label': {
						type: 'raster',
						tiles: tiandituTiles(labelLayer, settings.tiandituKey),
						tileSize: 256
					}
				},
				layers: [
					{ id: 'tianditu-base', type: 'raster', source: 'tianditu-base' },
					{ id: 'tianditu-label', type: 'raster', source: 'tianditu-label' }
				]
			};
		}

		return {
			version: 8,
			sources: {
				osm: {
					type: 'raster',
					tiles: [OSM_TILE_URL],
					tileSize: 256,
					attribution: '&copy; OpenStreetMap contributors'
				}
			},
			layers: [{ id: 'osm', type: 'raster', source: 'osm' }]
		};
	}

	function markerKey(place: FootprintPlaceGroup, index: number) {
		return `${place.id}:${place.slug}:${index}`;
	}

	function renderMarkers(nextPlaces: FootprintPlaceGroup[], nextSelectedSlug: string | null) {
		const currentMap = map;
		if (!currentMap || !mapReady) return;
		const activeKeys: string[] = [];

		nextPlaces.forEach((place, index) => {
			const key = markerKey(place, index);
			activeKeys.push(key);
			let marker = markers[key];
			if (!marker) {
				const element = document.createElement('button');
				element.type = 'button';
				element.className = 'footprint-map-marker';
				element.setAttribute('aria-label', `${place.cityName}，${place.stats.journeyCount} 次行程`);
				const dot = document.createElement('span');
				dot.className = 'footprint-map-marker__dot';
				const label = document.createElement('span');
				label.className = 'footprint-map-marker__label';
				label.textContent = place.cityName;
				element.append(dot, label);
				element.addEventListener('click', () => onSelect(place.slug));
				marker = new Marker({ element, anchor: 'center' })
					.setLngLat([place.longitude, place.latitude])
					.addTo(currentMap);
				markers[key] = marker;
			}
			marker.getElement().dataset.selected = String(place.slug === nextSelectedSlug);
		});

		for (const [key, marker] of Object.entries(markers)) {
			if (!activeKeys.includes(key)) {
				marker.remove();
				delete markers[key];
			}
		}
	}

	function resetView() {
		if (!map) return;
		if (places.length === 0) {
			map.flyTo({ center: DEFAULT_CENTER, zoom: DEFAULT_ZOOM });
			return;
		}
		if (places.length === 1) {
			map.flyTo({ center: [places[0].longitude, places[0].latitude], zoom: 7 });
			return;
		}
		const bounds = places.reduce(
			(result, place) => result.extend([place.longitude, place.latitude]),
			new LngLatBounds(
				[places[0].longitude, places[0].latitude],
				[places[0].longitude, places[0].latitude]
			)
		);
		map.fitBounds(bounds, { padding: 72, maxZoom: 7, duration: 500 });
	}

	$effect(() => {
		renderMarkers(places, selectedSlug);
	});

	onMount(() => {
		let loadTimer: number | undefined;
		let resizeObserver: ResizeObserver | undefined;
		try {
			setWorkerUrl(maplibreWorkerUrl);
			map = new MapLibreMap({
				container: mapElement,
				style: createMapStyle(mapSettings),
				center: DEFAULT_CENTER,
				zoom: DEFAULT_ZOOM
			});
			resizeObserver = new ResizeObserver(() => map?.resize());
			resizeObserver.observe(mapElement);
			map.addControl(new NavigationControl({ showCompass: false }), 'bottom-right');
			loadTimer = window.setTimeout(() => {
				if (!mapReady) loadError = '地图加载超时，请检查地图服务配置或网络连接';
			}, 12000);
			map.once('style.load', () => {
				window.clearTimeout(loadTimer);
				map?.resize();
				mapReady = true;
				renderMarkers(places, selectedSlug);
				resetView();
			});
		} catch (error) {
			loadError = error instanceof Error ? error.message : '地图初始化失败';
		}

		return () => {
			window.clearTimeout(loadTimer);
			resizeObserver?.disconnect();
			for (const marker of Object.values(markers)) marker.remove();
			for (const key of Object.keys(markers)) delete markers[key];
			map?.remove();
			map = null;
		};
	});
</script>

<div
	class="group/map relative aspect-[16/8.5] min-h-72 overflow-hidden border border-ink-200/80 bg-[#eef2ed] dark:border-ink-800 dark:bg-[#111916]"
>
	<div bind:this={mapElement} class="map-host" aria-label="已到达城市分布地图"></div>

	<div
		class="pointer-events-none absolute left-4 top-4 z-10 flex items-center gap-2 bg-ink-50/90 px-3 py-2 font-mono text-[10px] uppercase tracking-[0.16em] text-ink-500 shadow-sm backdrop-blur dark:bg-ink-950/90 dark:text-ink-400"
	>
		<LocateFixed size={14} class="text-jade-600 dark:text-jade-400" />
		{places.length} 座城市已点亮
	</div>

	{#if mapReady}
		<button
			type="button"
			class="absolute bottom-[6.75rem] right-[10px] z-10 flex h-[29px] w-[29px] items-center justify-center rounded-sm bg-white text-ink-700 shadow hover:text-jade-700"
			title="重置地图"
			aria-label="重置地图"
			onclick={resetView}><RotateCcw size={15} /></button
		>
	{:else}
		<div
			class="absolute inset-0 z-20 flex items-center justify-center bg-ink-50/85 backdrop-blur-sm dark:bg-ink-950/85"
			role={loadError ? 'alert' : undefined}
		>
			<p class="px-6 text-center text-xs text-ink-500 dark:text-ink-400">
				{loadError || '正在展开地图…'}
			</p>
		</div>
	{/if}
</div>

<style>
	.map-host {
		position: absolute !important;
		inset: 0;
		width: 100%;
		height: 100%;
	}
	:global(.footprint-map-marker) {
		display: flex;
		align-items: center;
		gap: 0.4rem;
		border: 0;
		background: transparent;
		padding: 0;
		cursor: pointer;
	}
	:global(.footprint-map-marker__dot) {
		display: block;
		width: 0.75rem;
		height: 0.75rem;
		border: 2px solid white;
		border-radius: 999px;
		background: var(--color-jade-600);
		box-shadow: 0 0 0 5px color-mix(in srgb, var(--color-jade-500) 22%, transparent);
		transition: transform 160ms ease;
	}
	:global(.footprint-map-marker__label) {
		display: none;
		white-space: nowrap;
		background: color-mix(in srgb, white 92%, transparent);
		padding: 0.2rem 0.4rem;
		font-family: var(--font-mono);
		font-size: 0.6875rem;
		color: var(--color-ink-900);
		box-shadow: 0 1px 4px rgb(0 0 0 / 14%);
	}
	:global(.footprint-map-marker[data-selected='true'] .footprint-map-marker__dot) {
		transform: scale(1.35);
	}
	:global(.footprint-map-marker[data-selected='true'] .footprint-map-marker__label) {
		display: block;
	}
	:global(.maplibregl-ctrl-attrib) {
		font-size: 9px;
	}
</style>

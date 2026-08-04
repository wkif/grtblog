<script lang="ts">
	import { onMount } from 'svelte';
	import type { Component } from 'svelte';
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

	type MapComponent = Component<{
		places: FootprintPlaceGroup[];
		mapSettings: FootprintMapSettings;
		selectedSlug: string | null;
		onSelect: (slug: string) => void;
	}>;

	let LoadedMap = $state<MapComponent | null>(null);
	let loadError = $state(false);

	onMount(() => {
		void import('./FootprintMapClient.svelte')
			.then((module) => {
				LoadedMap = module.default;
			})
			.catch((error: unknown) => {
				console.error('[footprints] failed to load map', error);
				loadError = true;
			});
	});
</script>

{#if LoadedMap}
	<LoadedMap {places} {mapSettings} {selectedSlug} {onSelect} />
{:else if loadError}
	<div
		class="relative flex aspect-[16/8.5] min-h-72 items-center justify-center overflow-hidden border border-ink-200/80 bg-ink-100 dark:border-ink-800 dark:bg-ink-900"
		role="alert"
	>
		<p class="relative text-xs text-ink-500 dark:text-ink-400">地图加载失败，请刷新页面重试</p>
	</div>
{:else}
	<div
		class="relative flex aspect-[16/8.5] min-h-72 items-center justify-center overflow-hidden border border-ink-200/80 bg-ink-100 dark:border-ink-800 dark:bg-ink-900"
		aria-label="足迹地图加载中"
	>
		<div
			class="absolute inset-0 opacity-25 [background-image:linear-gradient(to_right,currentColor_1px,transparent_1px),linear-gradient(to_bottom,currentColor_1px,transparent_1px)] [background-size:48px_48px]"
		></div>
		<p class="relative font-mono text-xs uppercase tracking-[0.18em] text-ink-400">正在展开地图…</p>
	</div>
{/if}

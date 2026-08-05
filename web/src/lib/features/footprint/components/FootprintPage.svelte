<script lang="ts">
	import { Clock3, MapPin, Route } from 'lucide-svelte';
	import PageHeader from '$lib/ui/common/PageHeader.svelte';
	import { footprintCtx } from '../context';
	import { formatDistance, formatDuration } from '../format';
	import type { FootprintOverview } from '../types';
	import FootprintJourneyList from './FootprintJourneyList.svelte';
	import FootprintMapShell from './FootprintMapShell.svelte';

	let { overview }: { overview: FootprintOverview } = $props();
	footprintCtx.mountModelData(() => overview);

	let selectedSlug = $state<string | null>(null);
	let selectedYear = $state<number | null>(null);

	const years = $derived(
		Array.from(
			new Set(
				overview.places.flatMap((place) =>
					place.journeys.map((journey) => Number(journey.journeyDate.slice(0, 4)))
				)
			)
		).sort((a, b) => b - a)
	);
	const visiblePlaces = $derived.by(() => {
		if (!selectedYear) return overview.places;
		return overview.places
			.map((place) => {
				const journeys = place.journeys.filter(
					(journey) => Number(journey.journeyDate.slice(0, 4)) === selectedYear
				);
				return {
					...place,
					journeys,
					stats: {
						journeyCount: journeys.length,
						totalDistanceMeters: journeys.reduce(
							(sum, journey) => sum + (journey.distanceMeters ?? 0),
							0
						),
						totalDurationSeconds: journeys.reduce(
							(sum, journey) => sum + (journey.durationSeconds ?? 0),
							0
						)
					}
				};
			})
			.filter((place) => place.journeys.length > 0);
	});
	const selectedPlace = $derived(
		visiblePlaces.find((place) => place.slug === selectedSlug) ?? visiblePlaces[0] ?? null
	);
	const visibleStats = $derived.by(() => ({
		cityCount: visiblePlaces.length,
		journeyCount: visiblePlaces.reduce((sum, place) => sum + place.stats.journeyCount, 0),
		totalDistanceMeters: visiblePlaces.reduce(
			(sum, place) => sum + place.stats.totalDistanceMeters,
			0
		),
		totalDurationSeconds: visiblePlaces.reduce(
			(sum, place) => sum + place.stats.totalDurationSeconds,
			0
		)
	}));

	function selectPlace(slug: string) {
		selectedSlug = slug;
	}

	function selectYear(year: number | null) {
		selectedYear = year;
		selectedSlug = null;
	}
</script>

<svelte:head>
	<title>足迹</title>
	<meta name="description" content="记录到达过的城市、徒步里程、轨迹与旅途中的相册。" />
</svelte:head>

<div class="mx-auto w-full max-w-[1500px] px-3.5 py-8 sm:px-6 sm:py-14 md:py-16">
	<PageHeader
		title="足迹"
		tag="Footprints"
		subtitle="把走过的路，留在地图上"
		description="城市、徒步、轨迹，以及一次次抵达时留下的照片。"
	/>

	{#if overview.places.length === 0}
		<div class="py-32 text-center font-serif text-lg text-ink-400/60">还没有点亮任何城市</div>
	{:else}
		<div
			class="mb-6 flex flex-col gap-5 border-y border-ink-200/80 py-5 dark:border-ink-800 sm:flex-row sm:items-center sm:justify-between"
		>
			<div class="grid grid-cols-2 gap-x-8 gap-y-4 sm:flex sm:items-center sm:gap-9">
				<div>
					<div class="font-serif text-2xl text-ink-900 dark:text-ink-50">
						{visibleStats.cityCount}
					</div>
					<div class="mt-1 font-mono text-[9px] uppercase tracking-[0.18em] text-ink-400">城市</div>
				</div>
				<div>
					<div class="font-serif text-2xl text-ink-900 dark:text-ink-50">
						{visibleStats.journeyCount}
					</div>
					<div class="mt-1 font-mono text-[9px] uppercase tracking-[0.18em] text-ink-400">行程</div>
				</div>
				<div>
					<div class="font-serif text-2xl text-ink-900 dark:text-ink-50">
						{formatDistance(visibleStats.totalDistanceMeters) ?? '—'}
					</div>
					<div class="mt-1 font-mono text-[9px] uppercase tracking-[0.18em] text-ink-400">
						累计徒步
					</div>
				</div>
				<div>
					<div class="font-serif text-2xl text-ink-900 dark:text-ink-50">
						{formatDuration(visibleStats.totalDurationSeconds) ?? '—'}
					</div>
					<div class="mt-1 font-mono text-[9px] uppercase tracking-[0.18em] text-ink-400">
						累计时长
					</div>
				</div>
			</div>
			{#if years.length > 1}
				<div class="flex max-w-full gap-1 overflow-x-auto" aria-label="按年份筛选">
					<button
						type="button"
						class:active={!selectedYear}
						class="year-filter"
						onclick={() => selectYear(null)}>全部</button
					>
					{#each years as year, yearIndex (`${year}-${yearIndex}`)}<button
							type="button"
							class:active={selectedYear === year}
							class="year-filter"
							onclick={() => selectYear(year)}>{year}</button
						>{/each}
				</div>
			{/if}
		</div>

		<div class="grid gap-6">
			<FootprintMapShell
				places={visiblePlaces}
				mapSettings={overview.map}
				selectedSlug={selectedPlace?.slug ?? null}
				onSelect={selectPlace}
			/>
			<aside
				class="border border-ink-200/80 bg-ink-50 p-5 dark:border-ink-800 dark:bg-ink-900/50"
				aria-live="polite"
			>
				{#if selectedPlace}
					<div class="border-b border-ink-200/80 pb-5 dark:border-ink-800">
						<div
							class="flex items-center gap-2 font-mono text-[10px] uppercase tracking-[0.18em] text-jade-700 dark:text-jade-300"
						>
							<MapPin size={14} />{selectedPlace.countryName}
						</div>
						<h2 class="mt-3 font-serif text-3xl text-ink-950 dark:text-ink-50">
							{selectedPlace.cityName}
						</h2>
						{#if selectedPlace.regionName}<p class="mt-1 text-sm text-ink-400">
								{selectedPlace.regionName}
							</p>{/if}
						<div class="mt-5 flex flex-wrap gap-x-4 gap-y-2 text-xs text-ink-500 dark:text-ink-400">
							<span>{selectedPlace.stats.journeyCount} 次行程</span>
							{#if formatDistance(selectedPlace.stats.totalDistanceMeters)}<span
									class="inline-flex items-center gap-1"
									><Route size={13} />{formatDistance(
										selectedPlace.stats.totalDistanceMeters
									)}</span
								>{/if}
							{#if formatDuration(selectedPlace.stats.totalDurationSeconds)}<span
									class="inline-flex items-center gap-1"
									><Clock3 size={13} />{formatDuration(
										selectedPlace.stats.totalDurationSeconds
									)}</span
								>{/if}
						</div>
					</div>
					<div class="mt-5 max-h-[520px] overflow-y-auto pr-1">
						<FootprintJourneyList journeys={selectedPlace.journeys} />
					</div>
				{/if}
			</aside>
		</div>

		<section class="mt-16 border-t border-ink-200/80 pt-8 dark:border-ink-800">
			<div class="mb-6 flex items-end justify-between gap-4">
				<div>
					<div
						class="font-mono text-[10px] uppercase tracking-[0.2em] text-jade-700 dark:text-jade-300"
					>
						Journey Index
					</div>
					<h2 class="mt-2 font-serif text-2xl text-ink-900 dark:text-ink-50">城市与行程</h2>
				</div>
				<span class="font-mono text-xs text-ink-400">{visiblePlaces.length} 座城市</span>
			</div>
			<div class="grid gap-x-10 gap-y-0 md:grid-cols-2">
				{#each visiblePlaces as place, placeIndex (`${place.id}-${place.slug}-${placeIndex}`)}
					<button
						type="button"
						class="group flex w-full items-center justify-between border-b border-ink-200/80 py-5 text-left dark:border-ink-800"
						onclick={() => selectPlace(place.slug)}
					>
						<div>
							<div
								class="font-serif text-xl text-ink-900 transition group-hover:text-jade-700 dark:text-ink-100 dark:group-hover:text-jade-300"
							>
								{place.cityName}
							</div>
							<div class="mt-1 text-xs text-ink-400">
								{place.regionName ? `${place.regionName} · ` : ''}{place.countryName}
							</div>
						</div>
						<div class="text-right">
							<div class="font-mono text-xs text-ink-600 dark:text-ink-300">
								{place.stats.journeyCount} 次
							</div>
							<div class="mt-1 text-[10px] text-ink-400">
								{formatDistance(place.stats.totalDistanceMeters) ?? '无里程记录'}
							</div>
						</div>
					</button>
				{/each}
			</div>
		</section>
	{/if}
</div>

<style>
	.year-filter {
		min-width: 3.25rem;
		height: 2rem;
		padding: 0 0.75rem;
		border: 1px solid color-mix(in srgb, currentColor 14%, transparent);
		font-family: var(--font-mono);
		font-size: 0.625rem;
		color: color-mix(in srgb, currentColor 60%, transparent);
		transition:
			border-color 180ms ease,
			color 180ms ease,
			background 180ms ease;
	}
	.year-filter:hover,
	.year-filter.active {
		border-color: var(--color-jade-500);
		color: var(--color-jade-700);
		background: color-mix(in srgb, var(--color-jade-500) 8%, transparent);
	}
	.year-filter:focus-visible {
		outline: 2px solid var(--color-jade-500);
		outline-offset: 2px;
	}
</style>

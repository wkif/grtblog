<script lang="ts">
	import { CalendarDays, Clock3, ExternalLink, Images, Route } from 'lucide-svelte';
	import { formatDistance, formatDuration, formatJourneyDate } from '../format';
	import type { FootprintJourney } from '../types';

	let { journeys }: { journeys: FootprintJourney[] } = $props();
</script>

<div class="divide-y divide-ink-200/80 dark:divide-ink-800">
	{#each journeys as journey, journeyIndex (`${journey.id}-${journeyIndex}`)}
		<article class="py-6 first:pt-0 last:pb-0">
			<div class="flex gap-4">
				{#if journey.cover || journey.albums[0]?.cover}
					<img
						src={journey.cover || journey.albums[0]?.cover || ''}
						alt=""
						class="h-20 w-24 shrink-0 object-cover sm:h-24 sm:w-32"
					/>
				{/if}
				<div class="min-w-0 flex-1">
					<div
						class="flex items-center gap-2 font-mono text-[10px] uppercase tracking-[0.14em] text-ink-400"
					>
						<CalendarDays size={13} />
						{formatJourneyDate(journey.journeyDate, journey.endedAt)}
					</div>
					<h3 class="mt-2 font-serif text-xl leading-snug text-ink-900 dark:text-ink-50">
						{journey.title}
					</h3>
					<div class="mt-3 flex flex-wrap gap-x-4 gap-y-2 text-xs text-ink-500 dark:text-ink-400">
						{#if formatDistance(journey.distanceMeters)}<span
								class="inline-flex items-center gap-1.5"
								><Route size={14} />{formatDistance(journey.distanceMeters)}</span
							>{/if}
						{#if formatDuration(journey.durationSeconds)}<span
								class="inline-flex items-center gap-1.5"
								><Clock3 size={14} />{formatDuration(journey.durationSeconds)}</span
							>{/if}
						{#if journey.albums.length}<span class="inline-flex items-center gap-1.5"
								><Images size={14} />{journey.albums.length} 个相册</span
							>{/if}
					</div>
				</div>
			</div>

			{#if journey.summary}
				<p class="mt-4 text-sm leading-7 text-ink-600 dark:text-ink-300">{journey.summary}</p>
			{/if}

			<div class="mt-4 flex flex-wrap gap-2">
				{#if journey.trackUrl}
					<a
						href={journey.trackUrl}
						target="_blank"
						rel="noopener noreferrer"
						class="inline-flex h-9 items-center gap-2 border border-jade-600/30 px-3 text-xs font-medium text-jade-700 transition hover:border-jade-600 hover:bg-jade-500/5 dark:text-jade-300"
					>
						<Route size={14} />查看轨迹<ExternalLink size={12} />
					</a>
				{/if}
				{#each journey.albums as album, albumIndex (`${album.id}-${albumIndex}`)}
					<a
						href="/albums/{album.shortUrl}"
						class="inline-flex h-9 items-center gap-2 border border-ink-200 px-3 text-xs text-ink-700 transition hover:border-jade-500/50 hover:text-jade-700 dark:border-ink-700 dark:text-ink-300 dark:hover:text-jade-300"
					>
						<Images size={14} />{album.title}<span class="text-ink-400">{album.photoCount}</span>
					</a>
				{/each}
			</div>
		</article>
	{/each}
</div>

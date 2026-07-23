<script lang="ts">
	import type { MediaRecord } from '../types';

	let { record, onOpen }: { record: MediaRecord; onOpen?: (record: MediaRecord) => void } =
		$props();

	const statusLabel = $derived(
		record.status === 'planned'
			? '想看'
			: record.status === 'watching'
				? '在看'
				: record.status === 'completed'
					? '看完'
					: '弃剧'
	);
	const year = $derived(record.releaseDate ? new Date(record.releaseDate).getFullYear() : '—');
	const progressPercent = $derived(
		record.status === 'completed'
			? 100
			: record.mediaType === 'movie'
				? Math.max(0, Math.min(100, Math.round(record.progress)))
				: record.progressTotal && record.progressTotal > 0
					? Math.min(100, Math.round((record.progress / record.progressTotal) * 100))
					: 0
	);
</script>

<button
	type="button"
	class="group relative block w-full overflow-hidden rounded-default border border-ink-200/70 bg-white/70 text-left transition duration-500 hover:-translate-y-1 hover:border-jade-300/70 hover:shadow-float focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-jade-500/70 dark:border-ink-800 dark:bg-ink-900/50 dark:hover:border-jade-700/70"
	onclick={() => onOpen?.(record)}
>
	<div class="grid grid-cols-[88px_1fr] gap-4 p-3.5 sm:grid-cols-[112px_1fr] sm:gap-5 sm:p-4">
		<div class="relative aspect-[2/3] overflow-hidden rounded-sm bg-ink-100 dark:bg-ink-800">
			{#if record.poster}
				<img
					src={record.poster}
					alt={record.title}
					loading="lazy"
					class="h-full w-full object-cover transition duration-700 group-hover:scale-105"
				/>
			{:else}
				<div
					class="flex h-full items-center justify-center px-2 text-center font-serif text-sm text-ink-400"
				>
					{record.title}
				</div>
			{/if}
			<div
				class="absolute left-2 top-2 rounded-full bg-ink-950/65 px-2 py-1 font-mono text-[9px] tracking-wider text-white/80 backdrop-blur"
			>
				{record.mediaType === 'tv' ? 'SERIES' : 'FILM'}
			</div>
		</div>

		<div class="flex min-w-0 flex-col py-1">
			<div class="flex items-start justify-between gap-3">
				<div class="min-w-0">
					<h2
						class="font-serif text-lg font-medium leading-snug text-ink-950 dark:text-ink-100 sm:text-xl"
					>
						{record.title}
					</h2>
					{#if record.originalTitle && record.originalTitle !== record.title}
						<p class="mt-1 truncate text-[11px] text-ink-400">{record.originalTitle}</p>
					{/if}
				</div>
				<span
					class="shrink-0 rounded-full border border-jade-500/20 bg-jade-500/5 px-2 py-1 text-[10px] text-jade-700 dark:text-jade-300"
					>{statusLabel}</span
				>
			</div>

			<div
				class="mt-3 flex flex-wrap items-center gap-x-3 gap-y-1 font-mono text-[10px] uppercase tracking-wider text-ink-400"
			>
				<span>{year}</span>
				{#if record.runtimeMinutes}<span>{record.runtimeMinutes} min</span>{/if}
				{#if record.rating}<span class="text-amber-600 dark:text-amber-400"
						>★ {record.rating.toFixed(1)}</span
					>{/if}
			</div>

			{#if record.overview}
				<p class="mt-3 line-clamp-3 text-xs leading-6 text-ink-500 dark:text-ink-400">
					{record.overview}
				</p>
			{/if}

			<div class="mt-auto pt-4">
				<div class="mb-1.5 flex items-center justify-between font-mono text-[10px] text-ink-400">
					<span
						>{record.status === 'completed'
							? '已看完'
							: record.mediaType === 'tv' && record.progressTotal
								? `第 ${record.progress} / ${record.progressTotal} 集`
								: '观看进度'}</span
					>
					<span>{progressPercent}%</span>
				</div>
				<div class="h-1 overflow-hidden rounded-full bg-ink-100 dark:bg-ink-800">
					<div
						class="h-full rounded-full bg-jade-500 transition-all duration-700 group-hover:bg-jade-400"
						style:width={`${progressPercent}%`}
					></div>
				</div>
			</div>
		</div>
	</div>
	{#if record.note}
		<div
			class="border-t border-ink-100 px-4 py-3 text-xs leading-5 text-ink-500 dark:border-ink-800 dark:text-ink-400"
		>
			“{record.note}”
		</div>
	{/if}
</button>

<script lang="ts">
	import type { MediaRecord } from '../types';

	let { record, onClose }: { record: MediaRecord; onClose: () => void } = $props();

	const statusLabel = $derived(
		record.status === 'planned'
			? '想看'
			: record.status === 'watching'
				? '在看'
				: record.status === 'completed'
					? '看完'
					: '弃剧'
	);
	const typeLabel = $derived(record.mediaType === 'tv' ? '剧集' : '电影');
	const year = $derived(record.releaseDate ? new Date(record.releaseDate).getFullYear() : null);
	const progressPercent = $derived(
		record.status === 'completed'
			? 100
			: record.mediaType === 'movie'
				? Math.max(0, Math.min(100, Math.round(record.progress)))
				: record.progressTotal && record.progressTotal > 0
					? Math.min(100, Math.round((record.progress / record.progressTotal) * 100))
					: 0
	);
	const runtimeLabel = $derived(
		record.runtimeMinutes
			? record.runtimeMinutes >= 60
				? `${Math.floor(record.runtimeMinutes / 60)} 小时 ${record.runtimeMinutes % 60} 分钟`
				: `${record.runtimeMinutes} 分钟`
			: null
	);
	const progressLabel = $derived(
		record.status === 'completed'
			? '已完整看完'
			: record.mediaType === 'movie'
				? `观看进度 ${progressPercent}%`
				: record.progressTotal
					? `第 ${record.progress} / ${record.progressTotal} 集`
					: `已观看 ${record.progress} 集`
	);

	function handleKeydown(event: KeyboardEvent) {
		if (event.key === 'Escape') onClose();
	}

	$effect(() => {
		const previousOverflow = document.body.style.overflow;
		document.body.style.overflow = 'hidden';
		return () => {
			document.body.style.overflow = previousOverflow;
		};
	});
</script>

<svelte:window onkeydown={handleKeydown} />

<div class="fixed inset-0 z-[1000] flex items-center justify-center p-3 sm:p-6">
	<button
		type="button"
		class="absolute inset-0 cursor-default bg-ink-950/75 backdrop-blur-sm"
		aria-label="关闭影视详情"
		onclick={onClose}
	></button>
	<div
		class="relative z-10 max-h-[min(860px,calc(100vh-1.5rem))] w-full max-w-4xl overflow-y-auto rounded-2xl border border-white/15 bg-ink-50 shadow-2xl dark:bg-ink-950 sm:max-h-[calc(100vh-3rem)]"
		role="dialog"
		aria-modal="true"
		aria-labelledby="media-record-detail-title"
	>
		{#if record.backdrop}
			<div class="absolute inset-x-0 top-0 h-64 overflow-hidden opacity-35">
				<img src={record.backdrop} alt="" class="h-full w-full object-cover" />
				<div
					class="absolute inset-0 bg-gradient-to-b from-ink-950/30 via-ink-950/70 to-ink-50 dark:to-ink-950"
				></div>
			</div>
		{/if}

		<button
			type="button"
			class="absolute right-4 top-4 z-10 flex h-9 w-9 items-center justify-center rounded-full border border-white/20 bg-ink-950/50 text-lg text-white/90 backdrop-blur transition hover:bg-ink-950/80 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-jade-400"
			aria-label="关闭影视详情"
			onclick={onClose}
		>
			×
		</button>

		<div class="relative grid gap-6 p-5 sm:grid-cols-[180px_1fr] sm:gap-8 sm:p-8">
			<div class="mx-auto w-36 sm:mx-0 sm:w-full">
				<div class="aspect-[2/3] overflow-hidden rounded-lg bg-ink-200 shadow-xl dark:bg-ink-800">
					{#if record.poster}
						<img src={record.poster} alt={record.title} class="h-full w-full object-cover" />
					{:else}
						<div
							class="flex h-full items-center justify-center p-4 text-center font-serif text-sm text-ink-400"
						>
							{record.title}
						</div>
					{/if}
				</div>
			</div>

			<div class="min-w-0 pt-1 sm:pt-8">
				<div
					class="flex flex-wrap items-center gap-2 font-mono text-[10px] uppercase tracking-[0.18em] text-jade-700 dark:text-jade-300"
				>
					<span>{typeLabel}</span>
					<span class="text-ink-300 dark:text-ink-700">/</span>
					<span>{statusLabel}</span>
					{#if year}<span class="text-ink-400">/ {year}</span>{/if}
				</div>
				<h2
					id="media-record-detail-title"
					class="mt-3 pr-8 font-serif text-3xl leading-tight text-ink-950 dark:text-ink-50 sm:text-4xl"
				>
					{record.title}
				</h2>
				{#if record.originalTitle && record.originalTitle !== record.title}
					<p class="mt-2 text-sm text-ink-500 dark:text-ink-400">{record.originalTitle}</p>
				{/if}

				<div
					class="mt-6 grid grid-cols-2 gap-x-5 gap-y-4 border-y border-ink-200/70 py-5 text-sm dark:border-ink-800 sm:grid-cols-3"
				>
					<div>
						<div class="font-mono text-[10px] uppercase tracking-wider text-ink-400">观看状态</div>
						<div class="mt-1 text-ink-800 dark:text-ink-200">{statusLabel}</div>
					</div>
					{#if runtimeLabel}
						<div>
							<div class="font-mono text-[10px] uppercase tracking-wider text-ink-400">片长</div>
							<div class="mt-1 text-ink-800 dark:text-ink-200">{runtimeLabel}</div>
						</div>
					{/if}
					{#if record.totalEpisodes}
						<div>
							<div class="font-mono text-[10px] uppercase tracking-wider text-ink-400">总集数</div>
							<div class="mt-1 text-ink-800 dark:text-ink-200">{record.totalEpisodes} 集</div>
						</div>
					{/if}
					{#if record.rating}
						<div>
							<div class="font-mono text-[10px] uppercase tracking-wider text-ink-400">
								个人评分
							</div>
							<div class="mt-1 text-amber-600 dark:text-amber-400">
								★ {record.rating.toFixed(1)} / 10
							</div>
						</div>
					{/if}
					{#if record.releaseDate}
						<div>
							<div class="font-mono text-[10px] uppercase tracking-wider text-ink-400">
								上映日期
							</div>
							<div class="mt-1 text-ink-800 dark:text-ink-200">
								{record.releaseDate.slice(0, 10)}
							</div>
						</div>
					{/if}
				</div>

				<div class="mt-6">
					<div class="mb-2 flex items-center justify-between gap-4">
						<span class="font-mono text-[10px] uppercase tracking-wider text-ink-400">观看进度</span
						>
						<span class="text-sm font-medium text-jade-700 dark:text-jade-300">{progressLabel}</span
						>
					</div>
					<div class="h-2 overflow-hidden rounded-full bg-ink-200 dark:bg-ink-800">
						<div
							class="h-full rounded-full bg-jade-500 transition-all duration-700"
							style:width={`${progressPercent}%`}
						></div>
					</div>
					<div class="mt-2 text-right font-mono text-xs text-ink-400">{progressPercent}%</div>
				</div>
			</div>
		</div>

		<div
			class="relative space-y-6 border-t border-ink-200/70 px-5 pb-6 pt-6 dark:border-ink-800 sm:px-8 sm:pb-8"
		>
			{#if record.overview}
				<section>
					<h3
						class="font-mono text-[10px] uppercase tracking-[0.2em] text-jade-700 dark:text-jade-300"
					>
						故事简介
					</h3>
					<p class="mt-3 text-sm leading-7 text-ink-700 dark:text-ink-300">{record.overview}</p>
				</section>
			{/if}
			{#if record.note}
				<section class="rounded-lg border-l-2 border-jade-500/60 bg-jade-500/5 px-4 py-3">
					<h3
						class="font-mono text-[10px] uppercase tracking-[0.2em] text-jade-700 dark:text-jade-300"
					>
						我的记录
					</h3>
					<p class="mt-2 font-serif text-base leading-7 text-ink-700 dark:text-ink-300">
						“{record.note}”
					</p>
				</section>
			{/if}
			<div
				class="flex flex-wrap gap-x-5 gap-y-2 font-mono text-[10px] uppercase tracking-wider text-ink-400"
			>
				{#if record.startedAt}<span>开始于 {record.startedAt.slice(0, 10)}</span>{/if}
				{#if record.completedAt}<span>完成于 {record.completedAt.slice(0, 10)}</span>{/if}
				<span>记录于 {record.updatedAt.slice(0, 10)}</span>
				{#if record.providerId}<span>来源 {record.provider.toUpperCase()} #{record.providerId}</span
					>{/if}
			</div>
		</div>
	</div>
</div>

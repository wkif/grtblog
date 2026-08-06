<script lang="ts">
	import type { AlbumDetail, PhotoItem } from '$lib/features/album/types';
	import { ArrowLeft, ChevronLeft, ChevronRight, Clock3, Maximize2, Video } from 'lucide-svelte';

	let {
		album,
		item,
		index,
		total,
		onBack,
		onPrev,
		onNext
	}: {
		album: AlbumDetail;
		item: PhotoItem;
		index: number;
		total: number;
		onBack: () => void;
		onPrev: () => void;
		onNext: () => void;
	} = $props();

	const duration = $derived.by(() => {
		if (!item.durationMs || item.durationMs <= 0) return null;
		const totalSeconds = Math.round(item.durationMs / 1000);
		const hours = Math.floor(totalSeconds / 3600);
		const minutes = Math.floor((totalSeconds % 3600) / 60);
		const seconds = totalSeconds % 60;
		return hours > 0
			? `${hours}:${String(minutes).padStart(2, '0')}:${String(seconds).padStart(2, '0')}`
			: `${minutes}:${String(seconds).padStart(2, '0')}`;
	});
</script>

<div
	class="fixed inset-x-0 bottom-0 top-[calc(env(safe-area-inset-top)+4.5rem)] z-40 flex flex-col overflow-y-auto bg-ink-950 md:inset-y-0 md:left-24 md:top-0 lg:flex-row lg:overflow-hidden"
>
	<main
		class="relative flex min-h-[60svh] flex-1 items-center justify-center bg-black px-3 py-16 sm:px-6 lg:min-h-0"
	>
		<button
			type="button"
			class="absolute left-3 top-3 z-20 inline-flex items-center gap-1.5 rounded-[3px] border border-white/10 bg-ink-900/80 px-3 py-2 text-[11px] text-white/70 backdrop-blur-xl transition-colors hover:text-white sm:left-4 sm:top-4"
			onclick={onBack}
		>
			<ArrowLeft size={15} />返回
		</button>

		{#if index > 0}
			<button
				type="button"
				aria-label="上一项"
				class="absolute left-2 top-1/2 z-20 -translate-y-1/2 rounded-full border border-white/10 bg-black/65 p-2.5 text-white/75 backdrop-blur-md transition hover:text-jade-300 sm:left-4"
				onclick={onPrev}
			>
				<ChevronLeft size={22} />
			</button>
		{/if}
		{#if index < total - 1}
			<button
				type="button"
				aria-label="下一项"
				class="absolute right-2 top-1/2 z-20 -translate-y-1/2 rounded-full border border-white/10 bg-black/65 p-2.5 text-white/75 backdrop-blur-md transition hover:text-jade-300 sm:right-4"
				onclick={onNext}
			>
				<ChevronRight size={22} />
			</button>
		{/if}

		<video
			poster={item.posterUrl || undefined}
			controls
			playsinline
			preload="metadata"
			class="max-h-[78svh] max-w-full bg-black object-contain shadow-2xl lg:max-h-[88vh]"
		>
			<source src={item.url} type={item.mimeType || undefined} />
		</video>
	</main>

	<aside
		class="noise-surface w-full shrink-0 border-t border-white/8 bg-ink-950/92 p-5 text-white lg:w-72 lg:overflow-y-auto lg:border-l lg:border-t-0"
	>
		<p class="font-mono text-[10px] uppercase tracking-[0.2em] text-white/30">Video</p>
		<h1 class="mt-2 font-serif text-lg leading-snug text-white/92">
			{item.caption || album.title}
		</h1>
		{#if item.description}
			<p class="mt-3 text-xs leading-relaxed text-white/55">{item.description}</p>
		{/if}

		<div class="mt-5 space-y-3 border-t border-white/8 pt-5 text-xs text-white/48">
			{#if duration}
				<div class="flex items-center gap-2.5">
					<Clock3 size={14} class="text-white/25" />{duration}
				</div>
			{/if}
			{#if item.width && item.height}
				<div class="flex items-center gap-2.5">
					<Maximize2 size={14} class="text-white/25" />{item.width} × {item.height}
				</div>
			{/if}
			{#if item.mimeType}
				<div class="flex items-center gap-2.5">
					<Video size={14} class="text-white/25" />{item.mimeType}
				</div>
			{/if}
		</div>

		<div class="mt-6 flex items-center justify-between border-t border-white/8 pt-4">
			<a
				href="/albums/{album.shortUrl}"
				class="text-xs text-white/35 transition-colors hover:text-jade-300">← {album.title}</a
			>
			<span class="font-mono text-[10px] text-white/25">{index + 1} / {total}</span>
		</div>
	</aside>
</div>

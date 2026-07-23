<script lang="ts">
	import StickyHeader from '$lib/ui/common/StickyHeader.svelte';
	import { postDetailCtx } from '$lib/features/post/context';
	import { buildImageExtInfoState, imageExtInfoCtx } from '$lib/shared/markdown/image-ext-info';
	import PostDetailHeader from './post-detail/PostDetailHeader.svelte';
	import PostDetailMain from './post-detail/PostDetailMain.svelte';
	import { detailHeroBgSrc } from '$lib/shared/stores/detailHeroBg';
	import { onDestroy } from 'svelte';
	import { resolve } from '$app/paths';
	import { FileText, Home, RefreshCw } from 'lucide-svelte';

	const hasPostStore = postDetailCtx.selectModelData((data) => Boolean(data));
	const postTitleStore = postDetailCtx.selectModelData((data) => data?.title ?? '');
	const postCoverStore = postDetailCtx.selectModelData((data) => data?.cover ?? '');
	const postExtInfoStore = postDetailCtx.selectModelData((data) => data?.extInfo ?? null, {
		equals: (a, b) => a === b
	});
	imageExtInfoCtx.mountModelData(() => buildImageExtInfoState($postExtInfoStore));

	$effect(() => {
		detailHeroBgSrc.set($postCoverStore);
	});
	onDestroy(() => detailHeroBgSrc.set(''));

	function reloadPage() {
		location.reload();
	}
</script>

{#if $hasPostStore}
	<StickyHeader title={$postTitleStore} />

	<article class="article-enter space-y-10">
		<PostDetailHeader />
		<PostDetailMain />
	</article>
{:else}
		<div class="mx-auto flex max-w-xl flex-col items-center rounded-default border border-ink-200/70 bg-ink-50/60 px-6 py-20 text-center dark:border-ink-800 dark:bg-ink-900/40">
			<div class="relative mb-5">
				<div class="absolute -inset-3 rounded-full bg-jade-500/10 blur-xl" aria-hidden="true"></div>
				<FileText size={34} class="relative text-ink-300 dark:text-ink-600" aria-hidden="true" />
			</div>
			<h1 class="font-serif text-xl font-medium text-ink-900 dark:text-ink-100">文章暂时无法打开</h1>
			<p class="mt-3 max-w-sm text-sm leading-6 text-ink-500 dark:text-ink-400">
				这篇内容可能已经被移动、取消发布，或暂时无法从服务器获取。你可以返回文章归档，或者稍后重试。
			</p>
			<div class="mt-7 flex flex-wrap justify-center gap-3">
				<a
					href={resolve('/posts', {})}
					class="inline-flex items-center gap-2 rounded-default bg-jade-600 px-4 py-2 text-sm text-white transition hover:-translate-y-0.5 hover:bg-jade-700 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-jade-500/50"
				>
					<Home size={14} aria-hidden="true" />
					文章归档
				</a>
				<button
					type="button"
					onclick={reloadPage}
					class="inline-flex items-center gap-2 rounded-default border border-ink-200 bg-white px-4 py-2 text-sm text-ink-700 transition hover:-translate-y-0.5 hover:bg-ink-100 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ink-400/50 dark:border-ink-700 dark:bg-ink-800 dark:text-ink-200 dark:hover:bg-ink-700"
				>
					<RefreshCw size={14} aria-hidden="true" />
					重新加载
				</button>
			</div>
		</div>
	{/if}

<script lang="ts">
	import { resolvePath } from '$lib/shared/utils/resolve-path';
	import { postDetailCtx } from '$lib/features/post/context';
	import { sameMetrics } from './selector-equals';
	import { formatDateCN, isDifferentDay } from '$lib/shared/utils/date';
	import { calculateReadingTime, formatReadingTime } from '$lib/shared/utils/reading-time';
	import { ArrowLeft, Calendar, Clock } from 'lucide-svelte';
	import Icon from '@iconify/svelte';
	import Button from '$lib/ui/primitives/button/Button.svelte';
	import Badge from '$lib/ui/primitives/badge/Badge.svelte';
	import ContentLikeButton from '$lib/features/analytics/components/ContentLikeButton.svelte';
	import TagList from '$lib/features/tag/components/TagList.svelte';
	import { buildCategoryPath } from '$lib/shared/utils/content-path';
	import { FadeIn, RollingNumber } from '$lib/ui/animation';

	const titleStore = postDetailCtx.selectModelData((data) => data?.title ?? '');
	const postIdStore = postDetailCtx.selectModelData((data) => data?.id ?? 0);
	const createdAtStore = postDetailCtx.selectModelData((data) => data?.createdAt ?? '');
	const contentUpdatedAtStore = postDetailCtx.selectModelData(
		(data) => data?.contentUpdatedAt ?? ''
	);
	const contentStore = postDetailCtx.selectModelData((data) => data?.content ?? '');
	const showUpdated = $derived(isDifferentDay($createdAtStore, $contentUpdatedAtStore));
	const isHotStore = postDetailCtx.selectModelData((data) => data?.isHot ?? false);
	const metricsStore = postDetailCtx.selectModelData((data) => data?.metrics ?? null, {
		equals: sameMetrics
	});
	const tagsStore = postDetailCtx.selectModelData((data) => data?.tags ?? []);
	const categoryNameStore = postDetailCtx.selectModelData((data) => data?.categoryName ?? '');
	const categoryShortUrlStore = postDetailCtx.selectModelData(
		(data) => data?.categoryShortUrl ?? ''
	);
	const categoryLabelStore = $derived.by(() => {
		const categoryName = ($categoryNameStore || '').trim();
		return categoryName || '未分类';
	});
	const readingTime = $derived(calculateReadingTime($contentStore));

	function goBack() {
		history.back();
	}
</script>

{#snippet backContent()}
	<ArrowLeft size={14} class="group-hover:-translate-x-1 transition-transform" />
	<span>返回</span>
{/snippet}

<header class="max-w-4xl space-y-6">
	<div class="flex items-center gap-4">
		<Button
			variant="ghost"
			class="!h-auto !p-0 font-mono text-[10px] font-semibold tracking-[0.2em] text-ink-400 uppercase hover:!bg-transparent hover:text-ink-900 group"
			onclick={goBack}
			content={backContent}
		/>
		<div class="h-px w-6 bg-ink-200/50 dark:bg-ink-800/50"></div>
	</div>

	<div class="space-y-4">
		<FadeIn y={8} duration={650}>
		<div class="flex items-center gap-3">
			<Badge variant="soft">文章</Badge>
			{#if $categoryShortUrlStore}
				<a
					href={resolvePath(buildCategoryPath($categoryShortUrlStore))}
					class="font-mono text-[9px] tracking-[0.3em] text-ink-400 uppercase hover:text-jade-600 dark:hover:text-jade-400 transition-colors"
				>
					{categoryLabelStore}
				</a>
			{:else}
				<span class="font-mono text-[9px] tracking-[0.3em] text-ink-400 uppercase"
					>{categoryLabelStore}</span
				>
			{/if}
		</div>
		</FadeIn>

		<FadeIn y={12} delay={80} duration={700}>
			<h1
				class="article-title font-serif text-2xl leading-[1.2] font-medium tracking-tight text-ink-950 md:text-3xl lg:text-4xl dark:text-ink-50"
			>
				{$titleStore}
			</h1>
		</FadeIn>

		<FadeIn y={10} delay={160} duration={650}>
		<div class="flex flex-col gap-4">
			<div
				class="article-meta-primary flex flex-wrap items-center gap-x-5 gap-y-2 font-mono text-[9px] tracking-widest text-ink-400 uppercase"
			>
				<span class="flex items-center gap-1.5">
					<Calendar size={12} />
					{formatDateCN($createdAtStore)}{#if showUpdated}<span class="text-ink-400/70"
							>（更新于 {formatDateCN($contentUpdatedAtStore)}）</span
						>{/if}
				</span>
				<span class="flex items-center gap-1.5"
					><Clock size={12} /> {formatReadingTime(readingTime)}</span
				>
			</div>

			<div class="article-meta-secondary flex flex-wrap items-center gap-x-4 gap-y-3 border-t border-ink-100 pt-3 font-mono text-[9px] tracking-widest text-ink-400 uppercase dark:border-ink-800">
				{#if $isHotStore}
					{#snippet hotIcon()}
						<Icon icon="ph:fire-fill" class="size-4 text-red-500" />
					{/snippet}
					<Badge
						variant="soft"
						class="!border-red-500/20 !bg-red-500/5 !text-red-600 dark:!text-red-400"
						icon={hotIcon}
					>
						热门
					</Badge>
				{/if}
				<span class="flex items-center gap-1.5">浏览 <RollingNumber value={$metricsStore?.views ?? 0} /></span>
				<ContentLikeButton
					contentType="article"
					contentId={$postIdStore}
					likes={$metricsStore?.likes ?? 0}
					className="inline-flex items-center gap-1.5"
				/>
				<span class="flex items-center gap-1.5">评论 <RollingNumber value={$metricsStore?.comments ?? 0} /></span>
			</div>

			<div class="article-tags pt-1">
				<TagList tags={$tagsStore} />
			</div>
		</div>
		</FadeIn>
	</div>
</header>

<style lang="postcss">
	@reference "$routes/layout.css";

	.article-title {
		text-wrap: balance;
	}

	.article-meta-secondary :global(button),
	.article-meta-secondary :global(a) {
		transition:
			color 180ms ease,
			transform 220ms cubic-bezier(0.16, 1, 0.3, 1);
	}

	.article-meta-secondary :global(button:hover),
	.article-meta-secondary :global(a:hover) {
		transform: translateY(-1px);
	}

	@media (prefers-reduced-motion: reduce) {
		.article-meta-secondary :global(button),
		.article-meta-secondary :global(a) {
			transition: none;
		}
	}
</style>

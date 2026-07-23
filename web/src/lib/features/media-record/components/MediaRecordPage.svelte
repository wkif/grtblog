<script lang="ts">
	import PageHeader from '$lib/ui/common/PageHeader.svelte';
	import StaggerList from '$lib/ui/animation/StaggerList.svelte';
	import type { MediaRecord, MediaStatus } from '../types';
	import MediaRecordCard from './MediaRecordCard.svelte';
	import MediaRecordDetailModal from './MediaRecordDetailModal.svelte';

	let { records }: { records: MediaRecord[] } = $props();
	let selectedRecord = $state<MediaRecord | null>(null);
	const groups: { status: MediaStatus; title: string; subtitle: string }[] = [
		{ status: 'watching', title: '在看', subtitle: '正在展开的故事' },
		{ status: 'completed', title: '看完', subtitle: '留在记忆里的片段' },
		{ status: 'planned', title: '想看', subtitle: '排队等候的夜晚' },
		{ status: 'dropped', title: '弃剧', subtitle: '没有继续的理由' }
	];
	const grouped = $derived(
		groups.map((group) => ({
			...group,
			items: records.filter((record) => record.status === group.status)
		}))
	);
</script>

<div class="mx-auto w-full max-w-[1200px] px-3.5 py-8 sm:px-6 sm:py-14 md:px-0 md:py-16">
	<PageHeader
		title="影视记录"
		tag="Watchlist"
		subtitle="把看过的故事留在这里"
		description="电影、剧集，以及观看过程中留下的时间、进度和一点私人感受。"
	/>

	{#if records.length === 0}
		<div class="py-32 text-center font-serif text-lg text-ink-400/60">还没有影视记录</div>
	{:else}
		<div class="space-y-14">
			{#each grouped as group (group.status)}
				{#if group.items.length > 0}
					<section>
						<div
							class="mb-5 flex items-end justify-between border-b border-ink-200/70 pb-3 dark:border-ink-800"
						>
							<div>
								<div
									class="font-mono text-[10px] uppercase tracking-[0.24em] text-jade-600 dark:text-jade-400"
								>
									{group.subtitle}
								</div>
								<h2 class="mt-2 font-serif text-2xl text-ink-900 dark:text-ink-100">
									{group.title}
								</h2>
							</div>
							<span class="font-mono text-xs text-ink-400">{group.items.length} 条</span>
						</div>
						<StaggerList class="grid grid-cols-1 gap-4 md:grid-cols-2" staggerDelay={80} y={14}>
							{#each group.items as record (record.id)}<MediaRecordCard
									{record}
									onOpen={(item) => (selectedRecord = item)}
								/>{/each}
						</StaggerList>
					</section>
				{/if}
			{/each}
		</div>
	{/if}
</div>

{#if selectedRecord}
	<MediaRecordDetailModal record={selectedRecord} onClose={() => (selectedRecord = null)} />
{/if}

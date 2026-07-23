<script lang="ts">
	import MediaRecordPage from '$lib/features/media-record/components/MediaRecordPage.svelte';
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();
</script>

<MediaRecordPage records={data.records} />

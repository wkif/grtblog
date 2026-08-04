<script lang="ts">
	import FootprintPage from '$lib/features/footprint/components/FootprintPage.svelte';
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();
</script>

<FootprintPage overview={data.overview} />

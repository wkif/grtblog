import { getFootprintOverview } from '$lib/features/footprint/api';
import { trackISRDeps } from '$lib/server/isr-deps';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async (event) => {
	trackISRDeps(event, 'footprints:list');
	return {
		overview: await getFootprintOverview(event.fetch)
	};
};

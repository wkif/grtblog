import { getMediaRecordList } from '$lib/features/media-record/api';
import { trackISRDeps } from '$lib/server/isr-deps';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async (event) => {
	trackISRDeps(event, 'media-records:list');
	return {
		records: (await getMediaRecordList(event.fetch, { pageSize: 100 })).items
	};
};

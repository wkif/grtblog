import { getApi } from '$lib/shared/clients/api';
import type { FootprintOverview } from './types';

const emptyOverview: FootprintOverview = {
	summary: {
		cityCount: 0,
		journeyCount: 0,
		totalDistanceMeters: 0,
		totalDurationSeconds: 0
	},
	places: [],
	map: {
		provider: 'osm',
		tiandituKey: '',
		tiandituLayer: 'vector'
	}
};

const defaultMap = emptyOverview.map;

export async function getFootprintOverview(fetcher?: typeof fetch): Promise<FootprintOverview> {
	const api = getApi(fetcher);
	const overview = await api<Partial<FootprintOverview>>('/footprints');
	if (!overview) return emptyOverview;
	return {
		summary: overview.summary ?? emptyOverview.summary,
		places: overview.places ?? [],
		map: {
			...defaultMap,
			...(overview.map ?? {})
		}
	};
}

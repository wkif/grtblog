import { getApi } from '$lib/shared/clients/api';
import type { MediaRecordListResponse, MediaStatus, MediaType } from './types';

export async function getMediaRecordList(
	fetcher?: typeof fetch,
	options: { status?: MediaStatus; mediaType?: MediaType; page?: number; pageSize?: number } = {}
): Promise<MediaRecordListResponse> {
	const api = getApi(fetcher);
	const query = new URLSearchParams({
		page: String(options.page ?? 1),
		pageSize: String(options.pageSize ?? 60)
	});
	if (options.status) query.set('status', options.status);
	if (options.mediaType) query.set('mediaType', options.mediaType);
	return (await api<MediaRecordListResponse>(`/media-records?${query.toString()}`)) ?? {
		items: [],
		total: 0,
		page: options.page ?? 1,
		size: options.pageSize ?? 60
	};
}

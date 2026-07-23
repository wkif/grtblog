export type MediaType = 'movie' | 'tv';
export type MediaStatus = 'planned' | 'watching' | 'completed' | 'dropped';

export type MediaRecord = {
	id: number;
	title: string;
	originalTitle?: string | null;
	mediaType: MediaType;
	provider: string;
	providerId?: string | null;
	poster?: string | null;
	backdrop?: string | null;
	overview?: string | null;
	releaseDate?: string | null;
	runtimeMinutes?: number | null;
	totalEpisodes?: number | null;
	status: MediaStatus;
	progress: number;
	progressTotal?: number | null;
	rating?: number | null;
	note?: string | null;
	startedAt?: string | null;
	completedAt?: string | null;
	isPublished: boolean;
	createdAt: string;
	updatedAt: string;
};

export type MediaRecordListResponse = {
	items: MediaRecord[];
	total: number;
	page: number;
	size: number;
};

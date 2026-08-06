export type PhotoExif = {
	make?: string;
	model?: string;
	lensModel?: string;
	focalLength?: string;
	fNumber?: string;
	exposureTime?: string;
	iso?: number;
	gpsLatitude?: number;
	gpsLongitude?: number;
	dateTimeOriginal?: string;
	imageWidth?: number;
	imageHeight?: number;
	dominantColor?: string;
	[key: string]: unknown;
};

export type AlbumMediaType = 'image' | 'video';

export type PhotoItem = {
	id: number;
	albumId?: number;
	url: string;
	mediaType: AlbumMediaType;
	mimeType?: string | null;
	posterUrl?: string | null;
	thumbnailUrl?: string;
	durationMs?: number | null;
	width?: number | null;
	height?: number | null;
	description?: string | null;
	caption?: string | null;
	exif?: PhotoExif | null;
	sortOrder: number;
	createdAt: string;
};

export type AlbumSummary = {
	id: number;
	title: string;
	description?: string | null;
	cover?: string | null;
	shortUrl: string;
	isPublished: boolean;
	photoCount: number;
	views: number;
	likes: number;
	comments: number;
	createdAt: string;
	updatedAt: string;
};

export type AlbumDetail = {
	id: number;
	title: string;
	description?: string | null;
	cover?: string | null;
	shortUrl: string;
	authorId: number;
	commentAreaId?: number | null;
	isPublished: boolean;
	allowComment: boolean;
	photoCount: number;
	metrics?: { views: number; likes: number; comments: number } | null;
	photos: PhotoItem[];
	createdAt: string;
	updatedAt: string;
};

export type AlbumListResponse = {
	items: AlbumSummary[];
	total: number;
	page: number;
	size: number;
};

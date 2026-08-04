export type FootprintPlace = {
	id: number;
	slug: string;
	cityName: string;
	regionName?: string | null;
	countryName: string;
	countryCode?: string | null;
	latitude: number;
	longitude: number;
};

export type FootprintAlbum = {
	id: number;
	title: string;
	shortUrl: string;
	cover?: string | null;
	photoCount: number;
	isPublished: boolean;
};

export type FootprintJourney = {
	id: number;
	placeId: number;
	place: FootprintPlace;
	title: string;
	journeyDate: string;
	endedAt?: string | null;
	summary?: string | null;
	cover?: string | null;
	distanceMeters?: number | null;
	durationSeconds?: number | null;
	trackUrl?: string | null;
	albums: FootprintAlbum[];
	isPublished: boolean;
	sortOrder: number;
	createdAt: string;
	updatedAt: string;
};

export type FootprintStats = {
	journeyCount: number;
	totalDistanceMeters: number;
	totalDurationSeconds: number;
};

export type FootprintPlaceGroup = FootprintPlace & {
	stats: FootprintStats;
	journeys: FootprintJourney[];
};

export type FootprintOverview = {
	summary: FootprintStats & { cityCount: number };
	places: FootprintPlaceGroup[];
	map: FootprintMapSettings;
};

export type FootprintMapSettings = {
	provider: 'osm' | 'tianditu';
	tiandituKey: string;
	tiandituLayer: 'vector' | 'imagery';
};

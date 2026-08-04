export function formatDistance(meters?: number | null): string | null {
	if (meters == null) return null;
	const kilometers = meters / 1000;
	return `${kilometers.toFixed(kilometers >= 10 ? 1 : 2)} km`;
}

export function formatDuration(seconds?: number | null): string | null {
	if (seconds == null) return null;
	const hours = Math.floor(seconds / 3600);
	const minutes = Math.round((seconds % 3600) / 60);
	if (hours === 0) return `${minutes} 分钟`;
	if (minutes === 0) return `${hours} 小时`;
	return `${hours} 小时 ${minutes} 分`;
}

export function formatJourneyDate(start: string, end?: string | null): string {
	const startLabel = start.slice(0, 10).replaceAll('-', '.');
	if (!end || end.slice(0, 10) === start.slice(0, 10)) return startLabel;
	return `${startLabel} — ${end.slice(0, 10).replaceAll('-', '.')}`;
}

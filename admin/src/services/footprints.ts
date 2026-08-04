import { request } from './http'

export interface FootprintPlace {
  id: number
  slug: string
  cityName: string
  regionName?: string | null
  countryName: string
  countryCode?: string | null
  latitude: number
  longitude: number
  createdAt?: string
  updatedAt?: string
}

export interface FootprintAlbum {
  id: number
  title: string
  shortUrl: string
  cover?: string | null
  photoCount: number
  isPublished: boolean
}

export interface FootprintJourney {
  id: number
  placeId: number
  place: FootprintPlace
  title: string
  journeyDate: string
  endedAt?: string | null
  summary?: string | null
  cover?: string | null
  distanceMeters?: number | null
  durationSeconds?: number | null
  trackUrl?: string | null
  albums: FootprintAlbum[]
  isPublished: boolean
  sortOrder: number
  createdAt: string
  updatedAt: string
}

export interface FootprintJourneyPayload {
  place: Omit<FootprintPlace, 'id' | 'createdAt' | 'updatedAt'>
  title: string
  journeyDate: string
  endedAt?: string | null
  summary?: string | null
  cover?: string | null
  distanceMeters?: number | null
  durationSeconds?: number | null
  trackUrl?: string | null
  albumIds: number[]
  isPublished: boolean
  sortOrder: number
}

export interface FootprintJourneyListResponse {
  items: FootprintJourney[]
  total: number
  page: number
  size: number
}

function stripEmpty<T extends object>(value: T): Record<string, unknown> {
  return Object.fromEntries(
    Object.entries(value).filter(
      ([, entry]) => entry !== undefined && entry !== null && entry !== '',
    ),
  )
}

export function listFootprints(params: {
  page?: number
  pageSize?: number
  search?: string
  published?: boolean
}) {
  return request<FootprintJourneyListResponse>('/admin/footprints', {
    method: 'GET',
    query: stripEmpty(params),
  })
}

export function getFootprint(id: number) {
  return request<FootprintJourney>(`/admin/footprints/${id}`, { method: 'GET' })
}

export function listFootprintPlaces() {
  return request<FootprintPlace[]>('/admin/footprint-places', { method: 'GET' })
}

export function createFootprint(payload: FootprintJourneyPayload) {
  return request<FootprintJourney>('/footprints', { method: 'POST', body: payload })
}

export function updateFootprint(id: number, payload: FootprintJourneyPayload) {
  return request<FootprintJourney>(`/footprints/${id}`, { method: 'PUT', body: payload })
}

export function deleteFootprint(id: number) {
  return request<void>(`/footprints/${id}`, { method: 'DELETE' })
}

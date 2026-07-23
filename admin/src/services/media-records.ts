import { request } from './http'

export type MediaType = 'movie' | 'tv'
export type MediaStatus = 'planned' | 'watching' | 'completed' | 'dropped'

export interface MediaRecord {
  id: number
  title: string
  originalTitle?: string | null
  mediaType: MediaType
  provider: string
  providerId?: string | null
  poster?: string | null
  backdrop?: string | null
  overview?: string | null
  releaseDate?: string | null
  runtimeMinutes?: number | null
  totalEpisodes?: number | null
  status: MediaStatus
  progress: number
  progressTotal?: number | null
  rating?: number | null
  note?: string | null
  startedAt?: string | null
  completedAt?: string | null
  isPublished: boolean
  createdAt: string
  updatedAt: string
}

export interface MediaSearchResult {
  providerId: string
  title: string
  originalTitle: string
  mediaType: MediaType
  poster?: string
  backdrop?: string
  overview?: string
  releaseDate?: string
  runtimeMinutes?: number
  totalEpisodes?: number
}

export interface MediaRecordPayload {
  title: string
  originalTitle?: string | null
  mediaType: MediaType
  provider: string
  providerId?: string | null
  poster?: string | null
  backdrop?: string | null
  overview?: string | null
  releaseDate?: string | null
  runtimeMinutes?: number | null
  totalEpisodes?: number | null
  status: MediaStatus
  progress: number
  progressTotal?: number | null
  rating?: number | null
  note?: string | null
  startedAt?: string | null
  completedAt?: string | null
  isPublished: boolean
}

export interface MediaRecordListResponse {
  items: MediaRecord[]
  total: number
  page: number
  size: number
}

const stripEmpty = (value: Record<string, unknown>) =>
  Object.fromEntries(Object.entries(value).filter(([, entry]) => entry !== undefined && entry !== null && entry !== ''))

export function listMediaRecords(params: { page?: number; pageSize?: number; search?: string; status?: string }) {
  return request<MediaRecordListResponse>('/admin/media-records', {
    method: 'GET',
    query: stripEmpty(params),
  })
}

export function getMediaRecord(id: number) {
  return request<MediaRecord>(`/admin/media-records/${id}`, { method: 'GET' })
}

export function searchMediaRecords(query: string, mediaType?: MediaType) {
  return request<MediaSearchResult[]>('/admin/media-records/search', {
    method: 'GET',
    query: stripEmpty({ q: query, mediaType }),
  })
}

export function getMediaRecordDetails(mediaType: MediaType, providerId: string) {
  return request<MediaSearchResult>(`/admin/media-records/details/${mediaType}/${providerId}`, {
    method: 'GET',
  })
}

export function createMediaRecord(payload: MediaRecordPayload) {
  return request<MediaRecord>('/media-records', { method: 'POST', body: payload })
}

export function updateMediaRecord(id: number, payload: MediaRecordPayload) {
  return request<MediaRecord>(`/media-records/${id}`, { method: 'PUT', body: payload })
}

export function deleteMediaRecord(id: number) {
  return request<void>(`/media-records/${id}`, { method: 'DELETE' })
}

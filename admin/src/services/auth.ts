import { ApiError, request } from './http'

export interface UserInfo {
  id: number
  username: string
  nickname: string
  email: string
  avatar: string
  isActive: boolean
  isAdmin: boolean
  createdAt: string
  updatedAt: string
  deletedAt?: string | null
}

export interface LoginResponse {
  token: string
  user: UserInfo
  roles: string[]
  permissions: string[]
}

export interface LoginPayload {
  credential: string
  password: string
  turnstileToken?: string
}

export interface RegisterPayload {
  username: string
  nickname?: string
  email: string
  password: string
  turnstileToken?: string
}

export function register(payload: RegisterPayload) {
  return request<UserInfo>('/auth/register', {
    method: 'POST',
    body: JSON.stringify(payload),
  })
}

export function login(payload: LoginPayload) {
  return request<LoginResponse>('/auth/login', {
    method: 'POST',
    body: JSON.stringify(payload),
  })
}

export interface SetupStateResponse {
  hasUser: boolean
  hasAdmin: boolean
  websiteInfoReady: boolean
  missingWebsiteInfoKeys: string[]
  needsSetup: boolean
}

/** Legacy API: only "at least one user exists" — used when `/auth/setup-state` is missing (older server). */
export interface InitStateResponse {
  initialized: boolean
}

export function getInitState() {
  return request<InitStateResponse>('/auth/init-state', {
    method: 'GET',
  })
}

function setupStateFromInitState(init: InitStateResponse): SetupStateResponse {
  const { initialized } = init
  return {
    hasUser: initialized,
    hasAdmin: initialized,
    websiteInfoReady: initialized,
    missingWebsiteInfoKeys: initialized
      ? []
      : ['website_name', 'public_url', 'description', 'keywords'],
    needsSetup: !initialized,
  }
}

/**
 * Prefer `/auth/setup-state` (user + admin + site info). Falls back to `/auth/init-state` on HTTP 404
 * so older server images still load the admin shell (init wizard may be less precise for site fields).
 */
export async function getSetupState(): Promise<SetupStateResponse> {
  try {
    return await request<SetupStateResponse>('/auth/setup-state', {
      method: 'GET',
    })
  } catch (e) {
    const notFound =
      e instanceof ApiError &&
      (e.status === 404 || e.bizErr === 'NOT_FOUND' || e.code === 404)
    if (!notFound) throw e
    try {
      const init = await getInitState()
      return setupStateFromInitState(init)
    } catch {
      throw e
    }
  }
}

export interface AccessInfoResponse {
  user: UserInfo
  roles: string[]
  permissions: string[]
}

export function getAccessInfo() {
  return request<AccessInfoResponse>('/auth/access-info', {
    method: 'GET',
  })
}

export interface UpdateProfilePayload {
  nickname?: string
  avatar?: string
  email?: string
}

export function updateProfile(payload: UpdateProfilePayload) {
  return request<UserInfo>('/auth/profile', {
    method: 'PUT',
    body: JSON.stringify(payload),
  })
}

export interface ChangePasswordPayload {
  oldPassword: string
  newPassword: string
}

export function changePassword(payload: ChangePasswordPayload) {
  return request<null>('/auth/password', {
    method: 'PUT',
    body: JSON.stringify(payload),
  })
}

export interface OAuthBinding {
  providerKey: string
  providerName: string
  boundAt: string
  expiresAt?: string | null
  providerScope?: string
}

export function getOAuthBindings() {
  return request<OAuthBinding[]>('/auth/oauth-bindings', {
    method: 'GET',
  })
}

export interface TurnstileStateResponse {
  enabled: boolean
  siteKey?: string
}

export function getTurnstileState() {
  return request<TurnstileStateResponse>('/auth/turnstile', {
    method: 'GET',
  })
}

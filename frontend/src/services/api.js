/**
 * API Service — все HTTP-запросы к бэкенду.
 */

const BASE = '/api/v1'

async function request(method, path, body = null, token = null) {
  const headers = { 'Content-Type': 'application/json' }
  if (token) headers['Authorization'] = `Bearer ${token}`

  const res = await fetch(`${BASE}${path}`, {
    method,
    headers,
    credentials: 'include', // отправляет HttpOnly-куки (refresh_token) автоматически
    body: body ? JSON.stringify(body) : null,
  })

  if (res.status === 204) return null

  const data = await res.json()
  if (!res.ok) throw { status: res.status, ...data }
  return data
}

// ── Auth ──
export const authApi = {
  register: (username, email, password) =>
    request('POST', '/auth/register', { username, email, password }),

  login: (identifier, password) =>
    request('POST', '/auth/login', { identifier, password }),

  refresh: () =>
    request('POST', '/auth/refresh'),

  logout: (token) =>
    request('POST', '/auth/logout', null, token),

  me: (token) =>
    request('GET', '/auth/me', null, token),

  forgotPassword: (email) =>
    request('POST', '/auth/forgot-password', { email }),

  resetPassword: (token, newPassword) =>
    request('POST', '/auth/reset-password', { token, new_password: newPassword }),
}

// ── User / Profile ──
export const userApi = {
  updateProfile: (username, email, token) =>
    request('PUT', '/users/profile', { username, email }, token),

  changePassword: (currentPassword, newPassword, token) =>
    request('POST', '/users/password', { current_password: currentPassword, new_password: newPassword }, token),
}

// ── Friends ──
export const friendsApi = {
  getFriends: (token) =>
    request('GET', '/users/friends', null, token),

  getRequests: (token) =>
    request('GET', '/users/friends/requests', null, token),

  sendRequest: (identifier, token) =>
    request('POST', '/users/friends/request', { identifier }, token),

  acceptRequest: (userId, token) =>
    request('POST', '/users/friends/accept', { user_id: userId }, token),

  declineRequest: (userId, token) =>
    request('POST', '/users/friends/decline', { user_id: userId }, token),

  deleteFriend: (id, token) =>
    request('DELETE', `/users/friends/${id}`, null, token),

  searchUsers: (q, token) =>
    request('GET', `/users/search?q=${encodeURIComponent(q)}`, null, token),
}

// ── WatchParty ──
export const watchpartyApi = {
  getActiveRooms: (userIds) =>
    request('POST', '/watchparty/rooms/active', { user_ids: userIds }),
}

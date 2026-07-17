import Taro from '@tarojs/taro'

const BASE_URL = 'http://localhost:8080'

let accessToken = ''

export function setToken(token: string) {
  accessToken = token
  Taro.setStorageSync('access_token', token)
}

export function getToken(): string {
  if (!accessToken) {
    accessToken = Taro.getStorageSync('access_token') || ''
  }
  return accessToken
}

/** 统一请求。网络重试复用相同 request_id 实现幂等。 */
export async function request<T = unknown>(
  path: string,
  options: { method?: string; data?: unknown; requestId?: string } = {},
): Promise<T> {
  const requestId = options.requestId || `req_${Date.now()}`
  const res = await Taro.request({
    url: `${BASE_URL}${path}`,
    method: (options.method as keyof Taro.request.Method) || 'GET',
    data: options.data,
    header: {
      'Content-Type': 'application/json',
      Authorization: getToken() ? `Bearer ${getToken()}` : '',
      'X-Request-Id': requestId,
    },
  })
  if (res.statusCode >= 400) {
    throw new Error((res.data as { message?: string })?.message || '请求失败')
  }
  return (res.data as { data?: T })?.data as T
}

export async function wechatLogin(code: string) {
  const res = await Taro.request({
    url: `${BASE_URL}/api/v1/auth/wechat/login`,
    method: 'POST',
    data: { code },
    header: { 'Content-Type': 'application/json' },
  })
  if (res.statusCode === 200 && (res.data as { data?: { access_token?: string } })?.data?.access_token) {
    const data = (res.data as { data: { access_token: string; user_id: string } }).data
    setToken(data.access_token)
    return data
  }
  throw new Error('登录失败')
}

export async function sendMessage(conversationId: string, message: string, requestId?: string) {
  return request(`/api/v1/conversations/${conversationId}/messages`, {
    method: 'POST',
    data: { message },
    requestId,
  })
}

/**
 * ZhiPath API Client
 *
 * 统一响应结构：{ code, message, data, trace_id }
 * 公网请求不携带可信 user_id，身份由 Bearer token 推导。
 */

export interface ApiResponse<T = unknown> {
  code: string;
  message: string;
  data: T;
  trace_id: string;
}

export interface ClientContext {
  client_type?: 'wechat_miniapp' | 'ios' | 'android' | 'web';
  app_version?: string;
  platform?: string;
  channel?: string;
}

export interface SendMessageOptions {
  message: string;
  consultation_type?: 'auto' | 'emotion' | 'career';
  attachments?: { file_id: string; file_type: string }[];
  request_id?: string;
  client_context?: ClientContext;
}

export class ZhiPathError extends Error {
  constructor(
    public code: string,
    message: string,
    public trace_id: string,
    public retryable = false,
  ) {
    super(message);
    this.name = 'ZhiPathError';
  }
}

export class ZhiPathClient {
  private readonly baseUrl: string;
  private readonly getToken: () => string | null;

  constructor(baseUrl: string, getToken: () => string | null = () => null) {
    this.baseUrl = baseUrl.replace(/\/$/, '');
    this.getToken = getToken;
  }

  private async request<T>(path: string, init: RequestInit = {}): Promise<T> {
    const token = this.getToken();
    const headers = new Headers(init.headers);
    if (token) headers.set('Authorization', `Bearer ${token}`);
    if (init.body && !headers.has('Content-Type')) {
      headers.set('Content-Type', 'application/json');
    }
    const res = await fetch(`${this.baseUrl}${path}`, { ...init, headers });
    const body = (await res.json()) as ApiResponse<T>;
    if (body.code !== 'SUCCESS') {
      throw new ZhiPathError(body.code, body.message, body.trace_id);
    }
    return body.data;
  }

  // ===== 健康 =====
  health() {
    return this.request<{ status: string }>('/healthz', { method: 'GET' });
  }

  // ===== 认证 =====
  wechatLogin(code: string, client_context?: ClientContext) {
    return this.request<{ access_token: string; refresh_token: string; expires_in: number; user_id: string }>(
      '/api/v1/auth/wechat/login',
      { method: 'POST', body: JSON.stringify({ code, client_context }) },
    );
  }

  // ===== 画像 =====
  getProfile() {
    return this.request<Record<string, unknown>>('/api/v1/me/profile', { method: 'GET' });
  }

  updateProfile(profile: Record<string, unknown>) {
    return this.request<Record<string, unknown>>('/api/v1/me/profile', {
      method: 'PUT',
      body: JSON.stringify(profile),
    });
  }

  // ===== MBTI =====
  getMbti() {
    return this.request<{ current: unknown; history: unknown[] }>('/api/v1/me/mbti', { method: 'GET' });
  }

  submitMbti(result: { result_type: string; assertiveness: 'A' | 'T'; dimensions?: unknown; tested_at?: string }) {
    return this.request<unknown>('/api/v1/me/mbti', { method: 'POST', body: JSON.stringify(result) });
  }

  ocrMbti(file_id: string) {
    return this.request<unknown>('/api/v1/me/mbti/ocr', { method: 'POST', body: JSON.stringify({ file_id }) });
  }

  confirmMbti(mbti_result_id: string) {
    return this.request<unknown>(`/api/v1/me/mbti/${mbti_result_id}/confirm`, { method: 'POST' });
  }

  // ===== 会话 =====
  listConversations(limit = 20, cursor?: string) {
    const q = new URLSearchParams({ limit: String(limit) });
    if (cursor) q.set('cursor', cursor);
    return this.request<{ items: unknown[]; next_cursor: string | null }>(
      `/api/v1/conversations?${q}`,
      { method: 'GET' },
    );
  }

  createConversation(input: { domain?: string; title?: string }) {
    return this.request<unknown>('/api/v1/conversations', { method: 'POST', body: JSON.stringify(input) });
  }

  sendMessage(conversation_id: string, options: SendMessageOptions) {
    return this.request<unknown>(`/api/v1/conversations/${conversation_id}/messages`, {
      method: 'POST',
      body: JSON.stringify(options),
    });
  }

  listMessages(conversation_id: string, limit = 50, before?: string) {
    const q = new URLSearchParams({ limit: String(limit) });
    if (before) q.set('before', before);
    return this.request<unknown[]>(`/api/v1/conversations/${conversation_id}/messages?${q}`, { method: 'GET' });
  }

  // ===== 文件 =====
  requestUpload(input: { file_type: string; mime_type: string; size_bytes: number; sha256?: string }) {
    return this.request<{ file_id: string; upload_url: string; expires_at: string }>('/api/v1/files', {
      method: 'POST',
      body: JSON.stringify(input),
    });
  }

  triggerOcr(file_id: string) {
    return this.request<{ file_id: string; ocr_status: string }>(`/api/v1/files/${file_id}/ocr`, {
      method: 'POST',
    });
  }

  getOcrResult(file_id: string) {
    return this.request<unknown>(`/api/v1/files/${file_id}/ocr-result`, { method: 'GET' });
  }

  // ===== 报告 =====
  listReports() {
    return this.request<unknown[]>('/api/v1/reports', { method: 'GET' });
  }

  createReport(conversation_id: string) {
    return this.request<unknown>('/api/v1/reports', {
      method: 'POST',
      body: JSON.stringify({ conversation_id }),
    });
  }

  getReport(report_id: string) {
    return this.request<unknown>(`/api/v1/reports/${report_id}`, { method: 'GET' });
  }

  // ===== 反馈 =====
  createFeedback(input: { conversation_id?: string; message_id?: string; rating: number; reason?: string; comment?: string }) {
    return this.request<unknown>('/api/v1/feedbacks', { method: 'POST', body: JSON.stringify(input) });
  }
}

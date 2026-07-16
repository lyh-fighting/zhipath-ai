import { describe, it, expect } from 'vitest';
import * as fs from 'node:fs';
import * as path from 'node:path';
import * as yaml from 'js-yaml';

const openapiPath = path.resolve(__dirname, '../../api-contracts/openapi.yaml');
const spec = yaml.load(fs.readFileSync(openapiPath, 'utf8')) as {
  openapi: string;
  info: { title: string };
  paths: Record<string, unknown>;
  components: {
    schemas: {
      ApiResponse: { required: string[] };
      Error: { properties: { code: { enum: string[] } } };
    };
    securitySchemes: Record<string, unknown>;
  };
};

describe('ZhiPath API 契约', () => {
  it('OpenAPI 规范可解析且版本为 3.0.3', () => {
    expect(spec.openapi).toBe('3.0.3');
    expect(spec.info.title).toBe('ZhiPath API');
  });

  it('包含健康检查接口', () => {
    expect(spec.paths['/healthz']).toBeDefined();
    expect(spec.paths['/readyz']).toBeDefined();
  });

  it('包含微信登录接口', () => {
    expect(spec.paths['/api/v1/auth/wechat/login']).toBeDefined();
  });

  it('包含画像接口', () => {
    expect(spec.paths['/api/v1/me/profile']).toBeDefined();
  });

  it('包含 MBTI 接口（查询/提交/OCR）', () => {
    expect(spec.paths['/api/v1/me/mbti']).toBeDefined();
    expect(spec.paths['/api/v1/me/mbti/ocr']).toBeDefined();
  });

  it('包含会话与消息接口', () => {
    expect(spec.paths['/api/v1/conversations']).toBeDefined();
    expect(spec.paths['/api/v1/conversations/{conversation_id}/messages']).toBeDefined();
  });

  it('包含文件上传与 OCR 接口', () => {
    expect(spec.paths['/api/v1/files']).toBeDefined();
    expect(spec.paths['/api/v1/files/{file_id}/ocr']).toBeDefined();
  });

  it('包含报告接口', () => {
    expect(spec.paths['/api/v1/reports']).toBeDefined();
  });

  it('包含反馈接口', () => {
    expect(spec.paths['/api/v1/feedbacks']).toBeDefined();
  });

  it('定义统一响应结构 {code,message,data,trace_id}', () => {
    const required = spec.components.schemas.ApiResponse.required;
    expect(required).toEqual(expect.arrayContaining(['code', 'message', 'data', 'trace_id']));
  });

  it('定义强制错误码', () => {
    const codes = spec.components.schemas.Error.properties.code.enum;
    expect(codes).toEqual(
      expect.arrayContaining([
        'AUTH_FAILED',
        'RESOURCE_FORBIDDEN',
        'INVALID_PARAM',
        'MBTI_MISSING',
        'MODEL_TIMEOUT',
        'OCR_FAILED',
        'CRISIS_HUMAN_HANDOFF',
      ]),
    );
  });

  it('公网登录接口不接收可信 user_id', () => {
    const loginPath = spec.paths['/api/v1/auth/wechat/login'] as {
      post: { requestBody: { content: { 'application/json': { schema: object } } } };
    };
    const schemaStr = JSON.stringify(loginPath.post.requestBody.content['application/json'].schema);
    expect(schemaStr).not.toMatch(/"user_id"\s*:/);
  });

  it('定义 Bearer 与内部 token 两种鉴权方案', () => {
    expect(spec.components.securitySchemes.bearerAuth).toBeDefined();
    expect(spec.components.securitySchemes.internalAuth).toBeDefined();
  });

  it('内部 Agent 接口使用 internalAuth', () => {
    const invoke = spec.paths['/internal/v1/agent/invoke'] as {
      post: { security: Record<string, unknown>[] };
    };
    expect(invoke.post.security[0]).toHaveProperty('internalAuth');
  });
});

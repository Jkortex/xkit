import { AppError, failure, Result, success } from '@/utils/result';

interface ParsedServerError {
  message: string;
  serverCode?: string;
}

/**
 * 极简 Fetch 封装，处理 Auth 和 RequestID
 */
export class FetchClient {
  private readonly baseUrl: string;

  constructor(baseUrl: string) {
    this.baseUrl = baseUrl;
  }

  private getHeaders(body?: BodyInit | null): HeadersInit {
    const headers: Record<string, string> = {
      'X-Request-ID': crypto.randomUUID(), // 自动生成全链路 Trace ID
    };

    // FormData 不能手动设置 Content-Type，交给浏览器自动补 boundary
    if (!(body instanceof FormData)) {
      headers['Content-Type'] = 'application/json';
    }

    return headers;
  }

  private async extractServerError(
    response: Response,
  ): Promise<ParsedServerError> {
    try {
      const payload = await response.clone().json();
      if (
        payload &&
        typeof payload === 'object' &&
        'error' in payload &&
        typeof payload.error === 'string' &&
        payload.error.trim() !== ''
      ) {
        const serverCode =
          'code' in payload && typeof payload.code === 'string'
            ? payload.code
            : undefined;
        return {
          message: payload.error,
          serverCode,
        };
      }
    } catch {
      // Ignore parse errors, fallback to status message
    }
    return { message: `服务器返回错误: ${response.status}` };
  }

  private mapServerCodeToErrorCode(
    serverCode?: string,
  ): 'AUTH_ERROR' | 'VALIDATION_ERROR' | 'SERVER_ERROR' | 'NOT_FOUND' {
    if (serverCode === 'INVALID_INPUT') return 'VALIDATION_ERROR';
    if (serverCode === 'UNAUTHORIZED' || serverCode === 'FORBIDDEN') {
      return 'AUTH_ERROR';
    }
    if (serverCode === 'NOT_FOUND') {
      return 'NOT_FOUND';
    }
    return 'SERVER_ERROR';
  }

  async request<T>(
    path: string,
    options: RequestInit = {},
  ): Promise<Result<T>> {
    const url = `${this.baseUrl}${path}`;
    const mergedOptions = {
      ...options,
      credentials: 'include' as const,
      headers: { ...this.getHeaders(options.body), ...options.headers },
    };

    let response: Response;
    try {
      response = await fetch(url, mergedOptions);
    } catch (error) {
      return failure(
        new AppError('NETWORK_ERROR', '网络连接失败，请检查网络', error),
      );
    }

    if (response.status === 401) {
      return failure(new AppError('AUTH_ERROR', '认证失败，请重新登录'));
    }

    if (!response.ok) {
      const parsed = await this.extractServerError(response);
      const errorCode = this.mapServerCodeToErrorCode(parsed.serverCode);
      return failure(
        new AppError(errorCode, parsed.message, undefined, parsed.serverCode),
      );
    }

    // 处理 204 No Content 或空响应
    if (response.status === 204) {
      return success(null as T);
    }

    try {
      const data = await response.json();
      return success(data);
    } catch (error) {
      return failure(new AppError('SERVER_ERROR', '响应解析失败', error));
    }
  }

  async requestBlob(
    path: string,
    options: RequestInit = {},
  ): Promise<Result<Blob>> {
    const url = `${this.baseUrl}${path}`;
    const mergedOptions = {
      ...options,
      credentials: 'include' as const,
      headers: { ...this.getHeaders(options.body), ...options.headers },
    };

    let response: Response;
    try {
      response = await fetch(url, mergedOptions);
    } catch (error) {
      return failure(
        new AppError('NETWORK_ERROR', '网络连接失败，请检查网络', error),
      );
    }

    if (response.status === 401) {
      return failure(new AppError('AUTH_ERROR', '认证失败，请重新登录'));
    }
    if (!response.ok) {
      const parsed = await this.extractServerError(response);
      const errorCode = this.mapServerCodeToErrorCode(parsed.serverCode);
      return failure(
        new AppError(errorCode, parsed.message, undefined, parsed.serverCode),
      );
    }

    return success(await response.blob());
  }
}

// 导出单例 (默认指向后端地址，可在 Vite Env 中配置)
export const httpClient = new FetchClient('/api/v1');

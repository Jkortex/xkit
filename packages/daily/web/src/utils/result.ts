/**
 * 通用的 Result 模式实现
 */
export type Result<T, E = AppError> = Success<T> | Failure<E>;

export class Success<T> {
  readonly kind = 'success';
  constructor(readonly value: T) {}
}

export class Failure<E> {
  readonly kind = 'failure';
  constructor(readonly error: E) {}
}

export const success = <T>(value: T): Success<T> => new Success(value);
export const failure = <E>(error: E): Failure<E> => new Failure(error);

/**
 * 结构化错误基类
 */
export type ErrorCode =
  | 'NETWORK_ERROR'
  | 'AUTH_ERROR'
  | 'VALIDATION_ERROR'
  | 'NOT_FOUND'
  | 'SERVER_ERROR'
  | 'UNKNOWN_ERROR';

export class AppError extends Error {
  constructor(
    public readonly code: ErrorCode,
    message: string,
    public readonly originalError?: unknown,
    public readonly serverCode?: string,
  ) {
    super(message);
    this.name = 'AppError';
  }
}

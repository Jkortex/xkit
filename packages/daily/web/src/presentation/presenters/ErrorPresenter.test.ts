import { describe, expect, it } from 'vitest';
import { ErrorPresenter } from './ErrorPresenter';
import { AppError } from '@/utils/result';

describe('ErrorPresenter', () => {
  it('maps auth/server keyword to auth message', () => {
    const err = new AppError('SERVER_ERROR', 'unauthorized');
    expect(ErrorPresenter.toMessage(err)).toBe(
      '登录已过期或凭证错误，请重新登录',
    );
  });

  it('maps not found keyword to not-found message', () => {
    const err = new AppError('SERVER_ERROR', 'resource not found');
    expect(ErrorPresenter.toMessage(err)).toBe('请求的资源不存在或已被删除');
  });

  it('prefers server code mapping over keyword heuristics', () => {
    const err = new AppError(
      'SERVER_ERROR',
      'some random backend message',
      undefined,
      'NOT_FOUND',
    );
    expect(ErrorPresenter.toMessage(err)).toBe('请求的资源不存在或已被删除');
  });

  it('maps invalid parameter keyword to validation-like message', () => {
    const err = new AppError('SERVER_ERROR', 'invalid date format');
    expect(ErrorPresenter.toMessage(err)).toBe(
      '请求参数不合法，请检查输入后重试',
    );
  });

  it('maps duplicate keyword to conflict message', () => {
    const err = new AppError('SERVER_ERROR', 'duplicate_by_id');
    expect(ErrorPresenter.toMessage(err)).toBe(
      '数据已存在或发生冲突，请刷新后重试',
    );
  });

  it('maps invite quota keyword to specific message', () => {
    const err = new AppError(
      'SERVER_ERROR',
      'active invite limit reached for role member',
    );
    expect(ErrorPresenter.toMessage(err)).toBe(
      '当前角色邀请码已达活跃上限，请先撤销或等待过期',
    );
  });

  it('maps CONFLICT server code to conflict message', () => {
    const err = new AppError(
      'SERVER_ERROR',
      'some conflict',
      undefined,
      'CONFLICT',
    );
    expect(ErrorPresenter.toMessage(err)).toBe(
      '数据已存在或发生冲突，请刷新后重试',
    );
  });

  it('maps import keyword to import guidance', () => {
    const err = new AppError('SERVER_ERROR', 'invalid zip file');
    expect(ErrorPresenter.toMessage(err)).toBe(
      '导入文件格式异常，请检查备份包后重试',
    );
  });

  it('keeps generic message when no keyword matched', () => {
    const err = new AppError('SERVER_ERROR', 'some unknown server fault');
    expect(ErrorPresenter.toMessage(err)).toBe(
      '服务器遇到点小麻烦，我们正在抢修中',
    );
  });
});

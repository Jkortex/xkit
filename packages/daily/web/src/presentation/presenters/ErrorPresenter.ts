import { AppError, ErrorCode } from '@/utils/result';

const SERVER_ERROR_HINTS: Array<{ pattern: RegExp; message: string }> = [
  {
    pattern: /(import|导入|memos_imported|resources_imported|zip file)/i,
    message: '导入文件格式异常，请检查备份包后重试',
  },
  {
    pattern: /(unauthorized|forbidden|token|认证失败|未授权|鉴权)/i,
    message: '登录已过期或凭证错误，请重新登录',
  },
  {
    pattern: /(not found|不存在|resource not found)/i,
    message: '请求的资源不存在或已被删除',
  },
  {
    pattern:
      /(invalid|out of range|cannot be empty|date format|from date|zip file|metadata|parameter)/i,
    message: '请求参数不合法，请检查输入后重试',
  },
  {
    pattern: /(duplicate|already exists|冲突|重复)/i,
    message: '数据已存在或发生冲突，请刷新后重试',
  },
  {
    pattern: /(active invite limit reached|邀请码活跃配额|invite.*limit)/i,
    message: '当前角色邀请码已达活跃上限，请先撤销或等待过期',
  },
];

const mapServerErrorMessage = (rawMessage: string): string => {
  const normalized = rawMessage.trim();
  if (!normalized) return '服务器遇到点小麻烦，我们正在抢修中';

  const matched = SERVER_ERROR_HINTS.find((item) =>
    item.pattern.test(normalized),
  );
  if (matched) return matched.message;

  return '服务器遇到点小麻烦，我们正在抢修中';
};

const SERVER_CODE_MESSAGES: Record<string, string> = {
  INVALID_INPUT: '请求参数不合法，请检查输入后重试',
  NOT_FOUND: '请求的资源不存在或已被删除',
  UNAUTHORIZED: '登录已过期或凭证错误，请重新登录',
  FORBIDDEN: '当前账号无权限执行该操作',
  CONFLICT: '数据已存在或发生冲突，请刷新后重试',
};

/**
 * 错误显示适配器
 * 职责：将结构化错误转换为用户界面的友好文字
 */
export const ErrorPresenter = {
  toMessage(error: unknown): string {
    if (!(error instanceof AppError)) {
      return '发生未知错误，请重试';
    }

    const messages: Record<ErrorCode, string> = {
      NETWORK_ERROR: '网络连接不稳定，请检查网络设置',
      AUTH_ERROR: '登录已过期或凭证错误，请重新登录',
      VALIDATION_ERROR: error.message || '输入内容不符合规则',
      NOT_FOUND: '请求的资源不存在或已被删除',
      SERVER_ERROR:
        (error.serverCode && SERVER_CODE_MESSAGES[error.serverCode]) ||
        mapServerErrorMessage(error.message),
      UNKNOWN_ERROR: '遇到了一些意外，请稍后再试',
    };

    return messages[error.code] || messages.UNKNOWN_ERROR;
  },
};

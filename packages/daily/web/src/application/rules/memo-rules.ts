/**
 * 业务规则验证函数
 */

const MAX_CONTENT_LENGTH = 100000;

// 验证内容是否有效
export const validateMemoContent = (content: string): string | null => {
  if (!content || content.trim().length === 0) {
    return '笔记内容不能为空';
  }
  if (content.length > MAX_CONTENT_LENGTH) {
    return `笔记内容不能超过 ${MAX_CONTENT_LENGTH} 个字符`;
  }
  return null;
};

// 验证日期逻辑
export const validateMemoDates = (
  created: Date,
  updated: Date,
): string | null => {
  if (updated < created) {
    return '更新时间不能早于创建时间';
  }
  return null;
};

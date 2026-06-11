import { MessagePlugin } from 'tdesign-vue-next';

export const copyAuditSummary = async (summary: string): Promise<void> => {
  try {
    await navigator.clipboard.writeText(summary);
    MessagePlugin.success('已复制');
  } catch {
    MessagePlugin.error('复制失败');
  }
};

export const formatRelativeAuditTime = (raw: string): string => {
  const date = new Date(raw);
  if (Number.isNaN(date.getTime())) return raw;
  const diffMs = Date.now() - date.getTime();
  const minute = 60 * 1000;
  const hour = 60 * minute;
  const day = 24 * hour;

  if (diffMs < minute) return '刚刚';
  if (diffMs < hour) return `${Math.floor(diffMs / minute)}分钟前`;
  if (diffMs < day) return `${Math.floor(diffMs / hour)}小时前`;
  if (diffMs < 7 * day) return `${Math.floor(diffMs / day)}天前`;
  return date.toLocaleDateString('zh-CN');
};

export const formatAbsoluteAuditTime = (raw: string): string => {
  const date = new Date(raw);
  if (Number.isNaN(date.getTime())) return raw;
  return date.toLocaleString('zh-CN', { hour12: false });
};

export const useTagGovernanceAudit = () => {
  return {
    copyAuditSummary,
    formatRelativeAuditTime,
    formatAbsoluteAuditTime,
  };
};

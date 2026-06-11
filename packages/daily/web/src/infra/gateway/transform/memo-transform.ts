import type { BackendMemoDTO, BackendResourceDTO } from '../dto/BackendMemoDTO';
import type { MemoDTO, ResourceDTO } from '@/application/ports/dto/Memo';
import { validateMemoDates } from '@/application/rules/memo-rules';

/**
 * 将后端原始资源映射为前端 DTO
 */
export function transformResource(raw: BackendResourceDTO): ResourceDTO {
  return {
    id: raw.id,
    filename: raw.filename,
    size: raw.size,
    mimeType: raw.mime_type,
    createdAt: new Date(raw.created_at),
  };
}

/**
 * 将后端原始笔记映射为前端稳定 DTO
 * 职责：类型转换、默认值填充、业务校验
 */
export function transformMemo(raw: BackendMemoDTO): MemoDTO {
  const createdAt = new Date(raw.created_at);
  const updatedAt = new Date(raw.updated_at);

  // 1. 执行业务规则校验 (Rule Check)
  const dateError = validateMemoDates(createdAt, updatedAt);
  if (dateError) {
    throw new Error(`Data Integrity Error: ${dateError}`);
  }

  // 2. 数据映射与归一化
  return {
    uuid: raw.uuid,
    content: raw.content,
    status: raw.row_status === 'archived' ? 'archived' : 'normal',
    tags: raw.tags || [], // 确保永远是数组
    resources: (raw.resources || []).map(transformResource),
    expiresAt: raw.expires_at ? new Date(raw.expires_at) : undefined,
    headline: raw.headline,
    createdAt,
    updatedAt,
  };
}

/**
 * 前端应用层使用的稳定笔记数据结构
 * 它是 Port 接口的输入输出类型
 */
export interface MemoDTO {
  readonly uuid: string;
  readonly content: string;
  readonly status: 'normal' | 'archived';
  readonly tags: string[];
  readonly resources: ResourceDTO[];
  readonly expiresAt?: Date;
  readonly headline?: string;
  readonly createdAt: Date;
  readonly updatedAt: Date;
}

export interface ResourceDTO {
  readonly id: string;
  readonly filename: string;
  readonly size: number;
  readonly mimeType: string;
  readonly createdAt: Date;
}

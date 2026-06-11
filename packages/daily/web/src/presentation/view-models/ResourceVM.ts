export interface ResourceVM {
  readonly id: string;
  readonly url: string; // 完整的访问地址
  readonly filename: string;
  readonly isImage: boolean; // 方便 UI 判断是否显示 <img>
}

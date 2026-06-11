import type { ResourceDTO } from '@/application/ports/dto/Memo';
import type { ResourceVM } from '../view-models/ResourceVM';

export const ResourcePresenter = {
  toViewModel(dto: ResourceDTO): ResourceVM {
    return {
      id: dto.id,
      // 核心：将基础设施细节封装在此，UI 层不再感知
      url: `/api/v1/resources/${dto.id}`,
      filename: dto.filename,
      isImage: dto.mimeType.startsWith('image/'),
    };
  },
};

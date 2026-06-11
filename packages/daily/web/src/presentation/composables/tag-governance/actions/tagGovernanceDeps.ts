import { useTagStore } from '@/infra/stores/useTagStore';
import { memoGateway } from '@/infra/gateway/HttpMemoGateway';
import { tagSetGateway } from '@/infra/gateway/HttpTagSetGateway';

export interface TagGovernanceDeps {
  tagStore: ReturnType<typeof useTagStore>;
  memoGateway: typeof memoGateway;
  tagSetGateway: typeof tagSetGateway;
}

export const useTagGovernanceDeps = (): TagGovernanceDeps => {
  return {
    tagStore: useTagStore(),
    memoGateway,
    tagSetGateway,
  };
};

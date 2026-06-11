export interface UiCommandMap {
  OpenTagGovernance: {
    tagName?: string;
  };
  OpenApiKeyManager: Record<string, never>;
}

type UiCommandHandler<K extends keyof UiCommandMap> = (
  payload: UiCommandMap[K],
) => void;

class UiCommandBus {
  private handlers: {
    [K in keyof UiCommandMap]?: Set<UiCommandHandler<K>>;
  } = {};

  on<K extends keyof UiCommandMap>(
    command: K,
    handler: UiCommandHandler<K>,
  ): () => void {
    const bucket =
      (this.handlers[command] as Set<UiCommandHandler<K>> | undefined) ??
      new Set();
    bucket.add(handler);
    this.handlers[command] = bucket as UiCommandBus['handlers'][K];
    return () => {
      bucket.delete(handler);
    };
  }

  emit<K extends keyof UiCommandMap>(
    command: K,
    payload: UiCommandMap[K],
  ): void {
    const bucket = this.handlers[command] as
      | Set<UiCommandHandler<K>>
      | undefined;
    if (!bucket) return;
    bucket.forEach((handler) => handler(payload));
  }
}

export const uiCommandBus = new UiCommandBus();

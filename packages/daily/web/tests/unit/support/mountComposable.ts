import { defineComponent, h } from 'vue';
import { mount, type VueWrapper } from '@vue/test-utils';

export interface MountedComposable<TApi> {
  readonly getApi: () => TApi;
  readonly unmount: () => void;
}

export interface MountComposableOptions<
  TProps extends Record<string, unknown>,
> {
  readonly props?: TProps;
  readonly plugins?: unknown[];
}

/** Mounts a composable in component setup so lifecycle hooks are active. */
export function mountComposable<
  TApi,
  TProps extends Record<string, unknown> = Record<string, never>,
>(
  factory: (props: TProps) => TApi,
  options: MountComposableOptions<TProps> = {},
): MountedComposable<TApi> {
  let api: TApi | null = null;
  const wrapper: VueWrapper = mount(
    defineComponent({
      name: 'ComposableHarness',
      setup() {
        const props = (options.props ?? {}) as TProps;
        api = factory(props);
        return () => h('div');
      },
    }),
    {
      global: {
        plugins: options.plugins ?? [],
      },
    },
  );

  const getApi = (): TApi => {
    if (!api) {
      throw new Error('Composable is not initialized');
    }
    return api;
  };

  const unmount = (): void => {
    wrapper.unmount();
    api = null;
  };

  return { getApi, unmount };
}

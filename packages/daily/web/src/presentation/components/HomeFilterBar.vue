<script setup lang="ts">
import { ref, watch, h, computed } from 'vue';
import { Tag as TTag, Dropdown as TDropdown } from 'tdesign-vue-next';
import {
  Search,
  ListFilter,
  X,
  Calendar as CalendarIcon,
  Hash,
  Paperclip,
  ArrowDownWideNarrow,
} from 'lucide-vue-next';
import type { FilterToken, SortMode } from '@/presentation/filters/types';

const props = defineProps<{
  tokens: FilterToken[];
  placeholder?: string;
}>();

const emit = defineEmits<{
  (e: 'submit'): void;
  (e: 'add-token', token: Omit<FilterToken, 'id'>): void;
  (e: 'remove-token', id: string): void;
  (e: 'clear-all'): void;
}>();

const inputText = defineModel<string>('inputText', { required: true });
const inputRef = ref<HTMLInputElement | null>(null);
const isFocused = ref(false);
const activeSuggestionIndex = ref(0);

const allFilterOptions = [
  { label: '全文搜索', value: 'text:', desc: 'text:keyword', icon: Search },
  { label: '标签 (包含任一)', value: 'tag:', desc: 'tag:work', icon: Hash },
  { label: '标签 (必须包含)', value: 'tag+:', desc: 'tag+:urgent', icon: Hash },
  { label: '排除标签', value: 'tag-:', desc: 'tag-:deprecated', icon: X },
  {
    label: '开始日期',
    value: 'from:',
    desc: 'from:YYYY/MM/DD',
    icon: CalendarIcon,
  },
  {
    label: '结束日期',
    value: 'to:',
    desc: 'to:YYYY/MM/DD',
    icon: CalendarIcon,
  },
  {
    label: '包含附件',
    value: 'has:resource',
    desc: 'has:resource',
    icon: Paperclip,
  },
  {
    label: '无附件',
    value: 'has:no-resource',
    desc: 'has:no-resource',
    icon: Paperclip,
  },
  {
    label: '排序',
    value: 'sort:',
    desc: 'sort:created_at_desc',
    icon: ArrowDownWideNarrow,
  },
];

const suggestions = computed(() => {
  const val = inputText.value.toLowerCase().trim();

  if (val === '?' || val === '？') return allFilterOptions;

  if (!val) return [];

  // Try to match partial filter prefix
  const matched = allFilterOptions.filter(
    (opt) =>
      opt.label.toLowerCase().includes(val) ||
      opt.value.toLowerCase().includes(val) ||
      opt.desc.toLowerCase().includes(val),
  );

  return matched;
});

const showSuggestions = computed(
  () => isFocused.value && suggestions.value.length > 0,
);

const focusInput = () => {
  inputRef.value?.focus();
};

const selectSuggestion = (opt: any) => {
  if (opt.value === '') {
    focusInput();
    return;
  }

  // Auto-remove '?' trigger
  if (inputText.value.startsWith('?') || inputText.value.startsWith('？')) {
    inputText.value = '';
  }

  if (opt.value.endsWith(':')) {
    inputText.value = opt.value;
    activeSuggestionIndex.value = 0;
    focusInput();
  } else {
    if (opt.value.startsWith('has:')) {
      const isRes = opt.value === 'has:resource';
      emit('add-token', {
        type: 'has_resource',
        value: isRes,
        label: opt.value,
      });
    } else if (opt.value.startsWith('sort:')) {
      const mode = opt.value.replace('sort:', '') as SortMode;
      emit('add-token', { type: 'sort', value: mode, label: `sort:${mode}` });
    }
    inputText.value = '';
    focusInput();
  }
};

const handleBlur = () => {
  setTimeout(() => {
    isFocused.value = false;
  }, 300);
};

const normalizeDate = (raw: string): string | null => {
  const match = raw.match(/^(\d{4})[./-](\d{1,2})[./-](\d{1,2})$/);
  if (!match) return null;
  const [_, y, m, d] = match;
  if (!y || !m || !d) return null;
  return `${y}-${m.padStart(2, '0')}-${d.padStart(2, '0')}`;
};

const tryFinalParse = () => {
  const trimmed = inputText.value.trim();
  if (!trimmed) return;

  const tagMatch = trimmed.match(/^tag:([\w-]+)$/);
  if (tagMatch) {
    emit('add-token', {
      type: 'tag',
      value: tagMatch[1],
      label: `tag:${tagMatch[1]}`,
    });
    inputText.value = '';
    return;
  }
  const tagAllMatch = trimmed.match(/^tag\+:([\w-]+)$/);
  if (tagAllMatch) {
    emit('add-token', {
      type: 'tags_all',
      value: tagAllMatch[1],
      label: `tag+:${tagAllMatch[1]}`,
    });
    inputText.value = '';
    return;
  }
  const tagExcludeMatch = trimmed.match(/^tag-:([\w-]+)$/);
  if (tagExcludeMatch) {
    emit('add-token', {
      type: 'tags_exclude',
      value: tagExcludeMatch[1],
      label: `tag-:${tagExcludeMatch[1]}`,
    });
    inputText.value = '';
    return;
  }
  const dateMatch = trimmed.match(/^(from|to):(.+)$/);
  if (dateMatch && dateMatch[1] && dateMatch[2]) {
    const normalized = normalizeDate(dateMatch[2]);
    if (normalized) {
      emit('add-token', {
        type: dateMatch[1] as any,
        value: normalized,
        label: `${dateMatch[1]}:${normalized}`,
      });
      inputText.value = '';
      return;
    }
  }

  if (!trimmed.endsWith(':')) {
    const textValue = trimmed.startsWith('text:')
      ? trimmed.replace('text:', '')
      : trimmed;
    emit('add-token', {
      type: 'text',
      value: textValue,
      label: `text:${textValue}`,
    });
    inputText.value = '';
  } else {
    inputText.value = '';
  }
};

const handleKeydown = (e: KeyboardEvent) => {
  if (showSuggestions.value) {
    if (e.key === 'ArrowDown') {
      e.preventDefault();
      activeSuggestionIndex.value =
        (activeSuggestionIndex.value + 1) % suggestions.value.length;
      return;
    } else if (e.key === 'ArrowUp') {
      e.preventDefault();
      activeSuggestionIndex.value =
        (activeSuggestionIndex.value - 1 + suggestions.value.length) %
        suggestions.value.length;
      return;
    } else if (
      e.key === 'Enter' &&
      inputText.value !== '' &&
      !inputText.value.endsWith(' ')
    ) {
      const opt = suggestions.value[activeSuggestionIndex.value];
      if (opt) {
        e.preventDefault();
        selectSuggestion(opt);
        return; // Selection handled, stop and don't submit search yet
      }
    }
  }

  if (
    e.key === 'Backspace' &&
    inputText.value === '' &&
    props.tokens.length > 0
  ) {
    const lastToken = props.tokens[props.tokens.length - 1];
    if (lastToken) {
      emit('remove-token', lastToken.id);
    }
  } else if (e.key === 'Enter') {
    tryFinalParse();
    emit('submit');
  }
};

watch(inputText, (newVal) => {
  if (newVal.endsWith(' ')) {
    const trimmed = newVal.trim();

    const textMatch = trimmed.match(/^text:(.+)$/);
    if (textMatch && textMatch[1]) {
      emit('add-token', {
        type: 'text',
        value: textMatch[1],
        label: `text:${textMatch[1]}`,
      });
      inputText.value = '';
      return;
    }

    const dateMatch = trimmed.match(/^(from|to):(.+)$/);
    if (dateMatch && dateMatch[1] && dateMatch[2]) {
      const type = dateMatch[1] as 'from' | 'to';
      const normalized = normalizeDate(dateMatch[2]);
      if (normalized) {
        emit('add-token', {
          type,
          value: normalized,
          label: `${type}:${normalized}`,
        });
        inputText.value = '';
        return;
      }
    }

    const tagAllMatch = trimmed.match(/^tag\+:([\w-]+)$/);
    if (tagAllMatch) {
      emit('add-token', {
        type: 'tags_all',
        value: tagAllMatch[1],
        label: `tag+:${tagAllMatch[1]}`,
      });
      inputText.value = '';
      return;
    }

    const tagExcludeMatch = trimmed.match(/^tag-:([\w-]+)$/);
    if (tagExcludeMatch) {
      emit('add-token', {
        type: 'tags_exclude',
        value: tagExcludeMatch[1],
        label: `tag-:${tagExcludeMatch[1]}`,
      });
      inputText.value = '';
      return;
    }

    const tagMatch = trimmed.match(/^tag:([\w-]+)$/);
    if (tagMatch) {
      emit('add-token', {
        type: 'tag',
        value: tagMatch[1],
        label: `tag:${tagMatch[1]}`,
      });
      inputText.value = '';
      return;
    }

    if (trimmed === 'has:resource') {
      emit('add-token', {
        type: 'has_resource',
        value: true,
        label: 'has:resource',
      });
      inputText.value = '';
      return;
    }

    if (trimmed === 'has:no-resource') {
      emit('add-token', {
        type: 'has_resource',
        value: false,
        label: 'has:no-resource',
      });
      inputText.value = '';
      return;
    }
  }
});

const handleMenuClick = (data: any) => {
  if (data.value === 'has-resource') {
    emit('add-token', {
      type: 'has_resource',
      value: true,
      label: 'has:resource',
    });
  } else if (data.value === 'no-resource') {
    emit('add-token', {
      type: 'has_resource',
      value: false,
      label: 'has:no-resource',
    });
  } else if (data.value === 'date-from') {
    inputText.value = 'from:';
    focusInput();
  } else if (data.value === 'date-to') {
    inputText.value = 'to:';
    focusInput();
  } else if (data.value.startsWith('sort:')) {
    const mode = data.value.replace('sort:', '') as SortMode;
    emit('add-token', { type: 'sort', value: mode, label: `sort:${mode}` });
  }
};

defineExpose({
  focusInput,
});
</script>

<template>
  <div class="relative w-full">
    <div
      class="ui-filter-bar"
      :class="{ 'is-focused': isFocused }"
      @click="focusInput"
    >
      <div class="ui-filter-bar-prefix">
        <Search :size="18" class="text-muted" />
      </div>

      <div class="ui-filter-tokens">
        <TTag
          v-for="token in tokens"
          :key="token.id"
          closable
          size="medium"
          variant="light"
          :theme="
            token.type === 'tags_all'
              ? 'primary'
              : token.type === 'tags_exclude'
                ? 'danger'
                : 'default'
          "
          class="ui-filter-token-chip"
          @close="emit('remove-token', token.id)"
        >
          <template #icon>
            <Hash
              v-if="
                token.type === 'tag' ||
                token.type === 'tags_all' ||
                token.type === 'tags_exclude'
              "
              :size="12"
            />
            <CalendarIcon
              v-else-if="token.type === 'from' || token.type === 'to'"
              :size="12"
            />
            <Paperclip v-else-if="token.type === 'has_resource'" :size="12" />
            <ArrowDownWideNarrow v-else-if="token.type === 'sort'" :size="12" />
            <Search
              v-else-if="token.type === 'text' || token.type === 'search'"
              :size="12"
            />
          </template>
          {{ token.label }}
        </TTag>

        <input
          ref="inputRef"
          v-model="inputText"
          type="text"
          class="ui-filter-input"
          :placeholder="tokens.length === 0 ? placeholder : ''"
          @keydown="handleKeydown"
          @focus="isFocused = true"
          @blur="handleBlur"
        />
      </div>

      <div class="ui-filter-bar-suffix">
        <button
          v-if="tokens.length > 0 || inputText"
          class="ui-btn-icon-xs mr-1"
          title="清空"
          @click.stop="emit('clear-all')"
        >
          <X :size="14" />
        </button>

        <div class="flex items-center">
          <TDropdown
            :options="[
              {
                content: '起始日期',
                value: 'date-from',
                prefixIcon: () => h(CalendarIcon, { size: 14 }),
              },
              {
                content: '结束日期',
                value: 'date-to',
                prefixIcon: () => h(CalendarIcon, { size: 14 }),
              },
              {
                content: '包含附件',
                value: 'has-resource',
                prefixIcon: () => h(Paperclip, { size: 14 }),
                divider: true,
              },
              {
                content: '无附件',
                value: 'no-resource',
                prefixIcon: () => h(Paperclip, { size: 14 }),
              },
              {
                content: '最新创建',
                value: 'sort:created_at_desc',
                divider: true,
              },
              { content: '最早创建', value: 'sort:created_at_asc' },
              { content: '最近更新', value: 'sort:updated_at_desc' },
            ]"
            trigger="click"
            @click="handleMenuClick"
          >
            <button class="ui-btn-icon-xs" title="添加筛选" @click.stop>
              <ListFilter :size="16" />
            </button>
          </TDropdown>
        </div>
      </div>
    </div>

    <!-- Suggestion List -->
    <div v-show="showSuggestions" class="ui-filter-suggestions">
      <div
        v-for="(opt, index) in suggestions"
        :key="opt.value"
        class="ui-suggestion-item"
        :class="{ 'is-active': index === activeSuggestionIndex }"
        @mousedown="selectSuggestion(opt)"
        @mouseenter="activeSuggestionIndex = index"
      >
        <component :is="opt.icon" :size="14" class="mr-2 text-muted" />
        <span class="flex-1 font-medium">{{ opt.label }}</span>
        <code class="text-[10px] opacity-50">{{ opt.desc }}</code>
      </div>
    </div>
  </div>
</template>

<style scoped>
.ui-filter-bar {
  display: flex;
  align-items: center;
  background: var(--td-bg-color-container);
  border: 1px solid var(--td-component-border);
  border-radius: var(--td-radius-large);
  padding: 4px 8px;
  min-height: 48px;
  cursor: text;
  transition: all 0.2s;
  box-shadow: var(--td-shadow-1);
}

.ui-filter-bar.is-focused {
  border-color: var(--td-brand-color);
  box-shadow: 0 0 0 2px var(--td-brand-color-focus);
}

.ui-filter-bar-prefix {
  display: flex;
  align-items: center;
  padding: 0 8px;
}

.ui-filter-tokens {
  flex: 1;
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 6px;
  padding: 4px 0;
}

.ui-filter-token-chip {
  height: 28px;
  font-weight: 500;
}

.ui-filter-input {
  flex: 1;
  min-width: 120px;
  border: none;
  outline: none;
  background: transparent;
  color: var(--td-text-color-primary);
  font-size: 15px;
  height: 28px;
}

.ui-filter-bar-suffix {
  display: flex;
  align-items: center;
  padding-left: 8px;
  border-left: 1px solid var(--td-component-border);
  margin-left: 4px;
}

.ui-filter-suggestions {
  position: absolute;
  top: calc(100% + 8px);
  left: 0;
  right: 0;
  background: var(--td-bg-color-container);
  border: 1px solid var(--td-component-border);
  border-radius: var(--td-radius-medium);
  box-shadow: var(--td-shadow-2);
  z-index: 1000;
  padding: 4px;
  max-height: 300px;
  overflow-y: auto;
}

.ui-suggestion-item {
  display: flex;
  align-items: center;
  padding: 8px 12px;
  border-radius: var(--td-radius-small);
  cursor: pointer;
  font-size: 13px;
  color: var(--td-text-color-primary);
  transition: all 0.2s;
}

.ui-suggestion-item.is-active {
  background: var(--td-bg-color-container-hover);
  color: var(--td-brand-color);
}

.ui-suggestion-item code {
  background: var(--td-bg-color-secondarycontainer);
  padding: 2px 4px;
  border-radius: 4px;
  margin-left: 8px;
}

.ui-btn-icon-xs {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  border-radius: 6px;
  border: none;
  background: transparent;
  cursor: pointer;
  color: var(--td-text-color-secondary);
  transition: background 0.2s;
}

.ui-btn-icon-xs:hover {
  background: var(--td-bg-color-container-hover);
  color: var(--td-brand-color-hover);
}
</style>

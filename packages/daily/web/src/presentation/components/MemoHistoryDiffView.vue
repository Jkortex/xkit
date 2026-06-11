<script setup lang="ts">
import { computed } from 'vue';
import * as Diff from 'diff';

const props = defineProps<{
  oldText: string;
  newText: string;
}>();

interface DiffLine {
  content: string;
  type: 'added' | 'removed' | 'unchanged';
}

const lines = computed(() => {
  const diffResult = Diff.diffLines(props.oldText, props.newText);
  const allLines: DiffLine[] = [];

  diffResult.forEach((part) => {
    // 拆分多行，确保每一行都有对应的符号
    const partLines = part.value.split('\n');
    // 如果最后一行是空的（因为 split 带来的），通常需要过滤掉，除非它是唯一的空行
    if (partLines.length > 1 && partLines[partLines.length - 1] === '') {
      partLines.pop();
    }

    partLines.forEach((line) => {
      allLines.push({
        content: line,
        type: part.added ? 'added' : part.removed ? 'removed' : 'unchanged',
      });
    });
  });

  return allLines;
});
</script>

<template>
  <div
    class="ui-diff-container font-mono text-[13px] border border-border rounded-xl overflow-hidden bg-surface"
  >
    <div class="ui-diff-scroll max-h-full overflow-y-auto">
      <table class="w-full border-collapse">
        <tbody>
          <tr
            v-for="(line, index) in lines"
            :key="index"
            :class="[
              'ui-diff-row group',
              line.type === 'added' ? 'is-added' : '',
              line.type === 'removed' ? 'is-removed' : '',
            ]"
          >
            <!-- Line Number -->
            <td
              class="ui-diff-num select-none text-right pr-3 pl-4 w-12 opacity-30 border-r border-border"
            >
              {{ index + 1 }}
            </td>

            <!-- Sign -->
            <td class="ui-diff-sign select-none text-center w-8 font-bold">
              <template v-if="line.type === 'added'">+</template>
              <template v-else-if="line.type === 'removed'">-</template>
              <template v-else>&nbsp;</template>
            </td>

            <!-- Content -->
            <td
              class="ui-diff-content py-0.5 px-4 break-all whitespace-pre-wrap"
            >
              {{ line.content }}
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<style scoped>
.ui-diff-container {
  box-shadow: inset 0 2px 4px 0 rgb(0 0 0 / 0.05);
}

.ui-diff-row {
  transition: background-color 0.1s ease;
}

.ui-diff-num {
  font-size: 11px;
  background: color-mix(in oklab, var(--color-page) 60%, transparent);
}

.ui-diff-sign {
  font-size: 14px;
}

/* Added State */
.is-added {
  background-color: color-mix(in oklab, #22c55e 12%, transparent);
}
.is-added .ui-diff-sign,
.is-added .ui-diff-content {
  color: #15803d;
}
:root[theme-mode='dark'] .is-added .ui-diff-sign,
:root[theme-mode='dark'] .is-added .ui-diff-content {
  color: #4ade80;
}

/* Removed State */
.is-removed {
  background-color: color-mix(in oklab, #ef4444 12%, transparent);
}
.is-removed .ui-diff-sign,
.is-removed .ui-diff-content {
  color: #b91c1c;
  text-decoration: line-through;
  text-decoration-thickness: 1px;
  text-decoration-color: rgba(185, 28, 28, 0.4);
}
:root[theme-mode='dark'] .is-removed .ui-diff-sign,
:root[theme-mode='dark'] .is-removed .ui-diff-content {
  color: #f87171;
  text-decoration-color: rgba(248, 113, 113, 0.4);
}

.ui-diff-row:hover {
  filter: brightness(0.98);
}

:root[theme-mode='dark'] .ui-diff-row:hover {
  filter: brightness(1.1);
}
</style>

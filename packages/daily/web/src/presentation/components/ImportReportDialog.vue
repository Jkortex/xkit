<script setup lang="ts">
import { toRef } from 'vue';
import { Dialog as TDialog } from 'tdesign-vue-next';
import { useDialogOpenFlag } from '@/presentation/composables/hotkeys/useDialogOpenFlag';
import type { ImportReportDTO } from '@/application/ports/dto/Resource';

interface ImportReportDialogProps {
  visible: boolean;
  report: ImportReportDTO | null;
  humanizeReason: (reason: string) => string;
}

const props = defineProps<ImportReportDialogProps>();

useDialogOpenFlag(toRef(props, 'visible'));

const emit = defineEmits<{
  (e: 'update:visible', value: boolean): void;
}>();
</script>

<template>
  <TDialog
    :visible="props.visible"
    header="导入报告"
    :footer="false"
    width="680px"
    class="import-report-dialog"
    destroy-on-close
    @update:visible="(value) => emit('update:visible', value)"
  >
    <div v-if="props.report" class="ui-dialog-body space-y-4 text-sm">
      <div class="ui-dialog-caption">该结果仅反映当前账号导入范围。</div>
      <div class="grid grid-cols-2 gap-3">
        <div class="ui-dialog-section">
          <div class="ui-dialog-caption">笔记</div>
          <div class="mt-1 font-semibold">
            导入 {{ props.report.report.memos.imported }} / 跳过
            {{ props.report.report.memos.skipped }}
          </div>
        </div>
        <div class="ui-dialog-section">
          <div class="ui-dialog-caption">资源</div>
          <div class="mt-1 font-semibold">
            导入 {{ props.report.report.resources.imported }} / 跳过
            {{ props.report.report.resources.skipped }}
          </div>
        </div>
      </div>

      <div class="space-y-2">
        <div class="ui-dialog-caption font-semibold">笔记跳过明细</div>
        <div
          v-if="props.report.report.memos.details.length === 0"
          class="ui-list-empty"
        >
          无跳过记录
        </div>
        <div v-else class="ui-list-shell max-h-40">
          <div
            v-for="item in props.report.report.memos.details"
            :key="`memo-${item.key}-${item.reason}`"
            class="ui-list-row"
          >
            <span class="truncate text-primary-text">{{ item.key }}</span>
            <span class="text-secondary">{{
              props.humanizeReason(item.reason)
            }}</span>
          </div>
        </div>
      </div>

      <div class="space-y-2">
        <div class="ui-dialog-caption font-semibold">资源跳过明细</div>
        <div
          v-if="props.report.report.resources.details.length === 0"
          class="ui-list-empty"
        >
          无跳过记录
        </div>
        <div v-else class="ui-list-shell max-h-40">
          <div
            v-for="item in props.report.report.resources.details"
            :key="`resource-${item.key}-${item.reason}`"
            class="ui-list-row"
          >
            <span class="truncate text-primary-text">{{ item.key }}</span>
            <span class="text-secondary">{{
              props.humanizeReason(item.reason)
            }}</span>
          </div>
        </div>
      </div>
    </div>
  </TDialog>
</template>

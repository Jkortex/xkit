<script setup lang="ts">
import { Loading as TLoading } from 'tdesign-vue-next';
import { X } from 'lucide-vue-next';
import type { ResourceVM } from '../view-models/ResourceVM';

defineProps<{
  resources: ResourceVM[];
  uploading: boolean;
}>();

defineEmits<{
  (e: 'remove', id: string): void;
}>();
</script>

<template>
  <div
    v-if="resources.length > 0 || uploading"
    class="flex flex-wrap gap-2 mt-3"
  >
    <div
      v-for="res in resources"
      :key="res.id"
      class="relative w-20 h-20 rounded-lg overflow-hidden border border-border group"
    >
      <img
        v-if="res.isImage"
        :src="res.url"
        class="w-full h-full object-cover"
      />
      <div
        v-else
        class="w-full h-full flex items-center justify-center bg-surface text-tiny text-center p-1"
      >
        {{ res.filename }}
      </div>

      <button
        class="absolute top-1 right-1 bg-black/50 text-white rounded-full p-0.5 opacity-0 group-hover:opacity-100 transition-opacity cursor-pointer"
        @click="$emit('remove', res.id)"
      >
        <X :size="12" />
      </button>
    </div>

    <!-- Uploading Placeholder -->
    <div
      v-if="uploading"
      class="w-20 h-20 rounded-lg border border-dashed border-border flex items-center justify-center"
    >
      <TLoading size="small" />
    </div>
  </div>
</template>

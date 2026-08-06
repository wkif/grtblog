<script setup lang="ts">
import { NButton, NPopconfirm, NSpace, NTooltip } from 'naive-ui'
import { computed } from 'vue'

import { formatMediaDuration } from '../composables/useMediaUtils'

import type { PhotoItem } from '@/services/albums'

const props = defineProps<{
  item: PhotoItem
  index: number
  total: number
  subtitle?: string
}>()

const emit = defineEmits<{
  edit: []
  delete: []
  move: [direction: -1 | 1]
}>()

const previewUrl = computed(() =>
  props.item.mediaType === 'video'
    ? props.item.posterUrl || ''
    : props.item.thumbnailUrl || props.item.url,
)
const duration = computed(() => formatMediaDuration(props.item.durationMs))
</script>

<template>
  <article
    class="group overflow-hidden rounded-lg border border-current/5 transition-all hover:border-current/15 hover:shadow-sm"
  >
    <div class="relative aspect-square overflow-hidden bg-current/3">
      <img
        v-if="previewUrl"
        :src="previewUrl"
        :alt="item.caption || ''"
        class="h-full w-full object-cover"
        loading="lazy"
      />
      <div
        v-else
        class="grid h-full w-full place-items-center bg-neutral-900 text-white/55"
      >
        <div class="iconify text-4xl ph--video-camera" />
      </div>

      <div
        class="absolute top-1.5 left-1.5 flex h-5 w-5 items-center justify-center rounded-full bg-black/55 text-[10px] font-medium text-white"
      >
        {{ index + 1 }}
      </div>
      <div
        v-if="item.mediaType === 'video'"
        class="absolute inset-0 grid place-items-center bg-black/10"
      >
        <div
          class="grid h-11 w-11 place-items-center rounded-full bg-black/60 text-xl text-white shadow-lg backdrop-blur-sm"
        >
          <div class="iconify ph--play-fill" />
        </div>
      </div>
      <span
        v-if="duration"
        class="absolute right-1.5 bottom-1.5 rounded bg-black/70 px-1.5 py-0.5 font-mono text-[10px] text-white"
      >
        {{ duration }}
      </span>
    </div>

    <div class="px-2.5 py-2">
      <button
        class="block min-h-5 w-full truncate text-left text-xs opacity-60"
        type="button"
        @click="emit('edit')"
      >
        {{ item.caption || subtitle || '点击编辑信息' }}
      </button>
      <div class="mt-1.5 flex items-center justify-between">
        <NSpace :size="2">
          <NTooltip>
            <template #trigger>
              <NButton
                quaternary
                circle
                size="tiny"
                :disabled="index === 0"
                @click="emit('move', -1)"
              >
                <template #icon><div class="iconify ph--caret-left" /></template>
              </NButton>
            </template>
            前移
          </NTooltip>
          <NTooltip>
            <template #trigger>
              <NButton
                quaternary
                circle
                size="tiny"
                :disabled="index === total - 1"
                @click="emit('move', 1)"
              >
                <template #icon><div class="iconify ph--caret-right" /></template>
              </NButton>
            </template>
            后移
          </NTooltip>
          <NTooltip>
            <template #trigger>
              <NButton
                quaternary
                circle
                size="tiny"
                @click="emit('edit')"
              >
                <template #icon><div class="iconify ph--pencil-simple" /></template>
              </NButton>
            </template>
            编辑
          </NTooltip>
        </NSpace>
        <NPopconfirm @positive-click="emit('delete')">
          <template #trigger>
            <NButton
              quaternary
              circle
              size="tiny"
              type="error"
            >
              <template #icon><div class="iconify ph--trash" /></template>
            </NButton>
          </template>
          确定删除这项媒体？
        </NPopconfirm>
      </div>
    </div>
  </article>
</template>

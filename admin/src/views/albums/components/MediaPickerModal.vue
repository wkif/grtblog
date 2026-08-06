<script setup lang="ts">
import {
  NButton,
  NButtonGroup,
  NCard,
  NEmpty,
  NModal,
  NPagination,
  NSpin,
  useMessage,
} from 'naive-ui'
import { ref, watch } from 'vue'

import { listUploads, type UploadFileResponse } from '@/services/uploads'

const props = defineProps<{ show: boolean }>()
const emit = defineEmits<{
  'update:show': [value: boolean]
  select: [item: UploadFileResponse]
}>()

const message = useMessage()
const loading = ref(false)
const type = ref<'picture' | 'video'>('picture')
const items = ref<UploadFileResponse[]>([])
const selected = ref<UploadFileResponse | null>(null)
const total = ref(0)
const page = ref(1)
const pageSize = 12

async function fetchList() {
  loading.value = true
  try {
    const result = await listUploads({ page: page.value, pageSize, type: type.value })
    items.value = result.items
    total.value = result.total
  } catch (error) {
    message.error(error instanceof Error ? error.message : '加载媒体列表失败')
  } finally {
    loading.value = false
  }
}

function changeType(next: 'picture' | 'video') {
  type.value = next
  page.value = 1
  selected.value = null
  fetchList()
}

function close() {
  emit('update:show', false)
}

function confirm() {
  if (!selected.value) return
  emit('select', selected.value)
  close()
}

watch(
  () => props.show,
  (show) => {
    if (!show) return
    page.value = 1
    selected.value = null
    fetchList()
  },
)
</script>

<template>
  <NModal
    :show="show"
    style="width: 760px; max-width: 92vw"
    @update:show="emit('update:show', $event)"
  >
    <NCard
      title="选择媒体"
      closable
      @close="close"
    >
      <div class="flex flex-col gap-4">
        <NButtonGroup size="small">
          <NButton
            :type="type === 'picture' ? 'primary' : 'default'"
            @click="changeType('picture')"
          >
            图片
          </NButton>
          <NButton
            :type="type === 'video' ? 'primary' : 'default'"
            @click="changeType('video')"
          >
            视频
          </NButton>
        </NButtonGroup>

        <NSpin :show="loading">
          <div
            v-if="items.length"
            class="grid grid-cols-2 gap-3 sm:grid-cols-4"
          >
            <button
              v-for="item in items"
              :key="item.id"
              type="button"
              class="overflow-hidden rounded-lg border-2 text-left transition-colors"
              :class="
                selected?.id === item.id
                  ? 'border-primary'
                  : 'border-transparent hover:border-current/20'
              "
              @click="selected = item"
            >
              <div
                class="relative grid aspect-square place-items-center overflow-hidden bg-neutral-100 dark:bg-neutral-800"
              >
                <img
                  v-if="item.type === 'picture'"
                  :src="item.thumbnailUrl || item.publicUrl"
                  :alt="item.name"
                  class="h-full w-full object-cover"
                  loading="lazy"
                />
                <div
                  v-else
                  class="iconify text-4xl opacity-35 ph--video-camera"
                />
                <div
                  v-if="selected?.id === item.id"
                  class="absolute inset-0 grid place-items-center bg-primary/20"
                >
                  <div class="iconify text-2xl text-white ph--check-circle-fill" />
                </div>
              </div>
              <div class="truncate px-2 py-1.5 text-xs opacity-70">{{ item.name }}</div>
            </button>
          </div>
          <NEmpty
            v-else-if="!loading"
            description="暂无媒体"
            class="py-10"
          />
        </NSpin>

        <div class="flex items-center justify-between">
          <NPagination
            v-if="total > pageSize"
            v-model:page="page"
            :page-size="pageSize"
            :item-count="total"
            @update:page="fetchList"
          />
          <div v-else />
          <div class="flex gap-2">
            <NButton @click="close">取消</NButton>
            <NButton
              type="primary"
              :disabled="!selected"
              @click="confirm"
              >添加</NButton
            >
          </div>
        </div>
      </div>
    </NCard>
  </NModal>
</template>

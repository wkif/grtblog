<script setup lang="ts">
import { NButton, NCard, NForm, NGrid, NInput, NInputNumber, NSelect, NSpace, NSwitch, NFormItemGi } from 'naive-ui'
import { computed, onMounted, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import { useDiscreteApi } from '@/composables/useDiscreteApi'
import { createMediaRecord, getMediaRecord, getMediaRecordDetails, searchMediaRecords, updateMediaRecord } from '@/services/media-records'

import type { MediaRecordPayload, MediaSearchResult } from '@/services/media-records'

const route = useRoute()
const router = useRouter()
const { message } = useDiscreteApi()
const isEdit = computed(() => Boolean(route.params.id))
const loading = ref(false)
const searching = ref(false)
const applyingResult = ref('')
const query = ref('')
const results = ref<MediaSearchResult[]>([])
const form = reactive<MediaRecordPayload>({ title: '', mediaType: 'movie', provider: 'tmdb', status: 'planned', progress: 0, isPublished: true })
const isTV = computed(() => form.mediaType === 'tv')
const progressLabel = computed(() => isTV.value ? '当前集数' : '观看进度 (%)')

const statusOptions = [{ label: '想看', value: 'planned' }, { label: '在看', value: 'watching' }, { label: '看完', value: 'completed' }, { label: '弃剧', value: 'dropped' }]
const typeOptions = [{ label: '电影', value: 'movie' }, { label: '剧集', value: 'tv' }]

async function search() {
  if (!query.value.trim()) return
  searching.value = true
  try { results.value = await searchMediaRecords(query.value, form.mediaType) } catch (error) { message.error(error instanceof Error ? error.message : '影视搜索失败') } finally { searching.value = false }
}

function fillResult(item: MediaSearchResult) {
  form.title = item.title
  form.originalTitle = item.originalTitle || null
  form.mediaType = item.mediaType
  form.providerId = item.providerId
  form.poster = item.poster || null
  form.backdrop = item.backdrop || null
  form.overview = item.overview || null
  form.releaseDate = item.releaseDate ? `${item.releaseDate}T00:00:00Z` : null
  form.runtimeMinutes = item.runtimeMinutes || null
  form.totalEpisodes = item.totalEpisodes || null
  if (item.mediaType === 'tv') form.progressTotal = item.totalEpisodes || null
  results.value = []
}

async function applyResult(item: MediaSearchResult) {
  const resultKey = `${item.mediaType}-${item.providerId}`
  applyingResult.value = resultKey
  try {
    fillResult(await getMediaRecordDetails(item.mediaType, item.providerId))
  } catch (error) {
    fillResult(item)
    message.warning(error instanceof Error ? `详情读取失败：${error.message}` : '详情读取失败，已使用搜索摘要')
  } finally {
    applyingResult.value = ''
  }
}

async function load() {
  if (!isEdit.value) return
  loading.value = true
  try {
    Object.assign(form, await getMediaRecord(Number(route.params.id)))
    const needsDetails = form.provider === 'tmdb' && form.providerId && (form.mediaType === 'tv' ? !form.totalEpisodes : !form.runtimeMinutes)
    if (needsDetails) {
      const details = await getMediaRecordDetails(form.mediaType, form.providerId!)
      if (!form.runtimeMinutes) form.runtimeMinutes = details.runtimeMinutes || null
      if (!form.totalEpisodes) form.totalEpisodes = details.totalEpisodes || null
    }
    if (form.mediaType === 'tv' && !form.progressTotal && form.totalEpisodes) form.progressTotal = form.totalEpisodes
  } catch (error) { message.error(error instanceof Error ? error.message : '加载影视记录失败') } finally { loading.value = false }
}

async function save() {
  if (!form.title.trim()) { message.warning('请先填写标题'); return }
  if (isTV.value) form.progressTotal = form.totalEpisodes || form.progressTotal || null
  else form.progressTotal = null
  loading.value = true
  try {
    if (isEdit.value) await updateMediaRecord(Number(route.params.id), form)
    else await createMediaRecord(form)
    message.success('影视记录已保存')
    router.push({ name: 'mediaRecordList' })
  } catch (error) { message.error(error instanceof Error ? error.message : '保存失败') } finally { loading.value = false }
}

onMounted(load)
</script>

<template>
  <div class="p-4">
    <NCard :title="isEdit ? '编辑影视记录' : '新增影视记录'" segmented>
      <div class="mb-5 flex gap-2">
        <NInput v-model:value="query" placeholder="从 TMDB 搜索电影或剧集" @keyup.enter="search" />
        <NButton type="primary" :loading="searching" @click="search">搜索</NButton>
      </div>
      <div v-if="results.length" class="mb-6 grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
        <button v-for="item in results" :key="`${item.mediaType}-${item.providerId}`" type="button" class="flex gap-3 rounded-lg border border-current/10 p-3 text-left transition hover:border-primary hover:bg-primary/5 disabled:cursor-wait disabled:opacity-60" :disabled="Boolean(applyingResult)" @click="applyResult(item)">
          <img v-if="item.poster" :src="item.poster" class="h-20 w-14 rounded object-cover" alt="" />
          <div class="min-w-0"><div class="truncate font-medium">{{ item.title }}</div><div class="mt-1 text-xs opacity-60">{{ item.mediaType === 'tv' ? '剧集' : '电影' }} · {{ item.releaseDate || '未知年份' }}</div><p class="mt-2 line-clamp-2 text-xs opacity-60">{{ applyingResult === `${item.mediaType}-${item.providerId}` ? '正在读取完整详情…' : item.overview }}</p></div>
        </button>
      </div>
      <NForm label-placement="left" label-width="100" :disabled="loading">
        <NGrid :cols="2" :x-gap="24" responsive="screen" item-responsive>
          <NFormItemGi :span="2" label="标题"><NInput v-model:value="form.title" /></NFormItemGi>
          <NFormItemGi label="类型"><NSelect v-model:value="form.mediaType" :options="typeOptions" /></NFormItemGi>
          <NFormItemGi label="状态"><NSelect v-model:value="form.status" :options="statusOptions" /></NFormItemGi>
          <NFormItemGi :label="progressLabel"><NInputNumber v-model:value="form.progress" :min="0" :max="isTV ? undefined : 100" /></NFormItemGi>
          <NFormItemGi v-if="isTV" label="总集数"><NInputNumber v-model:value="form.totalEpisodes" :min="0" /></NFormItemGi>
          <NFormItemGi v-else label="片长（分钟）"><NInputNumber v-model:value="form.runtimeMinutes" :min="0" /></NFormItemGi>
          <NFormItemGi label="评分"><NInputNumber v-model:value="form.rating" :min="0" :max="10" :step="0.5" /></NFormItemGi>
          <NFormItemGi label="公开展示"><NSwitch v-model:value="form.isPublished" /></NFormItemGi>
          <NFormItemGi :span="2" label="备注"><NInput v-model:value="form.note" type="textarea" :rows="4" /></NFormItemGi>
        </NGrid>
        <NSpace justify="end"><NButton @click="router.back()">取消</NButton><NButton type="primary" :loading="loading" @click="save">保存记录</NButton></NSpace>
      </NForm>
    </NCard>
  </div>
</template>

<script setup lang="ts">
import { NAlert, NButton, useMessage } from 'naive-ui'
import { ref } from 'vue'

import {
  listSysConfigs,
  testStorageConnection,
  updateSysConfigs,
  type SysConfigUpdateItem,
} from '@/services/sysconfig'

import ConfigPanel from '../ConfigPanel'

const emit = defineEmits<{ 'dirty-change': [dirty: boolean] }>()
const message = useMessage()
const testing = ref(false)
const panelRef = ref<{
  getPendingItems: () => SysConfigUpdateItem[]
  getValue: (key: string) => unknown
} | null>(null)

async function handleTestConnection() {
  const provider = String(panelRef.value?.getValue('storage.provider') ?? 'local').trim() || 'local'
  if (provider !== 'aliyun_oss') {
    message.warning('仅阿里云 OSS 存储模式支持连接自检')
    return
  }

  testing.value = true
  try {
    const pendingItems = panelRef.value?.getPendingItems() ?? []
    const result = await testStorageConnection(pendingItems)
    message.success(`连接成功：Bucket ${result.bucket}，Endpoint ${result.endpoint}`)
  } catch (err) {
    message.error(err instanceof Error ? err.message : 'OSS 连接测试失败')
  } finally {
    testing.value = false
  }
}
</script>

<template>
  <div class="space-y-4">
    <NAlert type="info" show-icon>
      切换存储后端或调整目录规则后，只影响新的上传和新的保存流程；历史资源仍按原始存储位置和原始公开 URL 兼容读取。
    </NAlert>

    <div class="flex justify-end">
      <NButton secondary :loading="testing" @click="handleTestConnection">连接自检 / 配置测试</NButton>
    </div>

    <ConfigPanel
      ref="panelRef"
      :list-fn="listSysConfigs"
      :update-fn="updateSysConfigs"
      title="资源存储"
      description="上传大小、草稿目录、正式目录和 OSS 存储配置"
      :filter-groups="['storage/upload', 'storage/provider', 'storage/oss']"
      :on-dirty-change="(dirty: boolean) => emit('dirty-change', dirty)"
    />
  </div>
</template>

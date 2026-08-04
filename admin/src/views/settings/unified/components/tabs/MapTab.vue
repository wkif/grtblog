<script setup lang="ts">
import { listSysConfigs, updateSysConfigs } from '@/services/sysconfig'

import ConfigPanel from '../ConfigPanel'

const emit = defineEmits<{ 'dirty-change': [dirty: boolean] }>()
</script>

<template>
  <ConfigPanel
    :list-fn="listSysConfigs"
    :update-fn="updateSysConfigs"
    title="地图服务"
    description="足迹地图的底图服务与访问凭证。"
    :filter-groups="['map']"
    :on-dirty-change="(dirty: boolean) => emit('dirty-change', dirty)"
  />
</template>

import { NButton, NCard, NDataTable, NInput, NPagination, NSpace, NTag } from 'naive-ui'
import { defineComponent, ref } from 'vue'
import { useRouter } from 'vue-router'

import { ScrollContainer } from '@/components'
import { useDiscreteApi } from '@/composables/useDiscreteApi'
import { deleteMediaRecord, listMediaRecords } from '@/services/media-records'

import type { MediaRecord } from '@/services/media-records'
import type { DataTableColumns } from 'naive-ui'

const statusLabels = { planned: '想看', watching: '在看', completed: '看完', dropped: '弃剧' }

export default defineComponent({
  name: 'MediaRecordList',
  setup() {
    const router = useRouter()
    const { message } = useDiscreteApi()
    const rows = ref<MediaRecord[]>([])
    const loading = ref(false)
    const search = ref('')
    const page = ref(1)
    const pageSize = ref(20)
    const total = ref(0)

    async function refresh() {
      loading.value = true
      try {
        const result = await listMediaRecords({ page: page.value, pageSize: pageSize.value, search: search.value })
        rows.value = result.items
        total.value = result.total
      } catch (error) {
        message.error(error instanceof Error ? error.message : '加载影视记录失败')
      } finally {
        loading.value = false
      }
    }

    function edit(id?: number) {
      router.push({ name: id ? 'mediaRecordEdit' : 'mediaRecordCreate', params: id ? { id } : undefined })
    }

    async function remove(row: MediaRecord) {
      try {
        await deleteMediaRecord(row.id)
        message.success('影视记录已删除')
        await refresh()
      } catch (error) {
        message.error(error instanceof Error ? error.message : '删除失败')
      }
    }

    const columns: DataTableColumns<MediaRecord> = [
      { title: '海报', key: 'poster', width: 64, render: row => row.poster ? <img src={row.poster} class='h-12 w-8 rounded object-cover' /> : <div class='h-12 w-8 rounded bg-current/5' /> },
      { title: '标题', key: 'title', minWidth: 220, ellipsis: { tooltip: true }, render: row => <NButton text type='primary' onClick={() => edit(row.id)}>{row.title}</NButton> },
      { title: '类型', key: 'mediaType', width: 72, render: row => row.mediaType === 'tv' ? '剧集' : '电影' },
      { title: '状态', key: 'status', width: 84, render: row => <NTag size='small' bordered={false}>{statusLabels[row.status]}</NTag> },
      { title: '进度', key: 'progress', width: 100, render: row => row.mediaType === 'movie' ? `${row.progress}%` : `${row.progress}${row.progressTotal ? ` / ${row.progressTotal}` : ''}` },
      { title: '评分', key: 'rating', width: 70, render: row => row.rating ?? '—' },
      { title: '发布', key: 'isPublished', width: 72, render: row => <NTag size='small' bordered={false} type={row.isPublished ? 'success' : 'default'}>{row.isPublished ? '公开' : '私密'}</NTag> },
      { title: '操作', key: 'actions', width: 140, render: row => <NSpace size={4}><NButton text type='primary' size='small' onClick={() => edit(row.id)}>编辑</NButton><NButton text type='error' size='small' onClick={() => remove(row)}>删除</NButton></NSpace> },
    ]

    refresh()

    return () => (
      <ScrollContainer>
        <NCard title='影视记录' segmented>
          {{ 'header-extra': () => <NSpace><NInput v-model:value={search.value} placeholder='搜索标题' clearable onKeyup={(event: KeyboardEvent) => event.key === 'Enter' && refresh()} /><NButton type='primary' onClick={() => edit()}>新增记录</NButton></NSpace> }}
          <NDataTable remote loading={loading.value} columns={columns} data={rows.value} pagination={false} />
          <div class='mt-4 flex justify-end'><NPagination v-model:page={page.value} v-model:page-size={pageSize.value} page-count={Math.max(1, Math.ceil(total.value / pageSize.value))} show-size-picker onUpdate:page={refresh} onUpdate:page-size={() => { page.value = 1; refresh() }} /></div>
        </NCard>
      </ScrollContainer>
    )
  },
})

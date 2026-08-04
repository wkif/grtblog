import { NButton, NCard, NDataTable, NInput, NPagination, NSpace, NTag } from 'naive-ui'
import { defineComponent, ref } from 'vue'
import { useRouter } from 'vue-router'

import { ScrollContainer } from '@/components'
import { useDiscreteApi } from '@/composables/useDiscreteApi'
import { deleteFootprint, listFootprints } from '@/services/footprints'

import type { FootprintJourney } from '@/services/footprints'
import type { DataTableColumns } from 'naive-ui'

function formatDistance(value?: number | null) {
  return value == null ? '—' : `${(value / 1000).toFixed(value >= 10000 ? 1 : 2)} km`
}

function formatDuration(value?: number | null) {
  if (value == null) return '—'
  const hours = Math.floor(value / 3600)
  const minutes = Math.round((value % 3600) / 60)
  return hours ? `${hours} 小时 ${minutes} 分` : `${minutes} 分钟`
}

export default defineComponent({
  name: 'FootprintList',
  setup() {
    const router = useRouter()
    const { message, dialog } = useDiscreteApi()
    const rows = ref<FootprintJourney[]>([])
    const loading = ref(false)
    const search = ref('')
    const page = ref(1)
    const pageSize = ref(20)
    const total = ref(0)

    async function refresh() {
      loading.value = true
      try {
        const result = await listFootprints({
          page: page.value,
          pageSize: pageSize.value,
          search: search.value,
        })
        rows.value = result.items
        total.value = result.total
      } catch (error) {
        message.error(error instanceof Error ? error.message : '加载足迹失败')
      } finally {
        loading.value = false
      }
    }

    function edit(id?: number) {
      router.push({
        name: id ? 'footprintEdit' : 'footprintCreate',
        params: id ? { id } : undefined,
      })
    }

    function remove(row: FootprintJourney) {
      dialog.warning({
        title: '删除足迹',
        content: `确定删除「${row.title}」？`,
        positiveText: '删除',
        negativeText: '取消',
        async onPositiveClick() {
          try {
            await deleteFootprint(row.id)
            message.success('足迹记录已删除')
            await refresh()
          } catch (error) {
            message.error(error instanceof Error ? error.message : '删除失败')
          }
        },
      })
    }

    const columns: DataTableColumns<FootprintJourney> = [
      {
        title: '行程',
        key: 'title',
        minWidth: 220,
        ellipsis: { tooltip: true },
        render: (row) => (
          <NButton
            text
            type='primary'
            onClick={() => edit(row.id)}
          >
            {row.title}
          </NButton>
        ),
      },
      {
        title: '城市',
        key: 'city',
        width: 150,
        render: (row) =>
          `${row.place.cityName}${row.place.regionName ? ` · ${row.place.regionName}` : ''}`,
      },
      {
        title: '日期',
        key: 'journeyDate',
        width: 118,
        render: (row) => row.journeyDate.slice(0, 10),
      },
      {
        title: '徒步里程',
        key: 'distanceMeters',
        width: 112,
        render: (row) => formatDistance(row.distanceMeters),
      },
      {
        title: '徒步时长',
        key: 'durationSeconds',
        width: 126,
        render: (row) => formatDuration(row.durationSeconds),
      },
      {
        title: '相册',
        key: 'albums',
        width: 76,
        render: (row) => `${row.albums.length} 个`,
      },
      {
        title: '轨迹',
        key: 'trackUrl',
        width: 76,
        render: (row) =>
          row.trackUrl ? (
            <NTag
              size='small'
              bordered={false}
              type='info'
            >
              已关联
            </NTag>
          ) : (
            '—'
          ),
      },
      {
        title: '发布',
        key: 'isPublished',
        width: 72,
        render: (row) => (
          <NTag
            size='small'
            bordered={false}
            type={row.isPublished ? 'success' : 'default'}
          >
            {row.isPublished ? '公开' : '私密'}
          </NTag>
        ),
      },
      {
        title: '操作',
        key: 'actions',
        width: 130,
        render: (row) => (
          <NSpace size={4}>
            <NButton
              text
              type='primary'
              size='small'
              onClick={() => edit(row.id)}
            >
              编辑
            </NButton>
            <NButton
              text
              type='error'
              size='small'
              onClick={() => remove(row)}
            >
              删除
            </NButton>
          </NSpace>
        ),
      },
    ]

    refresh()

    return () => (
      <ScrollContainer>
        <NCard
          title='足迹行程'
          segmented
        >
          {{
            'header-extra': () => (
              <NSpace>
                <NInput
                  v-model:value={search.value}
                  placeholder='搜索行程或城市'
                  clearable
                  onKeyup={(event: KeyboardEvent) => event.key === 'Enter' && refresh()}
                />
                <NButton
                  type='primary'
                  onClick={() => edit()}
                >
                  新增行程
                </NButton>
              </NSpace>
            ),
            default: () => (
              <>
                <NDataTable
                  remote
                  loading={loading.value}
                  columns={columns}
                  data={rows.value}
                  pagination={false}
                />
                <div class='mt-4 flex justify-end'>
                  <NPagination
                    v-model:page={page.value}
                    v-model:page-size={pageSize.value}
                    page-count={Math.max(1, Math.ceil(total.value / pageSize.value))}
                    show-size-picker
                    onUpdate:page={refresh}
                    onUpdate:page-size={() => {
                      page.value = 1
                      refresh()
                    }}
                  />
                </div>
              </>
            ),
          }}
        </NCard>
      </ScrollContainer>
    )
  },
})

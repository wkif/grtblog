<script setup lang="ts">
import {
  NButton,
  NButtonGroup,
  NDivider,
  NDrawer,
  NDrawerContent,
  NEmpty,
  NForm,
  NFormItem,
  NInput,
  NModal,
  NSpace,
  NSpin,
  NSwitch,
  NTag,
  NTooltip,
  useMessage,
} from 'naive-ui'
import { ref, watch } from 'vue'

import { ScrollContainer } from '@/components'
import ImageInput from '@/components/image-picker/ImageInput.vue'
import { uploadFile } from '@/services/uploads'

import AlbumMediaCard from './components/AlbumMediaCard.vue'
import MediaPickerModal from './components/MediaPickerModal.vue'
import { useAlbumForm } from './composables/useAlbumForm'
import {
  extractRemoteVideoMetadata,
  extractVideoMetadata,
  formatMediaDuration,
} from './composables/useMediaUtils'
import {
  extractExif,
  exifDevice,
  exifShootingInfo,
  exifLocation,
} from './composables/usePhotoUtils'

import type { CreatePhotoPayload, PhotoItem } from '@/services/albums'
import type { UploadFileResponse } from '@/services/uploads'

const {
  isEdit,
  loading,
  saving,
  form,
  photos,
  save,
  uploadPhotos,
  updatePhoto,
  deletePhoto,
  handleReorder,
  markDirty,
} = useAlbumForm()

const uploading = ref(false)
const message = useMessage()
const showMeta = ref(false)
const showMediaPicker = ref(false)
const editingPhoto = ref<PhotoItem | null>(null)
const showPhotoModal = ref(false)
const photoForm = ref({
  description: '',
  caption: '',
  posterUrl: '',
})
watch(
  () => [
    form.title,
    form.description,
    form.cover,
    form.shortUrl,
    form.isPublished,
    form.allowComment,
  ],
  () => markDirty(),
  { deep: true },
)

async function handleMediaUpload(e: Event) {
  const input = e.target as HTMLInputElement
  if (!input.files?.length) return
  uploading.value = true
  try {
    const items: Omit<CreatePhotoPayload, 'sortOrder'>[] = []
    for (const file of Array.from(input.files)) {
      const res = await uploadFile(file, 'cache')
      if (res.type === 'video') {
        const metadata = await extractVideoMetadata(file)
        let posterUrl: string | null = null
        if (metadata.poster) {
          posterUrl = (await uploadFile(metadata.poster, 'cache')).publicUrl
        }
        items.push({
          url: res.publicUrl,
          mediaType: 'video',
          mimeType: file.type || null,
          posterUrl,
          durationMs: metadata.durationMs,
          width: metadata.width,
          height: metadata.height,
          exif: null,
        })
        continue
      }

      const exif = await extractExif(file)
      const merged = { ...(exif || {}) }
      if (res.imageMeta?.width) merged.imageWidth = res.imageMeta.width
      if (res.imageMeta?.height) merged.imageHeight = res.imageMeta.height
      if (res.imageMeta?.dominantColor) merged.dominantColor = res.imageMeta.dominantColor
      items.push({
        url: res.publicUrl,
        mediaType: 'image',
        mimeType: file.type || null,
        width: res.imageMeta?.width ?? null,
        height: res.imageMeta?.height ?? null,
        exif: Object.keys(merged).length ? merged : null,
      })
    }
    await uploadPhotos(items)
  } catch (error) {
    message.error(error instanceof Error ? error.message : '媒体上传失败')
  } finally {
    uploading.value = false
    input.value = ''
  }
}

async function handlePickFromGallery(item: UploadFileResponse) {
  if (item.type === 'picture') {
    await uploadPhotos([
      {
        url: item.publicUrl,
        mediaType: 'image',
        width: item.imageMeta?.width ?? null,
        height: item.imageMeta?.height ?? null,
        exif: item.imageMeta
          ? {
              imageWidth: item.imageMeta.width,
              imageHeight: item.imageMeta.height,
              dominantColor: item.imageMeta.dominantColor,
            }
          : null,
      },
    ])
    return
  }

  let durationMs: number | null = null
  let width: number | null = null
  let height: number | null = null
  let posterUrl: string | null = null
  try {
    const metadata = await extractRemoteVideoMetadata(item.publicUrl)
    durationMs = metadata.durationMs
    width = metadata.width
    height = metadata.height
    if (metadata.poster) posterUrl = (await uploadFile(metadata.poster, 'cache')).publicUrl
  } catch {
    message.warning('已添加视频，但浏览器未能自动生成封面，可在媒体信息中手动选择')
  }
  await uploadPhotos([
    { url: item.publicUrl, mediaType: 'video', durationMs, width, height, posterUrl },
  ])
}

function openPhotoEdit(photo: PhotoItem) {
  editingPhoto.value = photo
  photoForm.value = {
    description: photo.description ?? '',
    caption: photo.caption ?? '',
    posterUrl: photo.posterUrl ?? '',
  }
  showPhotoModal.value = true
}

async function savePhotoEdit() {
  if (!editingPhoto.value) return
  await updatePhoto(editingPhoto.value.id, {
    description: photoForm.value.description || null,
    caption: photoForm.value.caption || null,
    posterUrl: photoForm.value.posterUrl || null,
  })
  showPhotoModal.value = false
}

function movePhoto(index: number, direction: -1 | 1) {
  const ids = photos.value.map((p) => p.id)
  const targetIndex = index + direction
  if (targetIndex < 0 || targetIndex >= ids.length) return
  const a = ids[index]!
  const b = ids[targetIndex]!
  ids[index] = b
  ids[targetIndex] = a
  handleReorder(ids)
}
</script>

<template>
  <div class="relative flex h-full min-h-0 flex-col">
    <div
      v-if="loading"
      class="bg-naive-body/35 absolute inset-0 z-30 grid place-items-center backdrop-blur-[1px]"
    >
      <NSpin size="large" />
    </div>
    <header
      class="z-10 flex shrink-0 flex-col gap-3 px-6 py-6 backdrop-blur sm:h-20 sm:flex-row sm:items-center sm:justify-between sm:px-10 sm:py-0"
    >
      <div class="flex min-w-0 flex-1 items-center gap-3">
        <NInput
          v-model:value="form.title"
          :bordered="false"
          placeholder="输入相册标题..."
          class="flex-1 text-xl font-bold sm:text-2xl"
          :style="{ '--n-caret-color': 'var(--primary-color)', backgroundColor: 'transparent' }"
        />
      </div>

      <div class="flex shrink-0 flex-wrap items-center gap-2 sm:flex-nowrap sm:gap-3">
        <div class="flex items-center gap-1 text-[11px] opacity-50">
          <div class="iconify ph--link-simple" />
          <span>/albums/</span>
          <input
            v-model="form.shortUrl"
            class="w-20 border-b border-current/20 bg-transparent transition-colors outline-none focus:border-current/50 sm:w-28"
            placeholder="auto"
          />
        </div>

        <NButtonGroup size="small">
          <NButton
            :type="form.isPublished ? 'default' : 'primary'"
            :ghost="form.isPublished"
            @click="form.isPublished = false"
          >
            草稿
          </NButton>
          <NButton
            :type="form.isPublished ? 'primary' : 'default'"
            :ghost="!form.isPublished"
            @click="form.isPublished = true"
          >
            发布
          </NButton>
        </NButtonGroup>

        <NTooltip>
          <template #trigger>
            <NButton
              quaternary
              circle
              size="small"
              @click="showMeta = !showMeta"
            >
              <template #icon><div class="iconify ph--sliders-horizontal" /></template>
            </NButton>
          </template>
          元数据
        </NTooltip>

        <NButton
          type="primary"
          :loading="saving"
          @click="save"
        >
          <template #icon><div class="iconify ph--floppy-disk" /></template>
          {{ isEdit ? '保存' : '创建' }}
        </NButton>
      </div>
    </header>

    <main class="flex min-h-0 flex-1 overflow-hidden">
      <div class="flex min-h-0 flex-1 flex-col px-6 sm:px-10">
        <div class="mb-6 shrink-0">
          <NInput
            v-model:value="form.description"
            type="textarea"
            :bordered="false"
            placeholder="添加相册描述..."
            :autosize="{ minRows: 2, maxRows: 5 }"
            :style="{ backgroundColor: 'transparent' }"
          />
        </div>

        <div class="mb-4 flex shrink-0 items-center justify-between">
          <div class="flex items-center gap-2">
            <div class="iconify text-lg opacity-50 ph--images" />
            <span class="text-sm font-medium">媒体</span>
            <NTag
              v-if="photos.length > 0"
              size="small"
              round
              :bordered="false"
            >
              {{ photos.length }}
            </NTag>
          </div>
          <NSpace
            v-if="isEdit"
            size="small"
          >
            <NButton
              size="small"
              quaternary
              @click="showMediaPicker = true"
            >
              <template #icon><div class="iconify ph--folder-open" /></template>
              媒体库
            </NButton>
            <NButton
              size="small"
              :loading="uploading"
              tag="label"
              style="cursor: pointer"
            >
              <template #icon><div class="iconify ph--upload-simple" /></template>
              上传
              <input
                type="file"
                accept="image/*,video/*"
                multiple
                style="display: none"
                @change="handleMediaUpload"
              />
            </NButton>
          </NSpace>
        </div>

        <NEmpty
          v-if="!isEdit"
          description="请先创建相册，再添加媒体"
          class="py-16"
        />
        <NEmpty
          v-else-if="photos.length === 0"
          description="点击上方按钮添加图片或视频"
          class="py-16"
        >
          <template #icon>
            <div class="iconify text-4xl opacity-20 ph--camera" />
          </template>
        </NEmpty>

        <div
          v-else
          class="min-h-0 flex-1 overflow-hidden"
        >
          <ScrollContainer
            class="h-full"
            wrapper-class="!p-0 !pr-1"
          >
            <div class="grid grid-cols-2 gap-3 sm:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5">
              <AlbumMediaCard
                v-for="(photo, index) in photos"
                :key="photo.id"
                :item="photo"
                :index="index"
                :total="photos.length"
                :subtitle="photo.mediaType === 'image' ? exifDevice(photo.exif) : '视频'"
                @edit="openPhotoEdit(photo)"
                @delete="deletePhoto(photo.id)"
                @move="movePhoto(index, $event)"
              />
            </div>
          </ScrollContainer>
        </div>
      </div>
    </main>

    <NDrawer
      v-model:show="showMeta"
      placement="right"
      :width="380"
    >
      <NDrawerContent
        title="相册设置"
        :native-scrollbar="false"
        closable
      >
        <div class="flex flex-col gap-6">
          <div class="space-y-4">
            <div class="flex items-center gap-2 text-sm font-medium">
              <div class="iconify ph--image" />
              <span>封面</span>
            </div>
            <ImageInput v-model:value="form.cover" />
          </div>

          <NDivider style="margin: 0" />

          <div class="space-y-4">
            <div class="flex items-center gap-2 text-sm font-medium">
              <div class="iconify ph--gear-six" />
              <span>属性</span>
            </div>
            <NForm
              label-placement="left"
              label-width="auto"
              :show-feedback="false"
              class="space-y-3"
            >
              <NFormItem label="允许评论">
                <NSwitch v-model:value="form.allowComment" />
              </NFormItem>
            </NForm>
          </div>
        </div>
      </NDrawerContent>
    </NDrawer>

    <NModal
      v-model:show="showPhotoModal"
      preset="card"
      title="编辑媒体信息"
      style="width: 520px; max-width: 90vw"
    >
      <div class="flex flex-col gap-4">
        <div
          v-if="editingPhoto?.mediaType === 'image' && editingPhoto.exif"
          class="rounded-lg bg-current/3 p-3 text-xs"
        >
          <div class="mb-2 flex items-center gap-1.5 text-sm font-medium opacity-70">
            <div class="iconify ph--info" />
            EXIF 信息
          </div>
          <div class="grid grid-cols-2 gap-x-4 gap-y-1.5 opacity-60">
            <div v-if="exifDevice(editingPhoto.exif)">
              <span class="mr-1 iconify inline-block align-[-2px] ph--camera" />{{
                exifDevice(editingPhoto.exif)
              }}
            </div>
            <div v-if="editingPhoto.exif.lensModel">
              <span class="mr-1 iconify inline-block align-[-2px] ph--aperture" />{{
                editingPhoto.exif.lensModel
              }}
            </div>
            <div
              v-if="exifShootingInfo(editingPhoto.exif)"
              class="col-span-2"
            >
              {{ exifShootingInfo(editingPhoto.exif) }}
            </div>
            <div v-if="exifLocation(editingPhoto.exif)">
              <span class="mr-1 iconify inline-block align-[-2px] ph--map-pin" />{{
                exifLocation(editingPhoto.exif)
              }}
            </div>
            <div v-if="editingPhoto.exif.dateTimeOriginal">
              <span class="mr-1 iconify inline-block align-[-2px] ph--clock" />{{
                editingPhoto.exif.dateTimeOriginal
              }}
            </div>
            <div v-if="editingPhoto.exif.imageWidth && editingPhoto.exif.imageHeight">
              {{ editingPhoto.exif.imageWidth }} × {{ editingPhoto.exif.imageHeight }}
            </div>
          </div>
        </div>

        <div
          v-if="editingPhoto?.mediaType === 'video'"
          class="grid grid-cols-2 gap-2 rounded-lg bg-current/3 p-3 text-xs opacity-70"
        >
          <div><span class="opacity-50">类型</span> {{ editingPhoto.mimeType || 'video' }}</div>
          <div>
            <span class="opacity-50">时长</span>
            {{ formatMediaDuration(editingPhoto.durationMs) || '-' }}
          </div>
          <div class="col-span-2">
            <span class="opacity-50">尺寸</span>
            {{
              editingPhoto.width && editingPhoto.height
                ? `${editingPhoto.width} × ${editingPhoto.height}`
                : '-'
            }}
          </div>
        </div>

        <NForm
          label-placement="top"
          :show-feedback="false"
          class="space-y-3"
        >
          <NFormItem label="说明文字">
            <NInput
              v-model:value="photoForm.caption"
              placeholder="给这项媒体加一句话..."
            />
          </NFormItem>
          <NFormItem
            v-if="editingPhoto?.mediaType === 'video'"
            label="视频封面"
          >
            <ImageInput v-model:value="photoForm.posterUrl" />
          </NFormItem>
          <NFormItem label="详细描述">
            <NInput
              v-model:value="photoForm.description"
              type="textarea"
              :rows="3"
              placeholder="更多描述..."
            />
          </NFormItem>
        </NForm>
      </div>
      <template #footer>
        <NSpace justify="end">
          <NButton @click="showPhotoModal = false">取消</NButton>
          <NButton
            type="primary"
            @click="savePhotoEdit"
          >
            <template #icon><div class="iconify ph--check" /></template>
            保存
          </NButton>
        </NSpace>
      </template>
    </NModal>

    <MediaPickerModal
      v-model:show="showMediaPicker"
      @select="handlePickFromGallery"
    />
  </div>
</template>

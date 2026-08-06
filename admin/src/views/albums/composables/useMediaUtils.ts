export interface VideoMetadata {
  durationMs: number | null
  width: number | null
  height: number | null
  poster: File | null
}

const MEDIA_TIMEOUT_MS = 15_000

function waitForMediaEvent(target: HTMLMediaElement, event: string): Promise<void> {
  return new Promise((resolve, reject) => {
    const timer = window.setTimeout(() => {
      cleanup()
      reject(new Error('读取视频信息超时'))
    }, MEDIA_TIMEOUT_MS)
    const onSuccess = () => {
      cleanup()
      resolve()
    }
    const onError = () => {
      cleanup()
      reject(new Error('浏览器无法读取该视频'))
    }
    const cleanup = () => {
      window.clearTimeout(timer)
      target.removeEventListener(event, onSuccess)
      target.removeEventListener('error', onError)
    }
    target.addEventListener(event, onSuccess, { once: true })
    target.addEventListener('error', onError, { once: true })
  })
}

function canvasToPoster(canvas: HTMLCanvasElement, filename: string): Promise<File | null> {
  return new Promise((resolve) => {
    canvas.toBlob(
      (blob) => resolve(blob ? new File([blob], filename, { type: 'image/jpeg' }) : null),
      'image/jpeg',
      0.84,
    )
  })
}

async function readVideoMetadata(source: string, posterName: string): Promise<VideoMetadata> {
  const video = document.createElement('video')
  video.preload = 'metadata'
  video.muted = true
  video.playsInline = true
  const metadataReady = waitForMediaEvent(video, 'loadedmetadata')
  video.src = source

  try {
    await metadataReady
    const durationSeconds = Number.isFinite(video.duration) ? video.duration : 0
    const width = video.videoWidth || null
    const height = video.videoHeight || null
    let poster: File | null = null

    if (width && height) {
      const targetTime =
        durationSeconds > 0
          ? Math.min(Math.max(durationSeconds * 0.08, 0.1), Math.max(durationSeconds - 0.1, 0))
          : 0
      if (targetTime > 0) {
        video.currentTime = targetTime
        await waitForMediaEvent(video, 'seeked')
      } else if (video.readyState < HTMLMediaElement.HAVE_CURRENT_DATA) {
        await waitForMediaEvent(video, 'loadeddata')
      }

      const maxWidth = 1600
      const scale = Math.min(1, maxWidth / width)
      const canvas = document.createElement('canvas')
      canvas.width = Math.max(1, Math.round(width * scale))
      canvas.height = Math.max(1, Math.round(height * scale))
      const ctx = canvas.getContext('2d')
      if (ctx) {
        try {
          ctx.drawImage(video, 0, 0, canvas.width, canvas.height)
          poster = await canvasToPoster(canvas, posterName)
        } catch {
          poster = null
        }
      }
    }

    return {
      durationMs: durationSeconds > 0 ? Math.round(durationSeconds * 1000) : null,
      width,
      height,
      poster,
    }
  } finally {
    video.removeAttribute('src')
    video.load()
  }
}

export async function extractVideoMetadata(file: File): Promise<VideoMetadata> {
  const objectUrl = URL.createObjectURL(file)
  try {
    return await readVideoMetadata(
      objectUrl,
      `${file.name.replace(/\.[^.]+$/, '') || 'video'}-poster.jpg`,
    )
  } finally {
    URL.revokeObjectURL(objectUrl)
  }
}

export function extractRemoteVideoMetadata(url: string): Promise<VideoMetadata> {
  return readVideoMetadata(url, 'video-poster.jpg')
}

export function formatMediaDuration(durationMs?: number | null): string {
  if (!durationMs || durationMs <= 0) return ''
  const totalSeconds = Math.round(durationMs / 1000)
  const hours = Math.floor(totalSeconds / 3600)
  const minutes = Math.floor((totalSeconds % 3600) / 60)
  const seconds = totalSeconds % 60
  if (hours > 0)
    return `${hours}:${String(minutes).padStart(2, '0')}:${String(seconds).padStart(2, '0')}`
  return `${minutes}:${String(seconds).padStart(2, '0')}`
}

export const MAX_IMAGE_INPUT_SIZE = 25 * 1024 * 1024

const MAX_DIMENSION = 2048
const TARGET_SIZE = 2 * 1024 * 1024
const MAX_UPLOAD_SIZE = 10 * 1024 * 1024
const MIN_QUALITY = 0.55
const QUALITY_STEP = 0.08

function canvasToBlob(canvas: HTMLCanvasElement, type: string, quality: number): Promise<Blob> {
  return new Promise((resolve, reject) => {
    canvas.toBlob(
      blob => blob ? resolve(blob) : reject(new Error('Image compression failed.')),
      type,
      quality,
    )
  })
}

export async function prepareImageForUpload(file: File): Promise<File> {
  if (!file.type.startsWith('image/')) throw new Error('Please select an image.')
  if (file.size > MAX_IMAGE_INPUT_SIZE) throw new Error('The original image must be 25 MB or smaller.')

  const bitmap = await createImageBitmap(file)

  try {
    const scale = Math.min(1, MAX_DIMENSION / Math.max(bitmap.width, bitmap.height))
    const canvas = document.createElement('canvas')
    canvas.width = Math.max(1, Math.round(bitmap.width * scale))
    canvas.height = Math.max(1, Math.round(bitmap.height * scale))

    const context = canvas.getContext('2d')
    if (!context) throw new Error('Image processing is unavailable.')

    context.imageSmoothingEnabled = true
    context.imageSmoothingQuality = 'high'
    context.drawImage(bitmap, 0, 0, canvas.width, canvas.height)

    let quality = 0.82
    let blob = await canvasToBlob(canvas, 'image/webp', quality)
    let extension = 'webp'

    if (blob.type !== 'image/webp') {
      context.globalCompositeOperation = 'destination-over'
      context.fillStyle = '#ffffff'
      context.fillRect(0, 0, canvas.width, canvas.height)
      context.globalCompositeOperation = 'source-over'

      quality = 0.84
      blob = await canvasToBlob(canvas, 'image/jpeg', quality)
      extension = 'jpg'
    }

    while (blob.size > TARGET_SIZE && quality > MIN_QUALITY) {
      quality = Math.max(MIN_QUALITY, quality - QUALITY_STEP)
      blob = await canvasToBlob(canvas, blob.type, quality)
    }

    if (blob.size > MAX_UPLOAD_SIZE) {
      throw new Error('The compressed image is still larger than 10 MB.')
    }

    const nameWithoutExtension = file.name.replace(/\.[^.]+$/, '') || 'image'
    return new File([blob], `${nameWithoutExtension}.${extension}`, {
      type: blob.type,
      lastModified: Date.now(),
    })
  } finally {
    bitmap.close()
  }
}

<template>
  <div class="image-uploader">
    <div class="upload-area" @click="triggerUpload" v-if="!imageUrl && !uploading">
      <el-icon :size="32"><Plus /></el-icon>
      <span>点击上传图片</span>
      <input
          ref="fileInput"
          type="file"
          accept="image/jpeg,image/png,image/gif,image/webp"
          style="display: none"
          @change="handleFileChange"
      />
    </div>

    <div v-else-if="uploading" class="uploading-area">
      <el-icon class="is-loading" :size="32"><Loading /></el-icon>
      <span>上传中...</span>
    </div>

    <div v-else class="image-preview">
      <el-image :src="imageUrl" fit="cover" class="preview-img" />
      <div class="image-actions">
        <el-button type="danger" size="small" circle @click.stop="removeImage">
          <el-icon><Delete /></el-icon>
        </el-button>
        <el-button type="primary" size="small" circle @click.stop="triggerUpload">
          <el-icon><Edit /></el-icon>
        </el-button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { Plus, Delete, Edit, Loading } from '@element-plus/icons-vue'
import { uploadImage } from '../api/upload'

const props = defineProps({
  modelValue: {
    type: String,
    default: ''
  },
  batchId: {
    type: Number,
    default: null
  }
})

const emit = defineEmits(['update:modelValue', 'upload-success'])

const fileInput = ref(null)
const uploading = ref(false)
const imageUrl = ref(props.modelValue)

// 压缩图片，保留 GIF 原始文件以避免破坏动图
const compressImage = async (file) => {
  if (file.type === 'image/gif') {
    return file
  }

  const imageBitmap = await createImageBitmap(file)
  let { width, height } = imageBitmap
  const maxSize = 1920

  if (width > maxSize || height > maxSize) {
    if (width > height) {
      height = Math.round((height * maxSize) / width)
      width = maxSize
    } else {
      width = Math.round((width * maxSize) / height)
      height = maxSize
    }
  }

  const canvas = document.createElement('canvas')
  canvas.width = width
  canvas.height = height
  const ctx = canvas.getContext('2d')
  ctx.drawImage(imageBitmap, 0, 0, width, height)
  imageBitmap.close()

  const outputType = file.type === 'image/png' ? 'image/png' : file.type
  const quality = file.type === 'image/png' ? undefined : 0.8

  const blob = await new Promise((resolve, reject) => {
    canvas.toBlob((result) => {
      if (result) {
        resolve(result)
      } else {
        reject(new Error('压缩图片失败'))
      }
    }, outputType, quality)
  })

  return new File([blob], file.name, { type: blob.type })
}

// 触发文件选择
const triggerUpload = () => {
  fileInput.value?.click()
}

// 处理文件选择
const handleFileChange = async (event) => {
  const file = event.target.files[0]
  if (!file) return

  // 验证文件类型
  const allowedTypes = ['image/jpeg', 'image/png', 'image/gif', 'image/webp']
  if (!allowedTypes.includes(file.type)) {
    ElMessage.error('只支持 JPG、PNG、GIF、WEBP 格式的图片')
    return
  }

  // 验证文件大小（5MB）
  if (file.size > 5 * 1024 * 1024) {
    ElMessage.error('图片大小不能超过 5MB')
    return
  }

  uploading.value = true
  try {
    // 压缩图片后上传
    const compressedFile = await compressImage(file)
    const response = await uploadImage(compressedFile)
    const url = response.data.url

    imageUrl.value = url
    emit('update:modelValue', url)
    emit('upload-success', url)
    ElMessage.success('图片上传成功')
  } catch (error) {
    ElMessage.error(error.message || '上传失败')
    // 清空 input，允许重新上传
    if (fileInput.value) {
      fileInput.value.value = ''
    }
  } finally {
    uploading.value = false
  }
}

// 移除图片
const removeImage = () => {
  imageUrl.value = ''
  emit('update:modelValue', '')
  emit('upload-success', '')
}

// 监听外部变化
watch(() => props.modelValue, (newVal) => {
  if (newVal !== imageUrl.value) {
    imageUrl.value = newVal
  }
})
</script>

<style scoped>
.image-uploader {
  width: 100%;
}

.upload-area {
  width: 100%;
  min-height: 150px;
  border: 2px dashed #ddd;
  border-radius: 8px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 8px;
  cursor: pointer;
  transition: all 0.2s;
  background: #fafafa;
  padding: 20px;
}

.upload-area:hover {
  border-color: #409eff;
  background: #f0f7ff;
}

.upload-area .el-icon {
  color: #999;
}

.upload-area span {
  color: #999;
  font-size: 14px;
}

.uploading-area {
  width: 100%;
  min-height: 150px;
  border: 2px dashed #ddd;
  border-radius: 8px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 8px;
  background: #fafafa;
  padding: 20px;
}

.uploading-area .el-icon {
  color: #409eff;
}

.uploading-area span {
  color: #666;
  font-size: 14px;
}

.image-preview {
  position: relative;
  width: 100%;
  border-radius: 8px;
  overflow: hidden;
}

.preview-img {
  width: 100%;
  height: 200px;
  object-fit: cover;
}

.image-actions {
  position: absolute;
  top: 8px;
  right: 8px;
  display: flex;
  gap: 8px;
}
</style>
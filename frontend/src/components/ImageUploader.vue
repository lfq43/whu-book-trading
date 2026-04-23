<template>
  <div class="image-uploader">
    <!-- 网格预览区域 -->
    <div class="thumbnails-grid">
      <!-- 已有的图片 -->
      <div
          v-for="(url, index) in imageUrls"
          :key="url"
          class="thumbnail-item"
      >
        <el-image
            :src="url"
            fit="cover"
            class="thumbnail-image"
            :preview-src-list="imageUrls"
            :initial-index="index"
        />
        <div class="thumbnail-mask">
          <el-icon
              class="delete-icon"
              @click.stop="removeImage(index)"
          >
            <CircleClose />
          </el-icon>
        </div>
      </div>

      <!-- 上传中占位 -->
      <div
          v-for="item in uploadingCount"
          :key="'uploading-' + item"
          class="thumbnail-item uploading"
      >
        <div class="uploading-placeholder">
          <el-icon class="is-loading" :size="24"><Loading /></el-icon>
          <span>上传中</span>
        </div>
      </div>

      <!-- 添加上传按钮 -->
      <div
          class="upload-trigger"
          @click="triggerUpload"
          v-if="imageUrls.length < maxCount"
      >
        <el-icon :size="24"><Plus /></el-icon>
        <span>{{ imageUrls.length === 0 ? '上传图片' : '继续添加' }}</span>
        <span class="count-tip">{{ imageUrls.length }}/{{ maxCount }}</span>
      </div>
    </div>

    <!-- 隐藏的文件输入 -->
    <input
        ref="fileInput"
        type="file"
        accept="image/jpeg,image/png,image/gif,image/webp"
        style="display: none"
        multiple
        @change="handleFileChange"
    />

    <!-- 提示信息 -->
    <div class="upload-tip">
      <span>支持 JPG、PNG、GIF、WEBP 格式，最多 {{ maxCount }} 张，单张不超过 5MB</span>
    </div>
  </div>
</template>

<script setup>
import { ref, watch, computed } from 'vue'
import { ElMessage } from 'element-plus'
import { Plus, CircleClose, Loading } from '@element-plus/icons-vue'
import { uploadImage } from '../api/upload'

const props = defineProps({
  modelValue: {
    type: Array,  // 改为数组类型
    default: () => []
  },
  maxCount: {
    type: Number,
    default: 9  // 默认最多9张
  }
})

const emit = defineEmits(['update:modelValue', 'upload-success'])

const fileInput = ref(null)
const uploadingList = ref([])  // 正在上传的文件列表
const imageUrls = ref([...props.modelValue])

// 计算正在上传的数量
const uploadingCount = computed(() => uploadingList.value.length)

// 压缩图片
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

// 验证单个文件
const validateFile = (file) => {
  const allowedTypes = ['image/jpeg', 'image/png', 'image/gif', 'image/webp']
  if (!allowedTypes.includes(file.type)) {
    ElMessage.warning(`${file.name} 格式不支持，已跳过`)
    return false
  }

  if (file.size > 5 * 1024 * 1024) {
    ElMessage.warning(`${file.name} 超过 5MB，已跳过`)
    return false
  }

  return true
}

// 触发文件选择
const triggerUpload = () => {
  fileInput.value?.click()
}

// 处理文件选择（支持多选）
const handleFileChange = async (event) => {
  const files = Array.from(event.target.files)
  if (files.length === 0) return

  // 检查数量限制
  const availableSlots = props.maxCount - imageUrls.value.length - uploadingList.value.length
  if (availableSlots <= 0) {
    ElMessage.warning(`最多只能上传 ${props.maxCount} 张图片`)
    fileInput.value.value = ''
    return
  }

  // 过滤出有效的文件
  const validFiles = files
      .filter(file => validateFile(file))
      .slice(0, availableSlots)

  if (validFiles.length === 0) {
    fileInput.value.value = ''
    return
  }

  if (validFiles.length < files.length) {
    ElMessage.info(`已选择 ${validFiles.length} 个有效文件`)
  }

  // 添加到上传列表
  const uploadIds = validFiles.map(() => Symbol('upload'))
  uploadingList.value.push(...uploadIds)

  // 逐个上传
  const uploadPromises = validFiles.map(async (file, index) => {
    const uploadId = uploadIds[index]
    try {
      const compressedFile = await compressImage(file)
      const response = await uploadImage(compressedFile)
      const url = response.data.url

      // 添加到图片列表
      imageUrls.value.push(url)

      return { success: true, url }
    } catch (error) {
      ElMessage.error(`${file.name} 上传失败: ${error.message}`)
      return { success: false, error }
    } finally {
      // 从上传列表中移除
      const idx = uploadingList.value.indexOf(uploadId)
      if (idx > -1) {
        uploadingList.value.splice(idx, 1)
      }
    }
  })

  // 等待所有上传完成
  const results = await Promise.all(uploadPromises)
  const successCount = results.filter(r => r.success).length

  if (successCount > 0) {
    // 更新 v-model
    emit('update:modelValue', [...imageUrls.value])
    emit('upload-success', [...imageUrls.value])
    ElMessage.success(`成功上传 ${successCount} 张图片`)
  }

  // 清空 input
  fileInput.value.value = ''
}

// 移除单张图片
const removeImage = (index) => {
  imageUrls.value.splice(index, 1)
  emit('update:modelValue', [...imageUrls.value])
  emit('upload-success', [...imageUrls.value])
}

// 批量移除所有图片（可选，供父组件调用）
const clearAllImages = () => {
  imageUrls.value = []
  emit('update:modelValue', [])
  emit('upload-success', [])
}

// 监听外部变化
watch(() => props.modelValue, (newVal) => {
  if (JSON.stringify(newVal) !== JSON.stringify(imageUrls.value)) {
    imageUrls.value = [...newVal]
  }
}, { deep: true })

// 暴露方法给父组件
defineExpose({
  clearAllImages
})
</script>

<style scoped>
.image-uploader {
  width: 100%;
}

.thumbnails-grid {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
}

.thumbnail-item {
  position: relative;
  width: 100px;
  height: 100px;
  border-radius: 8px;
  overflow: hidden;
  border: 1px solid #e4e7ed;
  background: #f5f7fa;
}

.thumbnail-image {
  width: 100%;
  height: 100%;
}

.thumbnail-image :deep(img) {
  object-fit: cover;
}

.thumbnail-mask {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background: rgba(0, 0, 0, 0.3);
  display: flex;
  align-items: center;
  justify-content: center;
  opacity: 0;
  transition: opacity 0.2s;
}

.thumbnail-item:hover .thumbnail-mask {
  opacity: 1;
}

.delete-icon {
  font-size: 20px;
  color: #fff;
  background: rgba(245, 108, 108, 0.8);
  border-radius: 50%;
  padding: 4px;
  cursor: pointer;
  transition: transform 0.2s;
}

.delete-icon:hover {
  transform: scale(1.1);
  background: #f56c6c;
}

.upload-trigger {
  width: 100px;
  height: 100px;
  border: 1px dashed #dcdfe6;
  border-radius: 8px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 4px;
  color: #909399;
  cursor: pointer;
  transition: all 0.2s;
  background: #fafafa;
}

.upload-trigger:hover {
  border-color: #409eff;
  color: #409eff;
  background: #ecf5ff;
}

.count-tip {
  font-size: 12px;
  color: #c0c4cc;
}

.thumbnail-item.uploading {
  border-style: dashed;
}

.uploading-placeholder {
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 4px;
  color: #909399;
  font-size: 12px;
  background: #f5f7fa;
}

.upload-tip {
  margin-top: 12px;
  font-size: 12px;
  color: #999;
}
</style>
<template>
  <div class="image-uploader">
    <div class="upload-area" @click="triggerUpload" v-if="!imageUrl">
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
import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import { Plus, Delete, Edit } from '@element-plus/icons-vue'
import { uploadImage } from '../api/upload'

const props = defineProps({
  modelValue: {
    type: String,
    default: ''
  }
})

const emit = defineEmits(['update:modelValue'])

const fileInput = ref(null)
const uploading = ref(false)
const imageUrl = ref(props.modelValue)

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
    const response = await uploadImage(file)
    imageUrl.value = response.data.url
    emit('update:modelValue', response.data.url)
    ElMessage.success('上传成功')
  } catch (error) {
    ElMessage.error(error.message || '上传失败')
  } finally {
    uploading.value = false
    // 清空 input，允许重新上传同一文件
    if (fileInput.value) {
      fileInput.value.value = ''
    }
  }
}

// 移除图片
const removeImage = () => {
  imageUrl.value = ''
  emit('update:modelValue', '')
}

// 监听外部变化
import { watch } from 'vue'
watch(() => props.modelValue, (newVal) => {
  imageUrl.value = newVal
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
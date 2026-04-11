<template>
  <div class="user-space-container" v-loading="loading">
    <!-- 个人资料卡片 -->
    <el-card class="profile-card">
      <div class="profile-header">
        <div class="avatar-section">
          <div class="avatar-wrapper" @click="isOwner ? triggerAvatarUpload() : null">
            <el-avatar :size="100" :src="userInfo?.avatar" class="user-avatar">
              {{ userInfo?.username?.charAt(0) }}
            </el-avatar>
            <div v-if="isOwner" class="avatar-overlay">
              <el-icon><Camera /></el-icon>
              <span>更换头像</span>
            </div>
          </div>
          <input
              ref="avatarInput"
              type="file"
              accept="image/jpeg,image/png,image/gif,image/webp"
              style="display: none"
              @change="handleAvatarChange"
          />
        </div>

        <div class="info-section">
          <div class="username-section">
            <h2 v-if="!isEditingName">{{ userInfo?.username }}</h2>
            <el-input
                v-else
                v-model="editUsername"
                size="large"
                style="width: 200px"
                @keyup.enter="saveUsername"
                ref="usernameInput"
            />
            <el-button
                v-if="isOwner && !isEditingName"
                link
                @click="startEditName"
            >
              <el-icon><Edit /></el-icon>
            </el-button>
            <div v-if="isEditingName" class="edit-actions">
              <el-button type="primary" size="small" @click="saveUsername">保存</el-button>
              <el-button size="small" @click="cancelEditName">取消</el-button>
            </div>
          </div>

          <div class="user-meta">
            <span class="meta-item">
              <el-icon><Postcard /></el-icon>
              {{ userInfo?.email || '未设置邮箱' }}
            </span>
            <span class="meta-item">
              <el-icon><Calendar /></el-icon>
              加入于 {{ formatDate(userInfo?.created_at) }}
            </span>
          </div>
        </div>
      </div>
    </el-card>

    <!-- 统计卡片 -->
    <div class="stats-cards">
      <el-card class="stat-card">
        <div class="stat-content">
          <div class="stat-number">{{ batches.length }}</div>
          <div class="stat-label">发布批次</div>
        </div>
      </el-card>
      <el-card class="stat-card">
        <div class="stat-content">
          <div class="stat-number">{{ totalBooks }}</div>
          <div class="stat-label">书籍总数</div>
        </div>
      </el-card>
      <el-card class="stat-card">
        <div class="stat-content">
          <div class="stat-number">{{ soldBooksCount }}</div>
          <div class="stat-label">已售出</div>
        </div>
      </el-card>
    </div>

    <!-- 发布的批次列表 -->
    <div class="batches-section">
      <h3>发布的批次</h3>

      <div v-if="batches.length === 0" class="empty-batches">
        <el-empty description="还没有发布过任何批次" />
      </div>

      <div v-else class="batches-list">
        <el-card
            v-for="batch in batches"
            :key="batch.id"
            class="batch-card"
            @click="goToBatch(batch.id)"
        >
          <div class="batch-content">
            <div class="batch-image">
              <ImageViewer
                  :src="batch.image || '/placeholder.png'"
                  :title="batch.title"
                  width="80px"
                  height="80px"
              />
            </div>
            <div class="batch-info">
              <div class="batch-title">{{ batch.title }}</div>
              <div class="batch-books">
                <span v-for="(name, idx) in getBookNames(batch)" :key="idx" class="book-name">
                  {{ name }}
                  <span v-if="isBookSold(batch, idx)" class="sold-badge">✓</span>
                </span>
              </div>
              <div class="batch-status">
                <el-tag size="small" :type="getStatusType(batch.status)">
                  {{ getStatusText(batch.status) }}
                </el-tag>
                <span class="batch-time">{{ formatDate(batch.created_at) }}</span>
              </div>
            </div>
          </div>
        </el-card>
      </div>
    </div>

    <!-- 头像上传弹窗 -->
    <el-dialog v-model="avatarDialogVisible" title="裁剪头像" width="620px" @close="closeAvatarDialog">
      <div class="avatar-cropper">
                <VueCropper
            :key="cropperKey"
            ref="cropperRef"
            style="width: 100%; height: 100%;"
            :src="avatarTempUrl"
            :options="cropperOptions"
        />
      </div>
      <template #footer>
        <el-button @click="closeAvatarDialog">取消</el-button>
        <el-button type="primary" @click="confirmAvatar">确认上传</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Camera, Edit, Postcard, Calendar } from '@element-plus/icons-vue'
import { getUserProfile, updateProfile } from '../api/user'
import { uploadAvatar } from '../api/upload'
import { useUserStore } from '../stores/user'
import ImageViewer from '../components/ImageViewer.vue'
import { VueCropper } from 'vue-cropper-next'
import 'cropperjs/dist/cropper.css'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()

const loading = ref(false)
const userInfo = ref(null)
const batches = ref([])
const isOwner = ref(false)

// 编辑用户名
const isEditingName = ref(false)
const editUsername = ref('')
const usernameInput = ref(null)

// 头像上传
const avatarInput = ref(null)
const avatarDialogVisible = ref(false)
const avatarTempUrl = ref('')
const cropperRef = ref(null)
const cropperKey = ref(0)
const cropperOptions = {
  viewMode: 2,
  autoCropArea: 0.85,
  dragMode: 'move',
  aspectRatio: 1,
  preview: '',
  responsive: true,
  modal: true,
  guides: true,
  center: true,
  background: false,
  zoomable: true,
  zoomOnWheel: true,
  cropBoxMovable: true,
  cropBoxResizable: true,
  toggleDragModeOnDblclick: false,
  minCropBoxWidth: 120,
  minCropBoxHeight: 120,
}

// 计算总书籍数
const totalBooks = computed(() => {
  return batches.value.reduce((sum, batch) => {
    const bookNames = getBookNames(batch)
    return sum + bookNames.length
  }, 0)
})

// 计算已售出书籍数
const soldBooksCount = computed(() => {
  return batches.value.reduce((sum, batch) => {
    const soldStatus = getSoldStatus(batch)
    const soldCount = soldStatus.filter(s => s === true).length
    return sum + soldCount
  }, 0)
})

// 解析书名列表
const getBookNames = (batch) => {
  if (!batch.book_names) return []
  try {
    return typeof batch.book_names === 'string'
        ? JSON.parse(batch.book_names)
        : batch.book_names
  } catch {
    return []
  }
}

// 解析售出状态
const getSoldStatus = (batch) => {
  if (!batch.sold_status) return []
  try {
    const status = typeof batch.sold_status === 'string'
        ? JSON.parse(batch.sold_status)
        : batch.sold_status
    return status || []
  } catch {
    return []
  }
}

// 判断书是否已售出
const isBookSold = (batch, index) => {
  const soldStatus = getSoldStatus(batch)
  return soldStatus[index] === true
}

// 状态文本
const getStatusText = (status) => {
  const map = {
    'available': '在售',
    'partial': '部分售出',
    'sold': '已售完'
  }
  return map[status] || '在售'
}

// 状态标签类型
const getStatusType = (status) => {
  const map = {
    'available': 'success',
    'partial': 'warning',
    'sold': 'info'
  }
  return map[status] || 'success'
}

// 格式化日期
const formatDate = (dateStr) => {
  if (!dateStr) return ''
  const date = new Date(dateStr)
  return `${date.getFullYear()}-${date.getMonth() + 1}-${date.getDate()}`
}

// 加载个人空间数据
const loadUserSpace = async () => {
  const userId = route.params.id
  // 判断是否是自己的空间
  isOwner.value = Number(userStore.userInfo?.id) === Number(userId)
  console.log(isOwner.value)

  loading.value = true
  try {
    const response = await getUserProfile(userId)
    userInfo.value = response.data.user
    batches.value = response.data.batches || []
  } catch (error) {
    ElMessage.error('加载失败')
    if (error.response?.status === 404) {
      router.push('/')
    }
  } finally {
    loading.value = false
  }
}

// 开始编辑用户名
const startEditName = () => {
  editUsername.value = userInfo.value.username
  isEditingName.value = true
  setTimeout(() => {
    usernameInput.value?.focus()
  }, 100)
}

// 保存用户名
const saveUsername = async () => {
  if (!editUsername.value.trim()) {
    ElMessage.warning('用户名不能为空')
    return
  }
  if (editUsername.value.length < 2) {
    ElMessage.warning('用户名至少2个字符')
    return
  }

  try {
    await updateProfile({ username: editUsername.value.trim() })
    userInfo.value.username = editUsername.value.trim()
    userStore.setUserInfo(userInfo.value)
    ElMessage.success('用户名更新成功')
    isEditingName.value = false
  } catch (error) {
    ElMessage.error(error.response?.data?.message || '更新失败')
  }
}

// 取消编辑用户名
const cancelEditName = () => {
  isEditingName.value = false
  editUsername.value = ''
}

// 触发头像上传
const triggerAvatarUpload = () => {
  avatarInput.value?.click()
}

// 处理头像选择
const handleAvatarChange = async (event) => {
  const file = event.target.files[0]
  if (!file) return

  // 验证文件类型
  const allowedTypes = ['image/jpeg', 'image/png', 'image/gif', 'image/webp']
  if (!allowedTypes.includes(file.type)) {
    ElMessage.error('只支持 JPG、PNG、GIF、WEBP 格式的图片')
    return
  }

  // 验证文件大小（2MB）
  if (file.size > 2 * 1024 * 1024) {
    ElMessage.error('头像大小不能超过 2MB')
    return
  }

  // 清理旧临时 URL，避免旧图片缓存
  if (avatarTempUrl.value) {
    URL.revokeObjectURL(avatarTempUrl.value)
    avatarTempUrl.value = ''
  }

  // 创建临时 URL 用于裁剪
  avatarTempUrl.value = URL.createObjectURL(file)
  cropperKey.value += 1
  avatarDialogVisible.value = true

  // 清空 input
  avatarInput.value.value = ''
}

const closeAvatarDialog = () => {
  avatarDialogVisible.value = false
  if (avatarTempUrl.value) {
    URL.revokeObjectURL(avatarTempUrl.value)
    avatarTempUrl.value = ''
  }
  cropperKey.value += 1
}

// 确认上传头像
const confirmAvatar = async () => {
  if (!cropperRef.value) return

  const cropperInstance = cropperRef.value.cropper
  if (!cropperInstance || typeof cropperInstance.getCroppedCanvas !== 'function') {
    ElMessage.error('裁剪组件未就绪，请重试')
    return
  }

  // 获取裁剪后的图片数据
  const canvas = cropperInstance.getCroppedCanvas({ width: 400, height: 400 })
  if (!canvas) {
    ElMessage.error('裁剪失败，请重试')
    return
  }

  // 将 canvas 转换为 Blob
  canvas.toBlob(async (blob) => {
    if (!blob) {
      ElMessage.error('头像生成失败，请重试')
      return
    }

    const file = new File([blob], 'avatar.png', { type: 'image/png' })

    try {
      const response = await uploadAvatar(file)
      const avatarUrl = response.data.url

      // 更新用户信息
      await updateProfile({ avatar: avatarUrl })
      userInfo.value.avatar = avatarUrl
      userStore.setUserInfo(userInfo.value)

      ElMessage.success('头像更新成功')
      avatarDialogVisible.value = false
      URL.revokeObjectURL(avatarTempUrl.value)
    } catch (error) {
      ElMessage.error(error.response?.data?.message || '上传失败')
    }
  }, 'image/png', 0.9)
}

// 跳转到批次详情
const goToBatch = (id) => {
  router.push(`/batch/${id}`)
}

onMounted(() => {
  loadUserSpace()
})
</script>

<style scoped>
.user-space-container {
  max-width: 900px;
  margin: 0 auto;
  padding: 20px;
}

.profile-card {
  margin-bottom: 24px;
}

.profile-header {
  display: flex;
  gap: 30px;
  align-items: center;
}

.avatar-section {
  position: relative;
}

.avatar-wrapper {
  position: relative;
  cursor: pointer;
}

.user-avatar {
  width: 100px;
  height: 100px;
  font-size: 40px;
  background: #409eff;
  color: #fff;
}

.avatar-overlay {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background: rgba(0, 0, 0, 0.6);
  border-radius: 50%;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 4px;
  opacity: 0;
  transition: opacity 0.2s;
  color: #fff;
  font-size: 12px;
}

.avatar-wrapper:hover .avatar-overlay {
  opacity: 1;
}

.avatar-overlay .el-icon {
  font-size: 20px;
}

.info-section {
  flex: 1;
}

.username-section {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 12px;
  flex-wrap: wrap;
}

.username-section h2 {
  margin: 0;
  font-size: 24px;
}

.edit-actions {
  display: flex;
  gap: 8px;
}

.user-meta {
  display: flex;
  gap: 24px;
  color: #666;
  font-size: 14px;
}

.meta-item {
  display: flex;
  align-items: center;
  gap: 6px;
}

/* 统计卡片 */
.stats-cards {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 20px;
  margin-bottom: 30px;
}

.stat-card {
  text-align: center;
}

.stat-content {
  padding: 10px;
}

.stat-number {
  font-size: 32px;
  font-weight: bold;
  color: #409eff;
}

.stat-label {
  font-size: 14px;
  color: #666;
  margin-top: 8px;
}

/* 批次列表 */
.batches-section h3 {
  margin-bottom: 20px;
  font-size: 18px;
}

.batches-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.batch-card {
  cursor: pointer;
  transition: all 0.2s;
}

.batch-card:hover {
  transform: translateX(4px);
}

.batch-content {
  display: flex;
  gap: 16px;
}

.batch-image {
  width: 80px;
  flex-shrink: 0;
}

.batch-info {
  flex: 1;
}

.batch-title {
  font-size: 16px;
  font-weight: 500;
  margin-bottom: 8px;
}

.batch-books {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-bottom: 8px;
}

.book-name {
  background: #f0f2f5;
  padding: 2px 8px;
  border-radius: 12px;
  font-size: 12px;
}

.book-name .sold-badge {
  color: #67c23a;
  margin-left: 4px;
}

.batch-status {
  display: flex;
  align-items: center;
  gap: 12px;
  font-size: 12px;
  color: #999;
}

.empty-batches {
  padding: 40px;
}

/* 头像裁剪弹窗 */
.avatar-cropper {
  height: 480px;
}

:deep(.vue-cropper) {
  width: 100%;
  height: 100%;
}

/* 响应式 */
@media (max-width: 768px) {
  .profile-header {
    flex-direction: column;
    text-align: center;
  }

  .user-meta {
    justify-content: center;
  }

  .stats-cards {
    gap: 12px;
  }

  .batch-content {
    flex-direction: column;
  }

  .batch-image {
    width: 100%;
    height: 120px;
  }
}
</style>
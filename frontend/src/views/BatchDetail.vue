<template>
  <div class="batch-detail-container" v-loading="loading">
    <div v-if="batch" class="detail-card">
      <!-- 头部信息 -->
      <div class="detail-header">
        <h1>{{ batch.title }}</h1>
        <div class="meta">
          <el-tag :type="getStatusType(batch.status)">
            {{ getStatusText(batch.status) }}
          </el-tag>
          <div class="seller-info">
            <el-avatar :size="32" :src="batch.user?.avatar">
              {{ batch.user?.username?.charAt(0) }}
            </el-avatar>
            <span class="seller-name" @click="goToUserSpace">{{ batch.user?.username }}</span>
            <el-button
                v-if="batch.user_id !== currentUserId"
                type="primary"
                size="small"
                circle
                @click="openChat"
            >
              <el-icon><ChatDotRound /></el-icon>
            </el-button>
          </div>
          <span> {{ formatDate(batch.created_at) }}</span>
        </div>
      </div>

      <!-- 图片区域 - 使用图片预览组件 -->
      <div v-if="batch.image" class="detail-image">
        <ImageViewer
            :src="batch.image"
            :title="batch.title"
            width="30%"
            height="30%"
        />
      </div>

      <!-- 未售出书籍列表 -->
      <div class="book-list-section" v-if="bookNames.length > 0">
        <h3>📗 在售书籍（{{ bookNames.length }}本）</h3>
        <div class="book-list">
          <div
              v-for="(name, index) in bookNames"
              :key="index"
              class="book-item"
          >
            <div class="book-name">
              <span class="book-index">{{ index + 1 }}.</span>
              <span>{{ name }}</span>
            </div>
            <!-- 只有发布者本人可以看到勾选框 -->
            <el-checkbox
                v-if="isOwner"
                :model-value="false"
                @change="toggleSoldStatus(name, true)"
            >
              标记售出
            </el-checkbox>
          </div>
        </div>
      </div>

      <!-- 已售出书籍列表 -->
      <div class="book-list-section" v-if="soldBookNames.length > 0">
        <h3>✅ 已售出（{{ soldBookNames.length }}本）</h3>
        <div class="book-list sold-list">
          <div
              v-for="(name, index) in soldBookNames"
              :key="index"
              class="book-item sold-item"
          >
            <div class="book-name">
              <span class="book-index">{{ index + 1 }}.</span>
              <span class="strikethrough">{{ name }}</span>
            </div>
            <!-- 只有发布者本人可以看到取消按钮 -->
            <el-button
                v-if="isOwner"
                type="warning"
                size="small"
                plain
                @click="toggleSoldStatus(name, false)"
            >
              取消售出
            </el-button>
          </div>
        </div>
      </div>

      <!-- 空状态 -->
      <div v-if="bookNames.length === 0 && soldBookNames.length === 0" class="empty-books">
        <el-empty description="暂无书籍" />
      </div>

      <!-- 描述 -->
      <div v-if="batch.description" class="description-section">
        <h3>📝 描述</h3>
        <p class="description">{{ batch.description }}</p>
      </div>

      <!-- 联系方式 -->
      <div class="contact-section">
        <h3>📠 联系方式</h3>
        <div class="contact-box">
          <span class="contact-value">{{ batch.contact }}</span>
          <el-button type="primary" size="small" @click="copyContact">
            复制
          </el-button>
        </div>
      </div>

      <!-- 删除按钮 -->
      <div v-if="isOwner || isAdmin" class="actions">
        <el-button type="danger" @click="handleDelete">删除这个发布</el-button>
      </div>
    </div>

    <!-- 聊天窗口 -->
    <ChatWindow
        v-model="chatVisible"
        :other-user="batch?.user"
    />
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { ChatDotRound } from '@element-plus/icons-vue'
import { getBatchDetail, updateBookSoldStatus, deleteBatch } from '../api/batch'
import { useUserStore } from '../stores/user'
import ChatWindow from '../components/ChatWindow.vue'
import ImageViewer from "../components/ImageViewer.vue";

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()

const loading = ref(false)
const batch = ref(null)
const chatVisible = ref(false)

const currentUserId = computed(() => userStore.userInfo?.id)
const isOwner = computed(() => batch.value?.user_id === currentUserId.value)
const isAdmin = computed(() => userStore.isAdmin)

// 解析未售出书籍列表
const bookNames = computed(() => {
  if (!batch.value?.book_names) return []
  try {
    const names = typeof batch.value.book_names === 'string'
        ? JSON.parse(batch.value.book_names)
        : batch.value.book_names
    return names || []
  } catch {
    return []
  }
})

const goToUserSpace = () => {
  router.push(`/user/${batch.value.user_id}`)
}

// 解析已售出书籍列表
const soldBookNames = computed(() => {
  if (!batch.value?.sold_book_names) return []
  try {
    const names = typeof batch.value.sold_book_names === 'string'
        ? JSON.parse(batch.value.sold_book_names)
        : batch.value.sold_book_names
    return names || []
  } catch {
    return []
  }
})

const getStatusText = (status) => {
  const map = {
    'available': '在售',
    'partial': '部分售出',
    'sold': '已售完'
  }
  return map[status] || '在售'
}

const getStatusType = (status) => {
  const map = {
    'available': 'success',
    'partial': 'warning',
    'sold': 'info'
  }
  return map[status] || 'success'
}

const formatDate = (dateStr) => {
  if (!dateStr) return ''
  const date = new Date(dateStr)
  return `${date.getFullYear()}-${date.getMonth() + 1}-${date.getDate()}`
}

// 加载详情
const loadDetail = async () => {
  const id = route.params.id
  loading.value = true
  try {
    const response = await getBatchDetail(id)
    batch.value = response.data
  } catch (error) {
    ElMessage.error('加载失败')
  } finally {
    loading.value = false
  }
}

// 切换售出状态
const toggleSoldStatus = async (bookName, sold) => {
  try {
    await updateBookSoldStatus(batch.value.id, bookName, sold)
    ElMessage.success(sold ? '已标记为售出' : '已取消售出标记')
    await loadDetail() // 刷新页面
  } catch (error) {
    ElMessage.error('操作失败')
  }
}

// 复制联系方式
const copyContact = () => {
  if (!batch.value?.contact) return
  navigator.clipboard.writeText(batch.value.contact)
  ElMessage.success('联系方式已复制')
}

// 删除批次
const handleDelete = async () => {
  try {
    await ElMessageBox.confirm('确定要删除这个发布吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    await deleteBatch(batch.value.id)
    ElMessage.success('删除成功')
    router.push('/my-batches')
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败')
    }
  }
}

// 打开聊天
const openChat = () => {
  if (!userStore.isLoggedIn) {
    ElMessage.warning('请先登录')
    router.push('/login')
    return
  }
  chatVisible.value = true
}

onMounted(() => {
  loadDetail()
})
</script>

<style scoped>
.batch-detail-container {
  max-width: 800px;
  margin: 0 auto;
  padding: 20px;
}

.detail-card {
  background: #fff;
  border-radius: 12px;
  padding: 24px;
  box-shadow: 0 2px 12px rgba(0,0,0,0.1);
}

.detail-header {
  border-bottom: 1px solid #eee;
  padding-bottom: 16px;
  margin-bottom: 20px;
}

.detail-header h1 {
  margin: 0 0 12px 0;
  font-size: 24px;
}

.meta {
  display: flex;
  align-items: center;
  gap: 16px;
  color: #999;
  font-size: 14px;
}

.detail-image {
  text-align: center;
  margin-bottom: 24px;
}

.book-list-section h3,
.description-section h3,
.contact-section h3 {
  margin: 0 0 16px 0;
  font-size: 18px;
}

.book-list {
  border: 1px solid #eee;
  border-radius: 8px;
  overflow: hidden;
}

.book-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  border-bottom: 1px solid #f0f0f0;
}

.book-item:last-child {
  border-bottom: none;
}

.book-item.sold {
  background: #fafafa;
}

.book-name {
  font-size: 16px;
}

.book-index {
  color: #999;
  margin-right: 8px;
}

.seller-name {
  cursor: pointer;
  color: #409eff;
}

.seller-name:hover {
  text-decoration: underline;
}

.strikethrough {
  text-decoration: line-through;
  color: #999;
}

.description-section {
  margin-top: 24px;
}

.description {
  background: #f9f9f9;
  padding: 16px;
  border-radius: 8px;
  line-height: 1.6;
  white-space: pre-wrap;
}

.contact-section {
  margin-top: 24px;
}

.contact-box {
  display: flex;
  align-items: center;
  gap: 12px;
  background: #e6f7ff;
  padding: 12px 16px;
  border-radius: 8px;
}

.contact-value {
  font-size: 18px;
  font-weight: bold;
  color: #1890ff;
  flex: 1;
}

.contact-tip {
  font-size: 12px;
  color: #999;
  margin-top: 8px;
}

.actions {
  margin-top: 30px;
  text-align: center;
  padding-top: 20px;
  border-top: 1px solid #eee;
}

.seller-info {
  display: flex;
  align-items: center;
  gap: 12px;  /* 统一设置所有子元素之间的间距 */
}

.sold-list {
  background: #fafafa;
}

.sold-item {
  opacity: 0.7;
}

.strikethrough {
  text-decoration: line-through;
  color: #999;
}

.empty-books {
  padding: 40px 0;
}
</style>

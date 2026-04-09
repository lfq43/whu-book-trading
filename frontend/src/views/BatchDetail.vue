<template>
  <div class="batch-detail-container" v-loading="loading">
    <div v-if="batch" class="detail-card">
      <!-- 头部信息 -->
      <div class="detail-header">
        <h1>{{ batch.title }}</h1>
        <div class="meta">
          <el-tag :type="batch.status === 'available' ? 'success' : 'info'">
            {{ batch.status === 'available' ? '在售' : '已全部售出' }}
          </el-tag>
          <div class="seller-info">
            <el-avatar :size="24" :src="batch.user?.avatar">
              {{ batch.user?.username?.charAt(0) }}
            </el-avatar>
            <span class="seller-name">{{ batch.user?.username }}</span>
            <!-- 聊天按钮：不能给自己发消息 -->
            <el-button
                v-if="batch.user_id !== currentUserId"
                type="primary"
                size="small"
                circle
                :style="{ width: '24px', height: '24px' }"
                @click="openChat"
            >
              <el-icon :size="20"><ChatDotRound /></el-icon>
            </el-button>
          </div>
          <span>📅 {{ formatDate(batch.created_at) }}</span>
        </div>
      </div>

      <!-- 图片 -->
      <div v-if="batch.image" class="detail-image">
        <el-image :src="batch.image" fit="contain" style="max-height: 300px" />
      </div>

      <!-- 书籍列表（带勾选框，仅自己可见编辑功能） -->
      <div class="book-list-section">
        <h3>📚 书籍列表（{{ bookNames.length }}本）</h3>
        <div class="book-list">
          <div
              v-for="(name, index) in bookNames"
              :key="index"
              class="book-item"
              :class="{ sold: soldStatus[index] }"
          >
            <div class="book-name">
              <span class="book-index">{{ index + 1 }}.</span>
              <span :class="{ 'strikethrough': soldStatus[index] }">{{ name }}</span>
            </div>
            <!-- 只有发布者本人可以看到勾选框 -->
            <el-checkbox
                v-if="isOwner"
                v-model="soldStatus[index]"
                @change="toggleSoldStatus(index, $event)"
            >
              {{ soldStatus[index] ? '已售出' : '标记售出' }}
            </el-checkbox>
            <el-tag v-else-if="soldStatus[index]" type="success" size="small">已售出</el-tag>
          </div>
        </div>
      </div>

      <!-- 描述 -->
      <div v-if="batch.description" class="description-section">
        <h3>📝 描述</h3>
        <p class="description">{{ batch.description }}</p>
      </div>

      <!-- 联系方式 -->
      <div class="contact-section">
        <h3>📞 联系方式</h3>
        <div class="contact-box">
          <span class="contact-value">{{ batch.contact }}</span>
          <el-button
              type="primary"
              size="small"
              @click="copyContact"
          >
            复制
          </el-button>
        </div>
        <p class="contact-tip">点击复制后，打开微信/QQ 添加好友</p>
      </div>

      <!-- 删除按钮（仅自己可见） -->
      <div v-if="isOwner" class="actions">
        <el-button type="danger" @click="handleDelete">删除这个发布</el-button>
      </div>
    </div>

    <el-empty v-else-if="!loading" description="批次不存在" />

    <!-- 聊天窗口 -->
    <ChatWindow
        v-model="chatVisible"
        :other-user="batch?.user"
        @message-sent="onMessageSent"
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

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()

const currentUserId = computed(() => userStore.userInfo?.id)
const chatVisible = ref(false)

const loading = ref(false)
const batch = ref(null)

// 解析 JSON 数组
const bookNames = computed(() => {
  if (!batch.value?.book_names) return []
  try {
    return typeof batch.value.book_names === 'string'
        ? JSON.parse(batch.value.book_names)
        : batch.value.book_names
  } catch {
    return []
  }
})

const soldStatus = ref([])

// 是否当前用户发布的
const isOwner = computed(() => {
  return userStore.userInfo?.id === batch.value?.user_id
})

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

    // 解析售出状态
    let status = batch.value.sold_status
    if (typeof status === 'string') {
      status = JSON.parse(status)
    }
    soldStatus.value = status || new Array(bookNames.value.length).fill(false)
  } catch (error) {
    ElMessage.error('加载失败')
  } finally {
    loading.value = false
  }
}

// 切换售出状态
const toggleSoldStatus = async (index, sold) => {
  try {
    await updateBookSoldStatus(batch.value.id, index, sold)
    ElMessage.success(sold ? '已标记为售出' : '已取消售出标记')
    // 刷新整体状态
    await loadDetail()
  } catch (error) {
    ElMessage.error('操作失败')
    // 恢复勾选状态
    soldStatus.value[index] = !sold
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

onMounted(() => {
  loadDetail()
})

// 打开聊天窗口
const openChat = () => {
  if (!userStore.isLoggedIn) {
    ElMessage.warning('请先登录')
    router.push('/login')
    return
  }
  chatVisible.value = true
}

const onMessageSent = () => {
  // 消息发送后可以刷新未读数量等
}
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
</style>

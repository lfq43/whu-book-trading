<template>
  <el-drawer
      v-model="visible"
      :title="`与 ${otherUser?.username} 聊天`"
      direction="rtl"
      size="400px"
      @close="handleClose"
  >
    <div class="chat-container">
      <!-- 消息列表 -->
      <div
          class="message-list"
          ref="messageListRef"
          @scroll="handleScroll"
      >
        <!-- 加载更多指示器 -->
        <div v-if="loadingMore" class="loading-more">
          <el-icon class="is-loading"><Loading /></el-icon>
          加载历史消息...
        </div>

        <div v-if="!hasMore && messages.length > 0" class="no-more">
          没有更多消息了
        </div>

        <!-- 消息列表 -->
        <div
            v-for="msg in messages"
            :key="msg.id"
            class="message-item"
            :class="{ 'message-self': msg.from_user_id === currentUserId }"
        >
          <div class="message-bubble">
            <div class="message-content">{{ msg.content }}</div>
            <div class="message-time">{{ formatTime(msg.created_at) }}</div>
          </div>
        </div>

        <!-- 新消息分隔线（可选） -->
        <div v-if="newMessageCount > 0" class="new-message-divider">
          <span>{{ newMessageCount }} 条新消息</span>
        </div>
      </div>

      <!-- 输入框 -->
      <div class="input-area">
        <el-input
            v-model="inputContent"
            type="textarea"
            :rows="3"
            placeholder="输入消息... (Ctrl+Enter 发送)"
            @keyup.ctrl.enter="send"
        />
        <div class="input-actions">
          <el-button type="primary" @click="send" :disabled="!inputContent.trim()">
            发送
          </el-button>
        </div>
      </div>
    </div>
  </el-drawer>
</template>

<script setup>
import { ref, watch, nextTick, onUnmounted, computed } from 'vue'
import { ElMessage } from 'element-plus'
import { Loading } from '@element-plus/icons-vue'
import { sendMessage, getConversation } from '../api/message'
import { useUserStore } from '../stores/user'

const props = defineProps({
  modelValue: Boolean,
  otherUser: {
    type: Object,
    default: null
  }
})

const emit = defineEmits(['update:modelValue', 'message-sent'])

const userStore = useUserStore()
const currentUserId = computed(() => userStore.userInfo?.id)

// 状态
const visible = ref(false)
const messages = ref([])
const inputContent = ref('')
const loadingMore = ref(false)
const hasMore = ref(true)
const nextBeforeId = ref(0)
const messageListRef = ref(null)
let pollingTimer = null
let isFirstLoad = true

// 监听外部控制
watch(() => props.modelValue, async (val) => {
  visible.value = val
  if (val && props.otherUser) {
    await loadMessages(true) // 重置并加载最新消息
    startPolling()
  } else {
    stopPolling()
  }
})

watch(visible, (val) => {
  emit('update:modelValue', val)
  if (!val) {
    stopPolling()
  }
})

// 加载消息
// reset: 是否重置（清空现有消息，重新加载）
const loadMessages = async (reset = false) => {
  if (!props.otherUser) return

  if (reset) {
    messages.value = []
    hasMore.value = true
    nextBeforeId.value = 0
    isFirstLoad = true
  }

  // 如果是加载更多，且没有更多了，直接返回
  if (!reset && !hasMore.value) return

  // 如果是加载更多，设置加载状态
  if (!reset) {
    loadingMore.value = true
  }

  try {
    const beforeId = reset ? 0 : nextBeforeId.value
    const response = await getConversation(props.otherUser.id, beforeId, 20)

    const newMessages = response.data.messages || []
    hasMore.value = response.data.has_more
    nextBeforeId.value = response.data.next_before_id

    if (reset) {
      messages.value = newMessages
      // 滚动到底部
      await nextTick()
      scrollToBottom()
    } else {
      // 加载更多历史消息，添加到列表顶部
      // 保存滚动高度，用于保持位置
      const oldScrollHeight = messageListRef.value?.scrollHeight || 0
      const oldScrollTop = messageListRef.value?.scrollTop || 0

      messages.value = [...newMessages, ...messages.value]

      await nextTick()
      // 恢复滚动位置（新的内容在顶部，需要调整）
      if (messageListRef.value && newMessages.length > 0) {
        const newScrollHeight = messageListRef.value.scrollHeight
        messageListRef.value.scrollTop = newScrollHeight - oldScrollHeight + oldScrollTop
      }
    }
  } catch (error) {
    console.error('加载消息失败:', error)
    ElMessage.error('加载消息失败')
  } finally {
    loadingMore.value = false
    isFirstLoad = false
  }
}

// 发送消息
const send = async () => {
  if (!inputContent.value.trim()) return
  if (!props.otherUser) return

  try {
    await sendMessage(props.otherUser.id, inputContent.value)
    inputContent.value = ''
    // 重新加载最新消息
    await loadMessages(true)
    emit('message-sent')
  } catch (error) {
    ElMessage.error('发送失败: ' + (error.message || '未知错误'))
  }
}

// 滚动到底部
const scrollToBottom = () => {
  if (messageListRef.value) {
    messageListRef.value.scrollTop = messageListRef.value.scrollHeight
  }
}

// 处理滚动事件（滑动加载）
const handleScroll = () => {
  if (!messageListRef.value) return
  if (loadingMore.value) return
  if (!hasMore.value) return

  // 滚动到顶部时加载更多
  if (messageListRef.value.scrollTop <= 50) {
    loadMessages(false)
  }
}

// 格式化时间
const formatTime = (timeStr) => {
  if (!timeStr) return ''
  const date = new Date(timeStr)
  const now = new Date()
  const diff = now - date

  // 今天内显示时间
  if (diff < 24 * 60 * 60 * 1000 && date.getDate() === now.getDate()) {
    return `${date.getHours().toString().padStart(2, '0')}:${date.getMinutes().toString().padStart(2, '0')}`
  }
  // 昨天
  const yesterday = new Date(now)
  yesterday.setDate(now.getDate() - 1)
  if (date.getDate() === yesterday.getDate()) {
    return `昨天 ${date.getHours().toString().padStart(2, '0')}:${date.getMinutes().toString().padStart(2, '0')}`
  }
  // 更早
  return `${date.getMonth() + 1}/${date.getDate()} ${date.getHours().toString().padStart(2, '0')}:${date.getMinutes().toString().padStart(2, '0')}`
}

// 开始轮询（每3秒检查新消息）
const startPolling = () => {
  if (pollingTimer) clearInterval(pollingTimer)
  pollingTimer = setInterval(() => {
    if (visible.value && props.otherUser) {
      checkNewMessages()
    }
  }, 3000)
}

// 检查新消息（只获取最新的消息）
const checkNewMessages = async () => {
  if (!props.otherUser) return

  try {
    // 获取最新消息（不传 before_id，只取最新的1条判断是否有新消息）
    const response = await getConversation(props.otherUser.id, 0, 1)
    const newMessages = response.data.messages || []

    if (newMessages.length > 0) {
      const latestMessage = newMessages[0]
      // 如果最新消息不是当前用户发送的，且不在现有消息列表中
      const existingIds = messages.value.map(m => m.id)
      if (!existingIds.includes(latestMessage.id) && latestMessage.from_user_id !== currentUserId.value) {
        // 有新消息，重新加载
        await loadMessages(true)
      }
    }
  } catch (error) {
    console.error('检查新消息失败:', error)
  }
}

// 停止轮询
const stopPolling = () => {
  if (pollingTimer) {
    clearInterval(pollingTimer)
    pollingTimer = null
  }
}

const handleClose = () => {
  stopPolling()
}

onUnmounted(() => {
  stopPolling()
})
</script>

<style scoped>
.chat-container {
  display: flex;
  flex-direction: column;
  height: 100%;
}

.message-list {
  flex: 1;
  overflow-y: auto;
  padding: 16px;
  background: #f5f5f5;
  display: flex;
  flex-direction: column;
}

.message-item {
  display: flex;
  margin-bottom: 16px;
}

.message-self {
  justify-content: flex-end;
}

.message-bubble {
  max-width: 70%;
  padding: 8px 12px;
  border-radius: 12px;
  background: #fff;
  box-shadow: 0 1px 2px rgba(0,0,0,0.1);
}

.message-self .message-bubble {
  background: #409eff;
  color: #fff;
}

.message-content {
  word-wrap: break-word;
  font-size: 14px;
}

.message-time {
  font-size: 10px;
  color: #999;
  margin-top: 4px;
  text-align: right;
}

.message-self .message-time {
  color: rgba(255,255,255,0.7);
}

.input-area {
  padding: 16px;
  border-top: 1px solid #eee;
  background: #fff;
}

.input-actions {
  margin-top: 8px;
  text-align: right;
}

.loading-more {
  text-align: center;
  padding: 10px;
  color: #999;
  font-size: 12px;
}

.no-more {
  text-align: center;
  padding: 10px;
  color: #ccc;
  font-size: 12px;
}

.new-message-divider {
  text-align: center;
  margin: 10px 0;
  position: relative;
}

.new-message-divider span {
  background: #409eff;
  color: #fff;
  font-size: 12px;
  padding: 2px 10px;
  border-radius: 10px;
}
</style>
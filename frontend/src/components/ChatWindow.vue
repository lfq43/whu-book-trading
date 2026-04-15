<template>
  <el-drawer
      v-model="visible"
      :title="titleText"
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
import { getConversation } from '../api/message'
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
const titleText = computed(() => {
  return props.otherUser?.id === 4 ? `向管理员反馈` : `与 ${ props.otherUser?.username } 聊天`
})
// 状态
const visible = ref(false)
const messages = ref([])
const inputContent = ref('')
const loadingMore = ref(false)
const hasMore = ref(true)
const nextBeforeId = ref(0)
const messageListRef = ref(null)
const socket = ref(null)
const wsConnected = ref(false)
let isFirstLoad = true

// 监听外部控制
watch(
  () => props.modelValue,
  async (val) => {
    visible.value = val
    if (val && props.otherUser) {
      await loadMessages(true)
      initWebsocket()
    } else {
      closeWebsocket()
    }
  }
)

watch(
  () => props.otherUser,
  async (val) => {
    if (visible.value && val) {
      await loadMessages(true)
      if (!socket.value || socket.value.readyState !== WebSocket.OPEN) {
        initWebsocket()
      }
    }
  }
)

watch(visible, (val) => {
  emit('update:modelValue', val)
  if (!val) {
    closeWebsocket()
  }
})

const loadMessages = async (reset = false) => {
  if (!props.otherUser) return

  if (reset) {
    messages.value = []
    hasMore.value = true
    nextBeforeId.value = 0
    isFirstLoad = true
  }

  if (!reset && !hasMore.value) return
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
      await nextTick()
      scrollToBottom()
    } else {
      const oldScrollHeight = messageListRef.value?.scrollHeight || 0
      const oldScrollTop = messageListRef.value?.scrollTop || 0
      messages.value = [...newMessages, ...messages.value]
      await nextTick()
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

const getWebsocketUrl = () => {
  const scheme = window.location.protocol === 'https:' ? 'wss' : 'ws'
  const host = window.location.host
  const token = encodeURIComponent(userStore.token)
  return `${scheme}://${host}/api/ws?token=${token}`
}

const initWebsocket = () => {
  if (!userStore.token || !props.otherUser) return
  if (socket.value && socket.value.readyState === WebSocket.OPEN) return

  try {
    const ws = new WebSocket(getWebsocketUrl())
    socket.value = ws

    ws.onopen = () => {
      wsConnected.value = true
    }

    ws.onmessage = async (event) => {
      try {
        const packet = JSON.parse(event.data)
        if (packet.type === 'message_received' || packet.type === 'message_sent') {
          const message = packet.data
          if (!message) return
          if (
            message.from_user_id === props.otherUser.id ||
            message.to_user_id === props.otherUser.id
          ) {
            appendMessage(message)
          }
        }
      } catch (error) {
        console.error('WebSocket 消息解析失败:', error)
      }
    }

    ws.onclose = () => {
      wsConnected.value = false
      socket.value = null
    }

    ws.onerror = (error) => {
      console.error('WebSocket 错误:', error)
      wsConnected.value = false
    }
  } catch (error) {
    console.error('WebSocket 连接失败:', error)
  }
}

const handleClose = () => {
  closeWebsocket()
  visible.value = false
}
const closeWebsocket = () => {
  if (socket.value) {
    socket.value.close()
    socket.value = null
  }
  wsConnected.value = false
}

const appendMessage = async (message) => {
  if (!message || !props.otherUser) return
  const exists = messages.value.some((item) => item.id === message.id)
  if (!exists) {
    messages.value.push(message)
    await nextTick()
    scrollToBottom()
  }
}

const send = async () => {
  if (!inputContent.value.trim() || !props.otherUser) return
  if (!socket.value || socket.value.readyState !== WebSocket.OPEN) {
    ElMessage.error('聊天连接未就绪，请稍后重试')
    return
  }
  console.log(props.otherUser)
  const payload = {
    type: 'send_message',
    data: {
      to_user_id: props.otherUser.id,
      content: inputContent.value.trim(),
    },
  }
  console.log(payload)
  socket.value.send(JSON.stringify(payload))
  inputContent.value = ''
}

const scrollToBottom = () => {
  if (messageListRef.value) {
    messageListRef.value.scrollTop = messageListRef.value.scrollHeight
  }
}

const handleScroll = () => {
  if (!messageListRef.value || loadingMore.value || !hasMore.value) return
  if (messageListRef.value.scrollTop <= 50) {
    loadMessages(false)
  }
}

const formatTime = (timeStr) => {
  if (!timeStr) return ''
  const date = new Date(timeStr)
  const now = new Date()
  const diff = now - date

  if (diff < 24 * 60 * 60 * 1000 && date.getDate() === now.getDate()) {
    return `${date.getHours().toString().padStart(2, '0')}:${date.getMinutes().toString().padStart(2, '0')}`
  }

  const yesterday = new Date(now)
  yesterday.setDate(now.getDate() - 1)
  if (date.getDate() === yesterday.getDate()) {
    return `昨天 ${date.getHours().toString().padStart(2, '0')}:${date.getMinutes().toString().padStart(2, '0')}`
  }

  return `${date.getMonth() + 1}/${date.getDate()} ${date.getHours().toString().padStart(2, '0')}:${date.getMinutes().toString().padStart(2, '0')}`
}

onUnmounted(() => {
  closeWebsocket()
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
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
      <div class="message-list" ref="messageListRef">
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

        <div v-if="loading" class="loading-more">
          <el-icon class="is-loading"><Loading /></el-icon>
          加载中...
        </div>
      </div>

      <!-- 输入框 -->
      <div class="input-area">
        <el-input
            v-model="inputContent"
            type="textarea"
            :rows="3"
            placeholder="输入消息..."
            @keyup.ctrl.enter="send"
        />
        <div class="input-actions">
          <el-button type="primary" @click="send" :disabled="!inputContent.trim()">
            发送 (Ctrl+Enter)
          </el-button>
        </div>
      </div>
    </div>
  </el-drawer>
</template>

<script setup>
import { ref, watch, nextTick, onUnmounted } from 'vue'
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
const currentUserId = userStore.userInfo?.id

const visible = ref(false)
const messages = ref([])
const inputContent = ref('')
const loading = ref(false)
const messageListRef = ref(null)
let pollingTimer = null

// 监听外部控制
watch(() => props.modelValue, (val) => {
  visible.value = val
  if (val && props.otherUser) {
    loadMessages()
    startPolling()
  } else {
    stopPolling()
  }
})

// 监听可见性
watch(visible, (val) => {
  emit('update:modelValue', val)
  if (!val) {
    stopPolling()
  }
})

// 加载消息
const loadMessages = async (append = false) => {
  if (!props.otherUser) return

  loading.value = true
  try {
    const response = await getConversation(props.otherUser.id, 1, 100)
    if (append) {
      messages.value = [...response.data.messages, ...messages.value]
    } else {
      messages.value = response.data.messages || []
      await nextTick()
      scrollToBottom()
    }
  } catch (error) {
    console.error('加载消息失败:', error)
  } finally {
    loading.value = false
  }
}

// 发送消息
const send = async () => {
  if (!inputContent.value.trim()) return
  if (!props.otherUser) return

  try {
    await sendMessage(props.otherUser.id, inputContent.value)
    inputContent.value = ''
    await loadMessages()
    emit('message-sent')
    // 滚动到底部
    await nextTick()
    scrollToBottom()
  } catch (error) {
    ElMessage.error('发送失败: ' + error.message)
  }
}

// 滚动到底部
const scrollToBottom = () => {
  if (messageListRef.value) {
    messageListRef.value.scrollTop = messageListRef.value.scrollHeight
  }
}

// 格式化时间
const formatTime = (timeStr) => {
  if (!timeStr) return ''
  const date = new Date(timeStr)
  return `${date.getHours().toString().padStart(2, '0')}:${date.getMinutes().toString().padStart(2, '0')}`
}

// 开始轮询（每3秒检查新消息）
const startPolling = () => {
  if (pollingTimer) clearInterval(pollingTimer)
  pollingTimer = setInterval(() => {
    if (visible.value && props.otherUser) {
      loadMessages()
    }
  }, 3000)
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
</style>
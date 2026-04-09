<template>
  <el-drawer
      v-model="visible"
      title="消息列表"
      direction="rtl"
      size="350px"
  >
    <div class="conversation-list">
      <div
          v-for="conv in conversations"
          :key="conv.user_id"
          class="conversation-item"
          @click="selectConversation(conv)"
      >
        <div class="conv-avatar">
          <el-avatar :size="40">
            {{ conv.username?.charAt(0) }}
          </el-avatar>
          <div v-if="conv.unread_count > 0" class="unread-badge">
            {{ conv.unread_count > 99 ? '99+' : conv.unread_count }}
          </div>
        </div>
        <div class="conv-info">
          <div class="conv-name">{{ conv.username }}</div>
          <div class="conv-last-message">{{ conv.last_message }}</div>
        </div>
        <div class="conv-time">{{ formatTime(conv.last_time) }}</div>
      </div>

      <el-empty v-if="conversations.length === 0" description="暂无消息" />
    </div>
  </el-drawer>
</template>

<script setup>
import { ref, watch } from 'vue'
import { getConversationList } from '../api/message'

const props = defineProps({
  modelValue: Boolean
})

const emit = defineEmits(['update:modelValue', 'select-chat'])

const visible = ref(false)
const conversations = ref([])
let pollingTimer = null

// 加载对话列表
const loadConversations = async () => {
  try {
    const response = await getConversationList()
    conversations.value = response.data || []
  } catch (error) {
    console.error('加载对话列表失败:', error)
  }
}

// 格式化时间
const formatTime = (timeStr) => {
  if (!timeStr) return ''
  const date = new Date(timeStr)
  const now = new Date()
  const diff = now - date

  if (diff < 24 * 60 * 60 * 1000) {
    return `${date.getHours()}:${date.getMinutes().toString().padStart(2, '0')}`
  }
  return `${date.getMonth() + 1}/${date.getDate()}`
}

// 选择对话
const selectConversation = (conv) => {
  emit('select-chat', {
    id: conv.user_id,
    username: conv.username,
    avatar: conv.avatar
  })
  emit('update:modelValue', false)
}

// 开始轮询
const startPolling = () => {
  if (pollingTimer) clearInterval(pollingTimer)
  pollingTimer = setInterval(() => {
    if (visible.value) {
      loadConversations()
    }
  }, 5000)
}

// 停止轮询
const stopPolling = () => {
  if (pollingTimer) {
    clearInterval(pollingTimer)
    pollingTimer = null
  }
}

watch(() => props.modelValue, (val) => {
  visible.value = val
  if (val) {
    loadConversations()
    startPolling()
  } else {
    stopPolling()
  }
})

watch(visible, (val) => {
  emit('update:modelValue', val)
})
</script>

<style scoped>
.conversation-list {
  display: flex;
  flex-direction: column;
}

.conversation-item {
  display: flex;
  align-items: center;
  padding: 12px 16px;
  cursor: pointer;
  transition: background 0.2s;
  border-bottom: 1px solid #f0f0f0;
}

.conversation-item:hover {
  background: #f5f5f5;
}

.conv-avatar {
  position: relative;
  margin-right: 12px;
}

.unread-badge {
  position: absolute;
  top: -4px;
  right: -8px;
  background: #f56c6c;
  color: #fff;
  font-size: 10px;
  min-width: 16px;
  height: 16px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0 4px;
}

.conv-info {
  flex: 1;
  min-width: 0;
}

.conv-name {
  font-weight: 500;
  margin-bottom: 4px;
}

.conv-last-message {
  font-size: 12px;
  color: #999;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.conv-time {
  font-size: 10px;
  color: #bbb;
  margin-left: 8px;
}
</style>
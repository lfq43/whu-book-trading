<template>
  <div id="app">
    <nav class="navbar">
      <div class="nav-container">
        <div class="logo">
          <router-link to="/">📚 二手书交易平台</router-link>
        </div>
        <div class="nav-links">
          <router-link to="/">浏览书籍</router-link>
          <router-link v-if="userStore.isLoggedIn" to="/publish">发布书籍</router-link>
          <router-link v-if="userStore.isLoggedIn" to="/my-batches">我的发布</router-link>

          <!-- 消息图标 -->
          <div v-if="userStore.isLoggedIn" class="message-icon" @click="openMessages">
            <el-badge :value="unreadCount" :hidden="unreadCount === 0">
              <el-icon :size="20"><ChatDotRound /></el-icon>
            </el-badge>
          </div>

          <template v-if="userStore.isLoggedIn">
            <span class="user-info">{{ userStore.userInfo?.username }}</span>
            <el-button link @click="handleLogout">退出</el-button>
          </template>
          <router-link v-else to="/login">登录/注册</router-link>
        </div>
      </div>
    </nav>
    <router-view />

    <!-- 消息列表抽屉 -->
    <MessageList v-model="messageListVisible" @select-chat="onSelectChat" />

    <!-- 聊天窗口 -->
    <ChatWindow
        v-model="chatVisible"
        :other-user="selectedUser"
        @message-sent="onMessageSent"
    />
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { ChatDotRound } from '@element-plus/icons-vue'
import { useUserStore } from './stores/user'
import { useNotification } from './composables/useNotification'
import MessageList from './components/MessageList.vue'
import ChatWindow from './components/ChatWindow.vue'

const router = useRouter()
const userStore = useUserStore()
const { unreadCount, startPolling } = useNotification()

const messageListVisible = ref(false)
const chatVisible = ref(false)
const selectedUser = ref(null)

// 启动消息轮询
if (userStore.isLoggedIn) {
  startPolling()
}

// 打开消息列表
const openMessages = () => {
  messageListVisible.value = true
}

// 选择聊天对象
const onSelectChat = (user) => {
  selectedUser.value = user
  messageListVisible.value = false
  chatVisible.value = true
}

// 消息发送后刷新未读
const onMessageSent = () => {
  // 刷新未读数量
  const { checkUnread } = useNotification()
  checkUnread()
}

const handleLogout = () => {
  userStore.logout()
  ElMessage.success('已退出登录')
  router.push('/login')
}
</script>

<style>
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

body {
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
  background: #f5f5f5;
}

.navbar {
  background: #fff;
  box-shadow: 0 2px 8px rgba(0,0,0,0.1);
  position: sticky;
  top: 0;
  z-index: 100;
}

.nav-container {
  max-width: 1200px;
  margin: 0 auto;
  padding: 0 20px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  height: 60px;
}

.logo a {
  font-size: 20px;
  font-weight: bold;
  color: #409eff;
  text-decoration: none;
}

.nav-links {
  display: flex;
  gap: 20px;
  align-items: center;
}

.nav-links a {
  color: #333;
  text-decoration: none;
  transition: color 0.2s;
}

.nav-links a:hover {
  color: #409eff;
}

.user-info {
  color: #409eff;
  margin-right: 10px;
}

.message-icon {
  cursor: pointer;
  display: flex;
  align-items: center;
}
</style>

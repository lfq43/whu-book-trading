<template>
  <div id="app">
    <nav class="navbar">
      <div class="nav-container">
        <div class="logo">
          <router-link to="/">WHU书籍交易平台</router-link>
        </div>
        <div class="nav-links">
          <router-link to="/">浏览书籍</router-link>
          <router-link v-if="userStore.isLoggedIn" to="/publish">发布书籍</router-link>
          <router-link v-if="userStore.isLoggedIn" to="/my-batches">我的发布</router-link>
          <router-link v-if="userStore.isAdmin" to="/admin">管理员</router-link>

          <!-- 消息图标 -->
          <div v-if="userStore.isLoggedIn" class="message-icon" @click="openMessages">
            <el-badge :value="unreadCount" :hidden="unreadCount === 0">
              <el-icon :size="20"><ChatDotRound /></el-icon>
            </el-badge>
          </div>

          <!-- 个人空间入口 -->
          <div v-if="userStore.isLoggedIn" class="user-menu">
            <el-dropdown @command="handleMenuCommand">
              <div class="user-info">
                <el-avatar :size="28" :src="userStore.userInfo?.avatar">
                  {{ userStore.userInfo?.username?.charAt(0) }}
                </el-avatar>
                <span>{{ userStore.userInfo?.username }}</span>
                <el-icon><ArrowDown /></el-icon>
              </div>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item command="space">个人空间</el-dropdown-item>
                  <el-dropdown-item command="logout" divided>退出登录</el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
          </div>
          <router-link v-else to="/login">登录/注册</router-link>
          <router-link
              v-if="userStore.isLoggedIn && !userStore.isAdmin"
              to=""
              @click.prevent="handleFeedback"
          >
            反馈
          </router-link>
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
    />
  </div>
</template>

<script setup>
import {computed, ref, watch} from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import {ArrowDown, ChatDotRound} from '@element-plus/icons-vue'
import { useUserStore } from './stores/user'
import { useNotification } from './composables/useNotification'
import MessageList from './components/MessageList.vue'
import ChatWindow from './components/ChatWindow.vue'

const router = useRouter()
const userStore = useUserStore()
const { unreadCount, startPolling, stopPolling, checkUnread } = useNotification()

const messageListVisible = ref(false)
const chatVisible = ref(false)
const selectedUser = ref(null)

const currentUserId = computed(() => userStore.userInfo?.id)
const adminId = 4
const adminUser = {
  id: adminId,
}

watch(
  () => userStore.isLoggedIn,
  (isLoggedIn) => {
    if (isLoggedIn) {
      startPolling()
    } else {
      stopPolling()
    }
  },
  { immediate: true }
)

// 打开消息列表
const openMessages = () => {
  messageListVisible.value = true
  checkUnread()
}

// 选择聊天对象
const onSelectChat = (user) => {
  selectedUser.value = user
  messageListVisible.value = false
  chatVisible.value = true
}

const handleMenuCommand = (command) => {
  if (command === 'space') {
    router.push(`/user/${userStore.userInfo?.id}`)
  } else if (command === 'logout') {
    userStore.logout()
    ElMessage.success('已退出登录')
    router.push('/login')
  }
}

const handleFeedback = () => {
  console.log('feedback')
  chatVisible.value = false
  selectedUser.value = adminUser
  chatVisible.value = true
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

.user-menu {
  cursor: pointer;
}

.user-info {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 4px 8px;
  border-radius: 20px;
  transition: background 0.2s;
}

.user-info:hover {
  background: #f0f0f0;
}

.seller-name {
  cursor: pointer;
  color: #409eff;
}

.seller-name:hover {
  text-decoration: underline;
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

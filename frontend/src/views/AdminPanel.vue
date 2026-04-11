<template>
  <div class="admin-panel-container" v-loading="loading">
    <el-card class="admin-card">
      <div class="admin-header">
        <h2>管理员控制台</h2>
        <span>您可以查看用户列表，并封禁账号。</span>
      </div>

      <el-table
          :data="users"
          stripe
          style="width: 100%"
      >
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="account" label="账号" width="180" />
        <el-table-column prop="username" label="用户名" width="180" />
        <el-table-column prop="email" label="邮箱" />
        <el-table-column prop="created_at" label="注册时间" width="180" />
        <el-table-column label="状态" width="120">
          <template #default="{ row }">
            <el-tag :type="row.is_banned ? 'danger' : 'success'">
              {{ row.is_banned ? '已封禁' : '正常' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="180">
          <template #default="{ row }">
            <el-button
                type="danger"
                size="small"
                :disabled="row.is_banned || row.account === adminAccount"
                @click="handleBan(row.id)"
            >
              封禁
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getAllUsers, banUser } from '../api/user'
import { useUserStore } from '../stores/user'

const loading = ref(false)
const users = ref([])
const userStore = useUserStore()
const adminAccount = userStore.userInfo?.account || 'admin'

const loadUsers = async () => {
  loading.value = true
  try {
    const response = await getAllUsers()
    users.value = response.data
  } catch (error) {
    ElMessage.error('加载用户失败')
  } finally {
    loading.value = false
  }
}

const handleBan = async (userId) => {
  try {
    await ElMessageBox.confirm('确认封禁该账号吗？', '警告', {
      confirmButtonText: '确认',
      cancelButtonText: '取消',
      type: 'warning',
    })
    await banUser(userId)
    ElMessage.success('封禁成功')
    await loadUsers()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error(error.message || '封禁失败')
    }
  }
}

onMounted(() => {
  loadUsers()
})
</script>

<style scoped>
.admin-panel-container {
  max-width: 1100px;
  margin: 20px auto;
  padding: 0 20px;
}

.admin-card {
  padding: 24px;
}

.admin-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.admin-header h2 {
  margin: 0;
  font-size: 22px;
}

.admin-header span {
  color: #909399;
}
</style>

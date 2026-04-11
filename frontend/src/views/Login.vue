<template>
  <div class="login-container">
    <el-card class="login-card">
      <template #header>
        <div class="card-header">
          <h2>二手书交易平台</h2>
          <span>登录</span>
        </div>
      </template>

      <el-form
          ref="formRef"
          :model="form"
          :rules="rules"
          label-width="80px"
      >
        <el-form-item label="用户名" prop="username">
          <!--prop对应rules中对应规则-->
          <el-input
              v-model="form.username"
              placeholder="请输入用户名"
              prefix-icon="User"
          />
        </el-form-item>

        <el-form-item label="密码" prop="password">
          <el-input
              v-model="form.password"
              type="password"
              placeholder="请输入密码"
              prefix-icon="Lock"
              show-password
          />
        </el-form-item>

        <el-form-item>
          <el-button type="primary" :loading="loading" @click="handleLogin">
            登录
          </el-button>
          <el-button @click="goToRegister">
            去注册
          </el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { useUserStore } from '../stores/user'
import { login } from '../api/user'

const router = useRouter()
const userStore = useUserStore()
// 1. ref：用于基本类型和对象（需要 .value 访问）
const formRef = ref(null)
const loading = ref(false)

// 表单数据
// 2. reactive：只用于对象（直接访问，不需要 .value）
const form = reactive({
  username: '',
  password: ''
})
// 3. 什么时候用 ref？什么时候用 reactive？
// - 基本类型（string, number, boolean）必须用 ref
// - 对象可以用 ref 或 reactive
// - 推荐：简单的用 ref，复杂的对象用 reactive

// 表单验证规则
const rules = {
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' },
    { min: 3, max: 15, trigger: 'blur' }
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 6, message: '密码长度至少 6 位', trigger: 'blur' }
  ]
}

// 登录处理
const handleLogin = async () => {
  // 1. 验证表单
  if (!formRef.value) return
  //如果向validate里传入函数，它就会把本应该返回的valid给这个回调函数
  await formRef.value.validate(async (valid) => {//表单验证,检查组件是否和它的rules符合
    if (!valid) return

    loading.value = true
    try {
      // 2. 调用登录 API
      const response = await login({
        username: form.username,
        password: form.password
      })

      // 3. 保存 token 和用户信息到 store
      userStore.setToken(response.data.token)
      userStore.setUserInfo(response.data.user)

      ElMessage.success('登录成功！')

      // 4. 跳转到首页（后续会改成书籍列表页）
      router.push('/')
    } catch (error) {
      ElMessage.error(error.message || '登录失败，请检查用户名和密码')
    } finally {
      loading.value = false
    }
  })
}

// 跳转到注册页
const goToRegister = () => {
  router.push('/register')
}
</script>

<style scoped>
.login-container {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100vh;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.login-card {
  width: 450px;
}

.card-header {
  text-align: center;
}

.card-header h2 {
  margin: 0 0 8px 0;
  color: #333;
}

.card-header span {
  font-size: 14px;
  color: #666;
}

.el-form-item:last-child {
  margin-bottom: 0;
}

.el-button {
  width: 100px;
}
</style>
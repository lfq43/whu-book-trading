<template>
  <div class="register-container">
    <el-card class="register-card">
      <template #header>
        <div class="card-header">
          <h2>二手书交易平台</h2>
          <span>注册新账号</span>
        </div>
      </template>

      <el-form
          ref="formRef"
          :model="form"
          :rules="rules"
          label-width="80px"
      >
        <el-form-item label="账号名" prop="account">
          <el-input
              v-model="form.account"
              placeholder="请输入账号名"
              prefix-icon="User"
          />
        </el-form-item>

        <el-form-item label="用户名" prop="username">
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
              placeholder="请输入密码（至少6位）"
              prefix-icon="Lock"
              show-password
          />
        </el-form-item>

        <el-form-item label="确认密码" prop="confirmPassword">
          <el-input
              v-model="form.confirmPassword"
              type="password"
              placeholder="请再次输入密码"
              prefix-icon="Lock"
              show-password
          />
        </el-form-item>

        <el-form-item label="邮箱" prop="email">
          <el-input
              v-model="form.email"
              placeholder="请输入邮箱（选填）"
              prefix-icon="Message"
          >
            <template #append>
              <el-button
                type="primary"
                size="small"
                :loading="codeSending"
                :disabled="!form.email || countDown > 0"
                @click="handleSendCode"
              >
                {{ countDown > 0 ? `重新发送(${countDown}s)` : '发送验证码' }}
              </el-button>
            </template>
          </el-input>
        </el-form-item>

        <el-form-item label="邮箱验证码" prop="verificationCode">
          <el-input
              v-model="form.verificationCode"
              placeholder="请输入验证码"
              prefix-icon="Lock"
          />
        </el-form-item>

        <el-form-item>
          <el-button type="primary" :loading="loading" @click="handleRegister">
            注册
          </el-button>
          <el-button @click="goToLogin">
            去登录
          </el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script setup>
import { ref, reactive, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { register, sendVerificationCode } from '../api/user'

const router = useRouter()
const formRef = ref(null)
const loading = ref(false)
const codeSending = ref(false)
const countDown = ref(0)
let timer = null

// 表单数据
const form = reactive({
  account: '',
  username: '',
  password: '',
  confirmPassword: '',
  email: '',
  verificationCode: ''
})

// 自定义验证函数：确认密码
const validateConfirmPassword = (rule, value, callback) => {
  if (value !== form.password) {
    callback(new Error('两次输入的密码不一致'))
  } else {
    callback()
  }
}

// 表单验证规则
const rules = {
  account: [
    { required: true, message: '请输入账号名', trigger: 'blur' },
    { min: 3, max: 50, trigger: 'blur' }
  ],
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' },
    { min: 3, max: 10, trigger: 'blur' }
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 6, message: '密码长度至少 6 位', trigger: 'blur' }
  ],
  confirmPassword: [
    { required: true, message: '请再次输入密码', trigger: 'blur' },
    { validator: validateConfirmPassword, trigger: 'blur' }
  ],
  email: [
    { type: 'email', message: '请输入正确的邮箱格式', trigger: 'blur' }
  ],
  verificationCode: [
    { required: false, message: '请输入验证码', trigger: 'blur' },
    { len: 6, message: '验证码长度为 6 位', trigger: 'blur' }
  ]
}

const startCountdown = () => {
  countDown.value = 60
  timer = setInterval(() => {
    countDown.value -= 1
    if (countDown.value <= 0) {
      clearInterval(timer)
      timer = null
    }
  }, 1000)
}

const handleSendCode = async () => {
  if (!form.email) {
    ElMessage.warning('请先填写邮箱')
    return
  }
  if (countDown.value > 0) {
    return
  }

  codeSending.value = true
  try {
    await sendVerificationCode({ email: form.email })
    ElMessage.success('验证码已发送，请查收邮箱')
    startCountdown()
  } catch (error) {
    ElMessage.error(error.message || '发送验证码失败')
  } finally {
    codeSending.value = false
  }
}

onUnmounted(() => {
  if (timer) {
    clearInterval(timer)
  }
})

// 注册处理
const handleRegister = async () => {
  // 1. 验证表单
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (!valid) return

    if (form.email && !form.verificationCode) {
      ElMessage.warning('请输入邮箱验证码')
      return
    }

    loading.value = true
    try {
      // 2. 调用注册 API（注意：不需要传 confirmPassword）
      await register({
        account: form.account,
        username: form.username,
        password: form.password,
        email: form.email,
        verification_code: form.verificationCode
      })

      ElMessage.success('注册成功！请登录')

      // 3. 跳转到登录页
      router.push('/login')
    } catch (error) {
      ElMessage.error(error.message || '注册失败，请稍后重试')
    } finally {
      loading.value = false
    }
  })
}

// 跳转到登录页
const goToLogin = () => {
  router.push('/login')
}
</script>

<style scoped>
.register-container {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100vh;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.register-card {
  width: 500px;
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
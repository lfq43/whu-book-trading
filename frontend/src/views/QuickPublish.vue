<template>
  <div class="quick-publish-container">
    <div class="header">
      <h1>📦 快速批量出书</h1>
      <p>适合一次出多本书，只需填写整体信息</p>
    </div>

    <el-card class="publish-card">
      <el-form :model="form" :rules="rules" ref="formRef" label-width="100px">
        <el-form-item label="批次标题" prop="title">
          <el-input
              v-model="form.title"
              placeholder="例如：考研数学全套教材、大学英语四六级资料"
          />
          <div class="form-tip">买家搜索时会看到这个标题</div>
        </el-form-item>

        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="总价(元)" prop="total_price">
              <el-input-number
                  v-model="form.total_price"
                  :min="0.01"
                  :precision="2"
                  :step="10"
                  style="width: 100%"
                  placeholder="所有书的总价"
              />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="数量(本)" prop="total_count">
              <el-input-number
                  v-model="form.total_count"
                  :min="1"
                  :max="100"
                  style="width: 100%"
              />
            </el-form-item>
          </el-col>
        </el-row>

        <el-form-item v-if="form.total_price > 0 && form.total_count > 0">
          <div class="price-info">
            💰 平均每本：<strong>¥{{ unitPrice.toFixed(2) }}</strong>
          </div>
        </el-form-item>

        <el-form-item label="新旧程度" prop="condition">
          <el-radio-group v-model="form.condition">
            <el-radio label="全新">全新</el-radio>
            <el-radio label="几乎全新">几乎全新</el-radio>
            <el-radio label="良好">良好</el-radio>
            <el-radio label="有笔记">有笔记</el-radio>
            <el-radio label="破损">破损</el-radio>
          </el-radio-group>
        </el-form-item>

        <el-form-item label="描述" prop="description">
          <el-input
              v-model="form.description"
              type="textarea"
              :rows="4"
              placeholder="描述这批书的情况，例如：包含哪些书、使用程度、是否有笔记等"
          />
        </el-form-item>

        <el-form-item label="图片" prop="image">
          <el-input
              v-model="form.image"
              placeholder="请输入图片URL（可选）"
          />
          <div class="form-tip">可以放一张这批书的合照</div>
        </el-form-item>

        <el-form-item>
          <el-button type="primary" :loading="submitting" @click="submitPublish" size="large">
            快速发布 {{ form.total_count }} 本书
          </el-button>
          <el-button @click="resetForm">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 预览区域 -->
    <el-card v-if="form.title" class="preview-card">
      <template #header>
        <span>预览效果</span>
      </template>
      <div class="preview">
        <div class="preview-image">
          <el-image :src="form.image || '/placeholder.png'" fit="cover" style="width: 100px; height: 100px;">
            <template #error>
              <div class="image-placeholder">📷</div>
            </template>
          </el-image>
        </div>
        <div class="preview-info">
          <h3>{{ form.title || '标题示例' }}</h3>
          <p>{{ form.description || '描述示例...' }}</p>
          <div class="preview-price">
            <span class="total-price">总价 ¥{{ form.total_price || 0 }}</span>
            <span class="unit-price">（{{ form.total_count || 0 }}本，平均¥{{ unitPrice.toFixed(2) }}）</span>
          </div>
          <el-tag size="small">{{ form.condition || '良好' }}</el-tag>
        </div>
      </div>
    </el-card>
  </div>
</template>

<script setup>
import { ref, reactive, computed } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { quickBatchCreate } from '../api/book'

const router = useRouter()
const formRef = ref(null)
const submitting = ref(false)

const form = reactive({
  title: '',
  total_price: 0,
  total_count: 1,
  condition: '良好',
  description: '',
  image: ''
})

const rules = {
  title: [
    { required: true, message: '请输入批次标题', trigger: 'blur' },
    { min: 2, max: 200, message: '标题长度 2-200 字符', trigger: 'blur' }
  ],
  total_price: [
    { required: true, message: '请输入总价', trigger: 'blur' },
    { type: 'number', min: 0.01, message: '总价必须大于0', trigger: 'blur' }
  ],
  total_count: [
    { required: true, message: '请输入数量', trigger: 'blur' },
    { type: 'number', min: 1, max: 100, message: '数量在 1-100 之间', trigger: 'blur' }
  ]
}

const unitPrice = computed(() => {
  if (form.total_price > 0 && form.total_count > 0) {
    return form.total_price / form.total_count
  }
  return 0
})

const submitPublish = async () => {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (!valid) return

    submitting.value = true
    try {
      const response = await quickBatchCreate({
        title: form.title,
        total_price: form.total_price,
        total_count: form.total_count,
        condition: form.condition,
        description: form.description,
        image: form.image
      })
      ElMessage.success(response.message)
      // 跳转到我的批次页面
      router.push('/my-batches')
    } catch (error) {
      ElMessage.error(error.message || '发布失败，请稍后重试')
    } finally {
      submitting.value = false
    }
  })
}

const resetForm = () => {
  form.title = ''
  form.total_price = 0
  form.total_count = 1
  form.condition = '良好'
  form.description = ''
  form.image = ''
  ElMessage.info('已重置')
}
</script>

<style scoped>
.quick-publish-container {
  max-width: 800px;
  margin: 0 auto;
  padding: 20px;
}

.header {
  text-align: center;
  margin-bottom: 30px;
}

.header h1 {
  color: #333;
  margin-bottom: 10px;
}

.header p {
  color: #666;
}

.publish-card {
  margin-bottom: 20px;
}

.form-tip {
  font-size: 12px;
  color: #999;
  margin-top: 4px;
}

.price-info {
  background: #e6f7ff;
  padding: 10px;
  border-radius: 4px;
  color: #1890ff;
}

.price-info strong {
  font-size: 18px;
  color: #f56c6c;
}

.preview-card {
  background: #fafafa;
}

.preview {
  display: flex;
  gap: 20px;
}

.preview-image {
  width: 100px;
  height: 100px;
  background: #f0f0f0;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 40px;
}

.preview-info {
  flex: 1;
}

.preview-info h3 {
  margin: 0 0 8px 0;
}

.preview-price {
  margin: 10px 0;
}

.total-price {
  font-size: 18px;
  font-weight: bold;
  color: #f56c6c;
}

.unit-price {
  color: #999;
  font-size: 14px;
}
</style>
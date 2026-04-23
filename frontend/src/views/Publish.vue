<template>
  <div class="publish-container">
    <div class="header">
      <h1>📦 发布一批书</h1>
      <p>填写书名列表</p>
    </div>

    <el-card class="publish-card">
      <el-form :model="form" :rules="rules" ref="formRef" label-width="100px">
        <el-form-item label="批次标题" prop="title">
          <el-input v-model="form.title" placeholder="默认: 出书出书" />
        </el-form-item>

        <!-- 动态书名列表 -->
        <el-form-item label="书籍列表" prop="book_names" required>
          <div class="book-list-editor">
            <div
                v-for="(book, index) in form.book_names"
                :key="index"
                class="book-name-item"
            >
              <el-input
                  v-model="form.book_names[index]"
                  placeholder="书名"
                  style="flex: 1"
              />
              <el-button
                  type="danger"
                  circle
                  size="small"
                  @click="removeBook(index)"
                  :disabled="form.book_names.length === 1"
              >
                <el-icon><Delete /></el-icon>
              </el-button>
            </div>
            <el-button type="primary" plain @click="addBook" style="margin-top: 10px">
              <el-icon><Plus /></el-icon>
              添加一本书
            </el-button>
            <el-button type="primary" plain @click="quick_add = true" style="margin-top: 10px">
              <el-icon><Plus /></el-icon>
              快捷添加多本书名
            </el-button>
            <el-button type="primary" plain v-if="quick_add" @click="addBooks" style="margin-top: 10px">
              <el-icon><Check /></el-icon>
              提交
            </el-button>
            <el-input
                v-if="quick_add"
                v-model="booknames"
                type="textarea"
                :rows="5"
                placeholder="书名之间用空格分隔"
            />
          </div>
          <div class="form-tip">列出每一本书的名字</div>
        </el-form-item>

        <el-form-item label="描述" prop="description">
          <el-input
              v-model="form.description"
              type="textarea"
              :rows="4"
              placeholder="补充说明：书籍情况、是否有笔记、交接书籍方式等"
          />
        </el-form-item>

        <!-- 图片上传组件（修改后） -->
        <el-form-item label="图片">
          <ImageUploader
              v-model="form.images"
              :max-count="6"
              @upload-success="onImagesUploaded"
              ref="imageUploaderRef"
          />
          <div class="form-tip">支持多张图片，最多6张，点击图片可放大预览</div>
        </el-form-item>

        <el-form-item label="联系方式" prop="contact">
          <el-input
              v-model="form.contact"
              placeholder="微信号 / QQ号 / 手机号"
          />
          <div class="form-tip">买家会通过这个联系方式找到你</div>
        </el-form-item>

        <el-form-item>
          <el-button type="primary" :loading="submitting" @click="submitPublish" size="large">
            发布 {{ form.book_names.length }} 本书
          </el-button>
          <el-button @click="resetForm">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import {Plus, Delete, Check} from '@element-plus/icons-vue'
import { createBatch, updateBatchImage } from '../api/batch'
import ImageUploader from '../components/ImageUploader.vue'

const router = useRouter()
const formRef = ref(null)
const submitting = ref(false)
const batchId = ref(null)  // 存储创建后的批次ID
const booknames = ref('')
const quick_add = ref(false)

const form = reactive({
  title: '出书出书',
  book_names: [''],
  description: '',
  image: [],
  contact: ''
})

const rules = {
  title: [
    { required: true, message: '请输入标题', trigger: 'blur' },
    { min: 2, max: 200, message: '标题长度 2-200 字符', trigger: 'blur' }
  ],
  book_names: [
    {
      validator: (rule, value, callback) => {
        if (!value || value.length === 0) {
          callback(new Error('至少添加一本书'))
        } else if (value.some(name => !name.trim())) {
          callback(new Error('请填写所有书名'))
        } else {
          callback()
        }
      },
      trigger: 'blur'
    }
  ],
  contact: [
    { required: true, message: '请输入联系方式', trigger: 'blur' },
    { min: 3, max: 100, message: '联系方式长度 3-100 字符', trigger: 'blur' }
  ]
}

const addBook = () => {
  if (form.book_names.length > 50) {
    ElMessage.warning('最多添加50本书')
    return
  }
  form.book_names.push('')
}

const addBooks = () => {
  const names = booknames.value
      .trim()
      .split(/\s+/)
      .filter(name => name !== '')
  if (names.length + form.book_names.length > 50) {
    ElMessage.warning('最多添加50本书')
    return
  }
  if (form.book_names.length === 1 && form.book_names[0] === '' && names.length !== 0) {
    form.book_names.shift()
  }
  for (const book of names) {
    form.book_names.push(book)
  }
  booknames.value = ''
  quick_add.value = false
}

const removeBook = (index) => {
  if (form.book_names.length === 1) {
    ElMessage.warning('至少保留一本书')
    return
  }
  form.book_names.splice(index, 1)
}

const submitPublish = async () => {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (!valid) return

    submitting.value = true
    try {
      // 第一步：创建批次（不带图片）
      const response = await createBatch({
        title: form.title,
        book_names: form.book_names.filter(name => name.trim()),
        description: form.description,
        contact: form.contact
        // 注意：这里不传 image
      })

      batchId.value = response.data.id
      ElMessage.success('发布成功！正在上传图片...')
      // 第二步：如果有图片，上传图片并更新批次
      if (form.images) {
        // 图片已经通过 ImageUploader 上传了，URL 已经存在
        // 只需要更新批次的 image 字段
        await updateBatchImage(batchId.value, form.images)
        ElMessage.success('图片上传成功')
      }

      router.push('/my-batches')
    } catch (error) {
      ElMessage.error(error.message || '发布失败')
    } finally {
      submitting.value = false
    }
  })
}

// 图片上传成功回调
const onImagesUploaded = (urls) => {
  form.images = urls
  console.log('当前图片列表:', urls)
}

const resetForm = () => {
  form.title = ''
  form.book_names = ['']
  form.description = ''
  form.image = []
  form.contact = ''
  batchId.value = null
}
</script>

<style scoped>
/* 样式保持不变 */
.publish-container {
  max-width: 700px;
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

.book-list-editor {
  width: 100%;
}

.book-name-item {
  display: flex;
  gap: 10px;
  margin-bottom: 10px;
  align-items: center;
}

.form-tip {
  font-size: 12px;
  color: #999;
  margin-top: 4px;
}

</style>

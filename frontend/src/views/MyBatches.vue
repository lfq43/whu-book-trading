<template>
  <div class="my-batches-container">
    <div class="header">
      <h1>我的发布</h1>
      <el-button type="primary" @click="$router.push('/publish')">+ 发布新书</el-button>
    </div>

    <div v-loading="loading" class="batch-list">
      <el-card v-for="batch in batches" :key="batch.id" class="batch-card" @click="goToDetail(batch.id)">
        <!-- 图片 -->
        <div class="batch-image">
          <ImageViewer
              :src="getFirstImage(batch)"
              :title="batch.title"
              width="100%"
              height="100px"
          />
        </div>
        <div class="batch-info">
          <div class="batch-header">
            <h3>{{ batch.title }}</h3>
          </div>

          <!-- 书籍列表预览 -->
          <div class="book-preview">
            <div
                v-for="(name, idx) in getBookNames(batch)"
                :key="idx"
                class="book-preview-item"
                :class="{ sold: getSoldStatus(batch)[idx] }"
            >
              <span class="book-preview-name">{{ name }}</span>
              <span v-if="getSoldStatus(batch)[idx]" class="sold-badge">✓</span>
            </div>
          </div>

          <div class="batch-footer">
            <span>📚 {{ getBookNames(batch).length }} 本书</span>
            <span>📅 {{ formatDate(batch.created_at) }}</span>
          </div>
        </div>
      </el-card>

      <el-empty v-if="!loading && batches.length === 0" description="还没有发布过">
        <el-button type="primary" @click="$router.push('/publish')">去发布</el-button>
      </el-empty>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { getMyBatches } from '../api/batch'
import ImageViewer from "../components/ImageViewer.vue";

const router = useRouter()
const loading = ref(false)
const batches = ref([])

const formatDate = (dateStr) => {
  if (!dateStr) return ''
  const date = new Date(dateStr)
  return `${date.getMonth() + 1}/${date.getDate()}`
}

const getBookNames = (batch) => {
  try {
    return typeof batch.book_names === 'string'
        ? JSON.parse(batch.book_names)
        : batch.book_names
  } catch {
    return []
  }
}

const getFirstImage = (batch) => {
  if (!batch.image) return '/placeholder.png'
  try {
    const images = typeof batch.image === 'string'
        ? JSON.parse(batch.image)
        : batch.image
    return images.length > 0 ? images[0] : '/placeholder.png'
  } catch {
    return '/placeholder.png'
  }
}

const getSoldStatus = (batch) => {
  try {
    const status = typeof batch.sold_status === 'string'
        ? JSON.parse(batch.sold_status)
        : batch.sold_status
    return status || []
  } catch {
    return []
  }
}

const goToDetail = (id) => {
  router.push(`/batch/${id}`)
}

const loadBatches = async () => {
  loading.value = true
  try {
    const response = await getMyBatches()
    batches.value = response.data
  } catch (error) {
    ElMessage.error('加载失败')
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  loadBatches()
})
</script>

<style scoped>
.my-batches-container {
  max-width: 800px;
  margin: 0 auto;
  padding: 20px;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 30px;
}

.header h1 {
  margin: 0;
}

.batch-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.batch-card {
  cursor: pointer;
  transition: all 0.2s;
}

.batch-card:hover {
  transform: translateX(4px);
}

.batch-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.batch-header h3 {
  margin: 0;
  font-size: 18px;
}

.book-preview {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-bottom: 12px;
}

.book-preview-item {
  background: #f5f5f5;
  padding: 4px 10px;
  border-radius: 16px;
  font-size: 13px;
  display: inline-flex;
  align-items: center;
  gap: 4px;
}

.book-preview-item.sold {
  background: #e8f5e9;
  color: #999;
}

.book-preview-name.sold {
  text-decoration: line-through;
}

.sold-badge {
  color: #67c23a;
  font-weight: bold;
}

.batch-footer {
  display: flex;
  gap: 20px;
  color: #999;
  font-size: 12px;
  padding-top: 10px;
  border-top: 1px solid #eee;
}
</style>
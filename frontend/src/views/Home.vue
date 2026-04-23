<template>
  <div class="home-container">
    <!-- 搜索栏 -->
    <div class="search-section">
      <div class="search-box">
        <el-input
            v-model="searchKeyword"
            placeholder="搜索书名、描述..."
            clearable
            size="large"
            @keyup.enter="handleSearch"
            @clear="handleSearch"
        >
          <template #prefix>
            <el-icon><Search /></el-icon>
          </template>
        </el-input>
        <el-button type="primary" size="large" @click="handleSearch">
          搜索
        </el-button>
      </div>
    </div>

    <!-- 统计信息 -->
    <div class="stats" v-if="total > 0">
      找到 <strong>{{ total }}</strong> 个发布
    </div>

    <!-- 书籍列表 -->
    <div v-loading="loading" class="batch-list">
      <el-card
          v-for="batch in batches"
          :key="batch.id"
          class="batch-card"
          shadow="hover"
          @click="goToDetail(batch.id)"
      >
        <div class="batch-content">
          <!-- 图片区域 - 使用图片预览组件 -->
          <div class="batch-image">
            <ImageViewer
                :src="getFirstImage(batch)"
                :title="batch.title"
                width="100%"
                height="120px"
            />
          </div>

          <!-- 信息区域 -->
          <div class="batch-info">
            <h3 class="batch-title">{{ batch.title }}</h3>

            <!-- 书籍列表预览 -->
            <div class="book-preview">
              <div
                  v-for="(name, idx) in getBookNames(batch)"
                  :key="idx"
                  class="book-tag"
              >
                {{ name }}
              </div>
            </div>

            <div class="info">{{ batch.description }}</div>

            <!-- 底部信息 -->
            <div class="batch-footer">
              <div class="seller-info">
                <el-avatar :size="24" :src="batch.user?.avatar">
                  {{ batch.user?.username?.charAt(0) }}
                </el-avatar>
                <span>{{ batch.user?.username }}</span>
              </div>
              <div class="batch-meta">
                <el-tag size="small" :type="getStatusType(batch.status)">
                  {{ getStatusText(batch.status) }}
                </el-tag>
                <span class="time">{{ formatDate(batch.created_at) }}</span>
              </div>
            </div>
          </div>
        </div>
      </el-card>

      <!-- 空状态 -->
      <el-empty
          v-if="!loading && batches.length === 0"
          description="暂无书籍"
      >
        <el-button type="primary" @click="$router.push('/publish')">
          去发布
        </el-button>
      </el-empty>
    </div>

    <!-- 分页 -->
    <div v-if="total > 0" class="pagination">
      <el-pagination
          v-model:current-page="currentPage"
          v-model:page-size="pageSize"
          :total="total"
          :page-sizes="[12, 24, 36]"
          layout="total, sizes, prev, pager, next"
          @size-change="handlePageChange"
          @current-change="handlePageChange"
      />
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { Search, Picture } from '@element-plus/icons-vue'
import { getBatchList } from '../api/batch'
import ImageViewer from "../components/ImageViewer.vue";

const router = useRouter()

// 数据
const loading = ref(false)
const batches = ref([])
const total = ref(0)
const currentPage = ref(1)
const pageSize = ref(12)
const searchKeyword = ref('')

// 搜索防抖
let searchTimer = null

// 解析书名列表（后端返回的是 JSON 字符串）
const getBookNames = (batch) => {
  if (!batch.book_names) return []
  try {
    return typeof batch.book_names === 'string'
        ? JSON.parse(batch.book_names)
        : batch.book_names
  } catch {
    return []
  }
}

// 获取第一张图片URL
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

// 状态文本
const getStatusText = (status) => {
  const map = {
    'available': '在售',
    'sold': '已售完'
  }
  return map[status] || '在售'
}

// 状态标签类型
const getStatusType = (status) => {
  const map = {
    'available': 'success',
    'sold': 'info'
  }
  return map[status] || 'success'
}

// 格式化日期
const formatDate = (dateStr) => {
  if (!dateStr) return ''
  const date = new Date(dateStr)
  const now = new Date()
  const diff = now - date

  // 今天内
  if (diff < 24 * 60 * 60 * 1000 && date.getDate() === now.getDate()) {
    return `今天 ${date.getHours().toString().padStart(2, '0')}:${date.getMinutes().toString().padStart(2, '0')}`
  }
  // 昨天
  const yesterday = new Date(now)
  yesterday.setDate(now.getDate() - 1)
  if (date.getDate() === yesterday.getDate()) {
    return '昨天'
  }
  // 更早
  return `${date.getMonth() + 1}/${date.getDate()}`
}

// 加载列表
const loadBatches = async () => {
  loading.value = true
  try {
    const params = {
      page: currentPage.value,
      page_size: pageSize.value
    }
    const keyword = searchKeyword.value
    if (keyword && typeof keyword === 'string' && keyword.trim()) {
      params.keyword = searchKeyword.value.trim()
    }

    const response = await getBatchList(params)
    batches.value = response.data.batches || []
    total.value = response.data.total || 0
  } catch (error) {
    console.error('加载失败:', error)
  } finally {
    loading.value = false
  }
}

// 搜索处理（带防抖）
const handleSearch = () => {
  if (searchTimer) clearTimeout(searchTimer)
  searchTimer = setTimeout(() => {
    currentPage.value = 1
    loadBatches()
  }, 300)
}

// 翻页处理
const handlePageChange = () => {
  loadBatches()
}

// 跳转详情
const goToDetail = (id) => {
  router.push(`/batch/${id}`)
}

// 监听搜索关键词变化（自动搜索）
watch(searchKeyword, () => {
  handleSearch()
})

onMounted(() => {
  loadBatches()
})
</script>

<style scoped>
.home-container {
  max-width: 1000px;
  margin: 0 auto;
  padding: 20px;
}

/* 搜索栏 */
.search-section {
  background: linear-gradient(135deg, #667eea 0%, #4b95a2 100%);
  margin: -20px -20px 20px -20px;
  padding: 40px 20px;
  border-radius: 0 0 20px 20px;
}

.search-box {
  max-width: 600px;
  margin: 0 auto;
  display: flex;
  gap: 12px;
}

.search-box .el-input {
  flex: 1;
}

/* 统计信息 */
.stats {
  margin-bottom: 20px;
  color: #666;
  font-size: 14px;
}

.stats strong {
  color: #409eff;
  font-size: 18px;
}

/* 列表 */
.batch-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.batch-card {
  cursor: pointer;
  transition: all 0.2s ease;
}

.batch-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0,0,0,0.15);
}

.batch-content {
  display: flex;
  gap: 20px;
}

/* 图片区域 */
.batch-image {
  width: 120px;
  height: 120px;
  flex-shrink: 0;
  border-radius: 8px;
  overflow: hidden;
  background: #f5f5f5;
}

.batch-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.image-placeholder {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #f0f0f0;
  color: #999;
  font-size: 32px;
}

/* 信息区域 */
.batch-info {
  flex: 1;
  min-width: 0;
}

.batch-title {
  margin: 0 0 12px 0;
  font-size: 18px;
  font-weight: 600;
  color: #333;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

/* 书籍标签预览 */
.book-preview {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-bottom: 12px;
}

.book-tag {
  background: #f0f2f5;
  padding: 4px 10px;
  border-radius: 16px;
  font-size: 13px;
  color: #333;
  display: inline-flex;
  align-items: center;
  gap: 4px;
  transition: all 0.2s;
}

.book-tag.book-sold {
  background: #e8f5e9;
  color: #999;
  text-decoration: line-through;
}

.sold-mark {
  color: #67c23a;
  font-weight: bold;
  font-size: 12px;
}

/* 底部信息 */
.batch-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 8px;
}

.seller-info {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
  color: #666;
}

.batch-meta {
  display: flex;
  align-items: center;
  gap: 12px;
}

.time {
  font-size: 12px;
  color: #999;
}

/* 分页 */
.pagination {
  margin-top: 30px;
  display: flex;
  justify-content: center;
}

/* 响应式 */
@media (max-width: 768px) {
  .batch-content {
    flex-direction: column;
  }

  .batch-image {
    width: 100%;
    height: 160px;
  }

  .batch-title {
    white-space: normal;
  }

  .batch-footer {
    flex-direction: column;
    align-items: flex-start;
    gap: 8px;
  }
}
</style>

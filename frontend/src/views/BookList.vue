<template>
  <div class="book-list-container">
    <!-- 搜索栏 -->
    <div class="search-bar">
      <el-input
          v-model="searchForm.keyword"
          placeholder="搜索书名或作者"
          clearable
          style="width: 300px"
          @keyup.enter="handleSearch"
      >
        <template #prefix>
          <el-icon><Search /></el-icon>
        </template>
      </el-input>

      <el-select v-model="searchForm.condition" placeholder="新旧程度" clearable style="width: 120px">
        <el-option label="全新" value="全新" />
        <el-option label="几乎全新" value="几乎全新" />
        <el-option label="良好" value="良好" />
        <el-option label="有笔记" value="有笔记" />
        <el-option label="破损" value="破损" />
      </el-select>

      <el-input-number v-model="searchForm.min_price" placeholder="最低价" :min="0" style="width: 120px" />
      <span> - </span>
      <el-input-number v-model="searchForm.max_price" placeholder="最高价" :min="0" style="width: 120px" />

      <el-button type="primary" @click="handleSearch">搜索</el-button>
      <el-button @click="resetSearch">重置</el-button>
    </div>

    <!-- 书籍网格 -->
    <div v-loading="loading" class="book-grid">
      <el-card
          v-for="book in books"
          :key="book.id"
          class="book-card"
          shadow="hover"
          @click="goToDetail(book.id)"
      >
        <div class="book-image">
          <el-image
              :src="getFirstImage(book.images)"
              fit="cover"
              class="book-cover"
          >
            <template #error>
              <div class="image-placeholder">
                <el-icon><Picture /></el-icon>
                <span>暂无封面</span>
              </div>
            </template>
          </el-image>
        </div>
        <div class="book-info">
          <h3 class="book-title">{{ book.title }}</h3>
          <p class="book-author">{{ book.author || '未知作者' }}</p>
          <div class="book-price">
            <span class="current-price">¥{{ book.price }}</span>
            <span v-if="book.original_price" class="original-price">¥{{ book.original_price }}</span>
          </div>
          <div class="book-meta">
            <el-tag size="small" :type="getConditionType(book.condition)">{{ book.condition }}</el-tag>
            <span class="publisher">发布于 {{ formatDate(book.created_at) }}</span>
          </div>
        </div>
      </el-card>
    </div>

    <!-- 空状态 -->
    <el-empty v-if="!loading && books.length === 0" description="暂无书籍" />

    <!-- 分页 -->
    <div v-if="total > 0" class="pagination">
      <el-pagination
          v-model:current-page="currentPage"
          v-model:page-size="pageSize"
          :total="total"
          :page-sizes="[12, 24, 36]"
          layout="total, sizes, prev, pager, next"
          @size-change="handleSearch"
          @current-change="handleSearch"
      />
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { Search, Picture } from '@element-plus/icons-vue'
import { getBookList } from '../api/book'

const router = useRouter()
const loading = ref(false)
const books = ref([])
const total = ref(0)
const currentPage = ref(1)
const pageSize = ref(12)

const searchForm = reactive({
  keyword: '',
  condition: '',
  min_price: null,
  max_price: null
})

// 获取第一张图片
const getFirstImage = (images) => {
  if (!images) return ''
  const imageList = images.split(',')
  return imageList[0]?.trim() || ''
}

// 新旧程度对应的标签类型
const getConditionType = (condition) => {
  const map = {
    '全新': 'success',
    '几乎全新': 'primary',
    '良好': '',
    '有笔记': 'warning',
    '破损': 'danger'
  }
  return map[condition] || 'info'
}

// 格式化日期
const formatDate = (dateStr) => {
  if (!dateStr) return ''
  const date = new Date(dateStr)
  return `${date.getMonth() + 1}/${date.getDate()}`
}

// 搜索书籍
const handleSearch = async () => {
  loading.value = true
  try {
    const params = {
      page: currentPage.value,
      page_size: pageSize.value,
      ...searchForm
    }
    // 过滤掉空值
    Object.keys(params).forEach(key => {
      if (params[key] === '' || params[key] === null || params[key] === undefined) {
        delete params[key]
      }
    })

    const response = await getBookList(params)
    books.value = response.data.books
    total.value = response.data.total
  } catch (error) {
    console.error('获取书籍列表失败:', error)
  } finally {
    loading.value = false
  }
}

// 重置搜索
const resetSearch = () => {
  searchForm.keyword = ''
  searchForm.condition = ''
  searchForm.min_price = null
  searchForm.max_price = null
  currentPage.value = 1
  handleSearch()
}

// 跳转详情页
const goToDetail = (id) => {
  router.push(`/book/${id}`)
}

onMounted(() => {
  handleSearch()
})
</script>

<style scoped>
.book-list-container {
  max-width: 1200px;
  margin: 0 auto;
  padding: 20px;
}

.search-bar {
  display: flex;
  gap: 15px;
  align-items: center;
  flex-wrap: wrap;
  margin-bottom: 30px;
  padding: 20px;
  background: #f5f5f5;
  border-radius: 8px;
}

.book-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(250px, 1fr));
  gap: 20px;
}

.book-card {
  cursor: pointer;
  transition: transform 0.2s;
}

.book-card:hover {
  transform: translateY(-4px);
}

.book-image {
  height: 200px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #fafafa;
  border-radius: 4px;
  overflow: hidden;
}

.book-cover {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.image-placeholder {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  color: #999;
}

.image-placeholder .el-icon {
  font-size: 48px;
  margin-bottom: 8px;
}

.book-info {
  padding: 12px 0;
}

.book-title {
  font-size: 16px;
  font-weight: 600;
  margin: 0 0 8px 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.book-author {
  color: #666;
  font-size: 14px;
  margin: 0 0 8px 0;
}

.book-price {
  margin-bottom: 8px;
}

.current-price {
  color: #f56c6c;
  font-size: 18px;
  font-weight: bold;
}

.original-price {
  color: #999;
  font-size: 12px;
  text-decoration: line-through;
  margin-left: 8px;
}

.book-meta {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 12px;
}

.publisher {
  color: #999;
}

.pagination {
  margin-top: 30px;
  display: flex;
  justify-content: center;
}
</style>
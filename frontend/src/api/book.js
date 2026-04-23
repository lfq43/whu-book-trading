import request from './user' // 复用之前的 axios 实例

// 批量发布书籍
export const batchCreateBooks = (books) => {
    return request.post('/books/batch', { books })
}

// 获取书籍列表（支持搜索筛选）
export const getBookList = (params) => {
    return request.get('/books', { params })
}

// 获取书籍详情
export const getBookDetail = (id) => {
    return request.get(`/books/${id}`)
}

// 获取我发布的书籍
export const getMyBooks = () => {
    return request.get('/user/books')
}

// 更新书籍状态
export const updateBookStatus = (id, status) => {
    return request.put(`/books/${id}/status`, { status })
}

// 删除书籍
export const deleteBook = (id) => {
    return request.delete(`/books/${id}`)
}

// 快速批量发布
export const quickBatchCreate = (data) => {
    return request.post('/books/quick', data)
}

// 获取我的批次列表
export const getMyBatches = () => {
    return request.get('/batches')
}

// 获取批次详情
export const getBatchDetail = (id) => {
    return request.get(`/batches/${id}`)
}
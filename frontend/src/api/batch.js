import request from './user'

// 发布批次
export const createBatch = (data) => {
    return request.post('/batches', data)
}

// 获取公开批次列表（只显示有未售出书籍的）
export const getBatchList = (params) => {
    return request.get('/batches', { params })
}

// 获取批次详情
export const getBatchDetail = (id) => {
    return request.get(`/batches/${id}`)
}

// 获取我的批次（全部）
export const getMyBatches = () => {
    return request.get('/user/batches')
}

// 更新单本书售出状态（传入书名）
export const updateBookSoldStatus = (batchId, bookName, sold) => {
    return request.put(`/batches/${batchId}/book-status`, {
        book_name: bookName,
        sold: sold
    })
}

// 删除批次
export const deleteBatch = (id) => {
    return request.delete(`/batches/${id}`)
}

// 更新批次图片
export const updateBatchImage = (batchId, imageUrl) => {
    return request.put(`/batches/${batchId}/image`, { image: imageUrl })
}
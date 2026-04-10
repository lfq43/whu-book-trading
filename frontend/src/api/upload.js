import request from './user'

// 上传图片
export const uploadImage = (file) => {
    const formData = new FormData()
    formData.append('image', file)

    return request.post('/upload/image', formData, {
        headers: {
            'Content-Type': 'multipart/form-data'
        }
    })
}
import request from './user'

// 上传图片（批次图片）
export const uploadImage = (file) => {
    const formData = new FormData()
    formData.append('image', file)

    return request.post('/upload/image', formData)
}

// 上传头像
export const uploadAvatar = (file) => {
    const formData = new FormData()
    formData.append('avatar', file)

    return request.post('/upload/avatar', formData)
}
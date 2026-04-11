import axios from 'axios'
import { ElMessage } from 'element-plus'
//axios用于使用js语句发送http请求
//请求方法返回promise对象，获取promise对象里的数据时需要 const 数据 = await 请求api函数
import { useUserStore } from '../stores/user'

// 创建 axios 实例
const request = axios.create({
    //请求配置
    baseURL: '/api',  // 后端 API 的基础地址
    timeout: 10000  // 请求超时时间（毫秒）
})

// 请求拦截器：在发送请求之前自动添加 token
request.interceptors.request.use(
    (config) => {
        // 获取 store 中的 token
        const userStore = useUserStore()
        const token = userStore.token

        // 如果存在 token，添加到请求头
        if (token) {
            config.headers.Authorization = `Bearer ${token}`
        }
        return config
    },
    (error) => {
        return Promise.reject(error)
    }
)

// 响应拦截器
request.interceptors.response.use(
    (response) => {
        const res = response.data
        // 根据你的后端，成功可能是 code === 0 或 200
        if (res.code !== 0 && res.code !== 200) {
            ElMessage.error(res.message || '请求失败')
            return Promise.reject(new Error(res.message || '请求失败'))
        }
        return res
    },
    (error) => {
        // 处理 HTTP 错误
        if (error.response) {
            const { status, data } = error.response

            switch (status) {
                case 400:
                    // 后端返回的 message 在这里
                    ElMessage.error(data?.message || '请求参数错误')
                    break
                case 401:
                    ElMessage.error('未登录或登录已过期')
                    // 跳转到登录页
                    router.push('/login')
                    break
                case 403:
                    ElMessage.error('没有权限')
                    break
                case 404:
                    ElMessage.error('请求的资源不存在')
                    break
                case 500:
                    ElMessage.error('服务器错误')
                    break
                default:
                    ElMessage.error(data?.message || '请求失败')
            }

            return Promise.reject(new Error(data?.message || '请求失败'))
        } else if (error.request) {
            // 请求已发出但没有收到响应
            ElMessage.error('网络连接失败')
        } else {
            // 其他错误
            ElMessage.error(error.message || '未知错误')
        }

        return Promise.reject(error)
    }
)

// 注册 API
export const register = (data) => {
    return request.post('/auth/register', data)
}

// 登录 API
export const login = (data) => {
    return request.post('/auth/login', data)
}

// 获取个人信息 API（需要认证）
export const getProfile = () => {
    return request.get('/user/profile')
}

// 获取指定用户的公开信息（用于个人空间）
export const getUserProfile = (userId) => {
    return request.get(`/users/${userId}`)
}

// 管理员：获取全部用户
export const getAllUsers = () => {
    return request.get('/admin/users')
}

// 管理员：封禁用户
export const banUser = (userId) => {
    return request.put(`/admin/users/${userId}/ban`)
}

// 更新个人资料
export const updateProfile = (data) => {
    return request.put('/user/profile', data)
}

export default request
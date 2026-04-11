import { defineStore } from 'pinia'

// 定义用户 store(token和用户信息储物柜) 导出一个生成用户store的函数
export const useUserStore = defineStore('user', {
    // 状态（数据）
    state: () => ({
        token: localStorage.getItem('token') || '',  // 从本地存储读取 token
        userInfo: (() => {
            try {
                const stored = localStorage.getItem('user_info')
                return stored ? JSON.parse(stored) : null
            } catch {
                return null
            }
        })()
    }),

    // 计算属性（类似 Vue 的 computed）
    getters: {
        isLoggedIn: (state) => !!state.token,  // 是否有 token，用于判断是否登录
        isAdmin: (state) => !!state.userInfo?.is_admin
    },

    // 方法（修改状态）
    actions: {
        // 登录成功时调用，保存 token 和用户信息
        setToken(token) {
            this.token = token
            localStorage.setItem('token', token)  // 保存到本地存储，刷新页面后不会丢失
        },

        setUserInfo(userInfo) {
            this.userInfo = userInfo
            localStorage.setItem('user_info', JSON.stringify(userInfo))
        },

        // 退出登录
        logout() {
            this.token = ''
            this.userInfo = null
            localStorage.removeItem('token')
            localStorage.removeItem('user_info')
        }
    }
})
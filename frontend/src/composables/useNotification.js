import { ref, onUnmounted } from 'vue'
import { getUnreadCount } from '../api/message'

let originalTitle = document.title
let timer = null
let isBlinking = false

// 闪烁标题
const blinkTitle = () => {
    if (isBlinking) return
    isBlinking = true

    let count = 0
    timer = setInterval(() => {
        if (count % 2 === 0) {
            document.title = '💬 新消息 - 二手书交易平台'
        } else {
            document.title = originalTitle
        }
        count++
        if (count > 10) {
            clearInterval(timer)
            document.title = originalTitle
            isBlinking = false
        }
    }, 500)
}

// 停止闪烁
const stopBlink = () => {
    if (timer) {
        clearInterval(timer)
        timer = null
    }
    document.title = originalTitle
    isBlinking = false
}

// 检查未读消息
export const useNotification = () => {
    const unreadCount = ref(0)

    const checkUnread = async () => {
        try {
            const response = await getUnreadCount()
            const newCount = response.data

            if (newCount > unreadCount.value && newCount > 0) {
                // 有新消息，闪烁标题
                blinkTitle()
            }

            unreadCount.value = newCount
        } catch (error) {
            console.error('获取未读数量失败:', error)
        }
    }

    const startPolling = () => {
        checkUnread()
    }

    const stopPolling = () => {
        stopBlink()
    }

    // 页面获得焦点时停止闪烁并刷新一次未读
    const onFocus = () => {
        stopBlink()
        checkUnread()
    }

    window.addEventListener('focus', onFocus)

    onUnmounted(() => {
        stopPolling()
        window.removeEventListener('focus', onFocus)
    })

    return {
        unreadCount,
        startPolling,
        stopPolling,
        checkUnread
    }
}
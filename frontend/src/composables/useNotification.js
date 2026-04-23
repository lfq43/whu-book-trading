import { ref, onUnmounted } from 'vue'
import { getUnreadCount, createUnreadEventSource } from '../api/message'
import { useUserStore } from '../stores/user'

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
    const userStore = useUserStore()
    const unreadCount = ref(0)
    const eventSource = ref(null)

    const checkUnread = async () => {
        try {
            const response = await getUnreadCount()
            const newCount = response.data

            if (newCount > unreadCount.value && newCount > 0) {
                blinkTitle()
            }

            unreadCount.value = newCount
        } catch (error) {
            console.error('获取未读数量失败:', error)
        }
    }

    const openSSE = () => {
        if (!userStore.token || eventSource.value) {
            return
        }

        try {
            const source = createUnreadEventSource(userStore.token)

            source.addEventListener('unread', (event) => {
                const newCount = Number(event.data)
                if (newCount > unreadCount.value && newCount > 0) {
                    blinkTitle()
                }
                unreadCount.value = newCount
            })

            source.onopen = () => {
                console.debug('SSE 未读数量已连接')
            }

            source.onerror = (event) => {
                console.error('SSE 未读数量连接错误:', event, 'readyState=', source.readyState)
                if (source.readyState === EventSource.CLOSED) {
                    closeSSE()
                    checkUnread()
                }
            }

            eventSource.value = source
        } catch (error) {
            console.error('创建 SSE 失败:', error)
            checkUnread()
        }
    }

    const closeSSE = () => {
        if (eventSource.value) {
            eventSource.value.close()
            eventSource.value = null
        }
    }

    const startPolling = () => {
        openSSE()
    }

    const stopPolling = () => {
        stopBlink()
        closeSSE()
    }

    onUnmounted(() => {
        stopPolling()
    })

    return {
        unreadCount,
        startPolling,
        stopPolling,
        checkUnread
    }
}
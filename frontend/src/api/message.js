import request from './user'

// 发送消息
export const sendMessage = (toUserId, content) => {
    return request.post('/messages', { to_user_id: toUserId, content })
}

// 获取未读消息数量
export const getUnreadCount = () => {
    return request.get('/messages/unread')
}

// 订阅未读消息数量更新
export const createUnreadEventSource = (token) => {
    const url = `/api/messages/unread/stream?token=${encodeURIComponent(token)}`
    return new EventSource(url)
}

// 获取对话列表
export const getConversationList = () => {
    return request.get('/messages/conversations')
}

// 获取与某人的聊天记录（滑动加载）
// beforeId: 查询比这个ID更旧的消息，传0或不传表示获取最新消息
export const getConversation = (userId, beforeId = 0, limit = 20) => {
    const params = { limit }//对象属性简写语法
    if (beforeId > 0) {
        params.before_id = beforeId//也是一种简写语法，简单创建属性
    }
    return request.get(`/messages/conversation/${userId}`, { params })
}
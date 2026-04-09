import request from './user'

// 发送消息
export const sendMessage = (toUserId, content) => {
    return request.post('/messages', { to_user_id: toUserId, content })
}

// 获取未读消息数量
export const getUnreadCount = () => {
    return request.get('/messages/unread')
}

// 获取对话列表
export const getConversationList = () => {
    return request.get('/messages/conversations')
}

// 获取与某人的聊天记录
export const getConversation = (userId, page = 1, pageSize = 50) => {
    return request.get(`/messages/conversation/${userId}`, {
        params: { page, page_size: pageSize }
    })
}
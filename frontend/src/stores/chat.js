import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { useUserStore } from './user'

export const useChatStore = defineStore('chat', () => {
  // 引入用户store
  const userStore = useUserStore()
  
  // 状态
  const privateChats = ref({}) // 私聊消息 {userId: [messages]}
  const groupChats = ref({})   // 群聊消息 {groupId: [messages]}
  const activeChat = ref(null) // 当前活跃的聊天 {type: 'private'|'group', id: userId|groupId}
  const socket = ref(null)     // WebSocket连接
  const isConnected = ref(false) // WebSocket连接状态
  
  // 计算属性
  const currentChatMessages = computed(() => {
    if (!activeChat.value) return []
    
    if (activeChat.value.type === 'private') {
      return privateChats.value[activeChat.value.id] || []
    } else {
      return groupChats.value[activeChat.value.id] || []
    }
  })
  
  // 方法
  // 初始化WebSocket连接
  function initSocket() {
    if (socket.value) {
      socket.value.close()
    }
    
    // 创建WebSocket连接
    const wsUrl = `${window.location.protocol === 'https:' ? 'wss:' : 'ws:'}//${window.location.host}/ws?token=${userStore.token}`
    socket.value = new WebSocket(wsUrl)
    
    // 连接建立
    socket.value.onopen = () => {
      console.log('WebSocket连接已建立')
      isConnected.value = true
    }
    
    // 接收消息
    socket.value.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data)
        handleIncomingMessage(data)
      } catch (error) {
        console.error('解析WebSocket消息失败:', error)
      }
    }
    
    // 连接关闭
    socket.value.onclose = () => {
      console.log('WebSocket连接已关闭')
      isConnected.value = false
      
      // 尝试重新连接
      if (userStore.isLoggedIn) {
        setTimeout(() => {
          initSocket()
        }, 3000)
      }
    }
    
    // 连接错误
    socket.value.onerror = (error) => {
      console.error('WebSocket连接错误:', error)
      isConnected.value = false
    }
  }
  
  // 处理接收到的消息
  function handleIncomingMessage(data) {
    const { type, message } = data
    
    if (type === 'private') {
      // 私聊消息
      const { from, to, content, timestamp } = message
      const chatUserId = from === userStore.userId ? to : from
      
      // 确保聊天记录数组存在
      if (!privateChats.value[chatUserId]) {
        privateChats.value[chatUserId] = []
      }
      
      // 添加消息
      privateChats.value[chatUserId].push({
        id: Date.now().toString(),
        senderId: from,
        content,
        timestamp
      })
    } else if (type === 'group') {
      // 群聊消息
      const { groupId, senderId, content, timestamp } = message
      
      // 确保聊天记录数组存在
      if (!groupChats.value[groupId]) {
        groupChats.value[groupId] = []
      }
      
      // 检查是否是自己发送的消息，避免重复添加
      // 如果是自己发送的消息，且本地已有该消息（通过sendGroupMessage函数添加），则不再添加
      if (senderId === userStore.userId) {
        // 检查最近的消息是否已经存在相同内容和时间戳的消息
        const recentMessages = groupChats.value[groupId];
        const isDuplicate = recentMessages.some(msg => 
          msg.senderId === senderId && 
          msg.content === content && 
          Math.abs(new Date(msg.timestamp) - new Date(timestamp)) < 1000 // 1秒内的消息视为重复
        );
        
        if (isDuplicate) {
          console.log('跳过重复的群聊消息');
          return;
        }
      }
      
      // 添加消息
      groupChats.value[groupId].push({
        id: Date.now().toString(),
        senderId,
        content,
        timestamp
      })
    }
  }
  
  // 发送私聊消息
  async function sendPrivateMessage(receiverId, content) {
    if (!isConnected.value) {
      return { success: false, message: 'WebSocket未连接' }
    }
    
    try {
      const message = {
        type: 'private',
        message: {
          from: userStore.userId,
          to: receiverId,
          content,
          timestamp: new Date().toISOString()
        }
      }
      
      // 通过WebSocket发送消息
      socket.value.send(JSON.stringify(message))
      
      // 同时通过API保存消息
      await userStore.http.post('/api/messages/private', {
        receiverId,
        content
      })
      
      // 更新本地聊天记录
      if (!privateChats.value[receiverId]) {
        privateChats.value[receiverId] = []
      }
      
      // privateChats.value[receiverId].push({
      //   id: Date.now().toString(),
      //   senderId: userStore.userId,
      //   content,
      //   timestamp: new Date().toISOString()
      // })
      
      return { success: true }
    } catch (error) {
      console.error('发送私聊消息失败:', error)
      return { 
        success: false, 
        message: '发送消息失败，请稍后再试'
      }
    }
  }
  
  // 发送群聊消息
  async function sendGroupMessage(groupId, content) {
    if (!isConnected.value) {
      return { success: false, message: 'WebSocket未连接' }
    }
    
    try {
      const message = {
        type: 'group',
        message: {
          groupId,
          senderId: userStore.userId,
          content,
          timestamp: new Date().toISOString()
        }
      }
      
      // 通过WebSocket发送消息
      socket.value.send(JSON.stringify(message))
      
      // 同时通过API保存消息
      await userStore.http.post('/api/messages/group', {
        groupId,
        content
      })
      
      // 更新本地聊天记录
      if (!groupChats.value[groupId]) {
        groupChats.value[groupId] = []
      }
      
      // groupChats.value[groupId].push({
      //   id: Date.now().toString(),
      //   senderId: userStore.userId,
      //   content,
      //   timestamp: new Date().toISOString()
      // })
      
      return { success: true }
    } catch (error) {
      console.error('发送群聊消息失败:', error)
      return { 
        success: false, 
        message: '发送消息失败，请稍后再试'
      }
    }
  }
  
  // 获取私聊历史消息
  async function fetchPrivateMessages(userId) {
    try {
      const response = await userStore.http.get(`/api/messages/private/${userId}`)
      privateChats.value[userId] = response.data.messages
      // console.log('私聊历史消息:', response.data.messages) // 调试信息，打印完整的响应数据
      return { success: true }
    } catch (error) {
      console.error('获取私聊历史失败:', error)
      return { success: false }
    }
  }
  
  // 获取群聊历史消息
  async function fetchGroupMessages(groupId) {
      // 增强验证逻辑
      if (!groupId) {
          console.error('获取群聊历史失败: 群组ID必须为有效数字', groupId)
          return { 
              success: false, 
              message: '群组ID必须为有效数字',
              receivedId: groupId
          }
      }
      
      try {
          const response = await userStore.http.get(`/api/messages/group/${groupId}`)
          groupChats.value[groupId] = response.data.messages
          // console.log('群聊历史消息:', response.data.messages) // 调试信息，打印完整的响应数据
          return { success: true }
      } catch (error) {
          console.error('获取群聊历史失败:', error)
          return { 
              success: false,
              message: error.response?.data?.error || '获取群聊历史失败'
          }
      }
  }
  
  // 设置当前活跃聊天
  function setActiveChat(type, id) {
    // 添加更严格的ID验证
    if (!id) {
        console.error(`设置活跃聊天失败: ${type}聊天ID无效`, id)
        return { 
            success: false, 
            message: `${type}聊天ID无效`,
            data: { receivedId: id } // 调试信息
        }
    }
    
    // 设置活跃聊天前先检查WebSocket连接状态
    if (!isConnected.value) {
        console.warn('WebSocket未连接，尝试重新连接...')
        initSocket()
    }
    
    activeChat.value = { type, id }
  
    // 仅在ID有效时加载历史消息
    if (type === 'private') {
        return fetchPrivateMessages(id)
    } else {
        return fetchGroupMessages(id)
    }
  }
    
  
  // 清除聊天数据（用于登出）
  function clearChatData() {
    privateChats.value = {}
    groupChats.value = {}
    activeChat.value = null
    
    // 关闭WebSocket连接
    if (socket.value && isConnected.value) {
      socket.value.close()
      socket.value = null
      isConnected.value = false
    }
  }
  
  return {
    // 状态
    privateChats,
    groupChats,
    activeChat,
    isConnected,
    // 计算属性
    currentChatMessages,
    // 方法
    initSocket,
    sendPrivateMessage,
    sendGroupMessage,
    fetchPrivateMessages,
    fetchGroupMessages,
    setActiveChat,
    clearChatData
  }
})
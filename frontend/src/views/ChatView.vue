<template>
  <div class="chat-view" v-if="chatStore.activeChat">
    <!-- 聊天头部 -->
    <div class="chat-header">
      <div class="chat-info">
        <n-avatar
          round
          :size="40"
          :src="currentChatAvatar"
          :fallback-src="defaultAvatar"
        />
        <div class="chat-details">
          <div class="chat-name">{{ currentChatName }}</div>
          <div class="chat-status">{{ currentChatStatus }}</div>
        </div>
      </div>
      <div class="chat-actions">
        <n-button quaternary circle>
          <template #icon>
            <n-icon><search /></n-icon>
          </template>
        </n-button>
        <n-button quaternary circle>
          <template #icon>
            <n-icon><EllipsisVertical /></n-icon>
          </template>
        </n-button>
      </div>
    </div>
    
    <!-- 消息列表 -->
    <div class="message-list" ref="messageListRef">
      <div v-if="chatStore.currentChatMessages.length === 0" class="no-messages">
        <n-empty description="暂无消息" />
        <div class="start-chat-tip">发送消息开始聊天吧</div>
      </div>
      
      <template v-else>
        <div
          v-for="message in chatStore.currentChatMessages"
          :key="message.id"
          class="message-item"
          :class="{ 'message-self': message.senderId === userStore.userId }"
        >
          <n-avatar
            v-if="message.senderId !== userStore.userId"
            round
            :size="36"
            :src="getSenderAvatar(message.senderId)"
            :fallback-src="defaultAvatar"
          />
          <div class="message-content">
            <div class="message-sender">
              {{ message.senderId === userStore.userId ? userStore.username : getSenderName(message.senderId) }}
            </div>
            <div class="message-bubble">
              {{ message.content }}
            </div>
            <div class="message-time">
              {{ formatMessageTime(message.timestamp) }}
            </div>
          </div>
        </div>
      </template>
    </div>
    
    <!-- 消息输入框 -->
    <div class="message-input">
      <n-input
        v-model:value="messageText"
        type="textarea"
        placeholder="输入消息..."
        :autosize="{ minRows: 1, maxRows: 5 }"
        @keydown.enter.prevent="sendMessage"
        :disabled="!chatStore.isConnected"
      />
      <n-button
        type="primary"
        :disabled="!messageText.trim() || !chatStore.isConnected"
        @click="sendMessage"
      >
        发送
      </n-button>
      <div v-if="!chatStore.isConnected" class="connection-warning">
        <n-alert type="warning" title="连接已断开" style="margin-top: 8px">
          正在尝试重新连接，请稍候...
        </n-alert>
      </div>
    </div>
  </div>
  
  <div v-else class="no-chat-selected">
    <n-empty description="选择一个聊天或开始新的对话" />
  </div>
</template>

<script setup>
import { ref, computed, watch, nextTick, onMounted } from 'vue'
import { useMessage } from 'naive-ui'
import { Search, EllipsisVertical } from '@vicons/ionicons5'
import { format } from 'date-fns'
import { useUserStore } from '../stores/user'
import { useChatStore } from '../stores/chat'

// 默认头像
const defaultAvatar = 'https://07akioni.oss-cn-beijing.aliyuncs.com/07akioni.jpeg'

// 消息提示
const message = useMessage()

// Store
const userStore = useUserStore()
const chatStore = useChatStore()

// 消息列表引用（用于滚动到底部）
const messageListRef = ref(null)

// 消息输入
const messageText = ref('')

// 当前是否为群聊
const isGroupChat = computed(() => {
  return chatStore.activeChat?.type === 'group'
})

// 当前聊天的头像
const currentChatAvatar = computed(() => {
  if (!chatStore.activeChat) return ''
  
  if (chatStore.activeChat.type === 'private') {
    const friend = userStore.friends.find(f => f.id === chatStore.activeChat.id)
    return friend?.avatar || ''
  } else {
    const group = userStore.groups.find(g => g.id === chatStore.activeChat.id)
    return group?.avatar || ''
  }
})

// 当前聊天的名称
const currentChatName = computed(() => {
  if (!chatStore.activeChat) return ''
  
  if (chatStore.activeChat.type === 'private') {
    const friend = userStore.friends.find(f => f.id === chatStore.activeChat.id)
    return friend?.nickname || ''
  } else {
    const group = userStore.groups.find(g => g.id === chatStore.activeChat.id)
    return group?.name || ''
  }
})

// 当前聊天的状态
const currentChatStatus = computed(() => {
  if (!chatStore.activeChat) return ''
  
  if (chatStore.activeChat.type === 'private') {
    return '在线' // 这里可以根据实际状态显示
  } else {
    const group = userStore.groups.find(g => g.id === chatStore.activeChat.id)
    return `${group?.memberCount || 0}位成员`
  }
})

// 获取发送者头像
function getSenderAvatar(senderId) {
  const friend = userStore.friends.find(f => f.id === senderId)
  return friend?.avatar || ''
}

// 获取发送者名称
function getSenderName(senderId) {
  const friend = userStore.friends.find(f => f.id === senderId)
  return friend?.username || '未知用户'
}

// 格式化消息时间
function formatMessageTime(timestamp) {
  try {
    const date = new Date(timestamp)
    return format(date, 'HH:mm')
  } catch (error) {
    return ''
  }
}

// 发送消息
async function sendMessage() {
  if (!messageText.value.trim()) return
  
  if (!chatStore.activeChat) {
    message.warning('请先选择聊天对象')
    return
  }
  
  try {
    let result
    
    if (chatStore.activeChat.type === 'private') {
      result = await chatStore.sendPrivateMessage(
        chatStore.activeChat.id,
        messageText.value
      )
    } else {
      result = await chatStore.sendGroupMessage(
        chatStore.activeChat.id,
        messageText.value
      )
    }
    
    if (result.success) {
      messageText.value = ''
      scrollToBottom()
    } else {
      message.error(result.message || '发送失败')
    }
  } catch (error) {
    console.error('Send message error:', error)
    message.error('发送消息时发生错误')
  }
}

// 滚动到底部
function scrollToBottom() {
  nextTick(() => {
    if (messageListRef.value) {
      messageListRef.value.scrollTop = messageListRef.value.scrollHeight
    }
  })
}

// 监听消息列表变化，自动滚动到底部
watch(() => chatStore.currentChatMessages, () => {
  scrollToBottom()
}, { deep: true })

// 监听活跃聊天变化，清空输入框
watch(() => chatStore.activeChat, () => {
  messageText.value = ''
  scrollToBottom()
})

// 组件挂载时滚动到底部
onMounted(() => {
  scrollToBottom()
})
</script>

<style scoped>
.chat-view {
  display: flex;
  flex-direction: column;
  height: 100%;
}

.chat-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  border-bottom: 1px solid #eee;
}

.chat-info {
  display: flex;
  align-items: center;
}

.chat-details {
  margin-left: 12px;
}

.chat-name {
  font-weight: 500;
  font-size: 16px;
}

.chat-status {
  font-size: 12px;
  color: #666;
}

.chat-actions {
  display: flex;
  gap: 8px;
}

.message-list {
  flex: 1;
  overflow-y: auto;
  padding: 16px;
  background-color: #f5f7fa;
}

.no-messages {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100%;
}

.start-chat-tip {
  margin-top: 16px;
  color: #666;
}

.message-item {
  display: flex;
  margin-bottom: 16px;
}

.message-self {
  flex-direction: row-reverse;
}

.message-content {
  margin: 0 12px;
  max-width: 70%;
}

.message-sender {
  font-size: 12px;
  color: #666;
  margin-bottom: 4px;
}

.message-self .message-sender {
  text-align: right;
}

.message-bubble {
  padding: 10px 14px;
  border-radius: 18px;
  background-color: #fff;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.1);
  word-break: break-word;
}

.message-self .message-bubble {
  background-color: #18a058;
  color: white;
}

.message-time {
  font-size: 12px;
  color: #999;
  margin-top: 4px;
  text-align: right;
}

.message-input {
  display: flex;
  align-items: flex-end;
  gap: 12px;
  padding: 12px 16px;
  border-top: 1px solid #eee;
  background-color: #fff;
}

.message-input .n-input {
  flex: 1;
}

.connection-warning {
  width: 100%;
  margin-top: 8px;
}

.no-chat-selected {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100%;
  background-color: #f5f7fa;
}
</style>
<template>
  <div class="group-view" v-if="currentGroup">
    <!-- 聊天头部 -->
    <div class="chat-header">
      <div class="chat-info">
        <n-avatar
          round
          :size="40"
          :src="currentGroup.avatar"
          :fallback-src="defaultGroupAvatar"
        />
        <div class="chat-details">
          <div class="chat-name">{{ currentGroup.name }}</div>
          <div class="chat-status">{{ currentGroup.memberCount }}位成员</div>
        </div>
      </div>
      <div class="chat-actions">
        <n-button quaternary circle @click="showMemberList = true">
          <template #icon>
            <n-icon><people /></n-icon>
          </template>
        </n-button>
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
        <div class="start-chat-tip">发送消息开始群聊吧</div>
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
            <div v-if="message.senderId !== userStore.userId" class="message-sender">
              {{ getSenderName(message.senderId) }}
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
      />
      <n-button
        type="primary"
        :disabled="!messageText.trim()"
        @click="sendMessage"
      >
        发送
      </n-button>
    </div>
    
    <!-- 群成员列表抽屉 -->
    <n-drawer v-model:show="showMemberList" :width="300" placement="right">
      <n-drawer-content title="群成员">
        <div class="member-list">
          <div v-for="member in groupMembers" :key="member.id" class="member-item">
            <n-avatar
              round
              :size="36"
              :src="member.avatar"
              :fallback-src="defaultAvatar"
            />
            <div class="member-info">
              <div class="member-name">{{ member.nickname }}</div>
              <div class="member-username">@{{ member.username }}</div>
            </div>
            <div class="member-role" v-if="member.isOwner">群主</div>
            <div class="member-role" v-else-if="member.isAdmin">管理员</div>
          </div>
          
          <div v-if="groupMembers.length === 0" class="no-members">
            加载群成员中...
          </div>
        </div>
      </n-drawer-content>
    </n-drawer>
  </div>
  
  <div v-else class="no-group-selected">
    <n-empty description="群组不存在或已被删除" />
  </div>
</template>

<script setup>
import { ref, computed, watch, nextTick, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { useMessage } from 'naive-ui'
import { Search, EllipsisVertical, People } from '@vicons/ionicons5'
import { format } from 'date-fns'
import { useUserStore } from '../stores/user'
import { useChatStore } from '../stores/chat'
import http from '../utils/request'

// 默认头像
const defaultAvatar = 'https://cn.bing.com/images/search?view=detailV2&ccid=StrDRqen&id=92D4568BD21B0D03661242697B510D4D295C4727&thid=OIP.StrDRqennoZNbzSPZapKZwAAAA&mediaurl=https%3a%2f%2fimg.shetu66.com%2f2023%2f06%2f26%2f1687770031227597.png&exph=265&expw=474&q=%e9%a3%8e%e6%99%af%e5%9b%be&simid=608050023063974373&FORM=IRPRST&ck=69A4339CE32B1BCBD039D689CE35ACC3&selectedIndex=9&itb=0'
const defaultGroupAvatar = 'https://07akioni.oss-cn-beijing.aliyuncs.com/07akioni.jpeg'

// 路由和消息
const route = useRoute()
const message = useMessage()

// Store
const userStore = useUserStore()
const chatStore = useChatStore()

// 消息列表引用（用于滚动到底部）
const messageListRef = ref(null)

// 消息输入
const messageText = ref('')

// 群成员列表
const showMemberList = ref(false)
const groupMembers = ref([])

// 当前群组
const currentGroup = computed(() => {
  const groupId = route.params.id
  return userStore.groups.find(g => g.id === groupId)
})

// 获取发送者头像
function getSenderAvatar(senderId) {
  const member = groupMembers.value.find(m => m.id === senderId)
  return member?.avatar || defaultAvatar
}

// 获取发送者名称
function getSenderName(senderId) {
  const member = groupMembers.value.find(m => m.id === senderId)
  return member?.nickname || '未知用户'
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
  
  if (!currentGroup.value) {
    message.warning('群组不存在或已被删除')
    return
  }
  
  try {
    const result = await chatStore.sendGroupMessage(
      currentGroup.value.id,
      messageText.value
    )
    
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

// 获取群成员列表
async function fetchGroupMembers() {
  if (!currentGroup.value) return
  
  try {
    const response = await http.get(`/api/groups/${currentGroup.value.id}/members`)
    groupMembers.value = response.data
  } catch (error) {
    console.error('Fetch group members error:', error)
    message.error('获取群成员列表失败')
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

// 监听群组变化，获取群成员
watch(() => route.params.id, (newId) => {
  // console.log('路由群组ID变化:', newId)
  if (currentGroup.value) {
    // console.log('当前群组ID:', currentGroup.value.id)
    chatStore.setActiveChat('group', currentGroup.value.id)
    fetchGroupMembers()
  } else {
    console.warn('currentGroup为空，可能原因:', {
      routeId: route.params.id,
      userGroups: userStore.groups
    })
  }
}, { immediate: true })

// 监听成员列表抽屉打开
watch(() => showMemberList.value, (newVal) => {
  if (newVal) {
    fetchGroupMembers()
  }
})

// 组件挂载时滚动到底部
onMounted(() => {
  scrollToBottom()
})
</script>

<style scoped>
.group-view {
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

.no-group-selected {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100%;
  background-color: #f5f7fa;
}

.member-list {
  padding: 8px 0;
}

.member-item {
  display: flex;
  align-items: center;
  padding: 12px 16px;
  border-bottom: 1px solid #f3f3f3;
}

.member-info {
  margin-left: 12px;
  flex: 1;
}

.member-name {
  font-weight: 500;
}

.member-username {
  font-size: 12px;
  color: #666;
}

.member-role {
  font-size: 12px;
  color: #18a058;
  font-weight: 500;
}

.no-members {
  padding: 24px 16px;
  text-align: center;
  color: #999;
}
</style>
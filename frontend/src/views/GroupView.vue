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
        <n-button quaternary circle @click="showSearchModal = true">
          <template #icon>
            <n-icon><search /></n-icon>
          </template>
        </n-button>
        <n-dropdown :options="groupMenuOptions" @select="handleGroupMenuSelect">
          <n-button quaternary circle>
            <template #icon>
              <n-icon><EllipsisVertical /></n-icon>
            </template>
          </n-button>
        </n-dropdown>
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
              <!-- <div class="member-name">{{ member.nickname }}</div> -->
              <div class="member-username">@{{ member.username }}</div>
            </div>

             <!-- 在线信息 -->
            <div class="member-status" v-if="member.status=='online'">
              <n-tag type="success" size="small">在线</n-tag>
            </div>
            <div class="member-status" v-else>
              <n-tag type="error" size="small">离线</n-tag>
            </div>

            <!-- 角色信息 -->
            <div class="member-role" v-if="member.role=='admin'">群主</div>
            <div class="member-role" v-else-if="member.role!=='admin'">成员</div>
          </div>
          
          <div v-if="groupMembers.length === 0" class="no-members">
            加载群成员中...
          </div>
        </div>
      </n-drawer-content>
    </n-drawer>
    
    <!-- 搜索消息对话框 -->
    <n-modal
      v-model:show="showSearchModal"
      title="搜索消息"
      preset="dialog"
      :show-icon="false"
    >
      <div class="search-container">
        <n-input
          v-model:value="searchQuery"
          placeholder="输入关键词搜索消息"
          clearable
          @keydown.enter="searchMessages"
        >
          <template #suffix>
            <n-button quaternary circle @click="searchMessages">
              <template #icon>
                <n-icon><search /></n-icon>
              </template>
            </n-button>
          </template>
        </n-input>
        
        <div class="search-results" v-if="searchResults.length > 0">
          <div v-for="result in searchResults" :key="result.id" class="search-result-item">
            <div class="search-result-sender">{{ getSenderName(result.senderId) }}</div>
            <div class="search-result-content">{{ result.content }}</div>
            <div class="search-result-time">{{ formatMessageTime(result.timestamp) }}</div>
          </div>
        </div>
        
        <div v-else-if="hasSearched" class="no-search-results">
          没有找到匹配的消息
        </div>
      </div>
    </n-modal>
    
    <!-- 邀请好友对话框 -->
    <n-modal
      v-model:show="showInviteModal"
      title="邀请好友加入群组"
      preset="dialog"
      positive-text="邀请"
      negative-text="取消"
      @positive-click="handleInviteFriend"
    >
      <n-form
        ref="inviteFormRef"
        :model="inviteForm"
        :rules="inviteRules"
        label-placement="left"
        label-width="80"
      >
        <n-form-item path="username" label="好友用户名">
          <n-select
            v-model:value="inviteForm.username"
            placeholder="选择要邀请的好友"
            :options="friendOptions"
            filterable
          />
        </n-form-item>
      </n-form>
    </n-modal>
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
import avatar from '@/assets/defaultAvatar.jpg'

// 默认头像
const defaultAvatar = avatar
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

// 搜索消息
const showSearchModal = ref(false)
const searchQuery = ref('')
const searchResults = ref([])
const hasSearched = ref(false)

// 邀请好友
const showInviteModal = ref(false)
const inviteFormRef = ref(null)
const inviteForm = ref({
  username: null
})
const inviteRules = {
  username: [
    { required: true, message: '请选择要邀请的好友', trigger: 'blur' }
  ]
}

// 群组菜单选项
const groupMenuOptions = [
  {
    label: '邀请好友',
    key: 'invite'
  },
  {
    label: '群组信息',
    key: 'info'
  }
]

// 当前群组
const currentGroup = computed(() => {
  const groupId = route.params.id
  return userStore.groups.find(g => g.id === groupId)
})

// 可邀请的好友选项
const friendOptions = computed(() => {
  // 获取当前群组成员的ID列表
  const memberIds = groupMembers.value.map(member => member.id)
  
  // 过滤出不在群组中的好友
  return userStore.friends
    .filter(friend => !memberIds.includes(friend.id))
    .map(friend => ({
      label: `${friend.nickname || friend.username} (@${friend.username})`,
      value: friend.username
    }))
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
    groupMembers.value = response.data.members
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

// 搜索消息
function searchMessages() {
  if (!searchQuery.value.trim()) return
  
  hasSearched.value = true
  const query = searchQuery.value.toLowerCase()
  
  // 在当前群聊消息中搜索
  searchResults.value = chatStore.currentChatMessages.filter(message => 
    message.content.toLowerCase().includes(query)
  )
}

// 处理群组菜单选择
function handleGroupMenuSelect(key) {
  switch (key) {
    case 'invite':
      showInviteModal.value = true
      break
    case 'info':
      message.info('群组信息功能正在开发中')
      break
  }
}

// 邀请好友加入群组
async function handleInviteFriend() {
  try {
    await inviteFormRef.value?.validate()
    
    if (!currentGroup.value) {
      message.error('群组不存在或已被删除')
      return
    }
    
    // 调用API邀请好友
    const response = await http.post(`/api/groups/${currentGroup.value.id}/members`, {
      username: inviteForm.value.username,
      role: 'member'
    })
    
    message.success('邀请好友成功')
    showInviteModal.value = false
    inviteForm.value.username = null
    
    // 刷新群成员列表
    fetchGroupMembers()
  } catch (error) {
    console.error('邀请好友失败:', error)
    message.error(error.response?.data?.error || '邀请好友失败')
  }
}
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

.member-status {
  margin-left: 8px;
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

/* 搜索消息样式 */
.search-container {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.search-results {
  max-height: 300px;
  overflow-y: auto;
  border: 1px solid #eee;
  border-radius: 8px;
}

.search-result-item {
  padding: 12px;
  border-bottom: 1px solid #f3f3f3;
}

.search-result-item:last-child {
  border-bottom: none;
}

.search-result-sender {
  font-size: 12px;
  color: #666;
  margin-bottom: 4px;
}

.search-result-content {
  font-size: 14px;
  margin-bottom: 4px;
}

.search-result-time {
  font-size: 12px;
  color: #999;
  text-align: right;
}

.no-search-results {
  padding: 24px 16px;
  text-align: center;
  color: #999;
}
</style>
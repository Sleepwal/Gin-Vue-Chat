<template>
  <div class="main-layout">
    <!-- 侧边栏 -->
    <div class="sidebar">
      <!-- 用户信息 -->
      <div class="user-info">
        <n-avatar
          round
          :size="40"
          :src="userStore.user?.avatar"
          :fallback-src="defaultAvatar"
        />
        <div class="user-details">
          <div class="username">{{ userStore.username }}</div>
          <div class="status">在线</div>
        </div>
        <div class="user-actions">
          <n-dropdown :options="userMenuOptions" @select="handleUserMenuSelect">
            <n-button quaternary circle>
              <template #icon>
                <n-icon><EllipsisVertical /></n-icon>
              </template>
            </n-button>
          </n-dropdown>
        </div>
      </div>
      
      <!-- 搜索框 -->
      <div class="search-box">
        <n-input
          v-model:value="searchQuery"
          placeholder="搜索"
          clearable
        >
          <template #prefix>
            <n-icon><search /></n-icon>
          </template>
        </n-input>
      </div>
      
      <!-- 标签页：聊天/好友/群组 -->
      <n-tabs
        v-model:value="activeTab"
        type="line"
        animated
        size="large"
        justify-content="space-evenly"
      >
        <n-tab-pane name="chats" tab="聊天">
          <!-- 聊天列表 -->
          <div class="chat-list">
            <div
              v-for="friend in filteredFriends"
              :key="friend.id"
              class="chat-item"
              :class="{ active: isActiveFriend(friend.id) }"
              @click="openPrivateChat(friend.id)"
            >
              <n-avatar
                round
                :size="40"
                :src="friend.avatar"
                :fallback-src="defaultAvatar"
              />
              <div class="chat-info">
                <div class="chat-name">{{ friend.nickname }}</div>
                <div class="chat-last-message">{{ getLastMessage('private', friend.id) }}</div>
              </div>
            </div>
            
            <div
              v-for="group in filteredGroups"
              :key="group.id"
              class="chat-item"
              :class="{ active: isActiveGroup(group.id) }"
              @click="openGroupChat(group.id)"
            >
              <n-avatar
                round
                :size="40"
                :src="group.avatar"
                :fallback-src="defaultAvatar"
              />
              <div class="chat-info">
                <div class="chat-name">{{ group.name }}</div>
                <div class="chat-last-message">{{ getLastMessage('group', group.id) }}</div>
              </div>
            </div>
            
            <div v-if="noChatsFound" class="no-chats">
              没有找到聊天记录
            </div>
          </div>
        </n-tab-pane>
        
        <n-tab-pane name="friends" tab="好友">
          <!-- 好友列表 -->
          <div class="friend-list">
            <div class="action-button">
              <n-button
                block
                @click="showAddFriendModal = true"
              >
                添加好友
              </n-button>
            </div>
            
            <div
              v-for="friend in filteredFriends"
              :key="friend.id"
              class="friend-item"
              @click="openPrivateChat(friend.id)"
            >
              <n-avatar
                round
                :size="40"
                :src="friend.avatar"
                :fallback-src="defaultAvatar"
              />
              <div class="friend-info">
                <div class="friend-name">{{ friend.nickname }}</div>
                <div class="friend-username">@{{ friend.username }}</div>
              </div>
            </div>
            
            <div v-if="filteredFriends.length === 0" class="no-friends">
              {{ searchQuery ? '没有找到匹配的好友' : '暂无好友，点击上方按钮添加' }}
            </div>
          </div>
        </n-tab-pane>
        
        <n-tab-pane name="groups" tab="群组">
          <!-- 群组列表 -->
          <div class="group-list">
            <div class="action-button">
              <n-button
                block
                @click="showCreateGroupModal = true"
              >
                创建群组
              </n-button>
            </div>
            
            <div
              v-for="group in filteredGroups"
              :key="group.id"
              class="group-item"
              @click="openGroupChat(group.id)"
            >
              <n-avatar
                round
                :size="40"
                :src="group.avatar"
                :fallback-src="defaultGroupAvatar"
              />
              <div class="group-info">
                <div class="group-name">{{ group.name }}</div>
                <div class="group-members">{{ group.memberCount }}位成员</div>
              </div>
            </div>
            
            <div v-if="filteredGroups.length === 0" class="no-groups">
              {{ searchQuery ? '没有找到匹配的群组' : '暂无群组，点击上方按钮创建' }}
            </div>
          </div>
        </n-tab-pane>
      </n-tabs>
    </div>
    
    <!-- 聊天区域 -->
    <div class="chat-area">
      <router-view />
    </div>
    
    <!-- 添加好友对话框 -->
    <n-modal
      v-model:show="showAddFriendModal"
      title="添加好友"
      preset="dialog"
      positive-text="添加"
      negative-text="取消"
      @positive-click="handleAddFriend"
    >
      <n-form
        ref="addFriendFormRef"
        :model="addFriendForm"
        :rules="addFriendRules"
        label-placement="left"
        label-width="80"
      >
        <n-form-item path="username" label="用户名">
          <n-input v-model:value="addFriendForm.username" placeholder="请输入用户名" />
        </n-form-item>
      </n-form>
    </n-modal>
    
    <!-- 创建群组对话框 -->
    <n-modal
      v-model:show="showCreateGroupModal"
      title="创建群组"
      preset="dialog"
      positive-text="创建"
      negative-text="取消"
      @positive-click="handleCreateGroup"
    >
      <n-form
        ref="createGroupFormRef"
        :model="createGroupForm"
        :rules="createGroupRules"
        label-placement="left"
        label-width="80"
      >
        <n-form-item path="name" label="群组名称">
          <n-input v-model:value="createGroupForm.name" placeholder="请输入群组名称" />
        </n-form-item>
        <n-form-item path="description" label="群组描述">
          <n-input
            v-model:value="createGroupForm.description"
            type="textarea"
            placeholder="请输入群组描述"
          />
        </n-form-item>
      </n-form>
    </n-modal>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useMessage } from 'naive-ui'
import { Search, EllipsisVertical } from '@vicons/ionicons5'
import { useUserStore } from '../stores/user'
import { useChatStore } from '../stores/chat'

// 默认头像
const defaultAvatar = 'https://07akioni.oss-cn-beijing.aliyuncs.com/07akioni.jpeg'
const defaultGroupAvatar = 'https://07akioni.oss-cn-beijing.aliyuncs.com/07akioni.jpeg'

// 路由和消息
const router = useRouter()
const message = useMessage()

// Store
const userStore = useUserStore()
const chatStore = useChatStore()

// 侧边栏状态
const activeTab = ref('chats')
const searchQuery = ref('')

// 添加好友表单
const showAddFriendModal = ref(false)
const addFriendFormRef = ref(null)
const addFriendForm = ref({
  username: ''
})
const addFriendRules = {
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' }
  ]
}

// 创建群组表单
const showCreateGroupModal = ref(false)
const createGroupFormRef = ref(null)
const createGroupForm = ref({
  name: '',
  description: ''
})
const createGroupRules = {
  name: [
    { required: true, message: '请输入群组名称', trigger: 'blur' },
    { max: 30, message: '群组名称不能超过30个字符', trigger: 'blur' }
  ],
  description: [
    { max: 200, message: '群组描述不能超过200个字符', trigger: 'blur' }
  ]
}

// 用户菜单选项
const userMenuOptions = [
  {
    label: '个人资料',
    key: 'profile'
  },
  {
    label: '设置',
    key: 'settings'
  },
  {
    label: '退出登录',
    key: 'logout'
  }
]

// 过滤后的好友列表
const filteredFriends = computed(() => {
  if (!searchQuery.value) return userStore.friends
  
  const query = searchQuery.value.toLowerCase()
  return userStore.friends.filter(friend => 
    friend.nickname.toLowerCase().includes(query) ||
    friend.username.toLowerCase().includes(query)
  )
})

// 过滤后的群组列表
const filteredGroups = computed(() => {
  // console.log("groups: " + JSON.stringify(userStore.groups))
  if (!searchQuery.value) return userStore.groups
  
  const query = searchQuery.value.toLowerCase()
  return userStore.groups.filter(group => 
    group.name.toLowerCase().includes(query)
  )
})

// 是否没有找到聊天
const noChatsFound = computed(() => {
  return filteredFriends.value.length === 0 && filteredGroups.value.length === 0
})

// 检查是否是当前活跃的私聊
function isActiveFriend(friendId) {
  return chatStore.activeChat?.type === 'private' && chatStore.activeChat?.id === friendId
}

// 检查是否是当前活跃的群聊
function isActiveGroup(groupId) {
  return chatStore.activeChat?.type === 'group' && chatStore.activeChat?.id === groupId
}

// 获取最后一条消息
function getLastMessage(type, id) {
  if (type === 'private') {
    const messages = chatStore.privateChats[id] || []
    if (messages.length === 0) return '暂无消息'
    
    const lastMessage = messages[messages.length - 1]
    // 添加对lastMessage是否存在以及是否有content属性的检查
    if (!lastMessage || !lastMessage.content) return '暂无消息内容'
    
    return lastMessage.content.length > 20
      ? lastMessage.content.substring(0, 20) + '...'
      : lastMessage.content
  } else {
    const messages = chatStore.groupChats[id] || []
    if (messages.length === 0) return '暂无消息'
    
    const lastMessage = messages[messages.length - 1]
    // 添加对lastMessage是否存在以及是否有content属性的检查
    if (!lastMessage || !lastMessage.content) return '暂无消息内容'
    
    return lastMessage.content.length > 20
      ? lastMessage.content.substring(0, 20) + '...'
      : lastMessage.content
  }
}

// 打开私聊
function openPrivateChat(friendId) {
  chatStore.setActiveChat('private', friendId)
  router.push('/')
}

// 打开群聊
function openGroupChat(groupId) {
  chatStore.setActiveChat('group', groupId)
  router.push(`/group/${groupId}`)
}

// 处理用户菜单选择
function handleUserMenuSelect(key) {
  switch (key) {
    case 'profile':
      router.push('/profile')
      break
    case 'settings':
      // 暂未实现设置页面
      message.info('设置功能正在开发中')
      break
    case 'logout':
      userStore.logout()
      chatStore.clearChatData()
      router.push('/login')
      message.success('已退出登录')
      break
  }
}

// 处理添加好友
async function handleAddFriend() {
  try {
    await addFriendFormRef.value?.validate()
    
    const result = await userStore.addFriend(addFriendForm.value.username)
    
    if (result.success) {
      message.success('添加好友成功')
      addFriendForm.value.username = ''
      return true
    } else {
      message.error(result.message || '添加好友失败')
      return false
    }
  } catch (error) {
    console.error('Add friend error:', error)
    return false
  }
}

// 处理创建群组
async function handleCreateGroup() {
  try {
    await createGroupFormRef.value?.validate()
    
    const result = await userStore.createGroup(createGroupForm.value)
    
    if (result.success) {
      message.success('创建群组成功')
      createGroupForm.value.name = ''
      createGroupForm.value.description = ''
      return true
    } else {
      message.error(result.message || '创建群组失败')
      return false
    }
  } catch (error) {
    console.error('Create group error:', error)
    return false
  }
}

// 组件挂载时初始化数据
onMounted(async () => {
  // 初始化WebSocket连接
  chatStore.initSocket()
  
  // 获取好友列表
  await userStore.fetchFriends()
  
  // 获取群组列表
  await userStore.fetchGroups()
  
  // 监听WebSocket连接状态
  if (!chatStore.isConnected) {
    console.log('正在尝试重新连接WebSocket...')
    // 如果初始连接失败，尝试再次连接
    setTimeout(() => {
      if (!chatStore.isConnected) {
        chatStore.initSocket()
      }
    }, 2000)
  }
})

// 监听路由变化，更新活跃聊天
watch(() => router.currentRoute.value, (route) => {
  if (route.name === 'group' && route.params.id) {
    chatStore.setActiveChat('group', route.params.id)
  }
}, { immediate: true })
</script>

<style scoped>
.main-layout {
  display: flex;
  height: 100vh;
  overflow: hidden;
}

.sidebar {
  width: 320px;
  display: flex;
  flex-direction: column;
  border-right: 1px solid #eee;
  background-color: #f7f7f7;
}

.user-info {
  display: flex;
  align-items: center;
  padding: 16px;
  border-bottom: 1px solid #eee;
}

.user-details {
  margin-left: 12px;
  flex: 1;
}

.username {
  font-weight: 500;
  font-size: 16px;
}

.status {
  font-size: 12px;
  color: #18a058;
}

.search-box {
  padding: 12px 16px;
  border-bottom: 1px solid #eee;
}

.chat-list,
.friend-list,
.group-list {
  flex: 1;
  overflow-y: auto;
  padding: 8px 0;
}

.chat-item,
.friend-item,
.group-item {
  display: flex;
  align-items: center;
  padding: 12px 16px;
  cursor: pointer;
  transition: background-color 0.2s;
}

.chat-item:hover,
.friend-item:hover,
.group-item:hover {
  background-color: #e9e9e9;
}

.chat-item.active {
  background-color: #e3f5e9;
}

.chat-info,
.friend-info,
.group-info {
  margin-left: 12px;
  flex: 1;
  overflow: hidden;
}

.chat-name,
.friend-name,
.group-name {
  font-weight: 500;
  margin-bottom: 4px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.chat-last-message,
.friend-username,
.group-members {
  font-size: 13px;
  color: #666;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.action-button {
  padding: 12px 16px;
  border-bottom: 1px solid #eee;
}

.no-chats,
.no-friends,
.no-groups {
  padding: 24px 16px;
  text-align: center;
  color: #999;
}

.chat-area {
  flex: 1;
  display: flex;
  flex-direction: column;
  background-color: #fff;
}
</style>
import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import http from '../utils/request'

export const useUserStore = defineStore('user', () => {
  // 状态
  const token = ref(localStorage.getItem('token') || '')
  const user = ref(JSON.parse(localStorage.getItem('user') || 'null'))
  const friends = ref([])
  const groups = ref([])
  
  // 计算属性
  const isLoggedIn = computed(() => !!token.value)
  const userId = computed(() => user.value?.id || null)
  const username = computed(() => user.value?.username || '')
  
  // 方法
  // 登录
  async function login(credentials) {
    try {
      // 登录请求不需要携带token
      const response = await http.post('/api/auth/login', credentials, { withToken: false })
      const { token: newToken, user: userData } = response.data
      
      // 保存到状态和本地存储
      token.value = newToken
      user.value = userData
      localStorage.setItem('token', newToken)
      localStorage.setItem('user', JSON.stringify(userData))
      
      return { success: true }
    } catch (error) {
      console.error('Login error:', error)
      return { 
        success: false, 
        message: error.response?.data?.message || '登录失败，请稍后再试'
      }
    }
  }
  
  // 注册
  async function register(userData) {
    try {
      // 注册请求不需要携带token
      const response = await http.post('/api/auth/register', userData, { withToken: false })
      return { success: true, data: response.data }
    } catch (error) {
      console.error('Register error:', error)
      return { 
        success: false, 
        message: error.response?.data?.message || '注册失败，请稍后再试'
      }
    }
  }
  
  // 登出
  function logout() {
    // 清除状态
    token.value = ''
    user.value = null
    friends.value = []
    groups.value = []
    
    // 清除本地存储
    localStorage.removeItem('token')
    localStorage.removeItem('user')
  }
  
  // 获取好友列表
  async function fetchFriends() {
    try {
      const response = await http.get('/api/friends')
      friends.value = response.data.friends
      return { success: true }
    } catch (error) {
      console.error('Fetch friends error:', error)
      return { success: false }
    }
  }
  
  // 获取群组列表
  async function fetchGroups() {
    try {
      const response = await http.get('/api/groups')
      groups.value = response.data.groups
      // console.log("groups: " + groups.value.groups[0].name)
      return { success: true }
    } catch (error) {
      console.error('Fetch groups error:', error)
      return { success: false }
    }
  }
  
  // 添加好友
  async function addFriend(friendId) {
    try {
      const response = await http.post('/api/friends/add', { friendId })
      await fetchFriends() // 重新获取好友列表
      return { success: true, data: response.data }
    } catch (error) {
      console.error('Add friend error:', error)
      return { 
        success: false, 
        message: error.response?.data?.message || '添加好友失败'
      }
    }
  }
  
  // 创建群组
  async function createGroup(groupData) {
    try {
      const response = await http.post('/api/groups/create', groupData)
      await fetchGroups() // 重新获取群组列表
      return { success: true, data: response.data }
    } catch (error) {
      console.error('Create group error:', error)
      return { 
        success: false, 
        message: error.response?.data?.message || '创建群组失败'
      }
    }
  }
  
  return { 
    // 状态
    token, 
    user, 
    friends,
    groups,
    // 计算属性
    isLoggedIn,
    userId,
    username,
    // 方法
    login,
    register,
    logout,
    fetchFriends,
    fetchGroups,
    addFriend,
    createGroup,
    // 导出请求对象
    http
  }
})
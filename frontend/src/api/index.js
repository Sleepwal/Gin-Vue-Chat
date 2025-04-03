import http from '../utils/request'

// 用户相关API
export const userApi = {
  // 登录
  login: (credentials) => http.post('/api/auth/login', credentials, { withToken: false }),
  // 注册
  register: (userData) => http.post('/api/auth/register', userData, { withToken: false }),
  // 获取用户资料
  getProfile: () => http.get('/api/user/profile'),
  // 更新用户资料
  updateProfile: (data) => http.put('/api/user/profile', data),
  // 修改密码
  changePassword: (data) => http.put('/api/user/password', data)
}

// 好友相关API
export const friendApi = {
  // 获取好友列表
  getFriends: () => http.get('/api/friends'),
  // 添加好友
  addFriend: (username) => http.post('/api/friends/add', { username }),
  // 删除好友
  removeFriend: (friendId) => http.delete(`/api/friends/${friendId}`)
}

// 群组相关API
export const groupApi = {
  // 获取群组列表
  getGroups: () => http.get('/api/groups'),
  // 创建群组
  createGroup: (groupData) => http.post('/api/groups/create', groupData),
  // 获取群组详情
  getGroupDetail: (groupId) => http.get(`/api/groups/${groupId}`),
  // 更新群组
  updateGroup: (groupId, data) => http.put(`/api/groups/${groupId}`, data),
  // 删除群组
  deleteGroup: (groupId) => http.delete(`/api/groups/${groupId}`),
  // 获取群组成员
  getGroupMembers: (groupId) => http.get(`/api/groups/${groupId}/members`),
  // 添加群组成员
  addGroupMember: (groupId, userId) => http.post(`/api/groups/${groupId}/members`, { userId }),
  // 移除群组成员
  removeGroupMember: (groupId, userId) => http.delete(`/api/groups/${groupId}/members/${userId}`)
}

// 消息相关API
export const messageApi = {
  // 获取私聊消息
  getPrivateMessages: (userId) => http.get(`/api/messages/private/${userId}`),
  // 发送私聊消息
  sendPrivateMessage: (receiverId, content) => http.post('/api/messages/private', { receiverId, content }),
  // 获取群聊消息
  getGroupMessages: (groupId) => http.get(`/api/messages/group/${groupId}`),
  // 发送群聊消息
  sendGroupMessage: (groupId, content) => http.post('/api/messages/group', { groupId, content })
}

// 导出所有API
export default {
  user: userApi,
  friend: friendApi,
  group: groupApi,
  message: messageApi
}
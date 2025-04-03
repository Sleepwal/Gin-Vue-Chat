<template>
  <div class="profile-view">
    <div class="profile-header">
      <h2>个人资料</h2>
    </div>
    
    <div class="profile-content">
      <n-card>
        <n-form
          ref="formRef"
          :model="formValue"
          :rules="rules"
          label-placement="left"
          label-width="100"
          require-mark-placement="right-hanging"
        >
          <n-form-item label="头像">
            <div class="avatar-upload">
              <n-avatar
                round
                :size="80"
                :src="formValue.avatar"
                :fallback-src="defaultAvatar"
              />
              <n-button class="upload-btn" @click="handleAvatarUpload">
                更换头像
              </n-button>
            </div>
          </n-form-item>
          
          <n-form-item path="username" label="用户名">
            <n-input
              v-model:value="formValue.username"
              placeholder="用户名"
              disabled
            />
            <template #help>
              用户名不可修改
            </template>
          </n-form-item>
          
          <n-form-item path="nickname" label="昵称">
            <n-input
              v-model:value="formValue.nickname"
              placeholder="请输入昵称"
            />
          </n-form-item>
          
          <n-form-item path="email" label="邮箱">
            <n-input
              v-model:value="formValue.email"
              placeholder="请输入邮箱"
            />
          </n-form-item>
          
          <n-form-item label="修改密码">
            <n-button @click="showPasswordModal = true">
              修改密码
            </n-button>
          </n-form-item>
          
          <div class="form-actions">
            <n-button
              type="primary"
              @click="handleSubmit"
              :loading="loading"
            >
              保存修改
            </n-button>
          </div>
        </n-form>
      </n-card>
    </div>
    
    <!-- 修改密码对话框 -->
    <n-modal
      v-model:show="showPasswordModal"
      title="修改密码"
      preset="dialog"
      positive-text="确认修改"
      negative-text="取消"
      @positive-click="handlePasswordChange"
    >
      <n-form
        ref="passwordFormRef"
        :model="passwordForm"
        :rules="passwordRules"
        label-placement="left"
        label-width="100"
      >
        <n-form-item path="oldPassword" label="当前密码">
          <n-input
            v-model:value="passwordForm.oldPassword"
            type="password"
            placeholder="请输入当前密码"
            show-password-on="click"
          />
        </n-form-item>
        
        <n-form-item path="newPassword" label="新密码">
          <n-input
            v-model:value="passwordForm.newPassword"
            type="password"
            placeholder="请输入新密码"
            show-password-on="click"
          />
        </n-form-item>
        
        <n-form-item path="confirmPassword" label="确认新密码">
          <n-input
            v-model:value="passwordForm.confirmPassword"
            type="password"
            placeholder="请再次输入新密码"
            show-password-on="click"
          />
        </n-form-item>
      </n-form>
    </n-modal>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useMessage } from 'naive-ui'
import { useUserStore } from '../stores/user'

// 默认头像
const defaultAvatar = 'https://07akioni.oss-cn-beijing.aliyuncs.com/07akioni.jpeg'

// 消息提示
const message = useMessage()

// Store
const userStore = useUserStore()

// 表单引用
const formRef = ref(null)
const passwordFormRef = ref(null)

// 加载状态
const loading = ref(false)

// 表单数据
const formValue = ref({
  username: '',
  nickname: '',
  email: '',
  avatar: ''
})

// 表单验证规则
const rules = {
  nickname: [
    { required: true, message: '请输入昵称', trigger: 'blur' },
    { max: 20, message: '昵称长度不能超过20个字符', trigger: 'blur' }
  ],
  email: [
    { type: 'email', message: '请输入有效的邮箱地址', trigger: 'blur' }
  ]
}

// 修改密码相关
const showPasswordModal = ref(false)
const passwordForm = ref({
  oldPassword: '',
  newPassword: '',
  confirmPassword: ''
})

// 密码表单验证规则
const passwordRules = {
  oldPassword: [
    { required: true, message: '请输入当前密码', trigger: 'blur' }
  ],
  newPassword: [
    { required: true, message: '请输入新密码', trigger: 'blur' },
    { min: 6, message: '密码长度不能少于6个字符', trigger: 'blur' }
  ],
  confirmPassword: [
    { required: true, message: '请确认新密码', trigger: 'blur' },
    {
      validator: (rule, value) => {
        return value === passwordForm.value.newPassword
      },
      message: '两次输入的密码不一致',
      trigger: ['blur', 'input']
    }
  ]
}

// 初始化表单数据
function initFormData() {
  if (userStore.user) {
    formValue.value = {
      username: userStore.user.username || '',
      nickname: userStore.user.nickname || '',
      email: userStore.user.email || '',
      avatar: userStore.user.avatar || ''
    }
  }
}

// 处理头像上传
function handleAvatarUpload() {
  message.info('头像上传功能正在开发中')
  // 这里应该实现文件上传功能
}

// 提交表单
async function handleSubmit() {
  await formRef.value?.validate()
  
  loading.value = true
  
  try {
    const response = await userStore.http.put('/api/user/profile', {
      nickname: formValue.value.nickname,
      email: formValue.value.email
    })
    
    // 更新本地用户数据
    const updatedUser = {
      ...userStore.user,
      nickname: formValue.value.nickname,
      email: formValue.value.email
    }
    
    localStorage.setItem('user', JSON.stringify(updatedUser))
    userStore.user = updatedUser
    
    message.success('个人资料已更新')
  } catch (error) {
    console.error('Update profile error:', error)
    message.error('更新个人资料失败')
  } finally {
    loading.value = false
  }
}

// 处理密码修改
async function handlePasswordChange() {
  try {
    await passwordFormRef.value?.validate()
    
    await userStore.http.put('/api/user/password', {
      oldPassword: passwordForm.value.oldPassword,
      newPassword: passwordForm.value.newPassword
    })
    
    message.success('密码已修改，请重新登录')
    
    // 清空表单
    passwordForm.value = {
      oldPassword: '',
      newPassword: '',
      confirmPassword: ''
    }
    
    // 登出用户
    userStore.logout()
    router.push('/login')
    
    return true
  } catch (error) {
    console.error('Change password error:', error)
    message.error(error.response?.data?.message || '修改密码失败')
    return false
  }
}

// 组件挂载时初始化数据
onMounted(() => {
  initFormData()
})
</script>

<style scoped>
.profile-view {
  padding: 24px;
  height: 100%;
  overflow-y: auto;
}

.profile-header {
  margin-bottom: 24px;
}

.avatar-upload {
  display: flex;
  align-items: center;
  gap: 16px;
}

.form-actions {
  margin-top: 24px;
  display: flex;
  justify-content: flex-end;
}
</style>
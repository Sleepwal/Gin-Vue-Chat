<template>
  <div class="register-container">
    <div class="register-card">
      <div class="register-header">
        <h1>Gin-Vue-Chat</h1>
        <p>创建新账号</p>
      </div>
      
      <n-form
        ref="formRef"
        :model="formValue"
        :rules="rules"
        label-placement="left"
        label-width="85"
        require-mark-placement="right-hanging"
        size="large"
        @submit.prevent="handleSubmit"
      >
        <n-form-item path="username" label="用户名">
          <n-input v-model:value="formValue.username" placeholder="请输入用户名" />
        </n-form-item>
        
        <n-form-item path="email" label="邮箱">
          <n-input v-model:value="formValue.email" placeholder="请输入邮箱" />
        </n-form-item>
        
        <n-form-item path="password" label="密码">
          <n-input
            v-model:value="formValue.password"
            type="password"
            placeholder="请输入密码"
            show-password-on="click"
          />
        </n-form-item>
        
        <n-form-item path="confirmPassword" label="确认密码">
          <n-input
            v-model:value="formValue.confirmPassword"
            type="password"
            placeholder="请再次输入密码"
            show-password-on="click"
          />
        </n-form-item>
        
        <div class="action-btns">
          <n-button
            type="primary"
            attr-type="submit"
            :loading="loading"
            block
          >
            注册
          </n-button>
          
          <div class="login-link">
            已有账号？
            <router-link to="/login">返回登录</router-link>
          </div>
        </div>
      </n-form>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useMessage } from 'naive-ui'
import { useUserStore } from '../stores/user'

const router = useRouter()
const message = useMessage()
const userStore = useUserStore()

// 表单引用
const formRef = ref(null)

// 表单数据
const formValue = ref({
  username: '',
  email: '',
  password: '',
  confirmPassword: ''
})

// 表单验证规则
const rules = {
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' },
    { min: 3, max: 20, message: '用户名长度应在3-20个字符之间', trigger: 'blur' }
  ],
  email: [
    { required: true, message: '请输入邮箱', trigger: 'blur' },
    { type: 'email', message: '请输入有效的邮箱地址', trigger: 'blur' }
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 6, message: '密码长度不能少于6个字符', trigger: 'blur' }
  ],
  confirmPassword: [
    { required: true, message: '请确认密码', trigger: 'blur' },
    {
      validator: (rule, value) => {
        return value === formValue.value.password
      },
      message: '两次输入的密码不一致',
      trigger: ['blur', 'input']
    }
  ]
}

// 加载状态
const loading = ref(false)

// 提交表单
async function handleSubmit() {
  await formRef.value?.validate()
  
  loading.value = true
  
  try {
    // 准备注册数据，移除确认密码字段
    const registerData = {
      username: formValue.value.username,
      email: formValue.value.email,
      password: formValue.value.password
    }
    
    const result = await userStore.register(registerData)
    
    if (result.success) {
      message.success('注册成功，请登录')
      router.push('/login')
    } else {
      message.error(result.error || '注册失败')
    }
  } catch (error) {
    console.error('Register error:', error)
    message.error('注册过程中发生错误')
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.register-container {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100vh;
  background-image: url('@/assets/auth-background.svg');
  background-size: cover;
  background-position: center;
  background-repeat: no-repeat;
}

.register-card {
  width: 430px;
  padding: 40px;
  background-color: rgba(255, 255, 255, 0.2);
  backdrop-filter: blur(10px);
  -webkit-backdrop-filter: blur(10px);
  border-radius: 16px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.1);
  border: 1px solid rgba(255, 255, 255, 0.3);
}

.register-header {
  text-align: center;
  margin-bottom: 30px;
}

.register-header p {
  color: white;
  text-shadow: 0 1px 2px rgba(0, 0, 0, 0.2);
}

.register-header h1 {
  margin-bottom: 8px;
  color: white;
  text-shadow: 0 2px 4px rgba(0, 0, 0, 0.2);
}

.action-btns {
  margin-top: 24px;
}

.login-link {
  margin-top: 16px;
  text-align: center;
  font-size: 14px;
  color: white;
}

.login-link a {
  color: #ffffff;
  font-weight: bold;
  text-decoration: none;
  text-shadow: 0 1px 2px rgba(0, 0, 0, 0.2);
}

.login-link a:hover {
  text-decoration: underline;
}
</style>
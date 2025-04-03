import axios from 'axios'
import router from '../router'

// 创建axios请求实例
const request = axios.create({
  baseURL: '',  // 可以设置API的基础URL
  timeout: 10000 // 请求超时时间
})

// 请求拦截器
request.interceptors.request.use(
  config => {
    // 从参数中获取是否需要携带token
    if (config.withToken !== false) {  
      const token = localStorage.getItem('token')
      if (token) {
        config.headers['Authorization'] = `Bearer ${token}`
      }
    }
    return config
  },
  error => {
    console.error('Request error:', error)
    return Promise.reject(error)
  }
)

// 响应拦截器
request.interceptors.response.use(
  response => {
    return response
  },
  error => {
    console.error('Response error:', error)
    // 处理常见错误
    if (error.response) {
      // 服务器返回错误状态码
      switch (error.response.status) {
        case 401:
          // 未授权，可能是token过期
          console.warn('Authentication failed, please login again')
          // 清除本地存储的token和用户信息
          localStorage.removeItem('token')
          localStorage.removeItem('user')
          // 跳转到登录页
          router.push('/login')
          break
        case 403:
          console.warn('Access forbidden')
          break
        case 404:
          console.warn('Resource not found')
          break
        case 500:
          console.warn('Server error')
          break
        default:
          break
      }
    }
    return Promise.reject(error)
  }
)

// 封装常用请求方法
const http = {
  // GET请求
  get(url, params, config = {}) {
    return request({
      method: 'get',
      url,
      params,
      ...config
    })
  },
  
  // POST请求
  post(url, data, config = {}) {
    return request({
      method: 'post',
      url,
      data,
      ...config
    })
  },
  
  // PUT请求
  put(url, data, config = {}) {
    return request({
      method: 'put',
      url,
      data,
      ...config
    })
  },
  
  // DELETE请求
  delete(url, params, config = {}) {
    return request({
      method: 'delete',
      url,
      params,
      ...config
    })
  }
}

export default http
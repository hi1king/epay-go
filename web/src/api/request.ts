// web/src/api/request.ts
import axios, { type AxiosInstance, type AxiosResponse } from 'axios'
import { Message } from '@arco-design/web-vue'

const request: AxiosInstance = axios.create({
  baseURL: '',
  timeout: 30000,
})

// 请求拦截器
request.interceptors.request.use(
  (config) => {
    // 根据当前 URL 判断是管理后台还是商户中心
    const path = window.location.pathname
    let token: string | null = null

    if (path.startsWith('/admin')) {
      token = localStorage.getItem('admin_token')
    } else if (path.startsWith('/merchant')) {
      token = localStorage.getItem('merchant_token')
    }

    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// 响应拦截器
request.interceptors.response.use(
  (response: AxiosResponse) => {
    const { data } = response
    if (data.code === 0) {
      return data
    }
    // 业务错误
    Message.error(data.msg || '请求失败')
    return Promise.reject(new Error(data.msg))
  },
  (error) => {
    if (error.response) {
      const { status, data } = error.response
      if (status === 401) {
        Message.error('登录已过期，请重新登录')
        const path = window.location.pathname
        if (path.startsWith('/admin')) {
          localStorage.removeItem('admin_token')
          window.location.href = '/admin/login'
        } else {
          localStorage.removeItem('merchant_token')
          window.location.href = '/merchant/login'
        }
      } else if (status === 403) {
        Message.error('无权访问')
      } else {
        Message.error(data?.msg || '服务器错误')
      }
    } else {
      Message.error('网络错误')
    }
    return Promise.reject(error)
  }
)

export default request

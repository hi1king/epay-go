<!-- web/src/views/merchant/Profile.vue -->
<template>
  <div>
    <a-card title="基本信息">
      <a-descriptions :column="2">
        <a-descriptions-item label="商户ID">{{ profile.id || '-' }}</a-descriptions-item>
        <a-descriptions-item label="用户名">{{ profile.username }}</a-descriptions-item>
        <a-descriptions-item label="邮箱">{{ profile.email || '-' }}</a-descriptions-item>
        <a-descriptions-item label="手机号">{{ profile.phone || '-' }}</a-descriptions-item>
        <a-descriptions-item label="注册时间">{{ profile.created_at }}</a-descriptions-item>
      </a-descriptions>
    </a-card>

    <a-card title="API密钥" style="margin-top: 16px">
      <div class="api-key-row">
        <a-input :model-value="profile.api_key" readonly style="flex: 1" />
        <a-popconfirm content="重置后原密钥将失效，确定要重置吗？" @ok="handleResetKey">
          <a-button type="outline" status="warning" style="margin-left: 12px">重置密钥</a-button>
        </a-popconfirm>
      </div>
    </a-card>

    <a-row :gutter="16" style="margin-top: 16px">
      <a-col :span="12">
        <a-card title="修改信息">
          <a-form :model="infoForm" layout="vertical">
            <a-form-item label="邮箱">
              <a-input v-model="infoForm.email" placeholder="请输入邮箱" />
            </a-form-item>
            <a-form-item label="手机号">
              <a-input v-model="infoForm.phone" placeholder="请输入手机号" />
            </a-form-item>
            <a-form-item>
              <a-button type="primary" @click="handleUpdateInfo" :loading="updatingInfo">保存</a-button>
            </a-form-item>
          </a-form>
        </a-card>
      </a-col>
      <a-col :span="12">
        <a-card title="修改密码">
          <a-form :model="pwdForm" layout="vertical">
            <a-form-item label="原密码">
              <a-input-password v-model="pwdForm.old_password" placeholder="请输入原密码" />
            </a-form-item>
            <a-form-item label="新密码">
              <a-input-password v-model="pwdForm.new_password" placeholder="请输入新密码" />
            </a-form-item>
            <a-form-item label="确认密码">
              <a-input-password v-model="pwdForm.confirm_password" placeholder="请确认新密码" />
            </a-form-item>
            <a-form-item>
              <a-button type="primary" @click="handleUpdatePassword" :loading="updatingPwd">修改密码</a-button>
            </a-form-item>
          </a-form>
        </a-card>
      </a-col>
    </a-row>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { Message } from '@arco-design/web-vue'
import { getProfile, updateProfile, updatePassword, resetApiKey } from '@/api/merchant'

const profile = reactive({
  id: 0,
  username: '',
  email: '',
  phone: '',
  api_key: '',
  created_at: '',
})

const infoForm = reactive({
  email: '',
  phone: '',
})

const pwdForm = reactive({
  old_password: '',
  new_password: '',
  confirm_password: '',
})

const updatingInfo = ref(false)
const updatingPwd = ref(false)

const fetchProfile = async () => {
  try {
    const res = await getProfile()
    Object.assign(profile, res.data)
    infoForm.email = res.data.email || ''
    infoForm.phone = res.data.phone || ''
  } catch (e) {
    // ignore
  }
}

const handleUpdateInfo = async () => {
  updatingInfo.value = true
  try {
    await updateProfile(infoForm)
    Message.success('信息已更新')
    fetchProfile()
  } catch (e) {
    // ignore
  } finally {
    updatingInfo.value = false
  }
}

const handleUpdatePassword = async () => {
  if (!pwdForm.old_password || !pwdForm.new_password) {
    Message.warning('请填写完整')
    return
  }
  if (pwdForm.new_password !== pwdForm.confirm_password) {
    Message.warning('两次密码输入不一致')
    return
  }
  if (pwdForm.new_password.length < 6) {
    Message.warning('新密码至少6个字符')
    return
  }
  updatingPwd.value = true
  try {
    await updatePassword({ old_password: pwdForm.old_password, new_password: pwdForm.new_password })
    Message.success('密码已修改')
    pwdForm.old_password = ''
    pwdForm.new_password = ''
    pwdForm.confirm_password = ''
  } catch (e) {
    // ignore
  } finally {
    updatingPwd.value = false
  }
}

const handleResetKey = async () => {
  try {
    const res = await resetApiKey()
    profile.api_key = res.data.api_key
    Message.success('密钥已重置')
  } catch (e) {
    // ignore
  }
}

onMounted(() => {
  fetchProfile()
})
</script>

<style scoped>
.api-key-row {
  display: flex;
  align-items: center;
}
</style>

<template>
  <div class="login-layout">
    <div class="form-section">
      <div class="form-panel">
        <div class="form-card">
          <div class="form-header">
            <h2 class="form-title">{{ $t('auth.login') }}</h2>
          </div>

          <div class="form-content">
            <t-form
              ref="formRef"
              :data="formData"
              :rules="formRules"
              @submit="handleLogin"
              layout="vertical"
            >
              <t-form-item label="用户名：" name="username">
                <t-input
                  v-model="formData.username"
                  placeholder="请输入域账户"
                  size="large"
                  :disabled="loading"
                />
              </t-form-item>

              <t-form-item label="密码：" name="password">
                <t-input
                  v-model="formData.password"
                  placeholder="请输入密码"
                  type="password"
                  size="large"
                  :disabled="loading"
                />
              </t-form-item>

              <t-button
                type="submit"
                theme="primary"
                size="large"
                block
                :loading="loading"
                class="submit-button"
              >
                {{ loading ? $t('auth.loggingIn') : $t('auth.login') }}
              </t-button>
            </t-form>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, nextTick, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { MessagePlugin } from 'tdesign-vue-next'
import { login } from '@/api/auth'
import { useAuthStore } from '@/stores/auth'
import { useI18n } from 'vue-i18n'

const router = useRouter()
const authStore = useAuthStore()
const { t } = useI18n()

const formRef = ref()
const loading = ref(false)

const formData = reactive<{ [key: string]: any }>({
  username: '',
  password: ''
})

const formRules = {
  username: [
    { required: true, message: '请输入域账户', type: 'error' },
    { min: 2, message: '用户名长度不能少于2位', type: 'error' },
    { max: 50, message: '用户名长度不能超过50位', type: 'error' }
  ],
  password: [
    { required: true, message: t('auth.passwordRequired'), type: 'error' }
  ]
}

onMounted(() => {
  if (authStore.isLoggedIn) {
    router.replace('/platform/tenant/knowledge-bases')
  }
})

const handleLogin = async () => {
  try {
    const valid = await formRef.value?.validate()
    if (valid !== true) return

    loading.value = true

    const response = await login({
      email: formData.username,
      password: formData.password
    })

    if (response.success) {
      if (response.user && response.tenant && response.token) {
        authStore.setUser({
          id: response.user.id || '',
          username: response.user.username || '',
          email: response.user.email || '',
          avatar: response.user.avatar,
          tenant_id: String(response.tenant.id) || '',
          can_access_all_tenants: response.user.can_access_all_tenants || false,
          system_role: response.user.system_role || 'user',
          created_at: response.user.created_at || new Date().toISOString(),
          updated_at: response.user.updated_at || new Date().toISOString()
        })

        authStore.setToken(response.token)

        if (response.refresh_token) {
          authStore.setRefreshToken(response.refresh_token)
        }

        authStore.setTenant({
          id: String(response.tenant.id) || '',
          name: response.tenant.name || '',
          api_key: response.tenant.api_key || '',
          owner_id: response.user.id || '',
          created_at: response.tenant.created_at || new Date().toISOString(),
          updated_at: response.tenant.updated_at || new Date().toISOString()
        })
      }

      MessagePlugin.success(t('auth.loginSuccess'))
      await nextTick()
      router.replace('/platform/knowledge-bases')
    } else {
      MessagePlugin.error(response.message || t('auth.loginError'))
    }
  } catch (error: any) {
    console.error('登录错误:', error)
    MessagePlugin.error(error.message || t('auth.loginErrorRetry'))
  } finally {
    loading.value = false
  }
}
</script>

<style lang="less" scoped>
.login-layout {
  position: relative;
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 24px;
  background: #f5f7fa;
}

.form-section {
  width: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
}

.form-panel {
  width: 100%;
  max-width: 480px;
}

.form-card {
  background: rgba(255, 255, 255, 0.96);
  backdrop-filter: blur(20px);
  border-radius: 16px;
  padding: 40px;
  box-shadow: 0 10px 40px rgba(0, 0, 0, 0.08);
  box-sizing: border-box;
  width: 100%;
}

.form-header {
  text-align: center;
  margin-bottom: 28px;
}

.form-title {
  font-size: 24px;
  font-weight: 600;
  color: #111827;
  margin: 0;
  font-family: 'PingFang SC', sans-serif;
}

.form-content {
  :deep(.t-form-item__label) {
    font-size: 14px;
    color: #111827;
    font-weight: 500;
    margin-bottom: 8px;
    font-family: 'PingFang SC', sans-serif;
    display: block;
    text-align: left;
  }

  :deep(.t-input) {
    border: 1px solid #e7e7e7;
    border-radius: 8px;
    background: #fff;
    transition: all 0.2s;

    &:focus-within {
      border-color: #07c05f;
      box-shadow: 0 0 0 3px rgba(7, 192, 95, 0.1);
    }

    &:hover {
      border-color: #07c05f;
    }

    .t-input__inner {
      border: none !important;
      box-shadow: none !important;
      outline: none !important;
      background: transparent;
      font-size: 15px;
      font-family: 'PingFang SC', sans-serif;
    }

    .t-input__wrap {
      border: none !important;
      box-shadow: none !important;
    }
  }

  :deep(.t-form-item) {
    margin-bottom: 18px;

    &:last-child {
      margin-bottom: 0;
    }
  }

  :deep(.t-form-item__control) {
    width: 100%;
  }
}

.submit-button {
  height: 46px;
  border-radius: 8px;
  font-size: 16px;
  font-weight: 500;
  font-family: 'PingFang SC', sans-serif;
  margin: 20px 0 0 0;
  transition: all 0.3s;

  :deep(.t-button) {
    background-color: #07c05f;
    border-color: #07c05f;

    &:hover {
      background-color: #06a855;
      border-color: #06a855;
      transform: translateY(-1px);
      box-shadow: 0 4px 12px rgba(7, 192, 95, 0.3);
    }

    &:active {
      transform: translateY(0);
    }
  }
}

@media (max-width: 768px) {
  .login-layout {
    padding: 20px;
  }

  .form-card {
    padding: 32px 24px;
  }
}

@media (max-width: 480px) {
  .login-layout {
    padding: 16px;
  }

  .form-card {
    padding: 28px 20px;
  }

  .form-title {
    font-size: 22px;
  }
}
</style>
<template>
  <Form @submit="handleSignIn">
    <fieldset class="uk-fieldset">
      <legend class="uk-legend">用户登录</legend>

      <div class="uk-margin">
        <label class="uk-form-label" for="name">电子邮箱</label>
        <Field v-model="signInForm.email" name="email" class="uk-input" type="text" rules="required|email"
               label="电子邮箱"/>
        <ErrorMessage class="field-error-message" name="email"/>
      </div>
      <div class="uk-margin">
        <label class="uk-form-label" for="password">密码</label>
        <Field v-model="signInForm.password" type="password" name="password" class="uk-input" rules="required"
               label="密码"/>
        <ErrorMessage class="field-error-message" name="password"/>
      </div>

      <div class="uk-margin">
        <button type="submit" class="uk-button uk-button-primary" :disabled="isLoading || !captchaReady">
          {{ isLoading ? '登录中...' : (captchaReady ? '登录' : '加载中...') }}
        </button>
        <button type="button" class="uk-button uk-button-default" @click="handleForgotPassword">忘记密码
        </button>
      </div>
    </fieldset>
  </Form>

  <Captcha ref="captchaRef"/>
</template>

<script setup lang="ts">
import {computed, ref} from 'vue'
import {Form, Field, ErrorMessage} from 'vee-validate';
import {api} from "@/api/client.ts";
import type {SignIn} from "@/api/api.ts";
import {useRoute, useRouter} from "vue-router";
import {ToastError, ToastSuccess} from "@/utils/notify.ts";
import {useAuthStore} from "@/store";
import Captcha from "@/components/Captcha.vue";

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()

const captchaRef = ref<InstanceType<typeof Captcha> | null>(null)
const captchaReady = computed(() => captchaRef.value?.ready ?? false)

const isLoading = ref<boolean>(false)
const signInForm = ref<SignIn>({
  email: '',
  password: '',
  captcha: '',
})

const handleSignIn = async () => {
  if (isLoading.value || !captchaReady.value) return
  if (!captchaRef.value) {
    ToastError('验证码加载失败，请刷新页面重试')
    return
  }

  isLoading.value = true
  try {
    try {
      signInForm.value.captcha = await captchaRef.value.acquire('sign_in')
    } catch {
      ToastError('验证码校验未完成，请重试')
      return
    }
    if (!signInForm.value.captcha.trim()) {
      ToastError('验证码获取失败，请稍后再试（可能是提交过于频繁）')
      return
    }
    const {data: res} = await api.auth.signIn(signInForm.value)
    ToastSuccess('登录成功，欢迎回来~')
    authStore.signIn(res.profile, res.sessionID)
    const target = typeof route.query.to === 'string' ? route.query.to : ''
    try {
      await router.push(target || {name: 'profile', params: {domain: res.profile.domain}})
    } catch {
      await router.push({name: 'profile', params: {domain: res.profile.domain}})
    }
  } catch {
    // API errors are displayed by the interceptor.
  } finally {
    isLoading.value = false
  }
}

const handleForgotPassword = () => {
  router.push({name: 'forgot-password'})
}
</script>

<style scoped>

</style>

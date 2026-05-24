<template>
  <Form @submit="handleForgotPassword">
    <fieldset class="uk-fieldset">
      <legend class="uk-legend">忘记密码</legend>
      <div class="uk-margin">
        <label class="uk-form-label" for="email">电子邮箱地址</label>
        <Field v-model="forgotPasswordForm.email" name="email" class="uk-input" type="text" rules="required|email"
               label="电子邮箱"/>
        <ErrorMessage class="field-error-message" name="email"/>
      </div>
      <div class="uk-margin">
        <button type="submit" class="uk-button uk-button-primary" :disabled="isLoading || !captchaReady">
          {{ isLoading ? '提交中...' : (captchaReady ? '找回密码' : '加载中...') }}
        </button>
      </div>
    </fieldset>
  </Form>

  <Captcha ref="captchaRef"/>
</template>

<script setup lang="ts">
import {computed, ref} from 'vue'
import {Form, Field, ErrorMessage} from "vee-validate";
import {type ForgotPasswordRequest, forgotPassword} from "@/api/auth.ts";
import {useRouter} from "vue-router";
import {ToastError, ToastSuccess} from "@/utils/notify.ts";
import Captcha from "@/components/Captcha.vue";

const router = useRouter()

const captchaRef = ref<InstanceType<typeof Captcha> | null>(null)
const captchaReady = computed(() => captchaRef.value?.ready ?? false)

const isLoading = ref<boolean>(false);
const forgotPasswordForm = ref<ForgotPasswordRequest>({
  email: '',
  captcha: '',
})

const handleForgotPassword = async () => {
  if (!captchaRef.value) {
    ToastError('验证码加载失败，请刷新页面重试')
    return
  }

  try {
    forgotPasswordForm.value.captcha = await captchaRef.value.acquire('forgot_password')
  } catch (error) {
    ToastError('验证码校验未完成，请重试')
    return
  }

  if (!forgotPasswordForm.value.captcha.trim()) {
    ToastError('验证码获取失败，请稍后再试（可能是提交过于频繁）')
    return
  }

  isLoading.value = true
  forgotPassword(forgotPasswordForm.value).then(res => {
    ToastSuccess(res)
    router.push({name: 'home'})
  }).finally(() => {
    isLoading.value = false
  })
}
</script>

<style scoped>

</style>

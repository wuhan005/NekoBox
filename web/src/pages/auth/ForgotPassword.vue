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
        <button type="submit" class="uk-button uk-button-primary" :disabled="isLoading || !recaptchaReady">
          {{ isLoading ? '提交中...' : (recaptchaReady ? '找回密码' : '加载中...') }}
        </button>
      </div>
    </fieldset>
  </Form>
</template>

<script setup lang="ts">
import {ref, onMounted} from 'vue'
import {Form, Field, ErrorMessage} from "vee-validate";
import {type ForgotPasswordRequest, forgotPassword} from "@/api/auth.ts";
import {useRouter} from "vue-router";
import {type IReCaptchaComposition, useReCaptcha} from "vue-recaptcha-v3";
import {ToastError, ToastSuccess} from "@/utils/notify.ts";
import {ensureRecaptchaReady, getRecaptchaToken} from "@/utils/recaptcha.ts";

const router = useRouter()
const {executeRecaptcha, recaptchaLoaded} = useReCaptcha() as IReCaptchaComposition

const isLoading = ref<boolean>(false);
const recaptchaReady = ref<boolean>(false)
const forgotPasswordForm = ref<ForgotPasswordRequest>({
  email: '',
  recaptcha: '',
})

onMounted(async () => {
  try {
    await ensureRecaptchaReady({executeRecaptcha, recaptchaLoaded})
    recaptchaReady.value = true
  } catch (error) {
    ToastError('无感验证码加载失败，请刷新页面重试')
  }
})
const handleForgotPassword = async () => {
  try {
    forgotPasswordForm.value.recaptcha = await getRecaptchaToken({executeRecaptcha, recaptchaLoaded})
  } catch (error) {
    ToastError('无感验证码加载失败，请刷新页面重试')
    return
  }

  // Check if recaptcha token is valid
  if (!forgotPasswordForm.value.recaptcha || forgotPasswordForm.value.recaptcha.trim() === '') {
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
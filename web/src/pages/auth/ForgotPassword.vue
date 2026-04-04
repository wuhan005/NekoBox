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
        <button type="submit" class="uk-button uk-button-primary" :disabled="isLoading">
          {{ isLoading ? '提交中...' : '找回密码' }}
        </button>
      </div>
    </fieldset>
  </Form>
</template>

<script setup lang="ts">
import {ref} from 'vue'
import {Form, Field, ErrorMessage} from "vee-validate";
import {type ForgotPasswordRequest, forgotPassword} from "@/api/auth.ts";
import {useRouter} from "vue-router";
import {type IReCaptchaComposition, useReCaptcha} from "vue-recaptcha-v3";
import {ToastError, ToastSuccess} from "@/utils/notify.ts";

const router = useRouter()
const {executeRecaptcha, recaptchaLoaded} = useReCaptcha() as IReCaptchaComposition

const isLoading = ref<boolean>(false);
const forgotPasswordForm = ref<ForgotPasswordRequest>({
  email: '',
  recaptcha: '',
})
const handleForgotPassword = async () => {
  try {
    await recaptchaLoaded()
    forgotPasswordForm.value.recaptcha = await executeRecaptcha('submit')
  } catch (error) {
    ToastError('无感验证码加载失败，请刷新页面重试')
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
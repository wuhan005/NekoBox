<template>
  <Form @submit="handleSignUp">
    <fieldset class="uk-fieldset">
      <legend class="uk-legend">新用户注册</legend>

      <div class="uk-margin">
        <label class="uk-form-label" for="email">电子邮箱地址</label>
        <Field v-model="signUpForm.email" name="email" class="uk-input" type="text" rules="required|email"
               label="电子邮箱"/>
        <ErrorMessage class="field-error-message" name="email"/>
      </div>
      <div class="uk-margin">
        <label class="uk-form-label" for="domain">个性域名 (你的问答箱网址将会是：
          <code>{{ ExternalURL }}/_/{{ signUpForm.domain }}</code>)</label>
        <Field v-model="signUpForm.domain" name="domain" class="uk-input" type="text"
               rules="required|alpha_dash|min:3|max:20" label="个性域名"/>
        <ErrorMessage class="field-error-message" name="domain"/>
      </div>
      <div class="uk-margin">
        <label class="uk-form-label" for="name">昵称</label>
        <Field v-model="signUpForm.name" name="name" class="uk-input" type="text" rules="required|max:20" label="昵称"/>
        <ErrorMessage class="field-error-message" name="name"/>
      </div>
      <div class="uk-margin">
        <label class="uk-form-label" for="password">密码</label>
        <Field v-model="signUpForm.password" type="password" name="password" class="uk-input"
               rules="required|min:8|max:30" label="密码"/>
        <ErrorMessage class="field-error-message" name="password"/>
      </div>
      <div class="uk-margin">
        <label class="uk-form-label" for="repeatPassword">确认密码</label>
        <Field v-model="signUpForm.repeatPassword" type="password" name="repeatPassword" class="uk-input"
               rules="required|confirmed:@password" label="确认密码"/>
        <ErrorMessage class="field-error-message" name="repeatPassword"/>
      </div>

      <div class="uk-margin">
        <button type="submit" class="uk-button uk-button-primary" :disabled="isLoading || !recaptchaReady">
          {{ isLoading ? '注册中...' : (recaptchaReady ? '注册' : '加载中...') }}
        </button>
      </div>
    </fieldset>
  </Form>
</template>

<script setup lang="ts">
import {ref, onMounted} from 'vue'
import {Form, Field, ErrorMessage} from 'vee-validate';
import {signUp, type SignUpRequest} from "@/api/auth.ts";
import {ToastError, ToastSuccess} from "@/utils/notify.ts";
import {type IReCaptchaComposition, useReCaptcha} from "vue-recaptcha-v3";
import {useRouter} from "vue-router";
import {ExternalURL} from "@/utils/consts.ts";

const router = useRouter()
const {executeRecaptcha, recaptchaLoaded} = useReCaptcha() as IReCaptchaComposition

const isLoading = ref<boolean>(false)
const recaptchaReady = ref<boolean>(false)
const signUpForm = ref<SignUpRequest>({
  email: '',
  domain: '',
  name: '',
  password: '',
  repeatPassword: '',
  recaptcha: '',
})

onMounted(async () => {
  try {
    await recaptchaLoaded()
    recaptchaReady.value = true
  } catch (error) {
    ToastError('无感验证码加载失败，请刷新页面重试')
  }
})

const handleSignUp = async () => {
  try {
    await recaptchaLoaded()
    signUpForm.value.recaptcha = await executeRecaptcha('submit')
  } catch (error) {
    ToastError('无感验证码加载失败，请刷新页面重试')
    return
  }

  // Check if recaptcha token is valid
  if (!signUpForm.value.recaptcha || signUpForm.value.recaptcha.trim() === '') {
    ToastError('验证码获取失败，请稍后再试（可能是提交过于频繁）')
    return
  }

  isLoading.value = true
  signUp(signUpForm.value)
      .then(res => {
        ToastSuccess(res)
        router.push({name: 'sign-in'})
      })
      .finally(() => {
        isLoading.value = false
      })
}
</script>

<style scoped>

</style>
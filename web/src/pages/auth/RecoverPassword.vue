<template>
  <Form @submit="handleRecoverPassword">
    <fieldset class="uk-fieldset">
      <legend class="uk-legend">重置密码</legend>
      <div class="uk-margin">
        {{ name }}，您正在重置您的密码：
      </div>
      <div class="uk-margin">
        <label class="uk-form-label" for="newPassword">新密码</label>
        <Field v-model="recoverPasswordForm.newPassword" name="newPassword" class="uk-input" type="password"
               rules="required|min:8|max:30"
               label="新密码"/>
        <ErrorMessage class="field-error-message" name="newPassword"/>
      </div>
      <div class="uk-margin">
        <label class="uk-form-label" for="repeatPassword">确认密码</label>
        <Field v-model="recoverPasswordForm.repeatPassword" name="repeatPassword" class="uk-input" type="password"
               rules="required|confirmed:@newPassword"
               label="确认密码"/>
        <ErrorMessage class="field-error-message" name="repeatPassword"/>
      </div>
      <div class="uk-margin">
        <button type="submit" class="uk-button uk-button-primary">重置密码</button>
      </div>
    </fieldset>
  </Form>
</template>

<script setup lang="ts">
import {ref, onMounted} from 'vue'
import {ErrorMessage, Field, Form} from "vee-validate";
import {useRoute, useRouter} from "vue-router";
import {api} from "@/api/client.ts";
import type {RecoverPassword} from "@/api/api.ts";
import {ToastSuccess} from "@/utils/notify.ts";

const route = useRoute()
const router = useRouter()
const name = ref<string>('')
const code = ref<string>(route.query.code as string || '')
const recoverPasswordForm = ref<RecoverPassword>({
  newPassword: '',
  repeatPassword: '',
  code: ''
})

const handleRecoverPassword = () => {
  recoverPasswordForm.value.code = code.value

  api.auth.recoverPassword(recoverPasswordForm.value).then(res => {
    ToastSuccess(res.data)
    router.push({name: 'home'})
  })
}

onMounted(() => {
  if (!code.value) {
    router.push({name: 'home'})
    return
  }

  api.auth.getRecoverPasswordCode({code: code.value}).then(res => {
    name.value = res.data.name
  }).catch(() => {
    router.push({name: 'home'})
  })
})
</script>

<style scoped>

</style>
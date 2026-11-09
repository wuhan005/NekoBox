<template>
  <fieldset class="uk-fieldset">
    <legend class="uk-legend">停用账号</legend>
    <div class="uk-margin">
      停用后，您的账号将无法登录，您的提问箱页面以及提问将无法访问，其他人也无法再给您发送新的提问。<b>该操作无法撤销！请谨慎操作！</b>
      <br>
      <br>
      <button type="button" class="uk-button uk-button-danger" @click="handleDeactivateAccount">我确认停用账号</button>
      <button type="button" class="uk-button uk-button-default" @click="handleBack">返回</button>
    </div>
  </fieldset>
</template>

<script setup lang="ts">
import {useRouter} from "vue-router";
import {api} from "@/api/client.ts";
import {ToastSuccess} from "@/utils/notify.ts";
import {useAuthStore} from "@/store";

const authStore = useAuthStore()
const router = useRouter()

const handleDeactivateAccount = () => {
  api.mine.deactivateAccount().then(res => {
    ToastSuccess(res.data)
    authStore.signOut()
    router.push({name: 'home'})
  })
}

const handleBack = () => {
  router.go(-1)
}
</script>

<style scoped>

</style>
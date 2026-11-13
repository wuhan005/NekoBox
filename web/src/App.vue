<template>
  <Header/>
  <div class="uk-container" :class="$route.name === 'pixel' ? 'uk-container-expand' : 'uk-container-xsmall'">
    <RouterView v-slot="{ Component, route }">
      <KeepAlive v-if="authStore.isSignedIn && (route.name === 'my-questions' || route.name === 'question')"
                 :key="appStore.questionsRevision">
        <component v-if="route.name === 'my-questions'" :is="Component"/>
      </KeepAlive>
      <KeepAlive v-if="route.name === 'profile' || route.name === 'question'"
                 :key="`${authStore.profile.uid ?? ''}:${appStore.publicQuestionsRevision}`" :max="1">
        <component v-if="route.name === 'profile'" :is="Component" :key="route.params.domain?.toString()"/>
      </KeepAlive>
      <component v-if="route.name !== 'my-questions' && route.name !== 'profile'"
                 :is="Component" :key="route.params.domain?.toString()"/>
    </RouterView>
    <Footer/>
  </div>
</template>

<script setup lang="ts">
import {onMounted} from "vue";
import Header from "@/components/Header.vue";
import Footer from "@/components/Footer.vue";
import {useAppStore, useAuthStore} from "@/store";
import {loadGlobalConfig} from "@/utils/config.ts";

const appStore = useAppStore()
const authStore = useAuthStore()

onMounted(() => {
  void loadGlobalConfig().catch(() => { /* Configuration-dependent forms can retry the request. */ })
})
</script>

<style scoped>

</style>

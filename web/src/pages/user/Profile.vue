<template>
  <template v-if="profile">
    <ProfileCard
        :name="profile.name"
        :intro="profile.intro"
        :avatar="profile.avatar"
        :background="profile.background"
    />

    <div>
      <div class="uk-card uk-card-default uk-card-small uk-card-body">
        <div style="display: flex;align-items: center;justify-content: space-between;">
          <div style="width: 36px"></div>
          <p class="uk-text-center uk-text-small">谁都可以以匿名的形式提问</p>
          <a href="#qrcode-modal" class="uk-icon-button" uk-icon="social" uk-toggle></a>
        </div>

        <ShareQRCode :name="profile.name" :avatar="profile.avatar" :qrcode="`${ExternalURL}/_/${profile.domain}`"/>

        <NewQuestion
            :page-profile-domain="profile.domain"
            :harassment-setting="profile.harassmentSetting"
        />

        <hr class="uk-divider-icon">
        <PageQuestions
            :page-profile-name="profile.name"
            :page-profile-domain="profile.domain"
        />
      </div>
    </div>
  </template>
  <Skeleton v-else-if="isLoading" :count="8"/>
  <div v-else class="uk-alert uk-alert-danger" role="alert">
    提问箱加载失败，请重试。
    <button type="button" class="uk-button uk-button-default" @click="fetchProfile">重新加载提问箱</button>
  </div>
</template>

<script setup lang="ts">
import {ref, onMounted, onUnmounted} from "vue";
import {api} from "@/api/client.ts";
import type {UserProfile} from "@/api/api.ts";
import {useRoute, useRouter} from "vue-router";
import {ExternalURL} from "@/utils/consts.ts";
import PageQuestions from "@/components/PageQuestions.vue";
import NewQuestion from "@/components/NewQuestion.vue";
import ShareQRCode from "@/components/ShareQRCode.vue";
import ProfileCard from "@/components/ProfileCard.vue";
import {Skeleton} from 'vue-loading-skeleton';

const route = useRoute()
const router = useRouter()
const domain = ref<string>(route.params.domain as string || '')

const isLoading = ref<boolean>(true)
const profile = ref<UserProfile | null>(null)
let isUnmounted = false
onUnmounted(() => { isUnmounted = true })

const fetchProfile = () => {
  isLoading.value = true
  return api.users.getUserProfile(domain.value)
      .then(res => {
        if (isUnmounted) return
        profile.value = res.data

        document.title = `${profile.value.name} - NekoBox`
      })
      .catch(err => {
        if (isUnmounted) return
        const statusCode = err.response?.status
        if (statusCode === 404) {
          router.push({name: 'home'})
        }
      })
      .finally(() => {
        isLoading.value = false
      })
}
onMounted(fetchProfile)
</script>

<style scoped>

</style>

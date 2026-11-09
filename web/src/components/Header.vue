<template>
  <nav class="uk-navbar-container uk-margin uk-navbar">
    <div class="uk-navbar-center">
      <div v-if="authStore.isSignedIn" class="uk-navbar-center-left">
        <ul class="uk-navbar-nav">
          <li><a :href="router.resolve({name: 'profile', params: {domain: authStore.profile.domain}}).href"
                 @click.prevent="router.push({name: 'profile', params: {domain: authStore.profile.domain}})">
            <span class="profile-name" :title="authStore.profile.name">{{ authStore.profile.name }}</span>
          </a></li>
        </ul>
      </div>
      <a class="uk-navbar-item uk-logo"
         :href="router.resolve({name: 'home'}).href"
         @click.prevent="router.push({name: 'home'})">NekoBox</a>

      <div v-if="authStore.isSignedIn" class="uk-navbar-center-right">
        <ul class="uk-navbar-nav">
          <li><a :href="router.resolve({name: 'my-questions'}).href"
                 @click.prevent="router.push({name: 'my-questions'})">提问</a></li>
        </ul>
        <ul class="uk-navbar-nav">
          <li><a :href="router.resolve({name: 'settings'}).href"
                 @click.prevent="router.push({name: 'settings'})">设置</a></li>
        </ul>
      </div>

      <div v-else class="uk-navbar-center-right">
        <ul class="uk-navbar-nav">
          <li><a :href="router.resolve({name: 'sign-up'}).href"
                 @click.prevent="router.push({name: 'sign-up'})">注册</a></li>
        </ul>
        <ul class="uk-navbar-nav">
          <li><a :href="router.resolve({name: 'sign-in'}).href"
                 @click.prevent="router.push({name: 'sign-in'})">登录</a></li>
        </ul>
      </div>
    </div>
  </nav>
</template>

<script setup lang="ts">
import {useRouter} from "vue-router";
import {useAuthStore} from "@/store";

const authStore = useAuthStore()
const router = useRouter()
</script>

<style scoped>
@media (max-width: 767px) {
  .uk-navbar-center {
    width: 100%;
    display: grid;
    grid-template-columns: minmax(0, 1fr) auto minmax(0, 1fr);
    gap: 0;
    padding: 0 10px;
    box-sizing: border-box;
  }
  .uk-navbar-center-left,
  .uk-navbar-center-right {
    position: static;
    min-width: 0;
    transform: none;
  }
  .uk-navbar-center-right {
    grid-column: 3;
    grid-row: 1;
    justify-self: end;
  }
  .uk-navbar-center-left {
    grid-column: 1;
    grid-row: 1;
    width: 100%;
  }
  .uk-navbar-center-left .uk-navbar-nav,
  .uk-navbar-center-left li,
  .profile-name {
    min-width: 0;
  }
  .profile-name {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .uk-navbar-nav > li > a,
  .uk-logo {
    padding: 0 6px;
  }
  .uk-logo {
    grid-column: 2;
    grid-row: 1;
  }
}
</style>

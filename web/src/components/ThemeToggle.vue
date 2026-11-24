<template>
  <a href="#" role="button" class="theme-toggle" :title="title" :aria-label="title" @click.prevent="handleToggle">
    <svg v-if="preference === 'light'" width="20" height="20" viewBox="0 0 20 20" xmlns="http://www.w3.org/2000/svg"
         fill="none" stroke="currentColor" stroke-width="1.1" stroke-linecap="round">
      <circle cx="10" cy="10" r="3.5"/>
      <path d="M10 1.5v2M10 16.5v2M1.5 10h2M16.5 10h2M4 4l1.4 1.4M14.6 14.6L16 16M4 16l1.4-1.4M14.6 5.4L16 4"/>
    </svg>
    <svg v-else-if="preference === 'dark'" width="20" height="20" viewBox="0 0 20 20"
         xmlns="http://www.w3.org/2000/svg" fill="none" stroke="currentColor" stroke-width="1.1"
         stroke-linejoin="round">
      <path d="M16.5 12.2A7 7 0 0 1 7.8 3.5a7 7 0 1 0 8.7 8.7z"/>
    </svg>
    <svg v-else width="20" height="20" viewBox="0 0 20 20" xmlns="http://www.w3.org/2000/svg">
      <circle cx="10" cy="10" r="7" fill="none" stroke="currentColor" stroke-width="1.1"/>
      <path d="M10 3a7 7 0 0 0 0 14z" fill="currentColor"/>
    </svg>
  </a>
</template>

<script setup lang="ts">
import {computed} from "vue";
import {useAppStore} from "@/store";
import {normalizeThemePreference, setThemePreference, THEME_LABELS, THEME_PREFERENCES} from "@/utils/theme.ts";

const appStore = useAppStore()

const preference = computed(() => normalizeThemePreference(appStore.theme))
const title = computed(() => `主题：${THEME_LABELS[preference.value]}（点击切换）`)

const handleToggle = () => {
  const index = THEME_PREFERENCES.indexOf(preference.value)
  setThemePreference(THEME_PREFERENCES[(index + 1) % THEME_PREFERENCES.length])
}
</script>
<template>
  <p class="uk-text-right uk-text-muted uk-text-small">
    Made with ❤️ by E99p1ant. {{ icp }}
    <br>
    <a @click="router.push({name: 'pixel'})">画板</a> |
    <a href="https://github.com/wuhan005/NekoBox" target="_blank">GitHub</a> | <a
      :href="`https://github.com/wuhan005/NekoBox/commit/${commitSHA}`" target="_blank">
    {{ commitSHA.slice(0, 7) }}
  </a>
    | <a href="https://support.qq.com/products/293656" target="_blank">吐槽反馈</a>
    | <a @click="router.push({name: 'sponsor'})">我要打钱</a>
  </p>
  <br/>
</template>

<script setup lang="ts">
import {useRouter} from "vue-router";

const icp = import.meta.env.VITE_ICP || ''
const commitSHA = import.meta.env.VITE_COMMIT_SHA || ''
const router = useRouter()
</script>

<style scoped>

</style>
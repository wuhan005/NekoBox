<template>
  <footer class="uk-margin-top uk-margin-medium-bottom uk-text-muted uk-text-small">
    <nav aria-label="页脚导航">
      <RouterLink :to="{name: 'pixel'}">画板</RouterLink> |
      <a href="https://github.com/wuhan005/NekoBox" target="_blank">GitHub</a> |
      <a href="https://support.qq.com/products/293656" target="_blank">吐槽反馈</a> |
      <RouterLink :to="{name: 'sponsor'}">我要打钱</RouterLink>
    </nav>
    <div class="footer-details uk-text-right">
      <p class="uk-margin-remove">Made with ❤️ by E99p1ant. {{ globalConfig?.icp }}</p>
      <nav class="uk-link-text" aria-label="条款和政策">
        <RouterLink :to="{name: 'terms'}">服务条款</RouterLink> |
        <RouterLink :to="{name: 'privacy'}">隐私政策</RouterLink>
      </nav>
    </div>
  </footer>
</template>

<script setup lang="ts">
import {globalConfig} from "@/utils/config.ts";
</script>

<style scoped>
footer {
  display: flex;
  flex-wrap: wrap;
  gap: 0.25rem 1rem;
}

.footer-details {
  margin-inline-start: auto;
}
</style>

<template>
  <div class="uk-card uk-card-default uk-card-body">
    <h1 class="uk-heading-small">NekoBox</h1>
    <p>一个究极简洁的匿名问答箱。没有多余的社交元素，只是一问一答。不提醒，不打扰。你不需要下载
      App，不需要绑定社交账号，甚至连邮箱都不需要是真的。（划掉</p>
    <br>
    <hr>

    <Skeleton :count="3" :loading="isLogLoading">
      <div>
        <p class=uk-article-meta>开发日记 - {{ humanizeDate(log.date) }}</p>
        <span class=uk-text-small v-html="log.content"></span>
      </div>
      <p class="uk-text-right uk-text-small">
        <a class="uk-link-text"
           :href="router.resolve({name: 'change-logs'}).href"
           @click.prevent="handleViewChangeLogs">查看更多...</a>
      </p>
    </Skeleton>
  </div>
</template>

<script setup lang="ts">
import {ref, onMounted} from "vue";
import {useRouter} from "vue-router";
import {type ChangeLogItem, getChangeLogs} from "@/utils/feeds.ts";
import {humanizeDate} from "@/utils/humanize.ts";
import {Skeleton} from 'vue-loading-skeleton';

const router = useRouter()
const isLogLoading = ref<boolean>(true)
const log = ref<ChangeLogItem>({
  date: '',
  content: ''
})
const handleViewChangeLogs = () => {
  router.push({name: 'change-logs'})
}

onMounted(() => {
  getChangeLogs().then(res => {
    log.value = res.data[0]
  }).finally(() => {
    isLogLoading.value = false
  })
})
</script>

<style scoped>

</style>
<template>
  <div class="uk-card uk-card-default uk-card-body">
    <h2>NekoBox 开发日记</h2>
    <div>
      <p>这里记录了 NekoBox 的功能迭代历史以及我的一些碎碎念。</p>
    </div>
    <dl v-if="!isLoading" class="uk-description-list uk-description-list-divider">
      <div v-for="(item, index) in logs" v-bind:key="index">
        <hr>
        <dt>{{ humanizeDate(item.date) }}</dt>
        <br>
        <dd class="uk-text-small" v-html="item.content"></dd>
      </div>
    </dl>
    <div v-else>
      <hr/>
      <span>加载中...</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import {ref, onMounted} from "vue";
import {humanizeDate} from "@/utils/humanize.ts";
import {type ChangeLogItem, getChangeLogs} from "@/utils/feeds.ts";

const isLoading = ref<boolean>(true)
const logs = ref<ChangeLogItem[]>([])

onMounted(() => {
  getChangeLogs().then(res => {
    logs.value = res.data
  }).finally(() => {
    isLoading.value = false
  })
})
</script>

<style scoped>

</style>
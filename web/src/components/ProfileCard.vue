<template>
  <div class="uk-card uk-card-default uk-text-center">
    <div class="uk-height-medium uk-flex uk-flex-center uk-flex-middle uk-background-cover uk-light"
         :style="{backgroundImage: `url(${props.background})`}">
      <div class="uk-card-body">
        <img alt="avatar" class="uk-border-circle uk-box-shadow-large" :src="props.avatar" width="100" height="100"/>
        <h3>
          <Skeleton>{{ props.name }}</Skeleton>
        </h3>
        <p>
          <Skeleton>{{ props.intro }}</Skeleton>
        </p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import {Skeleton} from 'vue-loading-skeleton';

const props = defineProps({
  name: String,
  intro: String,
  avatar: String,
  background: String,
})
</script>

<style scoped>

</style>
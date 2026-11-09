<template>
  <Header/>
  <div class="uk-container" :class="$route.name === 'pixel' ? 'uk-container-expand' : 'uk-container-xsmall'">
    <router-view :key="$route.params.domain?.toString()"/>
    <Footer/>
  </div>
</template>

<script setup lang="ts">
import Header from "@/components/Header.vue";
import Footer from "@/components/Footer.vue";
</script>

<style scoped>

</style>

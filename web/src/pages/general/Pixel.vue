<template>
  <div class="uk-container uk-container-large">
    <div class="uk-padding uk-padding-remove-vertical uk-margin-small-left uk-margin-small-right">
      <NekoPixelApp class="pixel-board" :editable="authStore.isSignedIn"/>
    </div>
  </div>

  <div class="uk-alert-warning uk-text-small" uk-alert>
    <p>
      请各位用户知悉并确认在使用画板功能时，不会发起及/或参与以下内容：
      1. 涉及地域、种族、性别、宗教、政治等方面的歧视性内容；
      2. 侮辱、诽谤他人，侵害他人合法权益的内容；
      3. 含有淫秽色情、血腥暴力、赌博的内容；
      4. 其他在法律法规中禁止的或不适宜的内容。
      如果您实施了违反上述规范的行为，我们将视情况采取相关措施，并保留移交司法机关或者行政机关的权利。
    </p>
  </div>
</template>

<script setup lang="ts">
import {onUnmounted, watchEffect} from 'vue'
import {NekoPixelApp, setAuthToken, setBaseURL} from '@e99p1ant/neko-pixel-umd/neko-pixel-app.es.js'
import '@e99p1ant/neko-pixel-umd/style.css'
import {useAuthStore} from '@/store'

const authStore = useAuthStore()
const apiOrigin = new URL(import.meta.env.VITE_BASE_URL || '/api', window.location.origin).origin
setBaseURL(`${apiOrigin}/service/pixel`)

watchEffect(() => setAuthToken(authStore.isSignedIn ? authStore.sessionID : ''))
onUnmounted(() => setAuthToken(''))
</script>

<style scoped>
.pixel-board :deep(.painting-canvas) {
  width: 100%;
  height: auto;
  image-rendering: pixelated;
}

.pixel-board :deep(.toolbar),
.pixel-board :deep(.color-picker) {
  flex-wrap: wrap;
}

.pixel-board :deep(.palette-panel) {
  flex: 1 1 20rem;
  width: auto;
  min-width: 0;
}

.pixel-board :deep(.paint .uk-button) {
  white-space: nowrap;
}
</style>

<template>
  <div v-if="loadError" class="uk-alert uk-alert-danger uk-text-small" role="alert">
    验证码加载失败，请重试。
    <button type="button" class="uk-button uk-button-default uk-button-small" @click="initialize">重新加载验证码</button>
  </div>
  <div v-if="config?.type === 'go_captcha'" class="captcha-popover" :class="{ 'captcha-popover-visible': popoverVisible }">
    <div class="captcha-popover-mask" @click="cancelChallenge"></div>
    <div class="captcha-popover-panel">
      <!-- :key forces Vue to remount the slide widget for each new challenge so its internal
           drag state (drag block / tile position) is always fresh; reset()/refresh() callbacks
           on the widget itself only clear part of the state and race with async data updates. -->
      <Slide :key="challengeKey" :config="slideConfig" :data="slideData" :events="slideEvents"/>
    </div>
  </div>
</template>

<script setup lang="ts">
import {onMounted, onUnmounted, ref, shallowRef} from 'vue'
import {Slide} from 'go-captcha-vue'
import 'go-captcha-vue/dist/style.css'
import {getCaptchaChallenge, verifyCaptchaSlide} from '@/api/captcha.ts'
import {E2E_BYPASS_TOKEN, isE2EMode, loadCaptchaConfig} from '@/utils/captcha.ts'
import type {CaptchaConfig} from '@/api/captcha.ts'
import {executeRecaptcha, loadRecaptcha} from '@/utils/recaptcha.ts'

interface SlidePoint {
  x: number;
  y: number;
}

const config = shallowRef<CaptchaConfig | null>(null)
const ready = ref<boolean>(false)
const initializing = ref<boolean>(false)
const loadError = ref<boolean>(false)
let isUnmounted = false
let challengeRequest = 0
let verifyingKey = ''

// State for the go-captcha slide widget.
const popoverVisible = ref<boolean>(false)
const slideData = ref({
  thumbX: 0,
  thumbY: 0,
  thumbWidth: 0,
  thumbHeight: 0,
  image: '',
  thumb: '',
})
const slideConfig = {
  width: 300,
  height: 220,
  showTheme: true,
  title: '请拖动滑块完成拼图',
}

// Promise controller for the in-flight acquire() call. Cleared on resolve/reject.
let pendingResolver: { resolve: (token: string) => void; reject: (err: Error) => void } | null = null

// Bound to the slide widget's :key so a new challenge remounts the widget and resets its drag state.
const challengeKey = ref('')

onMounted(initialize)
onUnmounted(() => {
  isUnmounted = true
  cancelChallenge()
})

async function initialize() {
  if (initializing.value) return
  initializing.value = true
  ready.value = false
  loadError.value = false
  try {
    config.value = await loadCaptchaConfig()
    if (!isE2EMode && config.value.type === 'recaptcha') await loadRecaptcha()
    if (!isUnmounted) ready.value = true
  } catch {
    if (!isUnmounted) loadError.value = true
  } finally {
    initializing.value = false
  }
}

async function loadChallenge() {
  const request = ++challengeRequest
  let data
  try {
    data = await getCaptchaChallenge()
  } catch (error) {
    if (isUnmounted || request !== challengeRequest) return
    throw error
  }
  if (isUnmounted || request !== challengeRequest) return
  challengeKey.value = data.key
  slideData.value = {
    image: data.image,
    thumb: data.thumb,
    thumbX: data.thumbX,
    thumbY: data.thumbY,
    thumbWidth: data.thumbWidth,
    thumbHeight: data.thumbHeight,
  }
}

async function refreshChallenge() {
  try {
    await loadChallenge()
  } catch {
    cancelChallenge()
    ready.value = false
    loadError.value = true
  }
}

const slideEvents = {
  refresh: () => {
    void refreshChallenge()
  },
  close: () => {
    cancelChallenge()
  },
  // The widget passes a `reset` callback as the second arg; we bump challengeKey instead, which
  // remounts the widget and is the only way to reliably reset both the drag block and the tile
  // when the underlying data is also being swapped asynchronously.
  confirm: (point: SlidePoint) => {
    const resolver = pendingResolver
    const key = challengeKey.value
    if (!resolver || verifyingKey === key) return
    verifyingKey = key
    void (async () => {
      try {
        const resp = await verifyCaptchaSlide({key, x: point.x, y: point.y})
        if (pendingResolver !== resolver || challengeKey.value !== key) return
        pendingResolver = null
        popoverVisible.value = false
        resolver.resolve(resp.token)
      } catch {
        if (pendingResolver === resolver) await refreshChallenge()
      } finally {
        if (verifyingKey === key) verifyingKey = ''
      }
    })()
  },
}

function cancelChallenge() {
  challengeRequest++
  verifyingKey = ''
  if (!pendingResolver) {
    popoverVisible.value = false
    return
  }
  const resolver = pendingResolver
  pendingResolver = null
  popoverVisible.value = false
  resolver.reject(new Error('captcha_canceled'))
}

// acquire runs a full captcha flow and returns a token consumable by business endpoints.
async function acquire(action = 'submit'): Promise<string> {
  if (isE2EMode) {
    return E2E_BYPASS_TOKEN
  }

  if (!ready.value || !config.value) throw new Error('验证码尚未就绪')

  if (config.value.type === 'recaptcha') {
    return await executeRecaptcha(action)
  }

  // go-captcha mode: open the slide panel and wait for user interaction.
  if (pendingResolver) {
    return Promise.reject(new Error('captcha_busy'))
  }

  return new Promise<string>((resolve, reject) => {
    pendingResolver = {resolve, reject}
    const resolver = pendingResolver
    void loadChallenge().then(() => {
      if (pendingResolver === resolver) popoverVisible.value = true
    }).catch(error => {
      if (pendingResolver !== resolver) return
      pendingResolver = null
      ready.value = false
      loadError.value = true
      reject(error)
    })
  })
}

defineExpose({
  ready,
  acquire,
  type: () => config.value?.type ?? 'recaptcha',
})
</script>

<style scoped>
.captcha-popover {
  display: none;
  position: fixed;
  inset: 0;
  z-index: 1080;
}

.captcha-popover-visible {
  display: block;
}

.captcha-popover-mask {
  position: absolute;
  inset: 0;
  background: rgba(0, 0, 0, 0.45);
}

.captcha-popover-panel {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
}
</style>

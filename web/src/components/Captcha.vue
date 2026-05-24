<template>
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
import {onMounted, ref, shallowRef} from 'vue'
import {Slide} from 'go-captcha-vue'
import 'go-captcha-vue/dist/style.css'
import {type IReCaptchaComposition, useReCaptcha} from 'vue-recaptcha-v3'
import {getCaptchaChallenge, verifyCaptchaSlide} from '@/api/captcha.ts'
import {E2E_BYPASS_TOKEN, isE2EMode, loadCaptchaConfig} from '@/utils/captcha.ts'
import type {CaptchaConfig} from '@/api/captcha.ts'

interface SlidePoint {
  x: number;
  y: number;
}

// useReCaptcha returns undefined when main.ts skipped registering the plugin (e.g. site_key empty);
// only used when type === 'recaptcha'.
const recaptchaApi = useReCaptcha() as IReCaptchaComposition | undefined

const config = shallowRef<CaptchaConfig | null>(null)
const ready = ref<boolean>(false)

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

onMounted(async () => {
  config.value = await loadCaptchaConfig()
  if (isE2EMode) {
    ready.value = true
    return
  }

  if (config.value.type === 'recaptcha') {
    if (!recaptchaApi) {
      console.warn('reCAPTCHA not initialized')
      return
    }
    try {
      await recaptchaApi.recaptchaLoaded()
      ready.value = true
    } catch (err) {
      console.warn('Failed to initialize reCAPTCHA', err)
    }
    return
  }

  ready.value = true
})

async function loadChallenge() {
  const data = await getCaptchaChallenge()
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

const slideEvents = {
  refresh: () => {
    void loadChallenge()
  },
  close: () => {
    cancelChallenge()
  },
  // The widget passes a `reset` callback as the second arg; we bump challengeKey instead, which
  // remounts the widget and is the only way to reliably reset both the drag block and the tile
  // when the underlying data is also being swapped asynchronously.
  confirm: (point: SlidePoint) => {
    void (async () => {
      try {
        const resp = await verifyCaptchaSlide({key: challengeKey.value, x: point.x, y: point.y})
        const resolver = pendingResolver
        pendingResolver = null
        popoverVisible.value = false
        resolver?.resolve(resp.token)
      } catch (err) {
        await loadChallenge()
      }
    })()
  },
}

function cancelChallenge() {
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

  if (!config.value) {
    config.value = await loadCaptchaConfig()
  }

  if (config.value.type === 'recaptcha') {
    if (!recaptchaApi) {
      throw new Error('reCAPTCHA 未初始化，请刷新页面重试')
    }
    await recaptchaApi.recaptchaLoaded()
    return await recaptchaApi.executeRecaptcha(action)
  }

  // go-captcha mode: open the slide panel and wait for user interaction.
  if (pendingResolver) {
    return Promise.reject(new Error('captcha_busy'))
  }

  await loadChallenge()
  popoverVisible.value = true

  return new Promise<string>((resolve, reject) => {
    pendingResolver = {resolve, reject}
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

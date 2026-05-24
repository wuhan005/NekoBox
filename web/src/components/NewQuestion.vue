<template>
  <div v-if="!canPostQuestion">
    <div uk-alert class="uk-text-center">
      <p>提问箱的主人设置了仅注册用户才能提问，你需要先登录 NekoBox 才能向他提问。</p>
    </div>
    <div class="uk-margin uk-text-center">
      <a class="uk-button uk-button-primary" :href="`/sign-in?to=${route.path}`">前往登录</a>
    </div>
  </div>

  <div v-else class="uk-margin">
    <div v-if="successMessageVisible" class="uk-alert-success" uk-alert>
      <a href="#" class="uk-alert-close" uk-close></a>
      <p v-html="successMessage"></p>
    </div>

    <Form @submit="handleSubmit">
      <div class="uk-margin uk-text-center">
      <textarea v-model="postQuestionForm.content"
                name="content" class="uk-textarea" rows="5" placeholder="在此处撰写你的问题..."
                maxlength="1000"></textarea>

        <div
            style="width: 100%; background-color: #f8f8f8; display: flex; padding-top: 5px; padding-bottom: 5px; align-items: center; justify-content: start;gap: 5px">
          <div class="uk-form-custom">
            <a href="#" class="uk-icon-link" uk-icon="image" style="margin-left: 10px"></a>
            <span style="font-size: 12px; margin-left: 5px">
            {{ postQuestionForm.images.length === 0 ? '添加图片' : postQuestionForm.images[0].name }}
          </span>
            <input ref="imageUploader" name="images" type="file" accept="image/*" @change="handleSelectImage">
          </div>
          <button v-if="postQuestionForm.images.length > 0" type="button" class="uk-icon-link" uk-icon="close"
                  style="margin-left: 10px;" @click="handleRemoveSelectedImage"></button>
        </div>
      </div>

      <div class="uk-margin uk-grid-small">
        <label class="uk-text-small">
          <input v-model="postQuestionForm.isPrivate" name="private" type="checkbox" class="uk-checkbox"/> 回复后不公开提问
        </label>
        <label class="uk-text-small">
          <input v-model="receiveReplyViaEmail" type="checkbox" class="uk-checkbox"/> 我想接收回复通知
        </label>
        <div v-show="receiveReplyViaEmail">
          <label>
            <label
                class="uk-form-label">接收回复通知的电子邮箱地址（当你的提问被提问箱主人回复时，你将收到一封邮件）</label>
            <input v-model="postQuestionForm.receiveReplyEmail" name="receiveReplyEmail" class="uk-input" type="text"
                   placeholder="电子邮箱地址"/>
          </label>
        </div>
        <br>
      </div>

      <div class="uk-margin uk-text-center">
        <button type="submit" class="uk-button uk-button-primary" :disabled="isLoading || !captchaReady">
          {{ isLoading ? '发送中...' : (captchaReady ? '发送提问' : '加载中...') }}
        </button>
      </div>
    </Form>

    <Captcha ref="captchaRef"/>
  </div>
</template>

<script setup lang="ts">
import {ref, computed, defineProps} from "vue";
import {useAuthStore} from "@/store";
import {postQuestion, type PostQuestionRequest} from "@/api/user.ts";
import {Form} from "vee-validate";
import {ToastError} from "@/utils/notify.ts";
import {useRoute} from "vue-router";
import Captcha from "@/components/Captcha.vue";

const route = useRoute();

const authStore = useAuthStore()

const captchaRef = ref<InstanceType<typeof Captcha> | null>(null)
const captchaReady = computed(() => captchaRef.value?.ready ?? false)

const canPostQuestion = computed(() => {
  return props.harassmentSetting !== 'register_only' || authStore.isSignedIn
})

const props = defineProps({
  harassmentSetting: {
    type: String,
    required: true
  },
  pageProfileDomain: {
    type: String,
    required: true
  }
})

const receiveReplyViaEmail = ref<boolean>(false)
const postQuestionForm = ref<PostQuestionRequest>({
  content: '',
  receiveReplyEmail: '',
  images: [],
  isPrivate: false,
  captcha: '',
})

const isLoading = ref<boolean>(false)
const successMessageVisible = ref<boolean>(false)
const successMessage = ref<string>('')

const handleSubmit = async () => {
  if (!captchaRef.value) {
    ToastError('验证码加载失败，请刷新页面重试')
    return
  }

  try {
    postQuestionForm.value.captcha = await captchaRef.value.acquire('post_question')
  } catch (error) {
    ToastError('验证码校验未完成，请重试')
    return
  }

  if (!postQuestionForm.value.captcha.trim()) {
    ToastError('验证码获取失败，请稍后再试（可能是提交过于频繁）')
    return
  }

  isLoading.value = true
  postQuestion(props.pageProfileDomain, postQuestionForm.value)
      .then(res => {
        successMessage.value = res
        successMessageVisible.value = true

        if (imageUploader.value) {
          imageUploader.value.value = ''
        }
        receiveReplyViaEmail.value = false
        postQuestionForm.value = {
          content: '',
          receiveReplyEmail: '',
          images: [],
          isPrivate: false,
          captcha: '',
        }
      })
      .finally(() => {
        isLoading.value = false
      })
}

const imageUploader = ref<HTMLInputElement | null>(null)
const handleSelectImage = (event: Event) => {
  const target = event.target as HTMLInputElement
  if (target.files && target.files.length > 0) {
    postQuestionForm.value.images = Array.from(target.files)
  } else {
    postQuestionForm.value.images = []
  }
}

const handleRemoveSelectedImage = () => {
  if (imageUploader.value) {
    imageUploader.value.value = ''
  }
  postQuestionForm.value.images = []
}
</script>

<style scoped>

</style>

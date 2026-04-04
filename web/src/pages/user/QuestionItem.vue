<template>
  <div>
    <div class="uk-card uk-card-default">
      <div class="uk-card-header">
        <div class="uk-text-left uk-text-small uk-text-muted">
          <Skeleton :loading="!question.createdAt" width="20%">{{ humanizeDate(question.createdAt) }}</Skeleton>
        </div>
        <h4 class="uk-text-center uk-margin-top uk-margin-bottom uk-text-break">
          <Skeleton>{{ question.content }}</Skeleton>
        </h4>
        <ul v-if="question.questionImageURLs.length > 0" class="uk-thumbnav" uk-margin>
          <li v-for="(imageURL, index) in question.questionImageURLs" v-bind:key="index" uk-lightbox>
            <a :href="imageURL">
              <img class="uk-border-rounded uk-object-fill" :src="imageURL" width="100" height="100"/>
            </a>
          </li>
        </ul>
      </div>

      <div v-if="question.answer" class="uk-card-body">
        <p class="uk-text-small uk-text-break">
          <Skeleton>{{ question.answer }}</Skeleton>
        </p>
        <ul v-if="question.answerImageURLs.length > 0" class="uk-thumbnav" uk-margin>
          <li v-for="(imageURL, index) in question.answerImageURLs" v-bind:key="index" uk-lightbox>
            <a :href="imageURL">
              <img class="uk-border-rounded uk-object-fill" :src="imageURL" width="100" height="100"/>
            </a>
          </li>
        </ul>
        <p class="uk-text-small uk-text-right uk-text-muted">-来自@{{ profile.name }}的回答</p>
      </div>

      <div class="uk-card-footer">
        <div v-if="question.isOwner">
          <div class="uk-float-right">
            <button type="button" class="uk-button uk-button-default uk-button-small" @click="handleSetVisible">
              设为{{ question.isPrivate ? '公开' : '私密' }}
            </button>
          </div>

          <a class="uk-button uk-button-danger uk-button-small" href="#">删除提问</a>
          <div class="uk-dropbar uk-card-default" uk-drop="mode: click">
            <h3 class="uk-card-title">危险！</h3>
            <p>你确定要删除这个提问吗？<br/>该操作不可恢复，请谨慎操作。</p>
            <button type="button" class="uk-button uk-button-danger" @click="handleDelete">确认删除</button>
          </div>

          <h5 class="uk-text-center">回答问题</h5>
          <Form @submit="handleAnswer">
            <div class="uk-margin uk-text-center">
              <textarea v-model="answerQuestionForm.answer"
                        name="answer" class="uk-textarea" rows="5" maxlength="1000"
                        placeholder="在此处撰写你的回答...">
              </textarea>

              <div
                  style="width: 100%; background-color: #f8f8f8; display: flex; padding-top: 5px; padding-bottom: 5px; align-items: center; justify-content: start;gap: 5px">
                <div class="uk-form-custom">
                  <a href="#" class="uk-icon-link" uk-icon="image" style="margin-left: 10px"></a>
                  <span style="font-size: 12px; margin-left: 5px">
                      {{ answerQuestionForm.images.length === 0 ? '添加图片' : answerQuestionForm.images[0].name }}
                    </span>
                  <input ref="imageUploader" name="images" type="file" accept="image/*" @change="handleSelectImage">
                </div>
                <button v-if="answerQuestionForm.images.length > 0" type="button" class="uk-icon-link" uk-icon="close"
                        style="margin-left: 10px;" @click="handleRemoveSelectedImage"></button>
              </div>
            </div>

            <div v-if="question.hasReplyEmail" class="uk-alert-warning uk-text-small" uk-alert>
              <p>提问人留下了自己的电子邮箱，在你第一次回复该问题后，提问人将会收到一封邮件通知。</p>
            </div>

            <div class="uk-margin uk-text-center">
              <button type="submit" class="uk-button uk-button-primary" :disabled="isSubmitting">
                <span v-if="isSubmitting">提交中...</span>
                <span v-else>
                  <span v-if="question.answer">更新</span>回答
                </span>
              </button>
            </div>
          </Form>
        </div>

        <div v-else-if="!isLoading">
          <h5 class="uk-text-center">再问点别的问题？</h5>
          <NewQuestion
              :page-profile-domain="profile.domain"
              :harassment-setting="profile.harassmentSetting"
          />
        </div>
        <Skeleton v-else :count="3"></Skeleton>

        <hr class="uk-divider-icon">
        <Skeleton :loading="isInitalLoading" :count="5">
          <PageQuestions
              :page-profile-name="profile.name"
              :page-profile-domain="profile.domain"
          />
        </Skeleton>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import {ref, onMounted, watch} from "vue";
import PageQuestions from "@/components/PageQuestions.vue";
import {getUserProfile, getUserQuestion, type PageQuestion, type Profile} from "@/api/user.ts";
import {useRoute, useRouter} from "vue-router";
import {humanizeDate} from "@/utils/humanize.ts";
import NewQuestion from "@/components/NewQuestion.vue";
import {answerQuestion, type AnswerQuestionRequest, deleteQuestion, setQuestionVisible} from "@/api/mine.ts";
import {ToastSuccess} from "@/utils/notify.ts";
import {Form} from "vee-validate";
import {Skeleton} from "vue-loading-skeleton";

const route = useRoute()
const router = useRouter()
const domain = ref<string>(route.params.domain as string || '')
const questionID = ref<string>(route.params.questionID as string || '')
const questionToken = ref<string>(route.query.t as string || '')

const isLoading = ref<boolean>(true)
const profile = ref<Profile>({} as Profile)
const EMPTY_QUESTION: PageQuestion = {
  id: 0,
  isOwner: false,
  createdAt: null,
  answeredAt: null,
  content: '',
  answer: '',
  questionImageURLs: [],
  answerImageURLs: [],
  isPrivate: false,
  hasReplyEmail: false,
}
const question = ref<PageQuestion>(EMPTY_QUESTION)

const isSubmitting = ref<boolean>(false)
const answerQuestionForm = ref<AnswerQuestionRequest>({
  answer: '',
  images: [],
})
const handleAnswer = () => {
  isSubmitting.value = true
  answerQuestion(questionID.value, answerQuestionForm.value)
      .then(res => {
        ToastSuccess(res)
        fetchQuestion()
      })
      .finally(() => {
        isSubmitting.value = false
      })
}

const imageUploader = ref<HTMLInputElement | null>(null)
const handleSelectImage = (event: Event) => {
  const target = event.target as HTMLInputElement
  if (target.files && target.files.length > 0) {
    answerQuestionForm.value.images = Array.from(target.files)
  } else {
    answerQuestionForm.value.images = []
  }
}

const handleRemoveSelectedImage = () => {
  if (imageUploader.value) {
    imageUploader.value.value = ''
  }
  answerQuestionForm.value.images = []
}

const handleDelete = () => {
  deleteQuestion(questionID.value).then(res => {
    ToastSuccess(res)
    router.push({
      name: 'profile',
      params: {
        domain: domain.value
      }
    })
  })
}

const handleSetVisible = () => {
  // isPrivate -> set public -> true
  // !isPrivate -> set private -> false
  setQuestionVisible(questionID.value, question.value.isPrivate).then(res => {
    question.value.isPrivate = !question.value.isPrivate
    ToastSuccess(res)
  })
}

watch(() => [route.params, route.query], () => {
  domain.value = route.params.domain as string || ''
  questionID.value = route.params.questionID as string || ''
  questionToken.value = route.query.t as string || ''
  fetchQuestion()
})

const fetchQuestion = () => {
  isLoading.value = true
  question.value = { ...EMPTY_QUESTION }

  getUserQuestion(domain.value, questionID.value, questionToken.value)
      .then(res => {
        question.value = res
        answerQuestionForm.value.answer = res.answer || ''
      })
      .catch(err => {
        const statusCode = err.response?.status
        if (statusCode === 404) {
          router.push({name: 'home'})
        }
      })
      .finally(() => {
        isLoading.value = false
      })
}

const isInitalLoading = ref<boolean>(true)
onMounted(() => {
  isInitalLoading.value = true
  getUserProfile(domain.value)
      .then(res => {
        profile.value = res
      })
      .catch(err => {
        const statusCode = err.response?.status
        if (statusCode === 404) {
          router.push({name: 'home'})
        }
      })
      .finally(() => {
        isInitalLoading.value = false
      })

  fetchQuestion()
})
</script>

<style scoped>

</style>
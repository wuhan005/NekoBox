<template>
  <Skeleton :loading="isInitLoading" :count="5"></Skeleton>
  <div v-if="!isInitLoading && questionCount > 0">
    <p class="uk-text-left uk-text-muted uk-text-small">@{{ props.pageProfileName }} 以前回答过的问题 ({{
        questionCount
      }})</p>

    <div id="question-list">
      <div v-for="question in questions" v-bind:key="question.id">
        <hr>
        <a class="uk-button uk-button-default uk-button-small uk-float-right"
           @click="handleViewQuestion(question)">查看回答</a>
        <div class="uk-text-left uk-text-small uk-text-muted">{{ humanizeDate(question.createdAt) }}</div>
        <p class="uk-text-small uk-text-break">{{ question.content }}</p>
      </div>
    </div>

    <div>
      <button v-if="hasMore" type="button" class="uk-button uk-button-default uk-width-1-1 uk-margin-small-bottom"
              :disabled="isLoading"
              @click="fetchQuestions">
        <span v-if="hasMore && !isLoading">加载更多</span>
        <span v-if="isLoading">加载中...</span>
      </button>
      <div v-else class="uk-text-meta uk-text-center">
        <hr>
        无更多提问
        <br><br>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import {defineProps, watch, ref} from "vue";
import {getUserQuestions, type PageQuestionItem} from '@/api/user.ts'
import {humanizeDate} from "@/utils/humanize.ts";
import {useRouter} from "vue-router";
import {Skeleton} from "vue-loading-skeleton";

const props = defineProps({
  pageProfileName: {
    type: String,
    required: true
  },
  pageProfileDomain: {
    type: String,
    required: true
  }
})

const router = useRouter()

const PAGE_SIZE = 20
const isInitLoading = ref<boolean>(true)
const isLoading = ref<boolean>(false)
const hasMore = ref<boolean>(true)
const questionCursor = ref<string>('')
const questions = ref<PageQuestionItem[]>([])
const questionCount = ref<number>(0);

const fetchQuestions = () => {
  isLoading.value = true

  getUserQuestions(props.pageProfileDomain, questionCursor.value, PAGE_SIZE)
      .then(res => {
        questionCount.value = res.total
        questionCursor.value = res.cursor
        questions.value = questions.value.concat(res.questions)
        if (res.questions.length < PAGE_SIZE) {
          hasMore.value = false
        }
      })
      .finally(() => {
        isLoading.value = false
        isInitLoading.value = false
      })
}

const handleViewQuestion = (question: PageQuestionItem) => {
  router.push({
    name: 'question',
    params: {
      domain: props.pageProfileDomain,
      questionID: question.id
    }
  })
}

watch(() => props.pageProfileDomain, (domain) => {
  if (domain) {
    questions.value = []
    questionCursor.value = ''
    hasMore.value = true
    fetchQuestions()
  }
}, {immediate: true})
</script>

<style scoped>

</style>
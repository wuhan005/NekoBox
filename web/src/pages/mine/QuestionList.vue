<template>
  <Skeleton :count="3" :loading="isInitLoading"></Skeleton>
  <div v-if="!isInitLoading">
    <a v-for="(question) in questions" v-bind:key="question.id" @click="handleView(question)">
      <div>
        <hr>
        <span v-if="!question.isAnswered" class="uk-label uk-float-right uk-margin-small-right">未回答</span>
        <span v-if="question.isPrivate"
              class="uk-label uk-label-warning uk-float-right uk-margin-small-right">私密</span>
        <div class="uk-text-left uk-text-small uk-text-muted">{{ humanizeDate(question.createdAt) }}</div>
        <p class="uk-text-small">{{ question.content }}</p>
      </div>
    </a>
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
</template>

<script setup lang="ts">
import {ref, onMounted} from "vue";
import {type MineQuestionItem, mineQuestions} from "@/api/mine.ts";
import {useRouter} from "vue-router";
import {humanizeDate} from "@/utils/humanize.ts";
import {useAuthStore} from "@/store";
import {Skeleton} from "vue-loading-skeleton";

const router = useRouter()
const authStore = useAuthStore()

const PAGE_SIZE = 20
const isInitLoading = ref<boolean>(true)
const isLoading = ref<boolean>(false)
const hasMore = ref<boolean>(true)
const questionCursor = ref<string>('')
const questions = ref<MineQuestionItem[]>([])

const fetchQuestions = () => {
  isLoading.value = true
  mineQuestions(questionCursor.value, PAGE_SIZE)
      .then(res => {
        questions.value = questions.value.concat(res.questions)
        questionCursor.value = res.cursor
        if (res.questions.length < PAGE_SIZE) {
          hasMore.value = false
        }
      })
      .finally(() => {
        isLoading.value = false
        isInitLoading.value = false
      })
}

const handleView = (question: MineQuestionItem) => {
  router.push({
    name: 'question',
    params: {
      domain: authStore.profile.domain,
      questionID: question.id
    }
  })
}

onMounted(() => {
  fetchQuestions()
})
</script>

<style scoped>

</style>
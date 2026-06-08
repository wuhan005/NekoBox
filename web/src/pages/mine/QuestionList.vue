<template>
  <UkTabs v-model="currentTab" :tabs="TABS">
    <template #received>
      <Skeleton :count="3" :loading="received.isInitLoading"></Skeleton>
      <div v-if="!received.isInitLoading">
        <a
            v-for="(question, index) in received.questions"
            :key="question.id"
            :href="router.resolve({name: 'question', params: {domain: authStore.profile.domain, questionID: question.id}}).href"
            @click.prevent="handleViewReceived(question)"
        >
          <div>
            <hr v-if="index > 0">
            <span v-if="!question.isAnswered" class="uk-label uk-float-right uk-margin-small-right">未回答</span>
            <span v-if="question.isPrivate" class="uk-label uk-label-warning uk-float-right uk-margin-small-right">私密</span>
            <div class="uk-text-left uk-text-small uk-text-muted">{{ humanizeDate(question.createdAt) }}</div>
            <p class="uk-text-small">{{ question.content }}</p>
          </div>
        </a>
      </div>

      <div>
        <button
            v-if="received.hasMore"
            type="button"
            class="uk-button uk-button-default uk-width-1-1 uk-margin-small-bottom"
            :disabled="received.isLoading"
            @click="fetchReceivedQuestions"
        >
          <span v-if="!received.isLoading">加载更多</span>
          <span v-else>加载中...</span>
        </button>
        <div v-else class="uk-text-meta uk-text-center">
          <hr>
          无更多提问
          <br><br>
        </div>
      </div>
    </template>

    <template #sent>
      <Skeleton :count="3" :loading="sent.isInitLoading"></Skeleton>
      <div v-if="!sent.isInitLoading">
        <a
            v-for="(question, index) in sent.questions"
            :key="question.id"
            :href="sentQuestionHref(question)"
            @click.prevent="handleViewSent(question)"
        >
          <div>
            <hr v-if="index > 0">
            <span v-if="!question.isAnswered" class="uk-label uk-float-right uk-margin-small-right">未回答</span>
            <span v-if="question.isPrivate" class="uk-label uk-label-warning uk-float-right uk-margin-small-right">私密</span>
            <div class="uk-text-left uk-text-small uk-text-muted">
              {{ humanizeDate(question.createdAt) }} · @{{ question.targetDomain || '已失效提问箱' }}
            </div>
            <p class="uk-text-small">{{ question.content }}</p>
          </div>
        </a>
      </div>

      <div>
        <button
            v-if="sent.hasMore"
            type="button"
            class="uk-button uk-button-default uk-width-1-1 uk-margin-small-bottom"
            :disabled="sent.isLoading"
            @click="fetchSentQuestions"
        >
          <span v-if="!sent.isLoading">加载更多</span>
          <span v-else>加载中...</span>
        </button>
        <div v-else class="uk-text-meta uk-text-center">
          <hr>
          无更多提问
          <br><br>
        </div>
      </div>
    </template>
  </UkTabs>
</template>

<script setup lang="ts">
import {computed, reactive, ref, onMounted} from "vue";
import {type MineQuestionItem, type MineSentQuestionItem, mineQuestions, mineSentQuestions} from "@/api/mine.ts";
import {useRouter} from "vue-router";
import {humanizeDate} from "@/utils/humanize.ts";
import {useAuthStore} from "@/store";
import {Skeleton} from "vue-loading-skeleton";
import UkTabs from "@/components/UkTabs.vue";

const router = useRouter()
const authStore = useAuthStore()

const PAGE_SIZE = 20
const receivedTotal = ref<number>(0)
const sentTotal = ref<number>(0)
const TABS = computed(() => [
  {name: 'received', label: `我收到的问题 (${receivedTotal.value})`},
  {name: 'sent', label: `我的提问 (${sentTotal.value})`},
])

const currentTab = ref<string>('received')

interface ListState<T> {
  isInitLoading: boolean;
  isLoading: boolean;
  hasMore: boolean;
  cursor: string;
  questions: T[];
}

const received = reactive<ListState<MineQuestionItem>>({
  isInitLoading: true,
  isLoading: false,
  hasMore: true,
  cursor: '',
  questions: [],
})

const sent = reactive<ListState<MineSentQuestionItem>>({
  isInitLoading: true,
  isLoading: false,
  hasMore: true,
  cursor: '',
  questions: [],
})

const fetchReceivedQuestions = () => {
  received.isLoading = true
  mineQuestions(received.cursor, PAGE_SIZE)
      .then(res => {
        receivedTotal.value = res.total
        received.questions = received.questions.concat(res.questions)
        received.cursor = res.cursor
        if (res.questions.length < PAGE_SIZE) {
          received.hasMore = false
        }
      })
      .finally(() => {
        received.isLoading = false
        received.isInitLoading = false
      })
}

const fetchSentQuestions = () => {
  sent.isLoading = true
  mineSentQuestions(sent.cursor, PAGE_SIZE)
      .then(res => {
        sentTotal.value = res.total
        sent.questions = sent.questions.concat(res.questions)
        sent.cursor = res.cursor
        if (res.questions.length < PAGE_SIZE) {
          sent.hasMore = false
        }
      })
      .finally(() => {
        sent.isLoading = false
        sent.isInitLoading = false
      })
}

const handleViewReceived = (question: MineQuestionItem) => {
  router.push({
    name: 'question',
    params: {
      domain: authStore.profile.domain,
      questionID: question.id
    }
  })
}

const sentQuestionHref = (question: MineSentQuestionItem) => {
  if (!question.targetDomain) {
    return '#'
  }

  return router.resolve({
    name: 'question',
    params: {
      domain: question.targetDomain,
      questionID: question.id,
    }
  }).href
}

const handleViewSent = (question: MineSentQuestionItem) => {
  if (!question.targetDomain) {
    return
  }

  router.push({
    name: 'question',
    params: {
      domain: question.targetDomain,
      questionID: question.id,
    }
  })
}

onMounted(() => {
  fetchReceivedQuestions()
  fetchSentQuestions()
})
</script>

<style scoped>

</style>
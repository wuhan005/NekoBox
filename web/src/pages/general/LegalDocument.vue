<template>
  <article class="uk-card uk-card-default uk-card-body uk-text-break legal-document" v-html="documents[page]"></article>
</template>

<script setup lang="ts">
import terms from '@/content/terms.md'
import privacy from '@/content/privacy.md'
import deactivationTerms from '@/content/deactivate-account-terms.md'

const documents = {terms, privacy, deactivationTerms}

defineProps<{page: keyof typeof documents}>()
</script>

<style scoped>
.legal-document {
  font-size: 0.9375rem;
  line-height: 1.75;
}

.legal-document :deep(:is(h1, h2, h3, h4, h5, h6)) {
  font-size: 1rem;
  font-weight: 600;
  line-height: 1.4;
  margin-bottom: 0.75rem;
}

.legal-document :deep(h1) {
  font-size: 1.25rem;
}

.legal-document :deep(h1:first-child) {
  font-size: 1.625rem;
}

.legal-document :deep(h2) {
  font-size: 1.125rem;
}

.legal-document :deep(h1 + h2) {
  margin-top: 0.75rem;
}

@media (max-width: 639px) {
  .legal-document {
    padding: 1.25rem;
  }
}
</style>

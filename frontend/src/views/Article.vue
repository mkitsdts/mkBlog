<template>
  <div class="article-detail" v-if="article">
    <h1 class="title">{{ article.title }}</h1>
    <div class="meta">
      <span>作者：{{ article.author }}</span>
      <span class="dot" />
      <span>更新时间：{{ article.updateAt }}</span>
    </div>
    <el-divider />
    <div class="content markdown-body" v-html="html"></div>
    <el-divider />
    <div class="footer">
      <el-button type="primary" link @click="$router.back()">返回</el-button>
    </div>
  </div>
  <div v-else class="loading">加载中...</div>
 </template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import api from '@/api'
import MarkdownIt from 'markdown-it'
import hljs from 'highlight.js'
import 'highlight.js/styles/github.min.css'

const route = useRoute()
const article = ref<any>(null)
const html = ref('')

const rawEscape = new MarkdownIt().utils.escapeHtml
const md = new MarkdownIt({
  html: false,
  linkify: true,
  typographer: true,
  highlight(code: string, lang: string): string {
    if (lang && hljs.getLanguage(lang)) {
      try {
        return '<pre class="hljs"><code>' + hljs.highlight(code, { language: lang, ignoreIllegals: true }).value + '</code></pre>'
      } catch {

      }
    }
    return '<pre class="hljs"><code>' + rawEscape(code) + '</code></pre>'
  }
})

onMounted(async () => {
  const title = route.params.title
  try {
    const res = await api.getArticleDetail(title as string)
    article.value = res.data
    const raw = article.value.content || ''
    html.value = md.render(stripDuplicateHeading(raw, article.value.title))
  } catch (e) {
    article.value = { title: '未找到', author: '', updateAt: '', content: '' }
    html.value = '<p>文章不存在</p>'
  }
})

function stripDuplicateHeading(raw: string, title: string): string {
  if (!raw || !title) return raw
  const lines = raw.split(/\r?\n/)
  let i = 0
  while (i < lines.length && lines[i].trim() === '') i++
  if (i < lines.length) {
    const m = lines[i].match(/^#{1,6}\s+(.*)$/)
    if (m) {
      const headingText = m[1].trim()
      if (headingText === title.trim()) {
        lines.splice(i, 1)
        if (i < lines.length && lines[i].trim() === '') lines.splice(i, 1)
        return lines.join('\n')
      }
    }
  }
  return raw
}
</script>

<style scoped>
.article-detail { padding: 28px 32px; max-width: 860px; margin: 0 auto; background: #fff; border-radius: 12px; box-shadow: 0 4px 18px rgba(0,0,0,0.05); }
.title { margin:0 0 8px; font-size: 2.1rem; font-weight: 600; line-height:1.25; }
.meta { color:#666; font-size: 13px; display:flex; align-items:center; gap:12px; margin-bottom: 8px; }
.meta .dot { width:4px; height:4px; background:#bbb; border-radius:50%; display:inline-block; }
.content { line-height:1.7; font-size:16px; color:#222; }
.content :deep(h2) { margin-top:2.2em; padding-bottom: .3em; border-bottom:1px solid #eee; font-size:1.5rem; }
.content :deep(pre) { background:#f6f8fa; padding:14px 16px; border-radius:8px; overflow:auto; font-size: 14px; }
.content :deep(code) { font-family: Menlo, Monaco, Consolas, 'Courier New', monospace; }
.content :deep(blockquote) { margin:1em 0; padding: .6em 1em; background:#f8f9fa; border-left:4px solid #d0d7de; color:#555; }
.content :deep(table) { border-collapse:collapse; margin:1.5em 0; }
.content :deep(th), .content :deep(td) { border:1px solid #e2e2e2; padding:6px 10px; }
.content :deep(img) { max-width:100%; border-radius:6px; }
.footer { text-align:right; margin-top:20px; }
.loading { padding:40px; text-align:center; color:#666; }
body { background:#f2f3f5; }
</style>

<template>
  <div class="article-detail" v-if="article">
    <h1 class="title">{{ article.title }}</h1>
    <div class="meta">
      <span>作者：{{ article.author }}</span>
      <span class="dot"></span>
      <span>更新时间：{{ formatDate(article.updateAt || article.UpdateAt) }}</span>
    </div>
    <el-divider />
    <div class="content markdown-body" v-html="html"></div>
    <el-divider />
    <!-- comment area -->
    <section class="comment-section" v-if="commentEnabled">
      <h2 class="comment-title">Comment</h2>
      <div class="comment-form" v-if="article">
        <el-input v-model="form.user" size="small" placeholder="Nickname" class="nick-input" />
        <el-input
          v-model="form.content"
          type="textarea"
          :rows="3"
          placeholder="Share your thoughts..."
          resize="none"
          class="content-input"
        />
        <div class="form-actions">
          <el-button type="primary" size="small" @click="submitComment()" :loading="submitting">发布</el-button>
          <el-button v-if="replyingTo" size="small" text @click="cancelReply">取消回复</el-button>
          <span v-if="replyingTo" class="replying-hint">正在回复 #{{ replyingTo.order }} @{{ replyingTo.comment_user }}</span>
        </div>
      </div>
      <div class="comment-list" v-loading="loadingComments">
        <div v-if="!loadingComments && flatComments.length === 0" class="empty">还没有评论，来抢沙发～</div>
        <ul class="root-list" v-else>
          <CommentNode
            v-for="node in tree"
            :key="node.order"
            :node="node"
            @reply="startReply"
          />
        </ul>
      </div>
    </section>
    <div class="footer">
      <el-button type="primary" link @click="$router.back()">返回</el-button>
    </div>
  </div>
  <div v-else class="loading">加载中...</div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import CommentNode from '@/components/CommentNode.vue'
import { loadConfig } from '@/config'
import { useRoute } from 'vue-router'
import api from '@/api'
import MarkdownIt from 'markdown-it'
import hljs from 'highlight.js'
import 'highlight.js/styles/github.min.css'

interface RawComment {
  id: number
  content: string
  comment_user: string
  comment_to_order: number
  title: string
  order: number
  created_at?: string
}

interface CommentNodeType extends RawComment {
  children: CommentNodeType[]
}

const route = useRoute()
const article = ref<any>(null)
const html = ref('')
const commentEnabled = ref(true)
const loadingComments = ref(false)
const flatComments = ref<RawComment[]>([])
const tree = ref<CommentNodeType[]>([])
const submitting = ref(false)
const replyingTo = ref<RawComment | null>(null)
const form = ref({ user: localStorage.getItem('comment_user') || '', content: '' })

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

function formatDate(dt?: string) {
  if (!dt) return ''
  try {
    const d = new Date(dt)
    if (isNaN(d.getTime())) return String(dt).replace('T',' ').substring(0,19)
    const pad = (n: number) => String(n).padStart(2,'0')
    return `${d.getFullYear()}-${pad(d.getMonth()+1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
  } catch {
    return String(dt).replace('T',' ').substring(0,19)
  }
}

onMounted(async () => {
  const title = route.params.title
  try {
    const site = await loadConfig()
    commentEnabled.value = !!site.comment_enabled
  } catch { commentEnabled.value = true }
  try {
    const res = await api.getArticleDetail(title as string)
    article.value = res.data
    const raw = article.value.content || ''
    html.value = md.render(stripDuplicateHeading(raw, article.value.title))
    if (commentEnabled.value) await fetchComments()
  } catch (e) {
    article.value = { title: '未找到', author: '', updateAt: '', content: '' }
    html.value = '<p>文章不存在</p>'
  }
})

async function fetchComments() {
  if (!article.value) return
  loadingComments.value = true
  try {
  const res = await (api as any).getComments(article.value.title)
    const list: RawComment[] = res.data.comments || []
    flatComments.value = list.sort((a, b) => a.order - b.order)
    buildTree()
  } catch (e) {
    // ignore
  } finally {
    loadingComments.value = false
  }
}

function buildTree() {
  const map = new Map<number, CommentNodeType>()
  const roots: CommentNodeType[] = []
  flatComments.value.forEach(c => {
    map.set(c.order, { ...c, children: [] })
  })
  map.forEach(node => {
    if (node.comment_to_order === -1) {
      roots.push(node)
    } else {
      const parent = map.get(node.comment_to_order)
      if (parent) parent.children.push(node)
      else roots.push(node) // fallback if invalid reference
    }
  })
  // optional: sort children by order
  const sortRec = (arr: CommentNodeType[]) => {
    arr.sort((a, b) => a.order - b.order)
    arr.forEach(n => sortRec(n.children))
  }
  sortRec(roots)
  tree.value = roots
}

function startReply(node: RawComment) {
  replyingTo.value = node
  form.value.content = `@${node.comment_user} `
}
function cancelReply() {
  replyingTo.value = null
}

async function submitComment() {
  if (!form.value.user.trim() || !form.value.content.trim() || !article.value) return
  submitting.value = true
  try {
  await (api as any).addComment({
      comment_user: form.value.user.trim(),
      content: form.value.content.trim(),
      comment_to: replyingTo.value ? replyingTo.value.order : -1,
      title: article.value.title
    })
    localStorage.setItem('comment_user', form.value.user.trim())
    form.value.content = ''
    replyingTo.value = null
    await fetchComments()
  } catch (e) {
    // add global message prompt if needed
  } finally {
    submitting.value = false
  }
}


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

// 评论组件已在外部文件中实现
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
.comment-section { margin-top: 28px; }
.comment-title { font-size: 20px; margin:0 0 12px; font-weight:600; }
.comment-form { background:#f7f8fa; padding:14px 16px 12px; border:1px solid #e5e6eb; border-radius:8px; margin-bottom:18px; }
.comment-form .nick-input { margin-bottom:8px; max-width:220px; }
.comment-form .content-input { margin-bottom:8px; }
.comment-form .form-actions { display:flex; align-items:center; gap:12px; font-size:12px; color:#666; }
.comment-form .replying-hint { color:#999; }
.comment-list { list-style:none; padding:0; margin:0; }
.root-list { list-style:none; padding:0; margin:0; }
.comment-item { list-style:none; padding:12px 0 6px; border-bottom:1px solid #f0f0f0; }
.comment-item:last-child { border-bottom:none; }
.comment-item .c-head { display:flex; align-items:center; gap:10px; font-size:13px; }
.comment-item .nick { font-weight:600; color:#333; }
.comment-item .order { color:#999; }
.comment-item .time { color:#bbb; font-size:12px; }
.comment-item .c-content { margin-top:4px; line-height:1.6; white-space:pre-wrap; word-break:break-word; }
.child-list { margin:6px 0 0 16px; padding-left:12px; border-left:2px solid #f0f0f0; }
.empty { padding:20px 0; text-align:center; color:#999; }
.footer { text-align:right; margin-top:20px; }
.loading { padding:40px; text-align:center; color:#666; }
body { background:#f2f3f5; }
</style>

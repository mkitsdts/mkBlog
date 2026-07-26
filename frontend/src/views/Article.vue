<template>
  <div class="article-layout" :class="{ 'has-toc': toc.length }" v-if="article">
    <main class="article-detail">
      <h1 class="title">{{ article.title }}</h1>
      <div class="meta">
        <span>作者：{{ article.author }}</span>
        <span class="dot"></span>
        <span>创建时间：{{ formatDate(article.createAt || article.CreateAt) }}</span>
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
    </main>
    <aside v-if="toc.length" class="toc-panel" aria-label="文章目录">
      <nav class="toc-nav">
        <h2 class="toc-title">目录</h2>
        <a
          v-for="item in toc"
          :key="item.id"
          class="toc-link"
          :class="`toc-level-${item.level}`"
          :href="`#${item.id}`"
          @click.prevent="scrollToHeading(item.id)"
        >
          {{ item.text }}
        </a>
      </nav>
    </aside>
  </div>
  <div v-else class="loading">加载中...</div>
</template>

<script setup lang="ts">
import { nextTick, onMounted, ref } from 'vue'
import CommentNode from '@/components/CommentNode.vue'
import { loadConfig } from '@/config'
import { useRoute } from 'vue-router'
import api from '@/api'
import MarkdownIt from 'markdown-it'
import type Token from 'markdown-it/lib/token.mjs'
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

interface TocItem {
  id: string
  text: string
  level: number
}

const route = useRoute()
const article = ref<any>(null)
const html = ref('')
const toc = ref<TocItem[]>([])
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
    const rendered = renderArticle(stripDuplicateHeading(raw, article.value.title))
    html.value = rendered.html
    toc.value = rendered.toc
    await nextTick()
    scrollToInitialHash()
    if (commentEnabled.value) await fetchComments()
  } catch (e) {
    article.value = { title: '未找到', author: '', createAt: '', updateAt: '', content: '' }
    html.value = '<p>文章不存在</p>'
    toc.value = []
  }
})

function renderArticle(raw: string): { html: string; toc: TocItem[] } {
  const env = {}
  const tokens = md.parse(raw, env)
  const items: TocItem[] = []
  const usedIds = new Map<string, number>()

  for (let i = 0; i < tokens.length; i++) {
    const token = tokens[i]
    if (token.type !== 'heading_open') continue

    const inlineToken = tokens[i + 1]
    if (!inlineToken || inlineToken.type !== 'inline') continue

    const text = headingText(inlineToken).trim()
    if (!text) continue

    const baseId = slugify(text) || `section-${items.length + 1}`
    const count = usedIds.get(baseId) || 0
    usedIds.set(baseId, count + 1)
    const id = count === 0 ? baseId : `${baseId}-${count + 1}`
    token.attrSet('id', id)

    items.push({
      id,
      text,
      level: Number(token.tag.slice(1))
    })
  }

  return {
    html: md.renderer.render(tokens, md.options, env),
    toc: items
  }
}

function headingText(token: Token): string {
  return (token.children || [])
    .filter(child => child.type !== 'html_inline')
    .map(child => child.content)
    .join('')
}

function slugify(text: string): string {
  return text
    .normalize('NFKC')
    .toLowerCase()
    .trim()
    .replace(/\s+/g, '-')
    .replace(/[^\p{Letter}\p{Number}\p{Mark}_-]/gu, '')
    .replace(/-+/g, '-')
    .replace(/^-|-$/g, '')
}

function scrollToHeading(id: string) {
  document.getElementById(id)?.scrollIntoView({ behavior: 'smooth', block: 'start' })
  window.history.replaceState(null, '', `${window.location.pathname}${window.location.search}#${encodeURIComponent(id)}`)
}

function scrollToInitialHash() {
  const hash = window.location.hash.slice(1)
  if (!hash) return
  try {
    document.getElementById(decodeURIComponent(hash))?.scrollIntoView({ block: 'start' })
  } catch {
    // Ignore malformed URL hashes.
  }
}

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
.article-layout { display:grid; grid-template-columns:minmax(0, 860px); align-items:start; max-width:860px; margin:0 auto; padding:0 24px; }
.article-layout.has-toc { grid-template-columns:minmax(0, 860px) 240px; gap:24px; max-width:1124px; }
.article-detail { box-sizing:border-box; width:100%; padding:28px 32px; background:#fff; border-radius:12px; box-shadow:0 4px 18px rgba(0,0,0,0.05); }
.title { margin:0 0 8px; font-size: 2.1rem; font-weight: 600; line-height:1.25; }
.meta { color:#666; font-size:13px; display:flex; align-items:center; flex-wrap:wrap; gap:8px 12px; margin-bottom:8px; }
.meta .dot { width:4px; height:4px; background:#bbb; border-radius:50%; display:inline-block; }
.content { line-height:1.7; font-size:16px; color:#222; }
.content :deep(h1), .content :deep(h2), .content :deep(h3), .content :deep(h4), .content :deep(h5), .content :deep(h6) { scroll-margin-top:24px; }
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
.toc-panel { position:sticky; top:24px; box-sizing:border-box; max-height:calc(100vh - 48px); overflow:auto; padding:18px 16px; background:rgba(255,255,255,.92); border-radius:12px; box-shadow:0 4px 18px rgba(0,0,0,.05); }
.toc-nav { display:flex; flex-direction:column; }
.toc-title { margin:0 0 12px; font-size:16px; font-weight:600; color:#303133; }
.toc-link { display:block; padding:5px 0; color:#606266; font-size:13px; line-height:1.45; text-decoration:none; overflow-wrap:anywhere; transition:color .2s; }
.toc-link:hover, .toc-link:focus-visible { color:#409eff; }
.toc-level-2 { padding-left:12px; }
.toc-level-3 { padding-left:24px; }
.toc-level-4 { padding-left:36px; }
.toc-level-5 { padding-left:48px; }
.toc-level-6 { padding-left:60px; }
.loading { padding:40px; text-align:center; color:#666; }
body { background:#f2f3f5; }

@media (max-width: 960px) {
  .article-layout { display:block; max-width:860px; padding:0 16px; }
  .toc-panel { display:none; }
}

@media (max-width: 600px) {
  .article-layout { padding:0 10px; }
  .article-detail { padding:22px 18px; }
  .title { font-size:1.8rem; }
  .meta .dot { display:none; }
  .meta span:not(.dot) { width:100%; }
}
</style>

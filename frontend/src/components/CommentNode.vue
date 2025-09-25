<template>
  <li class="comment-item">
    <div class="comment-body">
      <div class="c-head">
        <span class="nick">{{ node.comment_user }}</span>
        <span class="order">#{{ node.order }}</span>
        <span v-if="node.created_at" class="time">{{ formatTime(node.created_at) }}</span>
        <el-button link size="small" @click="$emit('reply', node)">回复</el-button>
      </div>
      <div class="c-content">{{ node.content }}</div>
    </div>
    <ul v-if="node.children.length" class="child-list">
      <CommentNode v-for="ch in node.children" :key="ch.order" :node="ch" @reply="$emit('reply', $event)" />
    </ul>
  </li>
</template>

<script setup lang="ts">
import type { PropType } from 'vue'

export interface RawComment {
  id: number
  content: string
  comment_user: string
  comment_to_order: number
  title: string
  order: number
  created_at?: string
}
export interface CommentNodeType extends RawComment { children: CommentNodeType[] }

const props = defineProps({
  node: { type: Object as PropType<CommentNodeType>, required: true }
})

function formatTime(t?: string) {
  if (!t) return ''
  try {
    const d = new Date(t)
    if (isNaN(d.getTime())) return ''
    return `${d.getFullYear()}-${String(d.getMonth()+1).padStart(2,'0')}-${String(d.getDate()).padStart(2,'0')} ${String(d.getHours()).padStart(2,'0')}:${String(d.getMinutes()).padStart(2,'0')}:${String(d.getSeconds()).padStart(2,'0')}`
  } catch { return '' }
}
</script>

<style scoped>
.comment-item { list-style:none; padding:12px 0 6px; border-bottom:1px solid #f0f0f0; }
.comment-item:last-child { border-bottom:none; }
.comment-item .c-head { display:flex; align-items:center; gap:10px; font-size:13px; }
.comment-item .nick { font-weight:600; color:#333; }
.comment-item .order { color:#999; }
.comment-item .time { color:#bbb; font-size:12px; }
.comment-item .c-content { margin-top:4px; line-height:1.6; white-space:pre-wrap; word-break:break-word; }
.child-list { margin:6px 0 0 16px; padding-left:12px; border-left:2px solid #f0f0f0; }
</style>

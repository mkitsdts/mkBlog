import axios from 'axios';
import type { AxiosInstance, AxiosResponse } from 'axios';
import { loadConfig } from '@/config';

// --- Data shape definitions (can be refined later) ---
export interface ArticleSummary { title: string; summary?: string; updateAt?: string; UpdateAt?: string }
export interface ArticleDetail { title: string; content: string; author?: string; updateAt?: string }
export interface CommentItem {
  id: number;
  content: string;
  comment_user: string;
  comment_to_order: number; // -1 means root
  title: string;
  order: number;
  created_at?: string;
}

export interface ArticlesResponse { articles: ArticleSummary[]; total: number }
export interface CommentsResponse { comments: CommentItem[] }

let clientPromise: Promise<AxiosInstance> | null = null;
async function getClient(): Promise<AxiosInstance> {
  if (!clientPromise) {
    clientPromise = (async () => {
      const site: any = await loadConfig();
      const isDev = typeof import.meta !== 'undefined' && (import.meta as any).env && (import.meta as any).env.DEV;
      const baseRoot = isDev ? '' : (site.server || '');
      const base = (baseRoot || '').replace(/\/$/, '');
      return axios.create({
        baseURL: `${base}/api`,
        headers: { 'Content-Type': 'application/json' },
      });
    })();
  }
  return clientPromise;
}

async function getArticles(page: number, pageSize: number, categories?: string[] | string, q?: string): Promise<AxiosResponse<ArticlesResponse>> {
  const apiClient = await getClient();
  const params = new URLSearchParams({ page: String(page), pageSize: String(pageSize) });
  if (Array.isArray(categories) && categories.length) {
    params.append('categories', categories.join(','));
  } else if (typeof categories === 'string' && categories) {
    params.append('category', categories);
  }
  if (q && typeof q === 'string') params.append('q', q);
  return apiClient.get(`/articles?${params.toString()}`);
}

async function getArticleDetail(title: string): Promise<AxiosResponse<ArticleDetail>> {
  const apiClient = await getClient();
  return apiClient.get(`/article/${encodeURIComponent(title)}`);
}

async function getCategories(): Promise<AxiosResponse<{ categories: string[] }>> {
  const apiClient = await getClient();
  return apiClient.get('/categories');
}

async function searchArticles(q: string, page: number, pageSize: number): Promise<AxiosResponse<ArticlesResponse>> {
  const apiClient = await getClient();
  const params = new URLSearchParams({ q: String(q || ''), page: String(page || 1), pageSize: String(pageSize || 10) });
  return apiClient.get(`/search?${params.toString()}`);
}

async function getFriends(): Promise<AxiosResponse<any>> { // refine later
  const apiClient = await getClient();
  return apiClient.get('/friends');
}

async function applyFriend(data: any): Promise<AxiosResponse<any>> { // refine later
  const apiClient = await getClient();
  return apiClient.post('/friends', data);
}

async function getComments(title: string): Promise<AxiosResponse<CommentsResponse>> {
  const apiClient = await getClient();
  const params = new URLSearchParams({ title: String(title || '') });
  return apiClient.get(`/comments?${params.toString()}`);
}

interface AddCommentPayload { comment_user: string; content: string; comment_to: number; title: string }
async function addComment(data: AddCommentPayload): Promise<AxiosResponse<any>> {
  const apiClient = await getClient();
  return apiClient.post('/comments', data);
}

export const api = {
  getArticles,
  getArticleDetail,
  getCategories,
  searchArticles,
  getFriends,
  applyFriend,
  getComments,
  addComment,
};

export type ApiType = typeof api;
export default api;

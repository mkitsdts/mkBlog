import axios from 'axios';
import { loadConfig } from '@/config';

let clientPromise = null;
async function getClient() {
  if (!clientPromise) {
    clientPromise = (async () => {
      const site = await loadConfig();
      // 开发环境使用同源相对路径，走 Vite 代理；生产使用 config.yaml 指定的 server
      const isDev = typeof import.meta !== 'undefined' && import.meta.env && import.meta.env.DEV;
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

export default {
  async getArticles(page, pageSize, categories, q) {
    const apiClient = await getClient();
    const params = new URLSearchParams({ page: String(page), pageSize: String(pageSize) });
    if (Array.isArray(categories) && categories.length) {
      params.append('categories', categories.join(','));
    } else if (typeof categories === 'string' && categories) {
      params.append('category', categories);
    }
    if (q && typeof q === 'string') {
      params.append('q', q);
    }
    return apiClient.get(`/articles?${params.toString()}`);
  },
  async getArticleDetail(title) {
    const apiClient = await getClient();
    return apiClient.get(`/article/${encodeURIComponent(title)}`);
  },
  async getCategories() {
    const apiClient = await getClient();
    return apiClient.get('/categories');
  },
  async searchArticles(q, page, pageSize) {
    const apiClient = await getClient();
    const params = new URLSearchParams({ q: String(q || ''), page: String(page || 1), pageSize: String(pageSize || 10) });
    return apiClient.get(`/search?${params.toString()}`);
  },
  async getFriends() {
    const apiClient = await getClient();
    return apiClient.get('/friends');
  },
  async applyFriend(data) {
    const apiClient = await getClient();
    return apiClient.post('/friends', data);
  },
};

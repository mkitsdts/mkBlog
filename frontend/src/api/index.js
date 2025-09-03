import axios from 'axios';
import { loadConfig } from '@/config';

let clientPromise = null;
async function getClient() {
  if (!clientPromise) {
    clientPromise = (async () => {
      const site = await loadConfig();
      const base = (site.server || '').replace(/\/$/, '');
      return axios.create({
        baseURL: `${base}/api`,
        headers: { 'Content-Type': 'application/json' },
      });
    })();
  }
  return clientPromise;
}

export default {
  async getArticles(page, pageSize, categories) {
    const apiClient = await getClient();
    const params = new URLSearchParams({ page: String(page), pageSize: String(pageSize) });
    if (Array.isArray(categories) && categories.length) {
      params.append('categories', categories.join(','));
    } else if (typeof categories === 'string' && categories) {
      params.append('category', categories);
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
  async getFriends() {
    const apiClient = await getClient();
    return apiClient.get('/friends');
  },
  async applyFriend(data) {
    const apiClient = await getClient();
    return apiClient.post('/friends', data);
  },
};

import axios from 'axios';

const apiClient = axios.create({
  baseURL: 'http://localhost:8080/api', // Adjust to your backend API URL
  headers: {
    'Content-Type': 'application/json',
  },
});

export default {
  getArticles(page, pageSize, categories) {
    const params = new URLSearchParams({ page: String(page), pageSize: String(pageSize) });
    if (Array.isArray(categories) && categories.length) {
      params.append('categories', categories.join(','));
    } else if (typeof categories === 'string' && categories) {
      params.append('category', categories);
    }
    return apiClient.get(`/articles?${params.toString()}`);
  },
  getArticleDetail(title) {
    return apiClient.get(`/article/${encodeURIComponent(title)}`);
  },
  getCategories() {
    return apiClient.get('/categories');
  },
  getFriends() {
    return apiClient.get('/friends');
  },
  applyFriend(data) {
    return apiClient.post('/friends', data);
  },
};

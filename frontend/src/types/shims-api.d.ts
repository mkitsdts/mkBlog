declare module '@/api' {
  interface ArticleSummary { title: string; summary?: string; UpdateAt?: string; updateAt?: string }
  interface ArticleDetail { title: string; content: string; author?: string; updateAt?: string }
  const api: {
    getArticles(page: number, pageSize: number): Promise<{ data: { articles: ArticleSummary[]; total: number } }>
    getArticleDetail(title: string): Promise<{ data: ArticleDetail }>
  getCategories(): Promise<{ data: { categories: string[] } }>
    getFriends(): Promise<any>
    applyFriend(data: any): Promise<any>
  }
  export default api
}

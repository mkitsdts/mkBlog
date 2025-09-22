import * as vscode from 'vscode';

export interface RawArticle {
  id: string | number;
  title: string;
}

export async function fetchArticles(listUrl: string, token?: string): Promise<RawArticle[]> {
  const headers: Record<string, string> = { 'Accept': 'application/json' };
  if (token) headers['Authorization'] = `Bearer ${token}`;
  const res = await fetch(listUrl, { method: 'GET', headers });
  if (!res.ok) {
    const text = await res.text().catch(() => '');
    throw new Error(`HTTP ${res.status} ${res.statusText} ${text}`);
  }
  const data = await res.json();
  // 兼容多种返回结构：
  // 1) 直接数组
  // 2) { articles: [...] }
  // 3) { data: [...] } 或 { data: { articles: [...] } }
  // 4) { items: [...] } / { list: [...] }
  let list: any[] = [];
  if (Array.isArray(data)) {
    list = data;
  } else if (Array.isArray((data as any)?.articles)) {
    list = (data as any).articles;
  } else if (Array.isArray((data as any)?.data)) {
    list = (data as any).data;
  } else if (Array.isArray((data as any)?.data?.articles)) {
    list = (data as any).data.articles;
  } else if (Array.isArray((data as any)?.items)) {
    list = (data as any).items;
  } else if (Array.isArray((data as any)?.list)) {
    list = (data as any).list;
  } else {
    vscode.window.showWarningMessage('mkBlog: 列表接口返回的不是数组，已尝试从 articles/data/items/list 等字段解析但未找到数组');
  }

  return list.map((it, i) => ({
    id: it?.id ?? it?._id ?? it?.slug ?? it?.title ?? i,
    title: it?.title ?? String(it?.id ?? it?._id ?? i),
  }));
}

export async function deleteArticleById(url: string, token?: string): Promise<void> {
  const headers: Record<string, string> = { 'Accept': 'application/json' };
  if (token) headers['Authorization'] = `Bearer ${token}`;
  const res = await fetch(url, { method: 'DELETE', headers });
  if (!res.ok) {
    const text = await res.text().catch(() => '');
    throw new Error(`HTTP ${res.status} ${res.statusText} ${text}`);
  }
}

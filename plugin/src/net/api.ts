import * as vscode from 'vscode';

export interface RawArticle {
  id: string | number;
  title: string;
}

function joinUrl(baseUrl: string, path: string): string {
  const base = (baseUrl || '').replace(/\/+$/, '');
  const suffix = path.startsWith('/') ? path : `/${path}`;
  return `${base}${suffix}`;
}

export async function fetchArticles(baseUrl: string, token?: string): Promise<RawArticle[]> {
  const listUrl = joinUrl(baseUrl, '/api/allarticles');
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

export function buildArticleEndpoint(baseUrl: string, title: string): string {
  const encoded = encodeURIComponent(title);
  return joinUrl(baseUrl, `/api/article/${encoded}`);
}

export function buildImageEndpoint(baseUrl: string): string {
  return joinUrl(baseUrl, '/api/image');
}

export async function deleteArticleByTitle(baseUrl: string, title: string, token?: string): Promise<void> {
  const url = buildArticleEndpoint(baseUrl, title);
  const headers: Record<string, string> = { 'Accept': 'application/json' };
  if (token) headers['Authorization'] = `Bearer ${token}`;
  const res = await fetch(url, { method: 'DELETE', headers });
  if (!res.ok) {
    const text = await res.text().catch(() => '');
    throw new Error(`HTTP ${res.status} ${res.statusText} ${text}`);
  }
}

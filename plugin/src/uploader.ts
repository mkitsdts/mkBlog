import * as vscode from 'vscode';
import * as path from 'path';
import { findMarkdownFilesWithImageFolders } from './utils/fs';
import { buildArticleEndpoint, buildImageEndpoint } from './net/api';

function extractMeta(md: string): { author?: string; category?: string; content: string } {
  let author: string | undefined;
  let category: string | undefined;
  let content = md;

  // 1) YAML front matter: --- ... --- at file start
  if (md.startsWith('---')) {
    const end = md.indexOf('\n---', 3);
    if (end !== -1) {
      const fm = md.slice(3, end).split(/\r?\n/);
      for (const line of fm) {
        const m = line.match(/^\s*(author|category)\s*:\s*(.+)\s*$/i);
        if (m) {
          const key = m[1].toLowerCase();
          const val = m[2].trim();
          if (key === 'author' && val) author = val;
          if (key === 'category' && val) category = val;
        }
      }
      // remove front matter block including closing --- line
      const after = md.slice(end + '\n---'.length);
      // trim a single leading newline if exists
      content = after.replace(/^\r?\n/, '');
      return { author, category, content };
    }
  }

  // 2) Simple top-of-file meta lines: e.g.
  // author: mkitsdts
  // category: language
  const lines = md.split(/\r?\n/);
  let i = 0;
  while (i < lines.length) {
    const line = lines[i];
    if (!line.trim()) { i++; continue; }
    const m = line.match(/^\s*(author|category)\s*:\s*(.+)\s*$/i);
    if (m) {
      const key = m[1].toLowerCase();
      const val = m[2].trim();
      if (key === 'author' && !author && val) author = val;
      if (key === 'category' && !category && val) category = val;
      i++;
      continue;
    }
    break; // stop at first non-meta line
  }
  if (i > 0) {
    content = lines.slice(i).join('\n');
  }
  return { author, category, content };
}

export async function uploadFolderAsBlog(uri?: vscode.Uri) {
  const cfg = vscode.workspace.getConfiguration();
  const baseUrl = (cfg.get<string>('mkBlog.baseUrl') || '').trim();
  const token = cfg.get<string>('mkBlog.authToken') || '';
  if (!baseUrl) {
    throw new Error('未配置 mkBlog.baseUrl');
  }

  // 优先使用传入目录；否则使用已打开的工作区根目录；若都没有，再让用户选择
  let dir: vscode.Uri | undefined = uri ?? vscode.workspace.workspaceFolders?.[0]?.uri;
  if (!dir) {
    const picked = await vscode.window.showOpenDialog({
      canSelectFiles: false,
      canSelectFolders: true,
      canSelectMany: false,
      openLabel: '选择包含 Markdown 的文件夹'
    });
    if (!picked || picked.length === 0) return;
    dir = picked[0];
  }

  const tasks = await findMarkdownFilesWithImageFolders(dir.fsPath);
  if (tasks.length === 0) {
    vscode.window.showInformationMessage('未找到任何 .md 文件');
    return;
  }

  const headers: Record<string, string> = {};
  if (token) headers['Authorization'] = `Bearer ${token}`;

  const imageEndpoint = buildImageEndpoint(baseUrl);

  for (const t of tasks) {
    const title = path.basename(t.mdPath, '.md');
    // 分离上传：文章 PUT /article/{title}，图片逐张 PUT /image
    const pad = (n: number) => (n < 10 ? `0${n}` : String(n));
    const now = new Date();
    const updateAt = `${now.getFullYear()}-${pad(now.getMonth() + 1)}-${pad(now.getDate())} ${pad(now.getHours())}:${pad(now.getMinutes())}:${pad(now.getSeconds())}`;

    // 解析 markdown 顶部的 author/category 元数据，如果不存在则不发送这些字段
    const meta = extractMeta(t.mdContent);
    const article: Record<string, any> = {
      title,
      update_at: updateAt,
      content: meta.content,
    };
    if (meta.author) article.author = meta.author;
    else article.author = cfg.get<string>('mkBlog.defaultAuthor') ?? cfg.get<string>('mkBlog.author');

    if (meta.category) article.category = meta.category;
    else article.category = cfg.get<string>('mkBlog.defaultCategory');

    try {
      const articleUrl = buildArticleEndpoint(baseUrl, title);
        const jsonHeaders: Record<string, string> = { ...headers, 'Content-Type': 'application/json' };
        const res = await fetch(articleUrl, { method: 'PUT', body: JSON.stringify(article), headers: jsonHeaders });
        if (!res.ok) {
          const text = await res.text().catch(() => '');
          vscode.window.showWarningMessage(`上传文章失败(JSON): ${path.basename(t.mdPath)} -> HTTP ${res.status} ${res.statusText} ${text}，将继续上传图片。`);
        }
    } catch (err: any) {
      vscode.window.showWarningMessage(`上传文章异常: ${path.basename(t.mdPath)} -> ${err?.message || err}，将继续上传图片。`);
    }

    if (imageEndpoint) {
      for (const img of t.images) {
        const base64 = Buffer.from(img.buffer).toString('base64');
        const payload = { title, data: base64, name: img.name };
        const jsonHeaders: Record<string, string> = { ...headers, 'Content-Type': 'application/json' };
        const res = await fetch(imageEndpoint, { method: 'PUT', body: JSON.stringify(payload), headers: jsonHeaders });
        if (!res.ok) {
          const text = await res.text().catch(() => '');
          throw new Error(`上传图片失败(JSON): ${img.name} -> HTTP ${res.status} ${res.statusText} ${text}`);
        }
      }
    }
  }
  vscode.window.showInformationMessage(`上传完成，共 ${tasks.length} 篇。`);
}



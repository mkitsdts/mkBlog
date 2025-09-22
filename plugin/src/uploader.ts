import * as vscode from 'vscode';
import * as path from 'path';
import { findMarkdownFilesWithImageFolders } from './utils/fs';

export async function uploadFolderAsBlog(uri?: vscode.Uri) {
  const cfg = vscode.workspace.getConfiguration();
  const uploadArticleUrl = cfg.get<string>('mkBlog.uploadArticleUrl');
  const uploadImageUrl = cfg.get<string>('mkBlog.uploadImageUrl');
  const defaultAuthor = cfg.get<string>('mkBlog.author') || '';
  const defaultCategory = cfg.get<string>('mkBlog.defaultCategory') || 'General';
  const token = cfg.get<string>('mkBlog.authToken') || '';
  if (!uploadArticleUrl && !uploadImageUrl) {
    throw new Error('未配置上传接口：mkBlog.uploadArticleUrl 或 mkBlog.uploadImageUrl');
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

  for (const t of tasks) {
    const title = path.basename(t.mdPath, '.md');
    // 分离上传：文章 PUT /article/{title}，图片逐张 PUT /image
      const pad = (n: number) => (n < 10 ? `0${n}` : String(n));
      const now = new Date();
      const updateAt = `${now.getFullYear()}-${pad(now.getMonth() + 1)}-${pad(now.getDate())} ${pad(now.getHours())}:${pad(now.getMinutes())}:${pad(now.getSeconds())}`;

      const article = {
        Author: defaultAuthor,
        Title: title,
        UpdateAt: updateAt,
        Category: defaultCategory,
        Content: t.mdContent,
      };

      if (uploadArticleUrl) {
        try {
          const articleUrl = uploadArticleUrl.replace('{title}', encodeURIComponent(title));
          const jsonHeaders: Record<string, string> = { ...headers, 'Content-Type': 'application/json' };
          const res = await fetch(articleUrl, { method: 'PUT', body: JSON.stringify(article), headers: jsonHeaders });
          if (!res.ok) {
            const text = await res.text().catch(() => '');
            vscode.window.showWarningMessage(`上传文章失败(JSON): ${path.basename(t.mdPath)} -> HTTP ${res.status} ${res.statusText} ${text}，将继续上传图片。`);
          }
        } catch (err: any) {
          vscode.window.showWarningMessage(`上传文章异常: ${path.basename(t.mdPath)} -> ${err?.message || err}，将继续上传图片。`);
        }
      }

      if (uploadImageUrl) {
        for (const img of t.images) {
          const base64 = Buffer.from(img.buffer).toString('base64');
          const payload = { title, data: base64, name: img.name };
          const jsonHeaders: Record<string, string> = { ...headers, 'Content-Type': 'application/json' };
          const res = await fetch(uploadImageUrl, { method: 'PUT', body: JSON.stringify(payload), headers: jsonHeaders });
          if (!res.ok) {
            const text = await res.text().catch(() => '');
            throw new Error(`上传图片失败(JSON): ${img.name} -> HTTP ${res.status} ${res.statusText} ${text}`);
          }
        }
      }
  }
  vscode.window.showInformationMessage(`上传完成，共 ${tasks.length} 篇。`);
}



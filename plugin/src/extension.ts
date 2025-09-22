import * as vscode from 'vscode';
import { ArticleProvider, ArticleTreeItem } from './tree/ArticleProvider';
import { fetchArticles, deleteArticleById } from './net/api';
import { uploadFolderAsBlog } from './uploader';

export async function activate(context: vscode.ExtensionContext) {
  const articleProvider = new ArticleProvider(async () => {
    const cfg = vscode.workspace.getConfiguration();
    const listUrl = cfg.get<string>('mkBlog.listUrl');
    const token = cfg.get<string>('mkBlog.authToken') || '';
    if (!listUrl) {
      vscode.window.showWarningMessage('mkBlog: 未配置 listUrl');
      return [];
    }
    try {
      const items = await fetchArticles(listUrl, token);
      return items;
    } catch (err: any) {
      vscode.window.showErrorMessage(`获取文章列表失败: ${err?.message || err}`);
      return [];
    }
  });

  const tree = vscode.window.createTreeView('mkBlog.articles', {
    treeDataProvider: articleProvider,
    showCollapseAll: false
  });
  context.subscriptions.push(tree);

  context.subscriptions.push(
    vscode.commands.registerCommand('mkBlog.refreshArticles', async () => {
      await articleProvider.refresh();
    })
  );

  context.subscriptions.push(
    vscode.commands.registerCommand('mkBlog.deleteArticle', async (item?: ArticleTreeItem) => {
      const picked = item
        ? item.article
        : (await vscode.window.showQuickPick(articleProvider.getQuickPickItems(), { placeHolder: '选择要删除的文章' }))?.value;
      if (!picked) return;
      const confirm = await vscode.window.showWarningMessage(
        `确定删除文章：${picked.title}?`,
        { modal: true },
        '删除'
      );
      if (confirm !== '删除') return;
      const cfg = vscode.workspace.getConfiguration();
      const template = cfg.get<string>('mkBlog.deleteUrl');
      const token = cfg.get<string>('mkBlog.authToken') || '';
      if (!template) {
        vscode.window.showWarningMessage('mkBlog: 未配置 deleteUrl');
        return;
      }
  // 同时支持 {id} 与 {title} 两种占位符
  let url = template;
  url = url.replace('{id}', encodeURIComponent(String(picked.id)));
  url = url.replace('{title}', encodeURIComponent(String(picked.title)));
      try {
        await deleteArticleById(url, token);
        vscode.window.showInformationMessage('删除成功');
        await articleProvider.refresh();
      } catch (err: any) {
        vscode.window.showErrorMessage(`删除失败: ${err?.message || err}`);
      }
    })
  );

  context.subscriptions.push(
    vscode.commands.registerCommand('mkBlog.uploadFolder', async (uri?: vscode.Uri) => {
      try {
        await uploadFolderAsBlog(uri);
        await articleProvider.refresh();
      } catch (err: any) {
        vscode.window.showErrorMessage(`上传失败: ${err?.message || err}`);
      }
    })
  );

  // 激活后自动加载一次
  articleProvider.refresh();
}

export function deactivate() {}

import * as vscode from 'vscode';
import { ArticleProvider, ArticleTreeItem } from './tree/ArticleProvider';
import { fetchArticles, deleteArticleByTitle } from './net/api';
import { uploadFolderAsBlog } from './uploader';

export async function activate(context: vscode.ExtensionContext) {
  const articleProvider = new ArticleProvider(async () => {
    const cfg = vscode.workspace.getConfiguration();
    const baseUrl = (cfg.get<string>('mkBlog.baseUrl') || '').trim();
    const token = cfg.get<string>('mkBlog.authToken') || '';
    if (!baseUrl) {
      vscode.window.showWarningMessage('mkBlog: 未配置 baseUrl');
      return [];
    }
    try {
      const items = await fetchArticles(baseUrl, token);
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
      const baseUrl = (cfg.get<string>('mkBlog.baseUrl') || '').trim();
      const token = cfg.get<string>('mkBlog.authToken') || '';
      if (!baseUrl) {
        vscode.window.showWarningMessage('mkBlog: 未配置 baseUrl');
        return;
      }
      try {
        await deleteArticleByTitle(baseUrl, String(picked.title), token);
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

import * as vscode from 'vscode';

export interface ArticleItem {
  id: string | number;
  title: string;
}

export class ArticleTreeItem extends vscode.TreeItem {
  constructor(public readonly article: ArticleItem) {
    super(article.title, vscode.TreeItemCollapsibleState.None);
    this.id = String(article.id);
    this.contextValue = 'mkBlog.article';
    this.tooltip = `${article.title} (ID: ${article.id})`;
    this.iconPath = new vscode.ThemeIcon('note');
  }
}

export class ArticleProvider implements vscode.TreeDataProvider<ArticleTreeItem> {
  private _onDidChangeTreeData = new vscode.EventEmitter<void>();
  readonly onDidChangeTreeData = this._onDidChangeTreeData.event;

  private items: ArticleItem[] = [];

  constructor(private readonly loader: () => Promise<ArticleItem[]>) {}

  async refresh() {
    this.items = await this.loader();
    this._onDidChangeTreeData.fire();
  }

  getTreeItem(element: ArticleTreeItem): vscode.TreeItem {
    return element;
  }

  getChildren(): Thenable<ArticleTreeItem[]> {
    return Promise.resolve(this.items.map((a) => new ArticleTreeItem(a)));
  }

  getQuickPickItems(): { label: string; value: ArticleItem }[] {
    return this.items.map((it) => ({ label: it.title, value: it }));
  }
}

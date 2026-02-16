import {
	ItemView,
	Notice,
	Plugin,
	setIcon,
	TFile,
	WorkspaceLeaf,
} from "obsidian";

export const MKBLOG_VIEW_TYPE = "mkblog-articles-view";

export interface ArticleItem {
	id: string | number;
	title: string;
}

export interface ArticleViewActions {
	refreshArticles: () => Promise<void>;
	deleteArticleByTitle: (title: string) => Promise<void>;
	uploadCurrentFile?: () => Promise<void>;
	uploadFolder?: () => Promise<void>;
}

type RootState = "idle" | "loading" | "error" | "empty" | "ready";

export class MkBlogArticleView extends ItemView {
	private plugin: Plugin;
	private actions: ArticleViewActions;

	private rootEl!: HTMLElement;
	private toolbarEl!: HTMLElement;
	private listWrapEl!: HTMLElement;
	private listEl!: HTMLElement;
	private stateEl!: HTMLElement;
	private searchInputEl!: HTMLInputElement;

	private allArticles: ArticleItem[] = [];
	private filteredArticles: ArticleItem[] = [];
	private query = "";
	private state: RootState = "idle";
	private errMsg = "";

	constructor(leaf: WorkspaceLeaf, plugin: Plugin, actions: ArticleViewActions) {
		super(leaf);
		this.plugin = plugin;
		this.actions = actions;
	}

	getViewType(): string {
		return MKBLOG_VIEW_TYPE;
	}

	getDisplayText(): string {
		return "mkBlog";
	}

	getIcon(): string {
		return "newspaper";
	}

	async onOpen(): Promise<void> {
		const container = this.containerEl.children[1] as HTMLElement;
		container.empty();

		this.rootEl = container.createDiv({ cls: "mkblog-view-root" });
		this.toolbarEl = this.rootEl.createDiv({ cls: "mkblog-toolbar" });

		this.buildToolbar();

		this.stateEl = this.rootEl.createDiv({ cls: "mkblog-state" });

		this.listWrapEl = this.rootEl.createDiv({ cls: "mkblog-list-wrap" });
		this.listEl = this.listWrapEl.createEl("ul", { cls: "mkblog-article-list" });

		this.injectStyles();
		this.renderState();
		await this.reload();
	}

	async onClose(): Promise<void> {
		// no-op
	}

	public setArticles(items: ArticleItem[]): void {
		this.allArticles = (items ?? []).slice().sort((a, b) => {
			const t1 = (a?.title ?? "").toLowerCase();
			const t2 = (b?.title ?? "").toLowerCase();
			return t1.localeCompare(t2, "zh-Hans-CN");
		});
		this.applyFilterAndRender();
	}

	public setLoading(isLoading: boolean): void {
		this.state = isLoading ? "loading" : "idle";
		this.renderState();
	}

	public setError(message: string): void {
		this.state = "error";
		this.errMsg = message || "未知错误";
		this.renderState();
	}

	public async reload(): Promise<void> {
		try {
			this.state = "loading";
			this.renderState();
			await this.actions.refreshArticles();
		} catch (e: any) {
			this.state = "error";
			this.errMsg = e?.message ?? String(e);
			this.renderState();
		}
	}

	private buildToolbar(): void {
		const left = this.toolbarEl.createDiv({ cls: "mkblog-toolbar-left" });
		const right = this.toolbarEl.createDiv({ cls: "mkblog-toolbar-right" });

		const refreshBtn = left.createEl("button", {
			cls: "clickable-icon",
			attr: { "aria-label": "刷新文章列表", title: "刷新文章列表" },
		});
		setIcon(refreshBtn, "refresh-cw");
		refreshBtn.addEventListener("click", async () => {
			await this.reload();
		});

		if (this.actions.uploadCurrentFile) {
			const uploadFileBtn = left.createEl("button", {
				cls: "clickable-icon",
				attr: { "aria-label": "上传当前文件", title: "上传当前文件" },
			});
			setIcon(uploadFileBtn, "cloud-upload");
			uploadFileBtn.addEventListener("click", async () => {
				try {
					await this.actions.uploadCurrentFile?.();
					await this.reload();
				} catch (e: any) {
					new Notice(`上传当前文件失败: ${e?.message ?? e}`);
				}
			});
		}

		if (this.actions.uploadFolder) {
			const uploadFolderBtn = left.createEl("button", {
				cls: "clickable-icon",
				attr: { "aria-label": "上传文件夹", title: "上传文件夹" },
			});
			setIcon(uploadFolderBtn, "folder-up");
			uploadFolderBtn.addEventListener("click", async () => {
				try {
					await this.actions.uploadFolder?.();
					await this.reload();
				} catch (e: any) {
					new Notice(`上传文件夹失败: ${e?.message ?? e}`);
				}
			});
		}

		this.searchInputEl = right.createEl("input", {
			type: "search",
			placeholder: "搜索标题...",
			cls: "mkblog-search",
		});
		this.searchInputEl.addEventListener("input", () => {
			this.query = this.searchInputEl.value.trim();
			this.applyFilterAndRender();
		});
	}

	private applyFilterAndRender(): void {
		const q = this.query.toLowerCase();
		if (!q) {
			this.filteredArticles = [...this.allArticles];
		} else {
			this.filteredArticles = this.allArticles.filter((a) =>
				(a?.title ?? "").toLowerCase().includes(q)
			);
		}

		if (this.allArticles.length === 0) {
			this.state = "empty";
		} else {
			this.state = "ready";
		}
		this.renderState();
		this.renderList();
	}

	private renderState(): void {
		if (!this.stateEl) return;
		this.stateEl.empty();

		switch (this.state) {
			case "loading":
				this.stateEl.setText("正在加载文章列表...");
				this.listWrapEl?.addClass("is-hidden");
				break;
			case "error":
				this.stateEl.setText(`加载失败: ${this.errMsg}`);
				this.listWrapEl?.addClass("is-hidden");
				break;
			case "empty":
				this.stateEl.setText("暂无文章");
				this.listWrapEl?.addClass("is-hidden");
				break;
			case "ready":
				this.stateEl.empty();
				this.listWrapEl?.removeClass("is-hidden");
				break;
			default:
				this.stateEl.empty();
				break;
		}
	}

	private renderList(): void {
		if (!this.listEl) return;
		this.listEl.empty();

		for (const article of this.filteredArticles) {
			const li = this.listEl.createEl("li", { cls: "mkblog-article-item" });

			const titleBtn = li.createEl("button", {
				cls: "mkblog-title-btn",
				text: article.title,
				attr: {
					title: article.title,
					"aria-label": `文章 ${article.title}`,
				},
			});

			titleBtn.addEventListener("click", async () => {
				await this.openLocalMarkdownByTitle(article.title);
			});

			const actions = li.createDiv({ cls: "mkblog-item-actions" });

			const delBtn = actions.createEl("button", {
				cls: "clickable-icon",
				attr: { "aria-label": `删除 ${article.title}`, title: `删除 ${article.title}` },
			});
			setIcon(delBtn, "trash-2");
			delBtn.addEventListener("click", async (evt) => {
				evt.preventDefault();
				evt.stopPropagation();
				await this.confirmAndDelete(article.title);
			});
		}
	}

	private async confirmAndDelete(title: string): Promise<void> {
		// 使用简洁确认方式（避免额外 Modal 依赖）
		const ok = window.confirm(`确定删除远端文章：${title} ?`);
		if (!ok) return;

		try {
			await this.actions.deleteArticleByTitle(title);
			new Notice(`删除成功: ${title}`);
			await this.reload();
		} catch (e: any) {
			new Notice(`删除失败: ${e?.message ?? e}`);
		}
	}

	private async openLocalMarkdownByTitle(title: string): Promise<void> {
		const normalized = title.toLowerCase();

		const candidates = this.plugin.app.vault
			.getMarkdownFiles()
			.filter((f) => f.basename.toLowerCase() === normalized);

		if (candidates.length === 0) {
			new Notice(`本地未找到同名文档: ${title}.md`);
			return;
		}

		// 优先当前叶子打开；多个候选取第一个
		const file = candidates[0];
		await this.leaf.openFile(file);
	}

	private injectStyles(): void {
		// 避免重复注入
		const id = "mkblog-article-view-style";
		if (document.getElementById(id)) return;

		const style = document.createElement("style");
		style.id = id;
		style.textContent = `
.mkblog-view-root {
	height: 100%;
	display: flex;
	flex-direction: column;
	gap: 8px;
	padding: 8px;
}
.mkblog-toolbar {
	display: flex;
	align-items: center;
	justify-content: space-between;
	gap: 8px;
}
.mkblog-toolbar-left,
.mkblog-toolbar-right {
	display: flex;
	align-items: center;
	gap: 6px;
}
.mkblog-search {
	width: 160px;
	max-width: 100%;
}
.mkblog-state {
	color: var(--text-muted);
	font-size: 12px;
	padding: 4px 2px;
}
.mkblog-list-wrap {
	flex: 1;
	overflow: auto;
}
.mkblog-list-wrap.is-hidden {
	display: none;
}
.mkblog-article-list {
	list-style: none;
	margin: 0;
	padding: 0;
	display: flex;
	flex-direction: column;
	gap: 4px;
}
.mkblog-article-item {
	display: flex;
	align-items: center;
	justify-content: space-between;
	gap: 6px;
	border: 1px solid var(--background-modifier-border);
	border-radius: 6px;
	padding: 4px 6px;
}
.mkblog-title-btn {
	flex: 1;
	text-align: left;
	background: transparent;
	border: none;
	color: var(--text-normal);
	cursor: pointer;
	overflow: hidden;
	text-overflow: ellipsis;
	white-space: nowrap;
	padding: 4px;
}
.mkblog-title-btn:hover {
	color: var(--text-accent);
}
.mkblog-item-actions {
	display: flex;
	align-items: center;
	gap: 4px;
}
`;
		document.head.appendChild(style);
	}
}

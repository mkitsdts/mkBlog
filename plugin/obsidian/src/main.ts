import {
  App,
  Notice,
  Plugin,
  PluginSettingTab,
  Setting,
  ItemView,
  WorkspaceLeaf,
  TFile,
  Modal,
  FuzzySuggestModal,
  TFolder,
  normalizePath,
} from "obsidian";

const VIEW_TYPE_MKBLOG = "mkblog-articles-view";

interface MkBlogSettings {
  baseUrl: string;
  defaultAuthor: string;
  defaultCategory: string;
  authToken: string;
}

interface RawArticle {
  id: string | number;
  title: string;
}

const DEFAULT_SETTINGS: MkBlogSettings = {
  baseUrl: "http://localhost:8080",
  defaultAuthor: "",
  defaultCategory: "General",
  authToken: "",
};

const IMG_EXT = new Set([".png", ".jpg", ".jpeg", ".gif", ".webp", ".svg"]);

function joinUrl(baseUrl: string, path: string): string {
  const base = (baseUrl || "").replace(/\/+$/, "");
  const suffix = path.startsWith("/") ? path : `/${path}`;
  return `${base}${suffix}`;
}

function buildArticleEndpoint(baseUrl: string, title: string): string {
  return joinUrl(baseUrl, `/api/article/${encodeURIComponent(title)}`);
}

function buildImageEndpoint(baseUrl: string): string {
  return joinUrl(baseUrl, "/api/image");
}

function nowAsUpdateAt(): string {
  const pad = (n: number) => (n < 10 ? `0${n}` : String(n));
  const d = new Date();
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(
    d.getMinutes(),
  )}:${pad(d.getSeconds())}`;
}

function extname(name: string): string {
  const idx = name.lastIndexOf(".");
  if (idx < 0) return "";
  return name.slice(idx).toLowerCase();
}

function basenameWithoutExt(path: string): string {
  const p = path.replace(/\\/g, "/");
  const name = p.split("/").pop() ?? p;
  const idx = name.lastIndexOf(".");
  return idx >= 0 ? name.slice(0, idx) : name;
}

function dirname(path: string): string {
  const p = path.replace(/\\/g, "/");
  const idx = p.lastIndexOf("/");
  if (idx < 0) return "";
  return p.slice(0, idx);
}

function removeFrontmatter(raw: string): string {
  if (!raw.startsWith("---")) return raw;
  const endIdx = raw.indexOf("\n---", 3);
  if (endIdx === -1) return raw;
  const after = raw.slice(endIdx + "\n---".length);
  return after.replace(/^\r?\n/, "");
}

function parseMeta(rawMd: string): {
  author?: string;
  category?: string;
  content: string;
} {
  let author: string | undefined;
  let category: string | undefined;
  let content = rawMd;

  // 1) YAML frontmatter at top
  if (rawMd.startsWith("---")) {
    const end = rawMd.indexOf("\n---", 3);
    if (end !== -1) {
      const fm = rawMd.slice(3, end).split(/\r?\n/);
      for (const line of fm) {
        const m = line.match(/^\s*(author|category)\s*:\s*(.+)\s*$/i);
        if (m) {
          const key = m[1].toLowerCase();
          const val = m[2].trim().replace(/^['"]|['"]$/g, "");
          if (key === "author" && val) author = val;
          if (key === "category" && val) category = val;
        }
      }
      content = removeFrontmatter(rawMd);
      return { author, category, content };
    }
  }

  // 2) fallback top lines: author: / category:
  const lines = rawMd.split(/\r?\n/);
  let i = 0;
  while (i < lines.length) {
    const line = lines[i];
    if (!line.trim()) {
      i++;
      continue;
    }
    const m = line.match(/^\s*(author|category)\s*:\s*(.+)\s*$/i);
    if (!m) break;
    const key = m[1].toLowerCase();
    const val = m[2].trim().replace(/^['"]|['"]$/g, "");
    if (key === "author" && val && !author) author = val;
    if (key === "category" && val && !category) category = val;
    i++;
  }
  if (i > 0) content = lines.slice(i).join("\n");

  return { author, category, content };
}

async function reqJson(url: string, init?: RequestInit): Promise<any> {
  const res = await fetch(url, init);
  if (!res.ok) {
    const text = await res.text().catch(() => "");
    throw new Error(`HTTP ${res.status} ${res.statusText} ${text}`);
  }
  const ct = res.headers.get("content-type") ?? "";
  if (ct.includes("application/json")) return res.json();
  const txt = await res.text();
  try {
    return JSON.parse(txt);
  } catch {
    return txt;
  }
}

class ArticlePickerModal extends FuzzySuggestModal<RawArticle> {
  private readonly items: RawArticle[];
  public onChoose: (it: RawArticle) => void;

  constructor(
    app: App,
    items: RawArticle[],
    onChoose: (it: RawArticle) => void,
  ) {
    super(app);
    this.items = items;
    this.onChoose = onChoose;
    this.setPlaceholder("选择要删除的文章...");
  }

  getItems(): RawArticle[] {
    return this.items;
  }

  getItemText(item: RawArticle): string {
    return item.title;
  }

  onChooseItem(item: RawArticle): void {
    this.onChoose(item);
  }
}

class FolderPickerModal extends FuzzySuggestModal<TFolder> {
  private readonly folders: TFolder[];
  private readonly onChooseCb: (folder: TFolder) => void;

  constructor(
    app: App,
    folders: TFolder[],
    onChoose: (folder: TFolder) => void,
  ) {
    super(app);
    this.folders = folders;
    this.onChooseCb = onChoose;
    this.setPlaceholder("选择要上传的文件夹...");
  }

  getItems(): TFolder[] {
    return this.folders;
  }

  getItemText(item: TFolder): string {
    return item.path || "/";
  }

  onChooseItem(item: TFolder): void {
    this.onChooseCb(item);
  }
}

class ConfirmModal extends Modal {
  private readonly message: string;
  private readonly onConfirm: () => void;

  constructor(app: App, message: string, onConfirm: () => void) {
    super(app);
    this.message = message;
    this.onConfirm = onConfirm;
  }

  onOpen(): void {
    const { contentEl } = this;
    contentEl.empty();
    contentEl.createEl("h3", { text: "确认操作" });
    contentEl.createEl("p", { text: this.message });

    const actions = contentEl.createDiv({ cls: "mkblog-modal-actions" });
    const cancelBtn = actions.createEl("button", { text: "取消" });
    const okBtn = actions.createEl("button", { text: "删除" });
    okBtn.addClass("mod-warning");

    cancelBtn.onclick = () => this.close();
    okBtn.onclick = () => {
      this.close();
      this.onConfirm();
    };
  }

  onClose(): void {
    this.contentEl.empty();
  }
}

class MkBlogArticlesView extends ItemView {
  private plugin: MkBlogPlugin;
  private listEl: HTMLElement | null = null;

  constructor(leaf: WorkspaceLeaf, plugin: MkBlogPlugin) {
    super(leaf);
    this.plugin = plugin;
  }

  getViewType(): string {
    return VIEW_TYPE_MKBLOG;
  }

  getDisplayText(): string {
    return "mkBlog";
  }

  getIcon(): string {
    return "notebook-pen";
  }

  async onOpen(): Promise<void> {
    this.contentEl.empty();
    this.contentEl.addClass("mkblog-view");

    const header = this.contentEl.createDiv({ cls: "mkblog-header" });
    header.createEl("h3", { text: "mkBlog 文章管理" });

    const actions = header.createDiv({ cls: "mkblog-actions" });
    const refreshBtn = actions.createEl("button", { text: "刷新" });
    const uploadFileBtn = actions.createEl("button", { text: "上传当前文件" });
    const uploadFolderBtn = actions.createEl("button", { text: "上传文件夹" });

    refreshBtn.onclick = async () => {
      await this.plugin.refreshArticles();
    };
    uploadFileBtn.onclick = async () => {
      await this.plugin.uploadCurrentFileAsBlog();
    };
    uploadFolderBtn.onclick = async () => {
      await this.plugin.pickAndUploadFolder();
    };

    this.listEl = this.contentEl.createDiv({ cls: "mkblog-list" });
    await this.renderList();
  }

  async renderList(): Promise<void> {
    if (!this.listEl) return;
    this.listEl.empty();

    const items = this.plugin.articles;
    if (!items.length) {
      this.listEl.createEl("div", {
        text: "暂无文章（可点击刷新）",
        cls: "mkblog-empty",
      });
      return;
    }

    for (const it of items) {
      const row = this.listEl.createDiv({ cls: "mkblog-row" });
      const titleEl = row.createDiv({ text: it.title, cls: "mkblog-title" });
      titleEl.setAttribute("title", `${it.title} (ID: ${String(it.id)})`);

      const delBtn = row.createEl("button", { text: "删除" });
      delBtn.addClass("mod-warning");
      delBtn.onclick = async () => {
        this.plugin.confirmDelete(it);
      };
    }
  }

  async onClose(): Promise<void> {
    this.contentEl.empty();
  }
}

export default class MkBlogPlugin extends Plugin {
  settings: MkBlogSettings = DEFAULT_SETTINGS;
  articles: RawArticle[] = [];

  async onload(): Promise<void> {
    await this.loadSettings();

    this.registerView(
      VIEW_TYPE_MKBLOG,
      (leaf) => new MkBlogArticlesView(leaf, this),
    );
    this.addSettingTab(new MkBlogSettingTab(this.app, this));

    this.addRibbonIcon(
      "cloud-upload",
      "mkBlog: 上传当前文件为博客",
      async () => {
        await this.uploadCurrentFileAsBlog();
      },
    );

    this.addCommand({
      id: "mkblog-open-view",
      name: "mkBlog: 打开管理视图",
      callback: async () => this.activateView(),
    });

    this.addCommand({
      id: "mkblog-upload-current-file",
      name: "mkBlog: 上传当前文件为博客",
      callback: async () => this.uploadCurrentFileAsBlog(),
    });

    this.addCommand({
      id: "mkblog-upload-folder",
      name: "mkBlog: 上传选择文件夹为博客",
      callback: async () => this.pickAndUploadFolder(),
    });

    this.addCommand({
      id: "mkblog-refresh-articles",
      name: "mkBlog: 刷新文章列表",
      callback: async () => this.refreshArticles(),
    });

    this.addCommand({
      id: "mkblog-delete-article",
      name: "mkBlog: 删除文章",
      callback: async () => this.pickAndDeleteArticle(),
    });

    await this.activateView();
    await this.refreshArticles().catch((e) => {
      console.error("[mkBlog] initial refresh failed", e);
      new Notice(`mkBlog 初始化拉取文章失败: ${e?.message ?? e}`);
    });
  }

  async onunload(): Promise<void> {
    this.app.workspace.detachLeavesOfType(VIEW_TYPE_MKBLOG);
  }

  private authHeaders(json = false): Record<string, string> {
    const h: Record<string, string> = { Accept: "application/json" };
    if (this.settings.authToken?.trim()) {
      h["Authorization"] = `Bearer ${this.settings.authToken.trim()}`;
    }
    if (json) h["Content-Type"] = "application/json";
    return h;
  }

  private ensureBaseUrl(): string {
    const base = (this.settings.baseUrl || "").trim();
    if (!base) throw new Error("未配置 Base URL");
    return base;
  }

  private async fileToArrayBuffer(vaultPath: string): Promise<ArrayBuffer> {
    const abs = (this.app.vault.adapter as any).getFullPath?.(vaultPath);
    if (abs && "requestUrl" in window === false) {
      // fallback, but usually not needed
    }
    return await this.app.vault.adapter.readBinary(vaultPath);
  }

  private async collectImagesForMarkdown(
    mdFile: TFile,
  ): Promise<{ name: string; dataBase64: string }[]> {
    const parent = dirname(mdFile.path);
    const title = basenameWithoutExt(mdFile.name);
    const imgFolderPath = normalizePath(parent ? `${parent}/${title}` : title);

    const folder = this.app.vault.getAbstractFileByPath(imgFolderPath);
    if (!folder || !(folder instanceof TFolder)) return [];

    const out: { name: string; dataBase64: string }[] = [];
    const stack: TFolder[] = [folder];

    while (stack.length > 0) {
      const cur = stack.pop()!;
      for (const child of cur.children) {
        if (child instanceof TFolder) {
          stack.push(child);
          continue;
        }
        if (!(child instanceof TFile)) continue;
        const ext = extname(child.name);
        if (!IMG_EXT.has(ext)) continue;

        const buf = await this.fileToArrayBuffer(child.path);
        const base64 = this.arrayBufferToBase64(buf);
        // name only keeps file name for backend compatibility
        out.push({ name: child.name, dataBase64: base64 });
      }
    }

    return out;
  }

  private arrayBufferToBase64(buf: ArrayBuffer): string {
    let binary = "";
    const bytes = new Uint8Array(buf);
    const chunk = 0x8000;
    for (let i = 0; i < bytes.length; i += chunk) {
      const sub = bytes.subarray(i, Math.min(i + chunk, bytes.length));
      binary += String.fromCharCode(...sub);
    }
    return btoa(binary);
  }

  async fetchArticles(): Promise<RawArticle[]> {
    const baseUrl = this.ensureBaseUrl();
    const listUrl = joinUrl(baseUrl, "/api/allarticles");
    const data = await reqJson(listUrl, {
      method: "GET",
      headers: this.authHeaders(false),
    });

    let list: any[] = [];
    if (Array.isArray(data)) list = data;
    else if (Array.isArray(data?.articles)) list = data.articles;
    else if (Array.isArray(data?.data)) list = data.data;
    else if (Array.isArray(data?.data?.articles)) list = data.data.articles;
    else if (Array.isArray(data?.items)) list = data.items;
    else if (Array.isArray(data?.list)) list = data.list;

    return list.map((it: any, i: number) => ({
      id: it?.id ?? it?._id ?? it?.slug ?? it?.title ?? i,
      title: String(it?.title ?? it?.id ?? it?._id ?? `untitled-${i}`),
    }));
  }

  async refreshArticles(): Promise<void> {
    try {
      this.articles = await this.fetchArticles();
      new Notice(`mkBlog: 已刷新，共 ${this.articles.length} 篇`);
    } catch (e: any) {
      new Notice(`mkBlog: 刷新失败 - ${e?.message ?? e}`);
      throw e;
    } finally {
      this.redrawView();
    }
  }

  private redrawView(): void {
    const leaves = this.app.workspace.getLeavesOfType(VIEW_TYPE_MKBLOG);
    for (const leaf of leaves) {
      const v = leaf.view;
      if (v instanceof MkBlogArticlesView) {
        v.renderList();
      }
    }
  }

  async uploadCurrentFileAsBlog(): Promise<void> {
    const file = this.app.workspace.getActiveFile();
    if (!file) {
      new Notice("请先打开一个 Markdown 文件");
      return;
    }
    if (!(file instanceof TFile) || file.extension.toLowerCase() !== "md") {
      new Notice("仅支持上传 Markdown 文件（.md）");
      return;
    }

    await this.uploadSingleFile(file);
    await this.refreshArticles().catch(() => {});
  }

  private async uploadSingleFile(mdFile: TFile): Promise<void> {
    const baseUrl = this.ensureBaseUrl();
    const title = basenameWithoutExt(mdFile.name);
    const mdRaw = await this.app.vault.cachedRead(mdFile);
    const meta = parseMeta(mdRaw);

    const payload: Record<string, any> = {
      title,
      update_at: nowAsUpdateAt(),
      content: meta.content,
    };
    payload.author = meta.author ?? this.settings.defaultAuthor ?? "";
    payload.category = meta.category ?? this.settings.defaultCategory ?? "";

    const articleUrl = buildArticleEndpoint(baseUrl, title);
    await reqJson(articleUrl, {
      method: "PUT",
      headers: this.authHeaders(true),
      body: JSON.stringify(payload),
    });

    const images = await this.collectImagesForMarkdown(mdFile);
    const imageUrl = buildImageEndpoint(baseUrl);

    for (const img of images) {
      const imgPayload = { title, name: img.name, data: img.dataBase64 };
      await reqJson(imageUrl, {
        method: "PUT",
        headers: this.authHeaders(true),
        body: JSON.stringify(imgPayload),
      });
    }

    new Notice(
      `上传完成：${title}${images.length ? `（图片 ${images.length} 张）` : ""}`,
    );
  }

  async pickAndUploadFolder(): Promise<void> {
    const allFolders = this.app.vault
      .getAllLoadedFiles()
      .filter((f) => f instanceof TFolder) as TFolder[];
    const rootPath = this.app.vault.getRoot().path;
    const candidates = allFolders.filter((f) => f.path !== rootPath);

    if (!candidates.length) {
      new Notice("未找到可上传的文件夹");
      return;
    }

    new FolderPickerModal(this.app, candidates, async (folder) => {
      try {
        await this.uploadFolderAsBlog(folder);
        await this.refreshArticles().catch(() => {});
      } catch (e: any) {
        new Notice(`上传文件夹失败: ${e?.message ?? e}`);
      }
    }).open();
  }

  private async uploadFolderAsBlog(folder: TFolder): Promise<void> {
    const mdFiles = this.collectMarkdownFiles(folder);
    if (!mdFiles.length) {
      new Notice("所选文件夹未找到 .md 文件");
      return;
    }

    let success = 0;
    for (const md of mdFiles) {
      try {
        await this.uploadSingleFile(md);
        success++;
      } catch (e: any) {
        console.error(`[mkBlog] upload failed for ${md.path}`, e);
        new Notice(`上传失败: ${md.path} - ${e?.message ?? e}`);
      }
    }

    new Notice(`文件夹上传完成：成功 ${success}/${mdFiles.length}`);
  }

  private collectMarkdownFiles(folder: TFolder): TFile[] {
    const out: TFile[] = [];
    const stack: TFolder[] = [folder];
    while (stack.length > 0) {
      const cur = stack.pop()!;
      for (const c of cur.children) {
        if (c instanceof TFolder) stack.push(c);
        else if (c instanceof TFile && c.extension.toLowerCase() === "md")
          out.push(c);
      }
    }
    return out;
  }

  async pickAndDeleteArticle(): Promise<void> {
    if (!this.articles.length) {
      await this.refreshArticles().catch(() => {});
    }
    if (!this.articles.length) {
      new Notice("暂无可删除文章");
      return;
    }

    new ArticlePickerModal(this.app, this.articles, (it) =>
      this.confirmDelete(it),
    ).open();
  }

  confirmDelete(item: RawArticle): void {
    new ConfirmModal(this.app, `确定删除文章「${item.title}」？`, async () => {
      try {
        await this.deleteArticleByTitle(item.title);
        new Notice(`删除成功：${item.title}`);
        await this.refreshArticles().catch(() => {});
      } catch (e: any) {
        new Notice(`删除失败：${e?.message ?? e}`);
      }
    }).open();
  }

  async deleteArticleByTitle(title: string): Promise<void> {
    const baseUrl = this.ensureBaseUrl();
    const url = buildArticleEndpoint(baseUrl, title);
    const res = await fetch(url, {
      method: "DELETE",
      headers: this.authHeaders(false),
    });
    if (!res.ok) {
      const text = await res.text().catch(() => "");
      throw new Error(`HTTP ${res.status} ${res.statusText} ${text}`);
    }
  }

  private async activateView(): Promise<void> {
    const { workspace } = this.app;
    let leaf: WorkspaceLeaf | null =
      workspace.getLeavesOfType(VIEW_TYPE_MKBLOG)[0] ?? null;

    if (!leaf) {
      leaf = workspace.getRightLeaf(false);
      if (!leaf) return;

      await leaf.setViewState({
        type: VIEW_TYPE_MKBLOG,
        active: true,
      });
    }

    workspace.revealLeaf(leaf);
  }

  async loadSettings(): Promise<void> {
    const loaded = (await this.loadData()) as Partial<MkBlogSettings> | null;
    this.settings = Object.assign({}, DEFAULT_SETTINGS, loaded ?? {});
  }

  async saveSettings(): Promise<void> {
    await this.saveData(this.settings);
  }
}

class MkBlogSettingTab extends PluginSettingTab {
  plugin: MkBlogPlugin;

  constructor(app: App, plugin: MkBlogPlugin) {
    super(app, plugin);
    this.plugin = plugin;
  }

  display(): void {
    const { containerEl } = this;
    containerEl.empty();

    containerEl.createEl("h2", { text: "mkBlog 插件设置" });

    new Setting(containerEl)
      .setName("Base URL")
      .setDesc("后端服务基础地址，例如 http://localhost:8080")
      .addText((text) =>
        text
          .setPlaceholder("http://localhost:8080")
          .setValue(this.plugin.settings.baseUrl)
          .onChange(async (value) => {
            this.plugin.settings.baseUrl = value.trim();
            await this.plugin.saveSettings();
          }),
      );

    new Setting(containerEl)
      .setName("Default Author")
      .setDesc("Markdown 未声明 author 时使用")
      .addText((text) =>
        text
          .setValue(this.plugin.settings.defaultAuthor)
          .onChange(async (value) => {
            this.plugin.settings.defaultAuthor = value;
            await this.plugin.saveSettings();
          }),
      );

    new Setting(containerEl)
      .setName("Default Category")
      .setDesc("Markdown 未声明 category 时使用")
      .addText((text) =>
        text
          .setValue(this.plugin.settings.defaultCategory)
          .onChange(async (value) => {
            this.plugin.settings.defaultCategory = value;
            await this.plugin.saveSettings();
          }),
      );

    new Setting(containerEl)
      .setName("Auth Token")
      .setDesc("可选 Bearer Token，将通过 Authorization 头发送")
      .addText((text) =>
        text
          .setPlaceholder("eyJhbGciOi...")
          .setValue(this.plugin.settings.authToken)
          .onChange(async (value) => {
            this.plugin.settings.authToken = value.trim();
            await this.plugin.saveSettings();
          }),
      );
  }
}

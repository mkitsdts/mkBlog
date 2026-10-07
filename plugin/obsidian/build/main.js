"use strict";
var __defProp = Object.defineProperty;
var __getOwnPropDesc = Object.getOwnPropertyDescriptor;
var __getOwnPropNames = Object.getOwnPropertyNames;
var __hasOwnProp = Object.prototype.hasOwnProperty;
var __export = (target, all) => {
  for (var name in all)
    __defProp(target, name, { get: all[name], enumerable: true });
};
var __copyProps = (to, from, except, desc) => {
  if (from && typeof from === "object" || typeof from === "function") {
    for (let key of __getOwnPropNames(from))
      if (!__hasOwnProp.call(to, key) && key !== except)
        __defProp(to, key, { get: () => from[key], enumerable: !(desc = __getOwnPropDesc(from, key)) || desc.enumerable });
  }
  return to;
};
var __toCommonJS = (mod) => __copyProps(__defProp({}, "__esModule", { value: true }), mod);

// src/main.ts
var main_exports = {};
__export(main_exports, {
  default: () => MkBlogPlugin
});
module.exports = __toCommonJS(main_exports);
var import_obsidian = require("obsidian");
var VIEW_TYPE_MKBLOG = "mkblog-articles-view";
var DEFAULT_SETTINGS = {
  baseUrl: "http://localhost:8080",
  defaultAuthor: "",
  defaultCategory: "General",
  authToken: "",
  openViewOnStartup: false,
  refreshOnStartup: true
};
var REQUEST_TIMEOUT_MS = 15e3;
var IMG_EXT = /* @__PURE__ */ new Set([".png", ".jpg", ".jpeg", ".gif", ".webp", ".svg"]);
function joinUrl(baseUrl, path) {
  const base = (baseUrl || "").replace(/\/+$/, "");
  const suffix = path.startsWith("/") ? path : `/${path}`;
  return `${base}${suffix}`;
}
function buildArticleEndpoint(baseUrl, title) {
  return joinUrl(baseUrl, `/api/article/${encodeURIComponent(title)}`);
}
function buildImageEndpoint(baseUrl) {
  return joinUrl(baseUrl, "/api/image");
}
function nowAsUpdateAt() {
  const pad = (n) => n < 10 ? `0${n}` : String(n);
  const d = /* @__PURE__ */ new Date();
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(
    d.getMinutes()
  )}:${pad(d.getSeconds())}`;
}
function extname(name) {
  const idx = name.lastIndexOf(".");
  if (idx < 0) return "";
  return name.slice(idx).toLowerCase();
}
function basenameWithoutExt(path) {
  var _a;
  const p = path.replace(/\\/g, "/");
  const name = (_a = p.split("/").pop()) != null ? _a : p;
  const idx = name.lastIndexOf(".");
  return idx >= 0 ? name.slice(0, idx) : name;
}
function dirname(path) {
  const p = path.replace(/\\/g, "/");
  const idx = p.lastIndexOf("/");
  if (idx < 0) return "";
  return p.slice(0, idx);
}
function removeFrontmatter(raw) {
  if (!raw.startsWith("---")) return raw;
  const endIdx = raw.indexOf("\n---", 3);
  if (endIdx === -1) return raw;
  const after = raw.slice(endIdx + "\n---".length);
  return after.replace(/^\r?\n/, "");
}
function parseMeta(rawMd) {
  let author;
  let category;
  let content = rawMd;
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
function errText(e) {
  if (e instanceof Error) return e.message;
  return String(e != null ? e : "\u672A\u77E5\u9519\u8BEF");
}
async function fetchWithTimeout(url, init = {}, timeoutMs = REQUEST_TIMEOUT_MS) {
  const controller = new AbortController();
  const timer = window.setTimeout(() => controller.abort(), timeoutMs);
  try {
    return await fetch(url, { ...init, signal: controller.signal });
  } catch (e) {
    if (controller.signal.aborted) {
      throw new Error(`\u8BF7\u6C42\u8D85\u65F6\uFF08${timeoutMs}ms\uFF09: ${url}`);
    }
    throw new Error(`\u8BF7\u6C42\u5931\u8D25: ${errText(e)} (${url})`);
  } finally {
    window.clearTimeout(timer);
  }
}
async function reqJson(url, init) {
  var _a;
  const res = await fetchWithTimeout(url, init);
  if (!res.ok) {
    const text = await res.text().catch(() => "");
    throw new Error(`HTTP ${res.status} ${res.statusText} ${text}`);
  }
  const ct = (_a = res.headers.get("content-type")) != null ? _a : "";
  if (ct.includes("application/json")) return res.json();
  const txt = await res.text();
  try {
    return JSON.parse(txt);
  } catch (e) {
    return txt;
  }
}
var ArticlePickerModal = class extends import_obsidian.FuzzySuggestModal {
  constructor(app, items, onChoose) {
    super(app);
    this.items = items;
    this.onChoose = onChoose;
    this.setPlaceholder("\u9009\u62E9\u8981\u5220\u9664\u7684\u6587\u7AE0...");
  }
  getItems() {
    return this.items;
  }
  getItemText(item) {
    return item.title;
  }
  onChooseItem(item) {
    this.onChoose(item);
  }
};
var FolderPickerModal = class extends import_obsidian.FuzzySuggestModal {
  constructor(app, folders, onChoose) {
    super(app);
    this.folders = folders;
    this.onChooseCb = onChoose;
    this.setPlaceholder("\u9009\u62E9\u8981\u4E0A\u4F20\u7684\u6587\u4EF6\u5939...");
  }
  getItems() {
    return this.folders;
  }
  getItemText(item) {
    return item.path || "/";
  }
  onChooseItem(item) {
    this.onChooseCb(item);
  }
};
var ConfirmModal = class extends import_obsidian.Modal {
  constructor(app, message, onConfirm) {
    super(app);
    this.message = message;
    this.onConfirm = onConfirm;
  }
  onOpen() {
    const { contentEl } = this;
    contentEl.empty();
    contentEl.createEl("h3", { text: "\u786E\u8BA4\u64CD\u4F5C" });
    contentEl.createEl("p", { text: this.message });
    const actions = contentEl.createDiv({ cls: "mkblog-modal-actions" });
    const cancelBtn = actions.createEl("button", { text: "\u53D6\u6D88" });
    const okBtn = actions.createEl("button", { text: "\u5220\u9664" });
    okBtn.addClass("mod-warning");
    cancelBtn.onclick = () => this.close();
    okBtn.onclick = () => {
      this.close();
      this.onConfirm();
    };
  }
  onClose() {
    this.contentEl.empty();
  }
};
var MkBlogArticlesView = class extends import_obsidian.ItemView {
  constructor(leaf, plugin) {
    super(leaf);
    this.listEl = null;
    this.plugin = plugin;
  }
  getViewType() {
    return VIEW_TYPE_MKBLOG;
  }
  getDisplayText() {
    return "mkBlog";
  }
  getIcon() {
    return "notebook-pen";
  }
  async onOpen() {
    this.contentEl.empty();
    this.contentEl.addClass("mkblog-view");
    const header = this.contentEl.createDiv({ cls: "mkblog-header" });
    header.createEl("h3", { text: "mkBlog \u6587\u7AE0\u7BA1\u7406" });
    const actions = header.createDiv({ cls: "mkblog-actions" });
    const refreshBtn = actions.createEl("button", { text: "\u5237\u65B0" });
    const uploadFileBtn = actions.createEl("button", { text: "\u4E0A\u4F20\u5F53\u524D\u6587\u4EF6" });
    const uploadFolderBtn = actions.createEl("button", { text: "\u4E0A\u4F20\u6587\u4EF6\u5939" });
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
    this.renderList();
  }
  renderList() {
    if (!this.listEl) return;
    this.listEl.empty();
    const items = this.plugin.articles;
    if (!items.length) {
      this.listEl.createEl("div", {
        text: "\u6682\u65E0\u6587\u7AE0\uFF08\u53EF\u70B9\u51FB\u5237\u65B0\uFF09",
        cls: "mkblog-empty"
      });
      return;
    }
    for (const it of items) {
      const row = this.listEl.createDiv({ cls: "mkblog-row" });
      const titleEl = row.createDiv({ text: it.title, cls: "mkblog-title" });
      titleEl.setAttribute("title", `${it.title} (ID: ${String(it.id)})`);
      const delBtn = row.createEl("button", { text: "\u5220\u9664" });
      delBtn.addClass("mod-warning");
      delBtn.onclick = async () => {
        this.plugin.confirmDelete(it);
      };
    }
  }
  async onClose() {
    this.contentEl.empty();
  }
};
var MkBlogPlugin = class extends import_obsidian.Plugin {
  constructor() {
    super(...arguments);
    this.settings = { ...DEFAULT_SETTINGS };
    this.articles = [];
  }
  async onload() {
    this.registerView(
      VIEW_TYPE_MKBLOG,
      (leaf) => new MkBlogArticlesView(leaf, this)
    );
    this.addSettingTab(new MkBlogSettingTab(this.app, this));
    this.addRibbonIcon("cloud-upload", "mkBlog: \u4E0A\u4F20\u5F53\u524D\u6587\u4EF6\u4E3A\u535A\u5BA2", () => {
      void this.uploadCurrentFileAsBlog();
    });
    this.addCommand({
      id: "mkblog-open-view",
      name: "mkBlog: \u6253\u5F00\u7BA1\u7406\u89C6\u56FE",
      callback: () => {
        void this.activateView();
      }
    });
    this.addCommand({
      id: "mkblog-upload-current-file",
      name: "mkBlog: \u4E0A\u4F20\u5F53\u524D\u6587\u4EF6\u4E3A\u535A\u5BA2",
      callback: () => {
        void this.uploadCurrentFileAsBlog();
      }
    });
    this.addCommand({
      id: "mkblog-upload-folder",
      name: "mkBlog: \u4E0A\u4F20\u9009\u62E9\u6587\u4EF6\u5939\u4E3A\u535A\u5BA2",
      callback: () => {
        void this.pickAndUploadFolder();
      }
    });
    this.addCommand({
      id: "mkblog-refresh-articles",
      name: "mkBlog: \u5237\u65B0\u6587\u7AE0\u5217\u8868",
      callback: () => {
        void this.refreshArticles();
      }
    });
    this.addCommand({
      id: "mkblog-delete-article",
      name: "mkBlog: \u5220\u9664\u6587\u7AE0",
      callback: () => {
        void this.pickAndDeleteArticle();
      }
    });
    await this.loadSettings();
    if (this.app.workspace.layoutReady) {
      void this.runStartupTasks();
    } else {
      this.app.workspace.onLayoutReady(() => {
        void this.runStartupTasks();
      });
    }
  }
  async runStartupTasks() {
    try {
      await this.healGhostLeaves();
      if (this.settings.openViewOnStartup) {
        await this.activateView();
      }
      if (this.settings.refreshOnStartup) {
        await this.refreshArticles({ silent: true });
      } else {
        this.redrawViews();
      }
    } catch (e) {
      console.error("[mkBlog] startup tasks failed", e);
    }
  }
  /**
   * 修复历史遗留的 ghost 叶子。
   *
   * 旧版本在恢复布局时 view type 还没注册，Obsidian 会为它创建一个
   * EmptyView 占位叶子并写回 workspace.json（icon=lucide-ghost）。
   * 这里把它重新实例化为真实视图，避免用户需要手动改 workspace.json。
   *
   * 注意：Obsidian 1.7.2+ 的 DeferredView 是正常状态（可见时才实例化），
   * 不能误伤。
   */
  async healGhostLeaves() {
    for (const leaf of this.app.workspace.getLeavesOfType(VIEW_TYPE_MKBLOG)) {
      if (leaf.view instanceof MkBlogArticlesView) continue;
      if ((0, import_obsidian.requireApiVersion)("1.7.2") && leaf.isDeferred) continue;
      try {
        await leaf.setViewState({ type: VIEW_TYPE_MKBLOG, active: false });
        console.log("[mkBlog] repaired ghost view leaf");
      } catch (e) {
        console.error("[mkBlog] failed to repair ghost view leaf", e);
      }
    }
  }
  onunload() {
    this.app.workspace.detachLeavesOfType(VIEW_TYPE_MKBLOG);
  }
  authHeaders(json = false) {
    var _a;
    const h = { Accept: "application/json" };
    if ((_a = this.settings.authToken) == null ? void 0 : _a.trim()) {
      h["Authorization"] = `Bearer ${this.settings.authToken.trim()}`;
    }
    if (json) h["Content-Type"] = "application/json";
    return h;
  }
  ensureBaseUrl() {
    const base = (this.settings.baseUrl || "").trim();
    if (!base) throw new Error("\u672A\u914D\u7F6E Base URL");
    return base;
  }
  async fileToArrayBuffer(vaultPath) {
    return await this.app.vault.adapter.readBinary(vaultPath);
  }
  async collectImagesForMarkdown(mdFile) {
    const parent = dirname(mdFile.path);
    const title = basenameWithoutExt(mdFile.name);
    const imgFolderPath = (0, import_obsidian.normalizePath)(parent ? `${parent}/${title}` : title);
    const folder = this.app.vault.getAbstractFileByPath(imgFolderPath);
    if (!folder || !(folder instanceof import_obsidian.TFolder)) return [];
    const out = [];
    const stack = [folder];
    while (stack.length > 0) {
      const cur = stack.pop();
      for (const child of cur.children) {
        if (child instanceof import_obsidian.TFolder) {
          stack.push(child);
          continue;
        }
        if (!(child instanceof import_obsidian.TFile)) continue;
        const ext = extname(child.name);
        if (!IMG_EXT.has(ext)) continue;
        const buf = await this.fileToArrayBuffer(child.path);
        const base64 = this.arrayBufferToBase64(buf);
        out.push({ name: child.name, dataBase64: base64 });
      }
    }
    return out;
  }
  arrayBufferToBase64(buf) {
    let binary = "";
    const bytes = new Uint8Array(buf);
    const chunk = 32768;
    for (let i = 0; i < bytes.length; i += chunk) {
      const sub = bytes.subarray(i, Math.min(i + chunk, bytes.length));
      binary += String.fromCharCode(...sub);
    }
    return btoa(binary);
  }
  async fetchArticles() {
    var _a;
    const baseUrl = this.ensureBaseUrl();
    const listUrl = joinUrl(baseUrl, "/api/allarticles");
    const data = await reqJson(listUrl, {
      method: "GET",
      headers: this.authHeaders(false)
    });
    let list = [];
    if (Array.isArray(data)) list = data;
    else if (Array.isArray(data == null ? void 0 : data.articles)) list = data.articles;
    else if (Array.isArray(data == null ? void 0 : data.data)) list = data.data;
    else if (Array.isArray((_a = data == null ? void 0 : data.data) == null ? void 0 : _a.articles)) list = data.data.articles;
    else if (Array.isArray(data == null ? void 0 : data.items)) list = data.items;
    else if (Array.isArray(data == null ? void 0 : data.list)) list = data.list;
    return list.map((it, i) => {
      var _a2, _b, _c, _d, _e, _f, _g;
      return {
        id: (_d = (_c = (_b = (_a2 = it == null ? void 0 : it.id) != null ? _a2 : it == null ? void 0 : it._id) != null ? _b : it == null ? void 0 : it.slug) != null ? _c : it == null ? void 0 : it.title) != null ? _d : i,
        title: String((_g = (_f = (_e = it == null ? void 0 : it.title) != null ? _e : it == null ? void 0 : it.id) != null ? _f : it == null ? void 0 : it._id) != null ? _g : `untitled-${i}`)
      };
    });
  }
  /**
   * 刷新文章列表。
   *
   * 不再向调用方抛出异常：命令 / Ribbon / 启动流程都会 invoke 它，
   * 抛出会造成未处理的 Promise rejection（Obsidian 会当成插件报错）。
   *
   * @returns 成功时返回文章列表，失败时返回 null。
   */
  async refreshArticles(opts = {}) {
    try {
      this.articles = await this.fetchArticles();
      this.redrawViews();
      if (!opts.silent) {
        new import_obsidian.Notice(`mkBlog: \u5DF2\u5237\u65B0\uFF0C\u5171 ${this.articles.length} \u7BC7`);
      }
      return this.articles;
    } catch (e) {
      if (opts.silent) {
        console.error("[mkBlog] refresh failed", e);
      } else {
        new import_obsidian.Notice(`mkBlog: \u5237\u65B0\u5931\u8D25 - ${errText(e)}`);
      }
      return null;
    }
  }
  redrawViews() {
    for (const leaf of this.app.workspace.getLeavesOfType(VIEW_TYPE_MKBLOG)) {
      if (leaf.view instanceof MkBlogArticlesView) {
        leaf.view.renderList();
      }
    }
  }
  async uploadCurrentFileAsBlog() {
    try {
      const file = this.app.workspace.getActiveFile();
      if (!file) {
        new import_obsidian.Notice("\u8BF7\u5148\u6253\u5F00\u4E00\u4E2A Markdown \u6587\u4EF6");
        return;
      }
      if (!(file instanceof import_obsidian.TFile) || file.extension.toLowerCase() !== "md") {
        new import_obsidian.Notice("\u4EC5\u652F\u6301\u4E0A\u4F20 Markdown \u6587\u4EF6\uFF08.md\uFF09");
        return;
      }
      await this.uploadSingleFile(file);
      await this.refreshArticles({ silent: true });
    } catch (e) {
      console.error("[mkBlog] upload current file failed", e);
      new import_obsidian.Notice(`\u4E0A\u4F20\u5931\u8D25: ${errText(e)}`);
    }
  }
  async uploadSingleFile(mdFile) {
    var _a, _b, _c, _d;
    const baseUrl = this.ensureBaseUrl();
    const title = basenameWithoutExt(mdFile.name);
    const mdRaw = await this.app.vault.cachedRead(mdFile);
    const meta = parseMeta(mdRaw);
    const payload = {
      title,
      update_at: nowAsUpdateAt(),
      content: meta.content
    };
    payload.author = (_b = (_a = meta.author) != null ? _a : this.settings.defaultAuthor) != null ? _b : "";
    payload.category = (_d = (_c = meta.category) != null ? _c : this.settings.defaultCategory) != null ? _d : "";
    const articleUrl = buildArticleEndpoint(baseUrl, title);
    await reqJson(articleUrl, {
      method: "PUT",
      headers: this.authHeaders(true),
      body: JSON.stringify(payload)
    });
    const images = await this.collectImagesForMarkdown(mdFile);
    const imageUrl = buildImageEndpoint(baseUrl);
    for (const img of images) {
      const imgPayload = { title, name: img.name, data: img.dataBase64 };
      await reqJson(imageUrl, {
        method: "PUT",
        headers: this.authHeaders(true),
        body: JSON.stringify(imgPayload)
      });
    }
    new import_obsidian.Notice(
      `\u4E0A\u4F20\u5B8C\u6210\uFF1A${title}${images.length ? `\uFF08\u56FE\u7247 ${images.length} \u5F20\uFF09` : ""}`
    );
  }
  async pickAndUploadFolder() {
    const allFolders = this.app.vault.getAllLoadedFiles().filter((f) => f instanceof import_obsidian.TFolder);
    const rootPath = this.app.vault.getRoot().path;
    const candidates = allFolders.filter((f) => f.path !== rootPath);
    if (!candidates.length) {
      new import_obsidian.Notice("\u672A\u627E\u5230\u53EF\u4E0A\u4F20\u7684\u6587\u4EF6\u5939");
      return;
    }
    new FolderPickerModal(this.app, candidates, async (folder) => {
      try {
        await this.uploadFolderAsBlog(folder);
        await this.refreshArticles({ silent: true });
      } catch (e) {
        new import_obsidian.Notice(`\u4E0A\u4F20\u6587\u4EF6\u5939\u5931\u8D25: ${errText(e)}`);
      }
    }).open();
  }
  async uploadFolderAsBlog(folder) {
    const mdFiles = this.collectMarkdownFiles(folder);
    if (!mdFiles.length) {
      new import_obsidian.Notice("\u6240\u9009\u6587\u4EF6\u5939\u672A\u627E\u5230 .md \u6587\u4EF6");
      return;
    }
    let success = 0;
    for (const md of mdFiles) {
      try {
        await this.uploadSingleFile(md);
        success++;
      } catch (e) {
        console.error(`[mkBlog] upload failed for ${md.path}`, e);
        new import_obsidian.Notice(`\u4E0A\u4F20\u5931\u8D25: ${md.path} - ${errText(e)}`);
      }
    }
    new import_obsidian.Notice(`\u6587\u4EF6\u5939\u4E0A\u4F20\u5B8C\u6210\uFF1A\u6210\u529F ${success}/${mdFiles.length}`);
  }
  collectMarkdownFiles(folder) {
    const out = [];
    const stack = [folder];
    while (stack.length > 0) {
      const cur = stack.pop();
      for (const c of cur.children) {
        if (c instanceof import_obsidian.TFolder) stack.push(c);
        else if (c instanceof import_obsidian.TFile && c.extension.toLowerCase() === "md")
          out.push(c);
      }
    }
    return out;
  }
  async pickAndDeleteArticle() {
    if (!this.articles.length) {
      await this.refreshArticles({ silent: true });
    }
    if (!this.articles.length) {
      new import_obsidian.Notice("\u6682\u65E0\u53EF\u5220\u9664\u6587\u7AE0");
      return;
    }
    new ArticlePickerModal(
      this.app,
      this.articles,
      (it) => this.confirmDelete(it)
    ).open();
  }
  confirmDelete(item) {
    new ConfirmModal(this.app, `\u786E\u5B9A\u5220\u9664\u6587\u7AE0\u300C${item.title}\u300D\uFF1F`, async () => {
      try {
        await this.deleteArticleByTitle(item.title);
        new import_obsidian.Notice(`\u5220\u9664\u6210\u529F\uFF1A${item.title}`);
        await this.refreshArticles({ silent: true });
      } catch (e) {
        new import_obsidian.Notice(`\u5220\u9664\u5931\u8D25\uFF1A${errText(e)}`);
      }
    }).open();
  }
  async deleteArticleByTitle(title) {
    const baseUrl = this.ensureBaseUrl();
    const url = buildArticleEndpoint(baseUrl, title);
    const res = await fetchWithTimeout(url, {
      method: "DELETE",
      headers: this.authHeaders(false)
    });
    if (!res.ok) {
      const text = await res.text().catch(() => "");
      throw new Error(`HTTP ${res.status} ${res.statusText} ${text}`);
    }
  }
  async activateView() {
    var _a;
    const { workspace } = this.app;
    let leaf = (_a = workspace.getLeavesOfType(VIEW_TYPE_MKBLOG)[0]) != null ? _a : null;
    if (!leaf) {
      leaf = workspace.getRightLeaf(false);
      if (!leaf) {
        new import_obsidian.Notice("mkBlog: \u65E0\u6CD5\u521B\u5EFA\u89C6\u56FE\uFF08\u53F3\u4FA7\u680F\u4E0D\u53EF\u7528\uFF09");
        return;
      }
      await leaf.setViewState({
        type: VIEW_TYPE_MKBLOG,
        active: true
      });
    }
    await workspace.revealLeaf(leaf);
    this.redrawViews();
  }
  async loadSettings() {
    try {
      const loaded = await this.loadData();
      this.settings = { ...DEFAULT_SETTINGS, ...loaded != null ? loaded : {} };
    } catch (e) {
      console.error("[mkBlog] failed to load settings, falling back to defaults", e);
      this.settings = { ...DEFAULT_SETTINGS };
    }
  }
  async saveSettings() {
    await this.saveData(this.settings);
  }
};
var MkBlogSettingTab = class extends import_obsidian.PluginSettingTab {
  constructor(app, plugin) {
    super(app, plugin);
    this.plugin = plugin;
  }
  display() {
    const { containerEl } = this;
    containerEl.empty();
    containerEl.createEl("h2", { text: "mkBlog \u63D2\u4EF6\u8BBE\u7F6E" });
    new import_obsidian.Setting(containerEl).setName("Base URL").setDesc("\u540E\u7AEF\u670D\u52A1\u57FA\u7840\u5730\u5740\uFF0C\u4F8B\u5982 http://localhost:8080").addText(
      (text) => text.setPlaceholder("http://localhost:8080").setValue(this.plugin.settings.baseUrl).onChange(async (value) => {
        this.plugin.settings.baseUrl = value.trim();
        await this.plugin.saveSettings();
      })
    );
    new import_obsidian.Setting(containerEl).setName("Default Author").setDesc("Markdown \u672A\u58F0\u660E author \u65F6\u4F7F\u7528").addText(
      (text) => text.setValue(this.plugin.settings.defaultAuthor).onChange(async (value) => {
        this.plugin.settings.defaultAuthor = value;
        await this.plugin.saveSettings();
      })
    );
    new import_obsidian.Setting(containerEl).setName("Default Category").setDesc("Markdown \u672A\u58F0\u660E category \u65F6\u4F7F\u7528").addText(
      (text) => text.setValue(this.plugin.settings.defaultCategory).onChange(async (value) => {
        this.plugin.settings.defaultCategory = value;
        await this.plugin.saveSettings();
      })
    );
    new import_obsidian.Setting(containerEl).setName("Auth Token").setDesc("\u53EF\u9009 Bearer Token\uFF0C\u5C06\u901A\u8FC7 Authorization \u5934\u53D1\u9001").addText(
      (text) => text.setPlaceholder("eyJhbGciOi...").setValue(this.plugin.settings.authToken).onChange(async (value) => {
        this.plugin.settings.authToken = value.trim();
        await this.plugin.saveSettings();
      })
    );
    containerEl.createEl("h3", { text: "\u542F\u52A8\u884C\u4E3A" });
    new import_obsidian.Setting(containerEl).setName("\u542F\u52A8\u65F6\u81EA\u52A8\u6253\u5F00\u7BA1\u7406\u89C6\u56FE").setDesc(
      "\u5173\u95ED\u540E\u4EC5\u5728\u70B9\u51FB\u5DE6\u4FA7 Ribbon \u56FE\u6807\u6216\u6267\u884C\u547D\u4EE4\u65F6\u6253\u5F00\uFF08\u63A8\u8350\u5173\u95ED\uFF0C\u907F\u514D\u6BCF\u6B21\u542F\u52A8\u90FD\u6539\u5199\u4F60\u7684\u5DE5\u4F5C\u533A\u5E03\u5C40\uFF09"
    ).addToggle(
      (toggle) => toggle.setValue(this.plugin.settings.openViewOnStartup).onChange(async (value) => {
        this.plugin.settings.openViewOnStartup = value;
        await this.plugin.saveSettings();
      })
    );
    new import_obsidian.Setting(containerEl).setName("\u542F\u52A8\u65F6\u81EA\u52A8\u5237\u65B0\u6587\u7AE0\u5217\u8868").setDesc("\u5728 Obsidian \u5E03\u5C40\u5C31\u7EEA\u540E\u518D\u53D1\u8D77\u8BF7\u6C42\uFF0C\u4E0D\u4F1A\u963B\u585E\u542F\u52A8\uFF1B\u8BF7\u6C42\u5E26 15s \u8D85\u65F6").addToggle(
      (toggle) => toggle.setValue(this.plugin.settings.refreshOnStartup).onChange(async (value) => {
        this.plugin.settings.refreshOnStartup = value;
        await this.plugin.saveSettings();
      })
    );
  }
};
//# sourceMappingURL=data:application/json;base64,ewogICJ2ZXJzaW9uIjogMywKICAic291cmNlcyI6IFsiLi4vc3JjL21haW4udHMiXSwKICAic291cmNlc0NvbnRlbnQiOiBbImltcG9ydCB7XG4gIEFwcCxcbiAgTm90aWNlLFxuICBQbHVnaW4sXG4gIFBsdWdpblNldHRpbmdUYWIsXG4gIFNldHRpbmcsXG4gIEl0ZW1WaWV3LFxuICBXb3Jrc3BhY2VMZWFmLFxuICBURmlsZSxcbiAgTW9kYWwsXG4gIEZ1enp5U3VnZ2VzdE1vZGFsLFxuICBURm9sZGVyLFxuICBub3JtYWxpemVQYXRoLFxuICByZXF1aXJlQXBpVmVyc2lvbixcbn0gZnJvbSBcIm9ic2lkaWFuXCI7XG5cbmNvbnN0IFZJRVdfVFlQRV9NS0JMT0cgPSBcIm1rYmxvZy1hcnRpY2xlcy12aWV3XCI7XG5cbmludGVyZmFjZSBNa0Jsb2dTZXR0aW5ncyB7XG4gIGJhc2VVcmw6IHN0cmluZztcbiAgZGVmYXVsdEF1dGhvcjogc3RyaW5nO1xuICBkZWZhdWx0Q2F0ZWdvcnk6IHN0cmluZztcbiAgYXV0aFRva2VuOiBzdHJpbmc7XG4gIC8qKiBcdTU0MkZcdTUyQThcdTY1RjZcdTgxRUFcdTUyQThcdTYyNTNcdTVGMDBcdTdCQTFcdTc0MDZcdTg5QzZcdTU2RkVcdUZGMDhcdTlFRDhcdThCQTRcdTUxNzNcdTk1RURcdUZGMENcdTkwN0ZcdTUxNERcdTdCRTFcdTY1MzlcdTc1MjhcdTYyMzdcdTc2ODRcdTVERTVcdTRGNUNcdTUzM0FcdTVFMDNcdTVDNDBcdUZGMDkgKi9cbiAgb3BlblZpZXdPblN0YXJ0dXA6IGJvb2xlYW47XG4gIC8qKiBcdTU0MkZcdTUyQThcdTY1RjZcdTgxRUFcdTUyQThcdTUyMzdcdTY1QjBcdTY1ODdcdTdBRTBcdTUyMTdcdTg4NjhcdUZGMDhcdTU3MjhcdTVFMDNcdTVDNDBcdTVDMzFcdTdFRUFcdTRFNEJcdTU0MEVcdTYyNjdcdTg4NENcdUZGMENcdTRFMERcdTk2M0JcdTU4NUVcdTU0MkZcdTUyQThcdUZGMDkgKi9cbiAgcmVmcmVzaE9uU3RhcnR1cDogYm9vbGVhbjtcbn1cblxuaW50ZXJmYWNlIFJhd0FydGljbGUge1xuICBpZDogc3RyaW5nIHwgbnVtYmVyO1xuICB0aXRsZTogc3RyaW5nO1xufVxuXG5jb25zdCBERUZBVUxUX1NFVFRJTkdTOiBNa0Jsb2dTZXR0aW5ncyA9IHtcbiAgYmFzZVVybDogXCJodHRwOi8vbG9jYWxob3N0OjgwODBcIixcbiAgZGVmYXVsdEF1dGhvcjogXCJcIixcbiAgZGVmYXVsdENhdGVnb3J5OiBcIkdlbmVyYWxcIixcbiAgYXV0aFRva2VuOiBcIlwiLFxuICBvcGVuVmlld09uU3RhcnR1cDogZmFsc2UsXG4gIHJlZnJlc2hPblN0YXJ0dXA6IHRydWUsXG59O1xuXG4vKiogXHU1MzU1XHU2QjIxXHU4QkY3XHU2QzQyXHU4RDg1XHU2NUY2XHU2NUY2XHU5NUY0XHUzMDAyXHU3RUREXHU0RTBEXHU1MTQxXHU4QkI4XHU4QkY3XHU2QzQyXHU2NUUwXHU5NjUwXHU2NzFGXHU2MzAyXHU4RDc3XHVGRjBDXHU1NDI2XHU1MjE5XHU0RjFBXHU2MkQ2XHU2QjdCIE9ic2lkaWFuIFx1NTQyRlx1NTJBOFx1NkQ0MVx1N0EwQlx1MzAwMiAqL1xuY29uc3QgUkVRVUVTVF9USU1FT1VUX01TID0gMTUwMDA7XG5cbmNvbnN0IElNR19FWFQgPSBuZXcgU2V0KFtcIi5wbmdcIiwgXCIuanBnXCIsIFwiLmpwZWdcIiwgXCIuZ2lmXCIsIFwiLndlYnBcIiwgXCIuc3ZnXCJdKTtcblxuZnVuY3Rpb24gam9pblVybChiYXNlVXJsOiBzdHJpbmcsIHBhdGg6IHN0cmluZyk6IHN0cmluZyB7XG4gIGNvbnN0IGJhc2UgPSAoYmFzZVVybCB8fCBcIlwiKS5yZXBsYWNlKC9cXC8rJC8sIFwiXCIpO1xuICBjb25zdCBzdWZmaXggPSBwYXRoLnN0YXJ0c1dpdGgoXCIvXCIpID8gcGF0aCA6IGAvJHtwYXRofWA7XG4gIHJldHVybiBgJHtiYXNlfSR7c3VmZml4fWA7XG59XG5cbmZ1bmN0aW9uIGJ1aWxkQXJ0aWNsZUVuZHBvaW50KGJhc2VVcmw6IHN0cmluZywgdGl0bGU6IHN0cmluZyk6IHN0cmluZyB7XG4gIHJldHVybiBqb2luVXJsKGJhc2VVcmwsIGAvYXBpL2FydGljbGUvJHtlbmNvZGVVUklDb21wb25lbnQodGl0bGUpfWApO1xufVxuXG5mdW5jdGlvbiBidWlsZEltYWdlRW5kcG9pbnQoYmFzZVVybDogc3RyaW5nKTogc3RyaW5nIHtcbiAgcmV0dXJuIGpvaW5VcmwoYmFzZVVybCwgXCIvYXBpL2ltYWdlXCIpO1xufVxuXG5mdW5jdGlvbiBub3dBc1VwZGF0ZUF0KCk6IHN0cmluZyB7XG4gIGNvbnN0IHBhZCA9IChuOiBudW1iZXIpID0+IChuIDwgMTAgPyBgMCR7bn1gIDogU3RyaW5nKG4pKTtcbiAgY29uc3QgZCA9IG5ldyBEYXRlKCk7XG4gIHJldHVybiBgJHtkLmdldEZ1bGxZZWFyKCl9LSR7cGFkKGQuZ2V0TW9udGgoKSArIDEpfS0ke3BhZChkLmdldERhdGUoKSl9ICR7cGFkKGQuZ2V0SG91cnMoKSl9OiR7cGFkKFxuICAgIGQuZ2V0TWludXRlcygpLFxuICApfToke3BhZChkLmdldFNlY29uZHMoKSl9YDtcbn1cblxuZnVuY3Rpb24gZXh0bmFtZShuYW1lOiBzdHJpbmcpOiBzdHJpbmcge1xuICBjb25zdCBpZHggPSBuYW1lLmxhc3RJbmRleE9mKFwiLlwiKTtcbiAgaWYgKGlkeCA8IDApIHJldHVybiBcIlwiO1xuICByZXR1cm4gbmFtZS5zbGljZShpZHgpLnRvTG93ZXJDYXNlKCk7XG59XG5cbmZ1bmN0aW9uIGJhc2VuYW1lV2l0aG91dEV4dChwYXRoOiBzdHJpbmcpOiBzdHJpbmcge1xuICBjb25zdCBwID0gcGF0aC5yZXBsYWNlKC9cXFxcL2csIFwiL1wiKTtcbiAgY29uc3QgbmFtZSA9IHAuc3BsaXQoXCIvXCIpLnBvcCgpID8/IHA7XG4gIGNvbnN0IGlkeCA9IG5hbWUubGFzdEluZGV4T2YoXCIuXCIpO1xuICByZXR1cm4gaWR4ID49IDAgPyBuYW1lLnNsaWNlKDAsIGlkeCkgOiBuYW1lO1xufVxuXG5mdW5jdGlvbiBkaXJuYW1lKHBhdGg6IHN0cmluZyk6IHN0cmluZyB7XG4gIGNvbnN0IHAgPSBwYXRoLnJlcGxhY2UoL1xcXFwvZywgXCIvXCIpO1xuICBjb25zdCBpZHggPSBwLmxhc3RJbmRleE9mKFwiL1wiKTtcbiAgaWYgKGlkeCA8IDApIHJldHVybiBcIlwiO1xuICByZXR1cm4gcC5zbGljZSgwLCBpZHgpO1xufVxuXG5mdW5jdGlvbiByZW1vdmVGcm9udG1hdHRlcihyYXc6IHN0cmluZyk6IHN0cmluZyB7XG4gIGlmICghcmF3LnN0YXJ0c1dpdGgoXCItLS1cIikpIHJldHVybiByYXc7XG4gIGNvbnN0IGVuZElkeCA9IHJhdy5pbmRleE9mKFwiXFxuLS0tXCIsIDMpO1xuICBpZiAoZW5kSWR4ID09PSAtMSkgcmV0dXJuIHJhdztcbiAgY29uc3QgYWZ0ZXIgPSByYXcuc2xpY2UoZW5kSWR4ICsgXCJcXG4tLS1cIi5sZW5ndGgpO1xuICByZXR1cm4gYWZ0ZXIucmVwbGFjZSgvXlxccj9cXG4vLCBcIlwiKTtcbn1cblxuZnVuY3Rpb24gcGFyc2VNZXRhKHJhd01kOiBzdHJpbmcpOiB7XG4gIGF1dGhvcj86IHN0cmluZztcbiAgY2F0ZWdvcnk/OiBzdHJpbmc7XG4gIGNvbnRlbnQ6IHN0cmluZztcbn0ge1xuICBsZXQgYXV0aG9yOiBzdHJpbmcgfCB1bmRlZmluZWQ7XG4gIGxldCBjYXRlZ29yeTogc3RyaW5nIHwgdW5kZWZpbmVkO1xuICBsZXQgY29udGVudCA9IHJhd01kO1xuXG4gIC8vIDEpIFlBTUwgZnJvbnRtYXR0ZXIgYXQgdG9wXG4gIGlmIChyYXdNZC5zdGFydHNXaXRoKFwiLS0tXCIpKSB7XG4gICAgY29uc3QgZW5kID0gcmF3TWQuaW5kZXhPZihcIlxcbi0tLVwiLCAzKTtcbiAgICBpZiAoZW5kICE9PSAtMSkge1xuICAgICAgY29uc3QgZm0gPSByYXdNZC5zbGljZSgzLCBlbmQpLnNwbGl0KC9cXHI/XFxuLyk7XG4gICAgICBmb3IgKGNvbnN0IGxpbmUgb2YgZm0pIHtcbiAgICAgICAgY29uc3QgbSA9IGxpbmUubWF0Y2goL15cXHMqKGF1dGhvcnxjYXRlZ29yeSlcXHMqOlxccyooLispXFxzKiQvaSk7XG4gICAgICAgIGlmIChtKSB7XG4gICAgICAgICAgY29uc3Qga2V5ID0gbVsxXS50b0xvd2VyQ2FzZSgpO1xuICAgICAgICAgIGNvbnN0IHZhbCA9IG1bMl0udHJpbSgpLnJlcGxhY2UoL15bJ1wiXXxbJ1wiXSQvZywgXCJcIik7XG4gICAgICAgICAgaWYgKGtleSA9PT0gXCJhdXRob3JcIiAmJiB2YWwpIGF1dGhvciA9IHZhbDtcbiAgICAgICAgICBpZiAoa2V5ID09PSBcImNhdGVnb3J5XCIgJiYgdmFsKSBjYXRlZ29yeSA9IHZhbDtcbiAgICAgICAgfVxuICAgICAgfVxuICAgICAgY29udGVudCA9IHJlbW92ZUZyb250bWF0dGVyKHJhd01kKTtcbiAgICAgIHJldHVybiB7IGF1dGhvciwgY2F0ZWdvcnksIGNvbnRlbnQgfTtcbiAgICB9XG4gIH1cblxuICAvLyAyKSBmYWxsYmFjayB0b3AgbGluZXM6IGF1dGhvcjogLyBjYXRlZ29yeTpcbiAgY29uc3QgbGluZXMgPSByYXdNZC5zcGxpdCgvXFxyP1xcbi8pO1xuICBsZXQgaSA9IDA7XG4gIHdoaWxlIChpIDwgbGluZXMubGVuZ3RoKSB7XG4gICAgY29uc3QgbGluZSA9IGxpbmVzW2ldO1xuICAgIGlmICghbGluZS50cmltKCkpIHtcbiAgICAgIGkrKztcbiAgICAgIGNvbnRpbnVlO1xuICAgIH1cbiAgICBjb25zdCBtID0gbGluZS5tYXRjaCgvXlxccyooYXV0aG9yfGNhdGVnb3J5KVxccyo6XFxzKiguKylcXHMqJC9pKTtcbiAgICBpZiAoIW0pIGJyZWFrO1xuICAgIGNvbnN0IGtleSA9IG1bMV0udG9Mb3dlckNhc2UoKTtcbiAgICBjb25zdCB2YWwgPSBtWzJdLnRyaW0oKS5yZXBsYWNlKC9eWydcIl18WydcIl0kL2csIFwiXCIpO1xuICAgIGlmIChrZXkgPT09IFwiYXV0aG9yXCIgJiYgdmFsICYmICFhdXRob3IpIGF1dGhvciA9IHZhbDtcbiAgICBpZiAoa2V5ID09PSBcImNhdGVnb3J5XCIgJiYgdmFsICYmICFjYXRlZ29yeSkgY2F0ZWdvcnkgPSB2YWw7XG4gICAgaSsrO1xuICB9XG4gIGlmIChpID4gMCkgY29udGVudCA9IGxpbmVzLnNsaWNlKGkpLmpvaW4oXCJcXG5cIik7XG5cbiAgcmV0dXJuIHsgYXV0aG9yLCBjYXRlZ29yeSwgY29udGVudCB9O1xufVxuXG5mdW5jdGlvbiBlcnJUZXh0KGU6IHVua25vd24pOiBzdHJpbmcge1xuICBpZiAoZSBpbnN0YW5jZW9mIEVycm9yKSByZXR1cm4gZS5tZXNzYWdlO1xuICByZXR1cm4gU3RyaW5nKGUgPz8gXCJcdTY3MkFcdTc3RTVcdTk1MTlcdThCRUZcIik7XG59XG5cbi8qKlxuICogXHU1RTI2XHU4RDg1XHU2NUY2XHU3Njg0IGZldGNoXHUzMDAyXG4gKlxuICogXHU1MzlGXHU1MTQ4XHU2M0QyXHU0RUY2XHU3NkY0XHU2M0E1XHU0RjdGXHU3NTI4XHU2NUUwXHU4RDg1XHU2NUY2XHU3Njg0IGZldGNoXHVGRjBDXHU0RTAwXHU2NUU2XHU1NDBFXHU3QUVGXHU0RTBEXHU1M0VGXHU4RkJFXHVGRjA4XHU0RTIyXHU1MzA1L1x1ODhBQlx1OTYzMlx1NzA2Qlx1NTg5OVx1OTc1OVx1OUVEOFx1NEUyMlx1NUYwM1x1RkYwOVx1RkYwQ1xuICogXHU4QkY3XHU2QzQyXHU1M0VGXHU4MEZEXHU2MzAyXHU4RDc3XHU2NTcwXHU1MjA2XHU5NDlGXHVGRjBDXHU4MDBDIE9ic2lkaWFuIFx1NEYxQVx1N0I0OVx1NUY4NSBwbHVnaW4ub25sb2FkKCkgXHU1QjhDXHU2MjEwXHU2MjREXHU3RUU3XHU3RUVEXHU1NDJGXHU1MkE4XHVGRjBDXG4gKiBcdTRFOEVcdTY2MkZcdTg4NjhcdTczQjBcdTRFM0FcdTMwMENPYnNpZGlhbiBcdTYyNTNcdTRFMERcdTVGMDAgLyBcdTk3MDBcdTg5ODFcdTVCODlcdTUxNjhcdTZBMjFcdTVGMEZcdTMwMERcdTMwMDJcbiAqL1xuYXN5bmMgZnVuY3Rpb24gZmV0Y2hXaXRoVGltZW91dChcbiAgdXJsOiBzdHJpbmcsXG4gIGluaXQ6IFJlcXVlc3RJbml0ID0ge30sXG4gIHRpbWVvdXRNcyA9IFJFUVVFU1RfVElNRU9VVF9NUyxcbik6IFByb21pc2U8UmVzcG9uc2U+IHtcbiAgY29uc3QgY29udHJvbGxlciA9IG5ldyBBYm9ydENvbnRyb2xsZXIoKTtcbiAgY29uc3QgdGltZXIgPSB3aW5kb3cuc2V0VGltZW91dCgoKSA9PiBjb250cm9sbGVyLmFib3J0KCksIHRpbWVvdXRNcyk7XG4gIHRyeSB7XG4gICAgcmV0dXJuIGF3YWl0IGZldGNoKHVybCwgeyAuLi5pbml0LCBzaWduYWw6IGNvbnRyb2xsZXIuc2lnbmFsIH0pO1xuICB9IGNhdGNoIChlKSB7XG4gICAgaWYgKGNvbnRyb2xsZXIuc2lnbmFsLmFib3J0ZWQpIHtcbiAgICAgIHRocm93IG5ldyBFcnJvcihgXHU4QkY3XHU2QzQyXHU4RDg1XHU2NUY2XHVGRjA4JHt0aW1lb3V0TXN9bXNcdUZGMDk6ICR7dXJsfWApO1xuICAgIH1cbiAgICB0aHJvdyBuZXcgRXJyb3IoYFx1OEJGN1x1NkM0Mlx1NTkzMVx1OEQyNTogJHtlcnJUZXh0KGUpfSAoJHt1cmx9KWApO1xuICB9IGZpbmFsbHkge1xuICAgIHdpbmRvdy5jbGVhclRpbWVvdXQodGltZXIpO1xuICB9XG59XG5cbmFzeW5jIGZ1bmN0aW9uIHJlcUpzb24odXJsOiBzdHJpbmcsIGluaXQ/OiBSZXF1ZXN0SW5pdCk6IFByb21pc2U8YW55PiB7XG4gIGNvbnN0IHJlcyA9IGF3YWl0IGZldGNoV2l0aFRpbWVvdXQodXJsLCBpbml0KTtcbiAgaWYgKCFyZXMub2spIHtcbiAgICBjb25zdCB0ZXh0ID0gYXdhaXQgcmVzLnRleHQoKS5jYXRjaCgoKSA9PiBcIlwiKTtcbiAgICB0aHJvdyBuZXcgRXJyb3IoYEhUVFAgJHtyZXMuc3RhdHVzfSAke3Jlcy5zdGF0dXNUZXh0fSAke3RleHR9YCk7XG4gIH1cbiAgY29uc3QgY3QgPSByZXMuaGVhZGVycy5nZXQoXCJjb250ZW50LXR5cGVcIikgPz8gXCJcIjtcbiAgaWYgKGN0LmluY2x1ZGVzKFwiYXBwbGljYXRpb24vanNvblwiKSkgcmV0dXJuIHJlcy5qc29uKCk7XG4gIGNvbnN0IHR4dCA9IGF3YWl0IHJlcy50ZXh0KCk7XG4gIHRyeSB7XG4gICAgcmV0dXJuIEpTT04ucGFyc2UodHh0KTtcbiAgfSBjYXRjaCB7XG4gICAgcmV0dXJuIHR4dDtcbiAgfVxufVxuXG5jbGFzcyBBcnRpY2xlUGlja2VyTW9kYWwgZXh0ZW5kcyBGdXp6eVN1Z2dlc3RNb2RhbDxSYXdBcnRpY2xlPiB7XG4gIHByaXZhdGUgcmVhZG9ubHkgaXRlbXM6IFJhd0FydGljbGVbXTtcbiAgcHVibGljIG9uQ2hvb3NlOiAoaXQ6IFJhd0FydGljbGUpID0+IHZvaWQ7XG5cbiAgY29uc3RydWN0b3IoXG4gICAgYXBwOiBBcHAsXG4gICAgaXRlbXM6IFJhd0FydGljbGVbXSxcbiAgICBvbkNob29zZTogKGl0OiBSYXdBcnRpY2xlKSA9PiB2b2lkLFxuICApIHtcbiAgICBzdXBlcihhcHApO1xuICAgIHRoaXMuaXRlbXMgPSBpdGVtcztcbiAgICB0aGlzLm9uQ2hvb3NlID0gb25DaG9vc2U7XG4gICAgdGhpcy5zZXRQbGFjZWhvbGRlcihcIlx1OTAwOVx1NjJFOVx1ODk4MVx1NTIyMFx1OTY2NFx1NzY4NFx1NjU4N1x1N0FFMC4uLlwiKTtcbiAgfVxuXG4gIGdldEl0ZW1zKCk6IFJhd0FydGljbGVbXSB7XG4gICAgcmV0dXJuIHRoaXMuaXRlbXM7XG4gIH1cblxuICBnZXRJdGVtVGV4dChpdGVtOiBSYXdBcnRpY2xlKTogc3RyaW5nIHtcbiAgICByZXR1cm4gaXRlbS50aXRsZTtcbiAgfVxuXG4gIG9uQ2hvb3NlSXRlbShpdGVtOiBSYXdBcnRpY2xlKTogdm9pZCB7XG4gICAgdGhpcy5vbkNob29zZShpdGVtKTtcbiAgfVxufVxuXG5jbGFzcyBGb2xkZXJQaWNrZXJNb2RhbCBleHRlbmRzIEZ1enp5U3VnZ2VzdE1vZGFsPFRGb2xkZXI+IHtcbiAgcHJpdmF0ZSByZWFkb25seSBmb2xkZXJzOiBURm9sZGVyW107XG4gIHByaXZhdGUgcmVhZG9ubHkgb25DaG9vc2VDYjogKGZvbGRlcjogVEZvbGRlcikgPT4gdm9pZDtcblxuICBjb25zdHJ1Y3RvcihcbiAgICBhcHA6IEFwcCxcbiAgICBmb2xkZXJzOiBURm9sZGVyW10sXG4gICAgb25DaG9vc2U6IChmb2xkZXI6IFRGb2xkZXIpID0+IHZvaWQsXG4gICkge1xuICAgIHN1cGVyKGFwcCk7XG4gICAgdGhpcy5mb2xkZXJzID0gZm9sZGVycztcbiAgICB0aGlzLm9uQ2hvb3NlQ2IgPSBvbkNob29zZTtcbiAgICB0aGlzLnNldFBsYWNlaG9sZGVyKFwiXHU5MDA5XHU2MkU5XHU4OTgxXHU0RTBBXHU0RjIwXHU3Njg0XHU2NTg3XHU0RUY2XHU1OTM5Li4uXCIpO1xuICB9XG5cbiAgZ2V0SXRlbXMoKTogVEZvbGRlcltdIHtcbiAgICByZXR1cm4gdGhpcy5mb2xkZXJzO1xuICB9XG5cbiAgZ2V0SXRlbVRleHQoaXRlbTogVEZvbGRlcik6IHN0cmluZyB7XG4gICAgcmV0dXJuIGl0ZW0ucGF0aCB8fCBcIi9cIjtcbiAgfVxuXG4gIG9uQ2hvb3NlSXRlbShpdGVtOiBURm9sZGVyKTogdm9pZCB7XG4gICAgdGhpcy5vbkNob29zZUNiKGl0ZW0pO1xuICB9XG59XG5cbmNsYXNzIENvbmZpcm1Nb2RhbCBleHRlbmRzIE1vZGFsIHtcbiAgcHJpdmF0ZSByZWFkb25seSBtZXNzYWdlOiBzdHJpbmc7XG4gIHByaXZhdGUgcmVhZG9ubHkgb25Db25maXJtOiAoKSA9PiB2b2lkO1xuXG4gIGNvbnN0cnVjdG9yKGFwcDogQXBwLCBtZXNzYWdlOiBzdHJpbmcsIG9uQ29uZmlybTogKCkgPT4gdm9pZCkge1xuICAgIHN1cGVyKGFwcCk7XG4gICAgdGhpcy5tZXNzYWdlID0gbWVzc2FnZTtcbiAgICB0aGlzLm9uQ29uZmlybSA9IG9uQ29uZmlybTtcbiAgfVxuXG4gIG9uT3BlbigpOiB2b2lkIHtcbiAgICBjb25zdCB7IGNvbnRlbnRFbCB9ID0gdGhpcztcbiAgICBjb250ZW50RWwuZW1wdHkoKTtcbiAgICBjb250ZW50RWwuY3JlYXRlRWwoXCJoM1wiLCB7IHRleHQ6IFwiXHU3ODZFXHU4QkE0XHU2NENEXHU0RjVDXCIgfSk7XG4gICAgY29udGVudEVsLmNyZWF0ZUVsKFwicFwiLCB7IHRleHQ6IHRoaXMubWVzc2FnZSB9KTtcblxuICAgIGNvbnN0IGFjdGlvbnMgPSBjb250ZW50RWwuY3JlYXRlRGl2KHsgY2xzOiBcIm1rYmxvZy1tb2RhbC1hY3Rpb25zXCIgfSk7XG4gICAgY29uc3QgY2FuY2VsQnRuID0gYWN0aW9ucy5jcmVhdGVFbChcImJ1dHRvblwiLCB7IHRleHQ6IFwiXHU1M0Q2XHU2RDg4XCIgfSk7XG4gICAgY29uc3Qgb2tCdG4gPSBhY3Rpb25zLmNyZWF0ZUVsKFwiYnV0dG9uXCIsIHsgdGV4dDogXCJcdTUyMjBcdTk2NjRcIiB9KTtcbiAgICBva0J0bi5hZGRDbGFzcyhcIm1vZC13YXJuaW5nXCIpO1xuXG4gICAgY2FuY2VsQnRuLm9uY2xpY2sgPSAoKSA9PiB0aGlzLmNsb3NlKCk7XG4gICAgb2tCdG4ub25jbGljayA9ICgpID0+IHtcbiAgICAgIHRoaXMuY2xvc2UoKTtcbiAgICAgIHRoaXMub25Db25maXJtKCk7XG4gICAgfTtcbiAgfVxuXG4gIG9uQ2xvc2UoKTogdm9pZCB7XG4gICAgdGhpcy5jb250ZW50RWwuZW1wdHkoKTtcbiAgfVxufVxuXG5jbGFzcyBNa0Jsb2dBcnRpY2xlc1ZpZXcgZXh0ZW5kcyBJdGVtVmlldyB7XG4gIHByaXZhdGUgcGx1Z2luOiBNa0Jsb2dQbHVnaW47XG4gIHByaXZhdGUgbGlzdEVsOiBIVE1MRWxlbWVudCB8IG51bGwgPSBudWxsO1xuXG4gIGNvbnN0cnVjdG9yKGxlYWY6IFdvcmtzcGFjZUxlYWYsIHBsdWdpbjogTWtCbG9nUGx1Z2luKSB7XG4gICAgc3VwZXIobGVhZik7XG4gICAgdGhpcy5wbHVnaW4gPSBwbHVnaW47XG4gIH1cblxuICBnZXRWaWV3VHlwZSgpOiBzdHJpbmcge1xuICAgIHJldHVybiBWSUVXX1RZUEVfTUtCTE9HO1xuICB9XG5cbiAgZ2V0RGlzcGxheVRleHQoKTogc3RyaW5nIHtcbiAgICByZXR1cm4gXCJta0Jsb2dcIjtcbiAgfVxuXG4gIGdldEljb24oKTogc3RyaW5nIHtcbiAgICByZXR1cm4gXCJub3RlYm9vay1wZW5cIjtcbiAgfVxuXG4gIGFzeW5jIG9uT3BlbigpOiBQcm9taXNlPHZvaWQ+IHtcbiAgICB0aGlzLmNvbnRlbnRFbC5lbXB0eSgpO1xuICAgIHRoaXMuY29udGVudEVsLmFkZENsYXNzKFwibWtibG9nLXZpZXdcIik7XG5cbiAgICBjb25zdCBoZWFkZXIgPSB0aGlzLmNvbnRlbnRFbC5jcmVhdGVEaXYoeyBjbHM6IFwibWtibG9nLWhlYWRlclwiIH0pO1xuICAgIGhlYWRlci5jcmVhdGVFbChcImgzXCIsIHsgdGV4dDogXCJta0Jsb2cgXHU2NTg3XHU3QUUwXHU3QkExXHU3NDA2XCIgfSk7XG5cbiAgICBjb25zdCBhY3Rpb25zID0gaGVhZGVyLmNyZWF0ZURpdih7IGNsczogXCJta2Jsb2ctYWN0aW9uc1wiIH0pO1xuICAgIGNvbnN0IHJlZnJlc2hCdG4gPSBhY3Rpb25zLmNyZWF0ZUVsKFwiYnV0dG9uXCIsIHsgdGV4dDogXCJcdTUyMzdcdTY1QjBcIiB9KTtcbiAgICBjb25zdCB1cGxvYWRGaWxlQnRuID0gYWN0aW9ucy5jcmVhdGVFbChcImJ1dHRvblwiLCB7IHRleHQ6IFwiXHU0RTBBXHU0RjIwXHU1RjUzXHU1MjREXHU2NTg3XHU0RUY2XCIgfSk7XG4gICAgY29uc3QgdXBsb2FkRm9sZGVyQnRuID0gYWN0aW9ucy5jcmVhdGVFbChcImJ1dHRvblwiLCB7IHRleHQ6IFwiXHU0RTBBXHU0RjIwXHU2NTg3XHU0RUY2XHU1OTM5XCIgfSk7XG5cbiAgICByZWZyZXNoQnRuLm9uY2xpY2sgPSBhc3luYyAoKSA9PiB7XG4gICAgICBhd2FpdCB0aGlzLnBsdWdpbi5yZWZyZXNoQXJ0aWNsZXMoKTtcbiAgICB9O1xuICAgIHVwbG9hZEZpbGVCdG4ub25jbGljayA9IGFzeW5jICgpID0+IHtcbiAgICAgIGF3YWl0IHRoaXMucGx1Z2luLnVwbG9hZEN1cnJlbnRGaWxlQXNCbG9nKCk7XG4gICAgfTtcbiAgICB1cGxvYWRGb2xkZXJCdG4ub25jbGljayA9IGFzeW5jICgpID0+IHtcbiAgICAgIGF3YWl0IHRoaXMucGx1Z2luLnBpY2tBbmRVcGxvYWRGb2xkZXIoKTtcbiAgICB9O1xuXG4gICAgdGhpcy5saXN0RWwgPSB0aGlzLmNvbnRlbnRFbC5jcmVhdGVEaXYoeyBjbHM6IFwibWtibG9nLWxpc3RcIiB9KTtcbiAgICB0aGlzLnJlbmRlckxpc3QoKTtcbiAgfVxuXG4gIHJlbmRlckxpc3QoKTogdm9pZCB7XG4gICAgaWYgKCF0aGlzLmxpc3RFbCkgcmV0dXJuO1xuICAgIHRoaXMubGlzdEVsLmVtcHR5KCk7XG5cbiAgICBjb25zdCBpdGVtcyA9IHRoaXMucGx1Z2luLmFydGljbGVzO1xuICAgIGlmICghaXRlbXMubGVuZ3RoKSB7XG4gICAgICB0aGlzLmxpc3RFbC5jcmVhdGVFbChcImRpdlwiLCB7XG4gICAgICAgIHRleHQ6IFwiXHU2NjgyXHU2NUUwXHU2NTg3XHU3QUUwXHVGRjA4XHU1M0VGXHU3MEI5XHU1MUZCXHU1MjM3XHU2NUIwXHVGRjA5XCIsXG4gICAgICAgIGNsczogXCJta2Jsb2ctZW1wdHlcIixcbiAgICAgIH0pO1xuICAgICAgcmV0dXJuO1xuICAgIH1cblxuICAgIGZvciAoY29uc3QgaXQgb2YgaXRlbXMpIHtcbiAgICAgIGNvbnN0IHJvdyA9IHRoaXMubGlzdEVsLmNyZWF0ZURpdih7IGNsczogXCJta2Jsb2ctcm93XCIgfSk7XG4gICAgICBjb25zdCB0aXRsZUVsID0gcm93LmNyZWF0ZURpdih7IHRleHQ6IGl0LnRpdGxlLCBjbHM6IFwibWtibG9nLXRpdGxlXCIgfSk7XG4gICAgICB0aXRsZUVsLnNldEF0dHJpYnV0ZShcInRpdGxlXCIsIGAke2l0LnRpdGxlfSAoSUQ6ICR7U3RyaW5nKGl0LmlkKX0pYCk7XG5cbiAgICAgIGNvbnN0IGRlbEJ0biA9IHJvdy5jcmVhdGVFbChcImJ1dHRvblwiLCB7IHRleHQ6IFwiXHU1MjIwXHU5NjY0XCIgfSk7XG4gICAgICBkZWxCdG4uYWRkQ2xhc3MoXCJtb2Qtd2FybmluZ1wiKTtcbiAgICAgIGRlbEJ0bi5vbmNsaWNrID0gYXN5bmMgKCkgPT4ge1xuICAgICAgICB0aGlzLnBsdWdpbi5jb25maXJtRGVsZXRlKGl0KTtcbiAgICAgIH07XG4gICAgfVxuICB9XG5cbiAgYXN5bmMgb25DbG9zZSgpOiBQcm9taXNlPHZvaWQ+IHtcbiAgICB0aGlzLmNvbnRlbnRFbC5lbXB0eSgpO1xuICB9XG59XG5cbmV4cG9ydCBkZWZhdWx0IGNsYXNzIE1rQmxvZ1BsdWdpbiBleHRlbmRzIFBsdWdpbiB7XG4gIHNldHRpbmdzOiBNa0Jsb2dTZXR0aW5ncyA9IHsgLi4uREVGQVVMVF9TRVRUSU5HUyB9O1xuICBhcnRpY2xlczogUmF3QXJ0aWNsZVtdID0gW107XG5cbiAgYXN5bmMgb25sb2FkKCk6IFByb21pc2U8dm9pZD4ge1xuICAgIC8vIC0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLVxuICAgIC8vIDEpIFx1NTNFQVx1NTA1QVx1MzAwQ1x1NTQwQ1x1NkI2NVx1NkNFOFx1NTE4Q1x1MzAwRFx1RkYwQ1x1NEUxNCByZWdpc3RlclZpZXcgXHU1RkM1XHU5ODdCXHU2NjJGXHU3QjJDXHU0RTAwXHU2NzYxXHU4QkVEXHU1M0U1XHUzMDAyXG4gICAgLy9cbiAgICAvLyBPYnNpZGlhbiBcdTYwNjJcdTU5MERcdTVERTVcdTRGNUNcdTUzM0FcdTVFMDNcdTVDNDBcdTY1RjZcdUZGMENcdTgyRTVcdTY3RDBcdTRFMkEgdmlldyB0eXBlIFx1OEZEOFx1NjcyQVx1NkNFOFx1NTE4Q1x1RkYwQ1x1NUMzMVx1NEYxQVx1NEUzQVx1OEJFNVx1NTNGNlx1NUI1MFxuICAgIC8vIFx1NTIxQlx1NUVGQVx1NEUwMFx1NEUyQSBFbXB0eVZpZXdcdUZGMDhcdTRGRERcdTVCNThcdTUyMzAgd29ya3NwYWNlLmpzb24gXHU1NDBFXHU4ODY4XHU3M0IwXHU0RTNBXG4gICAgLy8gXCJpY29uXCI6IFwibHVjaWRlLWdob3N0XCJcdTMwMDFcInRpdGxlXCI6IFwibWtibG9nLWFydGljbGVzLXZpZXdcIlx1RkYwOVx1MzAwMlxuICAgIC8vIFx1OEZEOVx1NEUyQSBnaG9zdCBcdTUzRjZcdTVCNTBcdTRGMUFcdTg4QUJcdTUzQ0RcdTU5MERcdTUxOTlcdTU2REVcdTc4QzFcdTc2RDhcdUZGMENcdTRFOEVcdTY2MkZcdTZCQ0ZcdTZCMjFcdTU0MkZcdTUyQThcdTkwRkRcdTYyQTVcdTg5QzZcdTU2RkUvXHU2M0QyXHU0RUY2XHU1MkEwXHU4RjdEXHU1OTMxXHU4RDI1XHVGRjBDXG4gICAgLy8gXHU0RTI1XHU5MUNEXHU2NUY2XHU3NkY0XHU2M0E1XHU1MzYxXHU0RjRGXHU1RTAzXHU1QzQwXHU2MDYyXHU1OTBEXHVGRjBDXHU1QkZDXHU4MUY0IE9ic2lkaWFuIFx1OTcwMFx1ODk4MVx1NEVFNVx1NUI4OVx1NTE2OFx1NkEyMVx1NUYwRlx1NTQyRlx1NTJBOFx1MzAwMlxuICAgIC8vXG4gICAgLy8gXHU0RTRCXHU1MjREIHJlZ2lzdGVyVmlldyBcdTRFNEJcdTUyNERcdTY3MDkgYGF3YWl0IHRoaXMubG9hZFNldHRpbmdzKClgXHVGRjBDXHU2Q0U4XHU1MThDXHU4OEFCXHU2M0E4XHU4RkRGXHU1MjMwXG4gICAgLy8gXHU0RTAwXHU2QjIxXHU3OEMxXHU3NkQ4IEkvTyBcdTRFNEJcdTU0MEVcdUZGMENcdTVCNThcdTU3MjhcdTYwNjJcdTU5MERcdTVFMDNcdTVDNDBcdTY1RjZcdTRFQ0RcdTY3MkFcdTZDRThcdTUxOENcdTc2ODRcdTdBREVcdTYwMDFcdTMwMDJcbiAgICAvLyAtLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS1cbiAgICB0aGlzLnJlZ2lzdGVyVmlldyhcbiAgICAgIFZJRVdfVFlQRV9NS0JMT0csXG4gICAgICAobGVhZikgPT4gbmV3IE1rQmxvZ0FydGljbGVzVmlldyhsZWFmLCB0aGlzKSxcbiAgICApO1xuXG4gICAgdGhpcy5hZGRTZXR0aW5nVGFiKG5ldyBNa0Jsb2dTZXR0aW5nVGFiKHRoaXMuYXBwLCB0aGlzKSk7XG5cbiAgICB0aGlzLmFkZFJpYmJvbkljb24oXCJjbG91ZC11cGxvYWRcIiwgXCJta0Jsb2c6IFx1NEUwQVx1NEYyMFx1NUY1M1x1NTI0RFx1NjU4N1x1NEVGNlx1NEUzQVx1NTM1QVx1NUJBMlwiLCAoKSA9PiB7XG4gICAgICB2b2lkIHRoaXMudXBsb2FkQ3VycmVudEZpbGVBc0Jsb2coKTtcbiAgICB9KTtcblxuICAgIHRoaXMuYWRkQ29tbWFuZCh7XG4gICAgICBpZDogXCJta2Jsb2ctb3Blbi12aWV3XCIsXG4gICAgICBuYW1lOiBcIm1rQmxvZzogXHU2MjUzXHU1RjAwXHU3QkExXHU3NDA2XHU4OUM2XHU1NkZFXCIsXG4gICAgICBjYWxsYmFjazogKCkgPT4ge1xuICAgICAgICB2b2lkIHRoaXMuYWN0aXZhdGVWaWV3KCk7XG4gICAgICB9LFxuICAgIH0pO1xuXG4gICAgdGhpcy5hZGRDb21tYW5kKHtcbiAgICAgIGlkOiBcIm1rYmxvZy11cGxvYWQtY3VycmVudC1maWxlXCIsXG4gICAgICBuYW1lOiBcIm1rQmxvZzogXHU0RTBBXHU0RjIwXHU1RjUzXHU1MjREXHU2NTg3XHU0RUY2XHU0RTNBXHU1MzVBXHU1QkEyXCIsXG4gICAgICBjYWxsYmFjazogKCkgPT4ge1xuICAgICAgICB2b2lkIHRoaXMudXBsb2FkQ3VycmVudEZpbGVBc0Jsb2coKTtcbiAgICAgIH0sXG4gICAgfSk7XG5cbiAgICB0aGlzLmFkZENvbW1hbmQoe1xuICAgICAgaWQ6IFwibWtibG9nLXVwbG9hZC1mb2xkZXJcIixcbiAgICAgIG5hbWU6IFwibWtCbG9nOiBcdTRFMEFcdTRGMjBcdTkwMDlcdTYyRTlcdTY1ODdcdTRFRjZcdTU5MzlcdTRFM0FcdTUzNUFcdTVCQTJcIixcbiAgICAgIGNhbGxiYWNrOiAoKSA9PiB7XG4gICAgICAgIHZvaWQgdGhpcy5waWNrQW5kVXBsb2FkRm9sZGVyKCk7XG4gICAgICB9LFxuICAgIH0pO1xuXG4gICAgdGhpcy5hZGRDb21tYW5kKHtcbiAgICAgIGlkOiBcIm1rYmxvZy1yZWZyZXNoLWFydGljbGVzXCIsXG4gICAgICBuYW1lOiBcIm1rQmxvZzogXHU1MjM3XHU2NUIwXHU2NTg3XHU3QUUwXHU1MjE3XHU4ODY4XCIsXG4gICAgICBjYWxsYmFjazogKCkgPT4ge1xuICAgICAgICB2b2lkIHRoaXMucmVmcmVzaEFydGljbGVzKCk7XG4gICAgICB9LFxuICAgIH0pO1xuXG4gICAgdGhpcy5hZGRDb21tYW5kKHtcbiAgICAgIGlkOiBcIm1rYmxvZy1kZWxldGUtYXJ0aWNsZVwiLFxuICAgICAgbmFtZTogXCJta0Jsb2c6IFx1NTIyMFx1OTY2NFx1NjU4N1x1N0FFMFwiLFxuICAgICAgY2FsbGJhY2s6ICgpID0+IHtcbiAgICAgICAgdm9pZCB0aGlzLnBpY2tBbmREZWxldGVBcnRpY2xlKCk7XG4gICAgICB9LFxuICAgIH0pO1xuXG4gICAgLy8gLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tXG4gICAgLy8gMikgXHU4QkZCXHU1M0Q2XHU5MTREXHU3RjZFXHUzMDAyXHU4RkQ5XHU2NjJGIG9ubG9hZCBcdTRFMkRcdTU1MkZcdTRFMDBcdTc2ODQgSS9PXHVGRjBDXHU0RTE0XHU1MTg1XHU5MEU4XHU1REYyIHRyeS9jYXRjaFx1RkYwQ1xuICAgIC8vICAgIG9ubG9hZCBcdTZDMzhcdThGRENcdTRFMERcdTRGMUEgcmVqZWN0XHVGRjA4XHU0RUU1XHU1MjREIGRhdGEuanNvbiBcdTYzNUZcdTU3NEZcdTRGMUFcdTc2RjRcdTYzQTVcdTVCRkNcdTgxRjRcdTYzRDJcdTRFRjZcdTUyQTBcdThGN0RcdTU5MzFcdThEMjVcdUZGMDlcdTMwMDJcbiAgICAvLyAtLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS1cbiAgICBhd2FpdCB0aGlzLmxvYWRTZXR0aW5ncygpO1xuXG4gICAgLy8gLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tXG4gICAgLy8gMykgXHU3RjUxXHU3RURDXHU4QkY3XHU2QzQyIC8gXHU1REU1XHU0RjVDXHU1MzNBXHU2NTM5XHU1MkE4XHU0RTAwXHU1RjhCXHU2M0E4XHU4RkRGXHU1MjMwXHU1RTAzXHU1QzQwXHU1QzMxXHU3RUVBXHU0RTRCXHU1NDBFXHUzMDAyXG4gICAgLy8gICAgXHU1Qjk4XHU2NUI5XHU2NTg3XHU2ODYzXHU2NjBFXHU3ODZFXHU4OTgxXHU2QzQyIG9ubG9hZCBcdTRFMERcdTVGOTdcdTUwNUFcdTY1NzBcdTYzNkVcdTYyQzlcdTUzRDZcdUZGMUJcdTU0MjZcdTUyMTlcdTRGMUFcdTY2M0VcdTg0NTdcdTYyRDZcdTYxNjJcdTc1MUFcdTgxRjNcdTk2M0JcdTU4NUVcdTU0MkZcdTUyQThcdTMwMDJcbiAgICAvLyAtLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS1cbiAgICBpZiAodGhpcy5hcHAud29ya3NwYWNlLmxheW91dFJlYWR5KSB7XG4gICAgICB2b2lkIHRoaXMucnVuU3RhcnR1cFRhc2tzKCk7XG4gICAgfSBlbHNlIHtcbiAgICAgIHRoaXMuYXBwLndvcmtzcGFjZS5vbkxheW91dFJlYWR5KCgpID0+IHtcbiAgICAgICAgdm9pZCB0aGlzLnJ1blN0YXJ0dXBUYXNrcygpO1xuICAgICAgfSk7XG4gICAgfVxuICB9XG5cbiAgcHJpdmF0ZSBhc3luYyBydW5TdGFydHVwVGFza3MoKTogUHJvbWlzZTx2b2lkPiB7XG4gICAgdHJ5IHtcbiAgICAgIGF3YWl0IHRoaXMuaGVhbEdob3N0TGVhdmVzKCk7XG4gICAgICBpZiAodGhpcy5zZXR0aW5ncy5vcGVuVmlld09uU3RhcnR1cCkge1xuICAgICAgICBhd2FpdCB0aGlzLmFjdGl2YXRlVmlldygpO1xuICAgICAgfVxuICAgICAgaWYgKHRoaXMuc2V0dGluZ3MucmVmcmVzaE9uU3RhcnR1cCkge1xuICAgICAgICBhd2FpdCB0aGlzLnJlZnJlc2hBcnRpY2xlcyh7IHNpbGVudDogdHJ1ZSB9KTtcbiAgICAgIH0gZWxzZSB7XG4gICAgICAgIHRoaXMucmVkcmF3Vmlld3MoKTtcbiAgICAgIH1cbiAgICB9IGNhdGNoIChlKSB7XG4gICAgICBjb25zb2xlLmVycm9yKFwiW21rQmxvZ10gc3RhcnR1cCB0YXNrcyBmYWlsZWRcIiwgZSk7XG4gICAgfVxuICB9XG5cbiAgLyoqXG4gICAqIFx1NEZFRVx1NTkwRFx1NTM4Nlx1NTNGMlx1OTA1N1x1NzU1OVx1NzY4NCBnaG9zdCBcdTUzRjZcdTVCNTBcdTMwMDJcbiAgICpcbiAgICogXHU2NUU3XHU3MjQ4XHU2NzJDXHU1NzI4XHU2MDYyXHU1OTBEXHU1RTAzXHU1QzQwXHU2NUY2IHZpZXcgdHlwZSBcdThGRDhcdTZDQTFcdTZDRThcdTUxOENcdUZGMENPYnNpZGlhbiBcdTRGMUFcdTRFM0FcdTVCODNcdTUyMUJcdTVFRkFcdTRFMDBcdTRFMkFcbiAgICogRW1wdHlWaWV3IFx1NTM2MFx1NEY0RFx1NTNGNlx1NUI1MFx1NUU3Nlx1NTE5OVx1NTZERSB3b3Jrc3BhY2UuanNvblx1RkYwOGljb249bHVjaWRlLWdob3N0XHVGRjA5XHUzMDAyXG4gICAqIFx1OEZEOVx1OTFDQ1x1NjI4QVx1NUI4M1x1OTFDRFx1NjVCMFx1NUI5RVx1NEY4Qlx1NTMxNlx1NEUzQVx1NzcxRlx1NUI5RVx1ODlDNlx1NTZGRVx1RkYwQ1x1OTA3Rlx1NTE0RFx1NzUyOFx1NjIzN1x1OTcwMFx1ODk4MVx1NjI0Qlx1NTJBOFx1NjUzOSB3b3Jrc3BhY2UuanNvblx1MzAwMlxuICAgKlxuICAgKiBcdTZDRThcdTYxMEZcdUZGMUFPYnNpZGlhbiAxLjcuMisgXHU3Njg0IERlZmVycmVkVmlldyBcdTY2MkZcdTZCNjNcdTVFMzhcdTcyQjZcdTYwMDFcdUZGMDhcdTUzRUZcdTg5QzFcdTY1RjZcdTYyNERcdTVCOUVcdTRGOEJcdTUzMTZcdUZGMDlcdUZGMENcbiAgICogXHU0RTBEXHU4MEZEXHU4QkVGXHU0RjI0XHUzMDAyXG4gICAqL1xuICBwcml2YXRlIGFzeW5jIGhlYWxHaG9zdExlYXZlcygpOiBQcm9taXNlPHZvaWQ+IHtcbiAgICBmb3IgKGNvbnN0IGxlYWYgb2YgdGhpcy5hcHAud29ya3NwYWNlLmdldExlYXZlc09mVHlwZShWSUVXX1RZUEVfTUtCTE9HKSkge1xuICAgICAgaWYgKGxlYWYudmlldyBpbnN0YW5jZW9mIE1rQmxvZ0FydGljbGVzVmlldykgY29udGludWU7XG4gICAgICBpZiAocmVxdWlyZUFwaVZlcnNpb24oXCIxLjcuMlwiKSAmJiBsZWFmLmlzRGVmZXJyZWQpIGNvbnRpbnVlO1xuICAgICAgdHJ5IHtcbiAgICAgICAgYXdhaXQgbGVhZi5zZXRWaWV3U3RhdGUoeyB0eXBlOiBWSUVXX1RZUEVfTUtCTE9HLCBhY3RpdmU6IGZhbHNlIH0pO1xuICAgICAgICBjb25zb2xlLmxvZyhcIltta0Jsb2ddIHJlcGFpcmVkIGdob3N0IHZpZXcgbGVhZlwiKTtcbiAgICAgIH0gY2F0Y2ggKGUpIHtcbiAgICAgICAgY29uc29sZS5lcnJvcihcIltta0Jsb2ddIGZhaWxlZCB0byByZXBhaXIgZ2hvc3QgdmlldyBsZWFmXCIsIGUpO1xuICAgICAgfVxuICAgIH1cbiAgfVxuXG4gIG9udW5sb2FkKCk6IHZvaWQge1xuICAgIC8vIFx1NjNEMlx1NEVGNlx1ODhBQlx1Nzk4MVx1NzUyOFx1NjVGNlx1NkUwNVx1NjM4OVx1ODFFQVx1NURGMVx1NzY4NFx1NTNGNlx1NUI1MFx1RkYwQ1x1OTA3Rlx1NTE0RFx1NzU1OVx1NEUwQlx1NjVFMFx1NkNENVx1ODlFM1x1Njc5MFx1NzY4NCBnaG9zdCBcdTg5QzZcdTU2RkVcdTMwMDJcbiAgICB0aGlzLmFwcC53b3Jrc3BhY2UuZGV0YWNoTGVhdmVzT2ZUeXBlKFZJRVdfVFlQRV9NS0JMT0cpO1xuICB9XG5cbiAgcHJpdmF0ZSBhdXRoSGVhZGVycyhqc29uID0gZmFsc2UpOiBSZWNvcmQ8c3RyaW5nLCBzdHJpbmc+IHtcbiAgICBjb25zdCBoOiBSZWNvcmQ8c3RyaW5nLCBzdHJpbmc+ID0geyBBY2NlcHQ6IFwiYXBwbGljYXRpb24vanNvblwiIH07XG4gICAgaWYgKHRoaXMuc2V0dGluZ3MuYXV0aFRva2VuPy50cmltKCkpIHtcbiAgICAgIGhbXCJBdXRob3JpemF0aW9uXCJdID0gYEJlYXJlciAke3RoaXMuc2V0dGluZ3MuYXV0aFRva2VuLnRyaW0oKX1gO1xuICAgIH1cbiAgICBpZiAoanNvbikgaFtcIkNvbnRlbnQtVHlwZVwiXSA9IFwiYXBwbGljYXRpb24vanNvblwiO1xuICAgIHJldHVybiBoO1xuICB9XG5cbiAgcHJpdmF0ZSBlbnN1cmVCYXNlVXJsKCk6IHN0cmluZyB7XG4gICAgY29uc3QgYmFzZSA9ICh0aGlzLnNldHRpbmdzLmJhc2VVcmwgfHwgXCJcIikudHJpbSgpO1xuICAgIGlmICghYmFzZSkgdGhyb3cgbmV3IEVycm9yKFwiXHU2NzJBXHU5MTREXHU3RjZFIEJhc2UgVVJMXCIpO1xuICAgIHJldHVybiBiYXNlO1xuICB9XG5cbiAgcHJpdmF0ZSBhc3luYyBmaWxlVG9BcnJheUJ1ZmZlcih2YXVsdFBhdGg6IHN0cmluZyk6IFByb21pc2U8QXJyYXlCdWZmZXI+IHtcbiAgICByZXR1cm4gYXdhaXQgdGhpcy5hcHAudmF1bHQuYWRhcHRlci5yZWFkQmluYXJ5KHZhdWx0UGF0aCk7XG4gIH1cblxuICBwcml2YXRlIGFzeW5jIGNvbGxlY3RJbWFnZXNGb3JNYXJrZG93bihcbiAgICBtZEZpbGU6IFRGaWxlLFxuICApOiBQcm9taXNlPHsgbmFtZTogc3RyaW5nOyBkYXRhQmFzZTY0OiBzdHJpbmcgfVtdPiB7XG4gICAgY29uc3QgcGFyZW50ID0gZGlybmFtZShtZEZpbGUucGF0aCk7XG4gICAgY29uc3QgdGl0bGUgPSBiYXNlbmFtZVdpdGhvdXRFeHQobWRGaWxlLm5hbWUpO1xuICAgIGNvbnN0IGltZ0ZvbGRlclBhdGggPSBub3JtYWxpemVQYXRoKHBhcmVudCA/IGAke3BhcmVudH0vJHt0aXRsZX1gIDogdGl0bGUpO1xuXG4gICAgY29uc3QgZm9sZGVyID0gdGhpcy5hcHAudmF1bHQuZ2V0QWJzdHJhY3RGaWxlQnlQYXRoKGltZ0ZvbGRlclBhdGgpO1xuICAgIGlmICghZm9sZGVyIHx8ICEoZm9sZGVyIGluc3RhbmNlb2YgVEZvbGRlcikpIHJldHVybiBbXTtcblxuICAgIGNvbnN0IG91dDogeyBuYW1lOiBzdHJpbmc7IGRhdGFCYXNlNjQ6IHN0cmluZyB9W10gPSBbXTtcbiAgICBjb25zdCBzdGFjazogVEZvbGRlcltdID0gW2ZvbGRlcl07XG5cbiAgICB3aGlsZSAoc3RhY2subGVuZ3RoID4gMCkge1xuICAgICAgY29uc3QgY3VyID0gc3RhY2sucG9wKCkhO1xuICAgICAgZm9yIChjb25zdCBjaGlsZCBvZiBjdXIuY2hpbGRyZW4pIHtcbiAgICAgICAgaWYgKGNoaWxkIGluc3RhbmNlb2YgVEZvbGRlcikge1xuICAgICAgICAgIHN0YWNrLnB1c2goY2hpbGQpO1xuICAgICAgICAgIGNvbnRpbnVlO1xuICAgICAgICB9XG4gICAgICAgIGlmICghKGNoaWxkIGluc3RhbmNlb2YgVEZpbGUpKSBjb250aW51ZTtcbiAgICAgICAgY29uc3QgZXh0ID0gZXh0bmFtZShjaGlsZC5uYW1lKTtcbiAgICAgICAgaWYgKCFJTUdfRVhULmhhcyhleHQpKSBjb250aW51ZTtcblxuICAgICAgICBjb25zdCBidWYgPSBhd2FpdCB0aGlzLmZpbGVUb0FycmF5QnVmZmVyKGNoaWxkLnBhdGgpO1xuICAgICAgICBjb25zdCBiYXNlNjQgPSB0aGlzLmFycmF5QnVmZmVyVG9CYXNlNjQoYnVmKTtcbiAgICAgICAgLy8gbmFtZSBvbmx5IGtlZXBzIGZpbGUgbmFtZSBmb3IgYmFja2VuZCBjb21wYXRpYmlsaXR5XG4gICAgICAgIG91dC5wdXNoKHsgbmFtZTogY2hpbGQubmFtZSwgZGF0YUJhc2U2NDogYmFzZTY0IH0pO1xuICAgICAgfVxuICAgIH1cblxuICAgIHJldHVybiBvdXQ7XG4gIH1cblxuICBwcml2YXRlIGFycmF5QnVmZmVyVG9CYXNlNjQoYnVmOiBBcnJheUJ1ZmZlcik6IHN0cmluZyB7XG4gICAgbGV0IGJpbmFyeSA9IFwiXCI7XG4gICAgY29uc3QgYnl0ZXMgPSBuZXcgVWludDhBcnJheShidWYpO1xuICAgIGNvbnN0IGNodW5rID0gMHg4MDAwO1xuICAgIGZvciAobGV0IGkgPSAwOyBpIDwgYnl0ZXMubGVuZ3RoOyBpICs9IGNodW5rKSB7XG4gICAgICBjb25zdCBzdWIgPSBieXRlcy5zdWJhcnJheShpLCBNYXRoLm1pbihpICsgY2h1bmssIGJ5dGVzLmxlbmd0aCkpO1xuICAgICAgYmluYXJ5ICs9IFN0cmluZy5mcm9tQ2hhckNvZGUoLi4uc3ViKTtcbiAgICB9XG4gICAgcmV0dXJuIGJ0b2EoYmluYXJ5KTtcbiAgfVxuXG4gIGFzeW5jIGZldGNoQXJ0aWNsZXMoKTogUHJvbWlzZTxSYXdBcnRpY2xlW10+IHtcbiAgICBjb25zdCBiYXNlVXJsID0gdGhpcy5lbnN1cmVCYXNlVXJsKCk7XG4gICAgY29uc3QgbGlzdFVybCA9IGpvaW5VcmwoYmFzZVVybCwgXCIvYXBpL2FsbGFydGljbGVzXCIpO1xuICAgIGNvbnN0IGRhdGEgPSBhd2FpdCByZXFKc29uKGxpc3RVcmwsIHtcbiAgICAgIG1ldGhvZDogXCJHRVRcIixcbiAgICAgIGhlYWRlcnM6IHRoaXMuYXV0aEhlYWRlcnMoZmFsc2UpLFxuICAgIH0pO1xuXG4gICAgbGV0IGxpc3Q6IGFueVtdID0gW107XG4gICAgaWYgKEFycmF5LmlzQXJyYXkoZGF0YSkpIGxpc3QgPSBkYXRhO1xuICAgIGVsc2UgaWYgKEFycmF5LmlzQXJyYXkoZGF0YT8uYXJ0aWNsZXMpKSBsaXN0ID0gZGF0YS5hcnRpY2xlcztcbiAgICBlbHNlIGlmIChBcnJheS5pc0FycmF5KGRhdGE/LmRhdGEpKSBsaXN0ID0gZGF0YS5kYXRhO1xuICAgIGVsc2UgaWYgKEFycmF5LmlzQXJyYXkoZGF0YT8uZGF0YT8uYXJ0aWNsZXMpKSBsaXN0ID0gZGF0YS5kYXRhLmFydGljbGVzO1xuICAgIGVsc2UgaWYgKEFycmF5LmlzQXJyYXkoZGF0YT8uaXRlbXMpKSBsaXN0ID0gZGF0YS5pdGVtcztcbiAgICBlbHNlIGlmIChBcnJheS5pc0FycmF5KGRhdGE/Lmxpc3QpKSBsaXN0ID0gZGF0YS5saXN0O1xuXG4gICAgcmV0dXJuIGxpc3QubWFwKChpdDogYW55LCBpOiBudW1iZXIpID0+ICh7XG4gICAgICBpZDogaXQ/LmlkID8/IGl0Py5faWQgPz8gaXQ/LnNsdWcgPz8gaXQ/LnRpdGxlID8/IGksXG4gICAgICB0aXRsZTogU3RyaW5nKGl0Py50aXRsZSA/PyBpdD8uaWQgPz8gaXQ/Ll9pZCA/PyBgdW50aXRsZWQtJHtpfWApLFxuICAgIH0pKTtcbiAgfVxuXG4gIC8qKlxuICAgKiBcdTUyMzdcdTY1QjBcdTY1ODdcdTdBRTBcdTUyMTdcdTg4NjhcdTMwMDJcbiAgICpcbiAgICogXHU0RTBEXHU1MThEXHU1NDExXHU4QzAzXHU3NTI4XHU2NUI5XHU2MjlCXHU1MUZBXHU1RjAyXHU1RTM4XHVGRjFBXHU1NDdEXHU0RUU0IC8gUmliYm9uIC8gXHU1NDJGXHU1MkE4XHU2RDQxXHU3QTBCXHU5MEZEXHU0RjFBIGludm9rZSBcdTVCODNcdUZGMENcbiAgICogXHU2MjlCXHU1MUZBXHU0RjFBXHU5MDIwXHU2MjEwXHU2NzJBXHU1OTA0XHU3NDA2XHU3Njg0IFByb21pc2UgcmVqZWN0aW9uXHVGRjA4T2JzaWRpYW4gXHU0RjFBXHU1RjUzXHU2MjEwXHU2M0QyXHU0RUY2XHU2MkE1XHU5NTE5XHVGRjA5XHUzMDAyXG4gICAqXG4gICAqIEByZXR1cm5zIFx1NjIxMFx1NTI5Rlx1NjVGNlx1OEZENFx1NTZERVx1NjU4N1x1N0FFMFx1NTIxN1x1ODg2OFx1RkYwQ1x1NTkzMVx1OEQyNVx1NjVGNlx1OEZENFx1NTZERSBudWxsXHUzMDAyXG4gICAqL1xuICBhc3luYyByZWZyZXNoQXJ0aWNsZXMoXG4gICAgb3B0czogeyBzaWxlbnQ/OiBib29sZWFuIH0gPSB7fSxcbiAgKTogUHJvbWlzZTxSYXdBcnRpY2xlW10gfCBudWxsPiB7XG4gICAgdHJ5IHtcbiAgICAgIHRoaXMuYXJ0aWNsZXMgPSBhd2FpdCB0aGlzLmZldGNoQXJ0aWNsZXMoKTtcbiAgICAgIHRoaXMucmVkcmF3Vmlld3MoKTtcbiAgICAgIGlmICghb3B0cy5zaWxlbnQpIHtcbiAgICAgICAgbmV3IE5vdGljZShgbWtCbG9nOiBcdTVERjJcdTUyMzdcdTY1QjBcdUZGMENcdTUxNzEgJHt0aGlzLmFydGljbGVzLmxlbmd0aH0gXHU3QkM3YCk7XG4gICAgICB9XG4gICAgICByZXR1cm4gdGhpcy5hcnRpY2xlcztcbiAgICB9IGNhdGNoIChlKSB7XG4gICAgICBpZiAob3B0cy5zaWxlbnQpIHtcbiAgICAgICAgY29uc29sZS5lcnJvcihcIltta0Jsb2ddIHJlZnJlc2ggZmFpbGVkXCIsIGUpO1xuICAgICAgfSBlbHNlIHtcbiAgICAgICAgbmV3IE5vdGljZShgbWtCbG9nOiBcdTUyMzdcdTY1QjBcdTU5MzFcdThEMjUgLSAke2VyclRleHQoZSl9YCk7XG4gICAgICB9XG4gICAgICByZXR1cm4gbnVsbDtcbiAgICB9XG4gIH1cblxuICBwcml2YXRlIHJlZHJhd1ZpZXdzKCk6IHZvaWQge1xuICAgIGZvciAoY29uc3QgbGVhZiBvZiB0aGlzLmFwcC53b3Jrc3BhY2UuZ2V0TGVhdmVzT2ZUeXBlKFZJRVdfVFlQRV9NS0JMT0cpKSB7XG4gICAgICAvLyBPYnNpZGlhbiAxLjcuMisgXHU3Njg0XHU4OUM2XHU1NkZFXHU1M0VGXHU4MEZEXHU2NjJGIERlZmVycmVkVmlld1x1RkYwQ1x1NUZDNVx1OTg3Qlx1NTA1QSBpbnN0YW5jZW9mIFx1NTIyNFx1NjVBRFxuICAgICAgaWYgKGxlYWYudmlldyBpbnN0YW5jZW9mIE1rQmxvZ0FydGljbGVzVmlldykge1xuICAgICAgICBsZWFmLnZpZXcucmVuZGVyTGlzdCgpO1xuICAgICAgfVxuICAgIH1cbiAgfVxuXG4gIGFzeW5jIHVwbG9hZEN1cnJlbnRGaWxlQXNCbG9nKCk6IFByb21pc2U8dm9pZD4ge1xuICAgIHRyeSB7XG4gICAgICBjb25zdCBmaWxlID0gdGhpcy5hcHAud29ya3NwYWNlLmdldEFjdGl2ZUZpbGUoKTtcbiAgICAgIGlmICghZmlsZSkge1xuICAgICAgICBuZXcgTm90aWNlKFwiXHU4QkY3XHU1MTQ4XHU2MjUzXHU1RjAwXHU0RTAwXHU0RTJBIE1hcmtkb3duIFx1NjU4N1x1NEVGNlwiKTtcbiAgICAgICAgcmV0dXJuO1xuICAgICAgfVxuICAgICAgaWYgKCEoZmlsZSBpbnN0YW5jZW9mIFRGaWxlKSB8fCBmaWxlLmV4dGVuc2lvbi50b0xvd2VyQ2FzZSgpICE9PSBcIm1kXCIpIHtcbiAgICAgICAgbmV3IE5vdGljZShcIlx1NEVDNVx1NjUyRlx1NjMwMVx1NEUwQVx1NEYyMCBNYXJrZG93biBcdTY1ODdcdTRFRjZcdUZGMDgubWRcdUZGMDlcIik7XG4gICAgICAgIHJldHVybjtcbiAgICAgIH1cblxuICAgICAgYXdhaXQgdGhpcy51cGxvYWRTaW5nbGVGaWxlKGZpbGUpO1xuICAgICAgYXdhaXQgdGhpcy5yZWZyZXNoQXJ0aWNsZXMoeyBzaWxlbnQ6IHRydWUgfSk7XG4gICAgfSBjYXRjaCAoZSkge1xuICAgICAgY29uc29sZS5lcnJvcihcIltta0Jsb2ddIHVwbG9hZCBjdXJyZW50IGZpbGUgZmFpbGVkXCIsIGUpO1xuICAgICAgbmV3IE5vdGljZShgXHU0RTBBXHU0RjIwXHU1OTMxXHU4RDI1OiAke2VyclRleHQoZSl9YCk7XG4gICAgfVxuICB9XG5cbiAgcHJpdmF0ZSBhc3luYyB1cGxvYWRTaW5nbGVGaWxlKG1kRmlsZTogVEZpbGUpOiBQcm9taXNlPHZvaWQ+IHtcbiAgICBjb25zdCBiYXNlVXJsID0gdGhpcy5lbnN1cmVCYXNlVXJsKCk7XG4gICAgY29uc3QgdGl0bGUgPSBiYXNlbmFtZVdpdGhvdXRFeHQobWRGaWxlLm5hbWUpO1xuICAgIGNvbnN0IG1kUmF3ID0gYXdhaXQgdGhpcy5hcHAudmF1bHQuY2FjaGVkUmVhZChtZEZpbGUpO1xuICAgIGNvbnN0IG1ldGEgPSBwYXJzZU1ldGEobWRSYXcpO1xuXG4gICAgY29uc3QgcGF5bG9hZDogUmVjb3JkPHN0cmluZywgYW55PiA9IHtcbiAgICAgIHRpdGxlLFxuICAgICAgdXBkYXRlX2F0OiBub3dBc1VwZGF0ZUF0KCksXG4gICAgICBjb250ZW50OiBtZXRhLmNvbnRlbnQsXG4gICAgfTtcbiAgICBwYXlsb2FkLmF1dGhvciA9IG1ldGEuYXV0aG9yID8/IHRoaXMuc2V0dGluZ3MuZGVmYXVsdEF1dGhvciA/PyBcIlwiO1xuICAgIHBheWxvYWQuY2F0ZWdvcnkgPSBtZXRhLmNhdGVnb3J5ID8/IHRoaXMuc2V0dGluZ3MuZGVmYXVsdENhdGVnb3J5ID8/IFwiXCI7XG5cbiAgICBjb25zdCBhcnRpY2xlVXJsID0gYnVpbGRBcnRpY2xlRW5kcG9pbnQoYmFzZVVybCwgdGl0bGUpO1xuICAgIGF3YWl0IHJlcUpzb24oYXJ0aWNsZVVybCwge1xuICAgICAgbWV0aG9kOiBcIlBVVFwiLFxuICAgICAgaGVhZGVyczogdGhpcy5hdXRoSGVhZGVycyh0cnVlKSxcbiAgICAgIGJvZHk6IEpTT04uc3RyaW5naWZ5KHBheWxvYWQpLFxuICAgIH0pO1xuXG4gICAgY29uc3QgaW1hZ2VzID0gYXdhaXQgdGhpcy5jb2xsZWN0SW1hZ2VzRm9yTWFya2Rvd24obWRGaWxlKTtcbiAgICBjb25zdCBpbWFnZVVybCA9IGJ1aWxkSW1hZ2VFbmRwb2ludChiYXNlVXJsKTtcblxuICAgIGZvciAoY29uc3QgaW1nIG9mIGltYWdlcykge1xuICAgICAgY29uc3QgaW1nUGF5bG9hZCA9IHsgdGl0bGUsIG5hbWU6IGltZy5uYW1lLCBkYXRhOiBpbWcuZGF0YUJhc2U2NCB9O1xuICAgICAgYXdhaXQgcmVxSnNvbihpbWFnZVVybCwge1xuICAgICAgICBtZXRob2Q6IFwiUFVUXCIsXG4gICAgICAgIGhlYWRlcnM6IHRoaXMuYXV0aEhlYWRlcnModHJ1ZSksXG4gICAgICAgIGJvZHk6IEpTT04uc3RyaW5naWZ5KGltZ1BheWxvYWQpLFxuICAgICAgfSk7XG4gICAgfVxuXG4gICAgbmV3IE5vdGljZShcbiAgICAgIGBcdTRFMEFcdTRGMjBcdTVCOENcdTYyMTBcdUZGMUEke3RpdGxlfSR7aW1hZ2VzLmxlbmd0aCA/IGBcdUZGMDhcdTU2RkVcdTcyNDcgJHtpbWFnZXMubGVuZ3RofSBcdTVGMjBcdUZGMDlgIDogXCJcIn1gLFxuICAgICk7XG4gIH1cblxuICBhc3luYyBwaWNrQW5kVXBsb2FkRm9sZGVyKCk6IFByb21pc2U8dm9pZD4ge1xuICAgIGNvbnN0IGFsbEZvbGRlcnMgPSB0aGlzLmFwcC52YXVsdFxuICAgICAgLmdldEFsbExvYWRlZEZpbGVzKClcbiAgICAgIC5maWx0ZXIoKGYpID0+IGYgaW5zdGFuY2VvZiBURm9sZGVyKSBhcyBURm9sZGVyW107XG4gICAgY29uc3Qgcm9vdFBhdGggPSB0aGlzLmFwcC52YXVsdC5nZXRSb290KCkucGF0aDtcbiAgICBjb25zdCBjYW5kaWRhdGVzID0gYWxsRm9sZGVycy5maWx0ZXIoKGYpID0+IGYucGF0aCAhPT0gcm9vdFBhdGgpO1xuXG4gICAgaWYgKCFjYW5kaWRhdGVzLmxlbmd0aCkge1xuICAgICAgbmV3IE5vdGljZShcIlx1NjcyQVx1NjI3RVx1NTIzMFx1NTNFRlx1NEUwQVx1NEYyMFx1NzY4NFx1NjU4N1x1NEVGNlx1NTkzOVwiKTtcbiAgICAgIHJldHVybjtcbiAgICB9XG5cbiAgICBuZXcgRm9sZGVyUGlja2VyTW9kYWwodGhpcy5hcHAsIGNhbmRpZGF0ZXMsIGFzeW5jIChmb2xkZXIpID0+IHtcbiAgICAgIHRyeSB7XG4gICAgICAgIGF3YWl0IHRoaXMudXBsb2FkRm9sZGVyQXNCbG9nKGZvbGRlcik7XG4gICAgICAgIGF3YWl0IHRoaXMucmVmcmVzaEFydGljbGVzKHsgc2lsZW50OiB0cnVlIH0pO1xuICAgICAgfSBjYXRjaCAoZSkge1xuICAgICAgICBuZXcgTm90aWNlKGBcdTRFMEFcdTRGMjBcdTY1ODdcdTRFRjZcdTU5MzlcdTU5MzFcdThEMjU6ICR7ZXJyVGV4dChlKX1gKTtcbiAgICAgIH1cbiAgICB9KS5vcGVuKCk7XG4gIH1cblxuICBwcml2YXRlIGFzeW5jIHVwbG9hZEZvbGRlckFzQmxvZyhmb2xkZXI6IFRGb2xkZXIpOiBQcm9taXNlPHZvaWQ+IHtcbiAgICBjb25zdCBtZEZpbGVzID0gdGhpcy5jb2xsZWN0TWFya2Rvd25GaWxlcyhmb2xkZXIpO1xuICAgIGlmICghbWRGaWxlcy5sZW5ndGgpIHtcbiAgICAgIG5ldyBOb3RpY2UoXCJcdTYyNDBcdTkwMDlcdTY1ODdcdTRFRjZcdTU5MzlcdTY3MkFcdTYyN0VcdTUyMzAgLm1kIFx1NjU4N1x1NEVGNlwiKTtcbiAgICAgIHJldHVybjtcbiAgICB9XG5cbiAgICBsZXQgc3VjY2VzcyA9IDA7XG4gICAgZm9yIChjb25zdCBtZCBvZiBtZEZpbGVzKSB7XG4gICAgICB0cnkge1xuICAgICAgICBhd2FpdCB0aGlzLnVwbG9hZFNpbmdsZUZpbGUobWQpO1xuICAgICAgICBzdWNjZXNzKys7XG4gICAgICB9IGNhdGNoIChlKSB7XG4gICAgICAgIGNvbnNvbGUuZXJyb3IoYFtta0Jsb2ddIHVwbG9hZCBmYWlsZWQgZm9yICR7bWQucGF0aH1gLCBlKTtcbiAgICAgICAgbmV3IE5vdGljZShgXHU0RTBBXHU0RjIwXHU1OTMxXHU4RDI1OiAke21kLnBhdGh9IC0gJHtlcnJUZXh0KGUpfWApO1xuICAgICAgfVxuICAgIH1cblxuICAgIG5ldyBOb3RpY2UoYFx1NjU4N1x1NEVGNlx1NTkzOVx1NEUwQVx1NEYyMFx1NUI4Q1x1NjIxMFx1RkYxQVx1NjIxMFx1NTI5RiAke3N1Y2Nlc3N9LyR7bWRGaWxlcy5sZW5ndGh9YCk7XG4gIH1cblxuICBwcml2YXRlIGNvbGxlY3RNYXJrZG93bkZpbGVzKGZvbGRlcjogVEZvbGRlcik6IFRGaWxlW10ge1xuICAgIGNvbnN0IG91dDogVEZpbGVbXSA9IFtdO1xuICAgIGNvbnN0IHN0YWNrOiBURm9sZGVyW10gPSBbZm9sZGVyXTtcbiAgICB3aGlsZSAoc3RhY2subGVuZ3RoID4gMCkge1xuICAgICAgY29uc3QgY3VyID0gc3RhY2sucG9wKCkhO1xuICAgICAgZm9yIChjb25zdCBjIG9mIGN1ci5jaGlsZHJlbikge1xuICAgICAgICBpZiAoYyBpbnN0YW5jZW9mIFRGb2xkZXIpIHN0YWNrLnB1c2goYyk7XG4gICAgICAgIGVsc2UgaWYgKGMgaW5zdGFuY2VvZiBURmlsZSAmJiBjLmV4dGVuc2lvbi50b0xvd2VyQ2FzZSgpID09PSBcIm1kXCIpXG4gICAgICAgICAgb3V0LnB1c2goYyk7XG4gICAgICB9XG4gICAgfVxuICAgIHJldHVybiBvdXQ7XG4gIH1cblxuICBhc3luYyBwaWNrQW5kRGVsZXRlQXJ0aWNsZSgpOiBQcm9taXNlPHZvaWQ+IHtcbiAgICBpZiAoIXRoaXMuYXJ0aWNsZXMubGVuZ3RoKSB7XG4gICAgICBhd2FpdCB0aGlzLnJlZnJlc2hBcnRpY2xlcyh7IHNpbGVudDogdHJ1ZSB9KTtcbiAgICB9XG4gICAgaWYgKCF0aGlzLmFydGljbGVzLmxlbmd0aCkge1xuICAgICAgbmV3IE5vdGljZShcIlx1NjY4Mlx1NjVFMFx1NTNFRlx1NTIyMFx1OTY2NFx1NjU4N1x1N0FFMFwiKTtcbiAgICAgIHJldHVybjtcbiAgICB9XG5cbiAgICBuZXcgQXJ0aWNsZVBpY2tlck1vZGFsKHRoaXMuYXBwLCB0aGlzLmFydGljbGVzLCAoaXQpID0+XG4gICAgICB0aGlzLmNvbmZpcm1EZWxldGUoaXQpLFxuICAgICkub3BlbigpO1xuICB9XG5cbiAgY29uZmlybURlbGV0ZShpdGVtOiBSYXdBcnRpY2xlKTogdm9pZCB7XG4gICAgbmV3IENvbmZpcm1Nb2RhbCh0aGlzLmFwcCwgYFx1Nzg2RVx1NUI5QVx1NTIyMFx1OTY2NFx1NjU4N1x1N0FFMFx1MzAwQyR7aXRlbS50aXRsZX1cdTMwMERcdUZGMUZgLCBhc3luYyAoKSA9PiB7XG4gICAgICB0cnkge1xuICAgICAgICBhd2FpdCB0aGlzLmRlbGV0ZUFydGljbGVCeVRpdGxlKGl0ZW0udGl0bGUpO1xuICAgICAgICBuZXcgTm90aWNlKGBcdTUyMjBcdTk2NjRcdTYyMTBcdTUyOUZcdUZGMUEke2l0ZW0udGl0bGV9YCk7XG4gICAgICAgIGF3YWl0IHRoaXMucmVmcmVzaEFydGljbGVzKHsgc2lsZW50OiB0cnVlIH0pO1xuICAgICAgfSBjYXRjaCAoZSkge1xuICAgICAgICBuZXcgTm90aWNlKGBcdTUyMjBcdTk2NjRcdTU5MzFcdThEMjVcdUZGMUEke2VyclRleHQoZSl9YCk7XG4gICAgICB9XG4gICAgfSkub3BlbigpO1xuICB9XG5cbiAgYXN5bmMgZGVsZXRlQXJ0aWNsZUJ5VGl0bGUodGl0bGU6IHN0cmluZyk6IFByb21pc2U8dm9pZD4ge1xuICAgIGNvbnN0IGJhc2VVcmwgPSB0aGlzLmVuc3VyZUJhc2VVcmwoKTtcbiAgICBjb25zdCB1cmwgPSBidWlsZEFydGljbGVFbmRwb2ludChiYXNlVXJsLCB0aXRsZSk7XG4gICAgY29uc3QgcmVzID0gYXdhaXQgZmV0Y2hXaXRoVGltZW91dCh1cmwsIHtcbiAgICAgIG1ldGhvZDogXCJERUxFVEVcIixcbiAgICAgIGhlYWRlcnM6IHRoaXMuYXV0aEhlYWRlcnMoZmFsc2UpLFxuICAgIH0pO1xuICAgIGlmICghcmVzLm9rKSB7XG4gICAgICBjb25zdCB0ZXh0ID0gYXdhaXQgcmVzLnRleHQoKS5jYXRjaCgoKSA9PiBcIlwiKTtcbiAgICAgIHRocm93IG5ldyBFcnJvcihgSFRUUCAke3Jlcy5zdGF0dXN9ICR7cmVzLnN0YXR1c1RleHR9ICR7dGV4dH1gKTtcbiAgICB9XG4gIH1cblxuICBwcml2YXRlIGFzeW5jIGFjdGl2YXRlVmlldygpOiBQcm9taXNlPHZvaWQ+IHtcbiAgICBjb25zdCB7IHdvcmtzcGFjZSB9ID0gdGhpcy5hcHA7XG4gICAgbGV0IGxlYWY6IFdvcmtzcGFjZUxlYWYgfCBudWxsID1cbiAgICAgIHdvcmtzcGFjZS5nZXRMZWF2ZXNPZlR5cGUoVklFV19UWVBFX01LQkxPRylbMF0gPz8gbnVsbDtcblxuICAgIGlmICghbGVhZikge1xuICAgICAgbGVhZiA9IHdvcmtzcGFjZS5nZXRSaWdodExlYWYoZmFsc2UpO1xuICAgICAgaWYgKCFsZWFmKSB7XG4gICAgICAgIG5ldyBOb3RpY2UoXCJta0Jsb2c6IFx1NjVFMFx1NkNENVx1NTIxQlx1NUVGQVx1ODlDNlx1NTZGRVx1RkYwOFx1NTNGM1x1NEZBN1x1NjgwRlx1NEUwRFx1NTNFRlx1NzUyOFx1RkYwOVwiKTtcbiAgICAgICAgcmV0dXJuO1xuICAgICAgfVxuXG4gICAgICBhd2FpdCBsZWFmLnNldFZpZXdTdGF0ZSh7XG4gICAgICAgIHR5cGU6IFZJRVdfVFlQRV9NS0JMT0csXG4gICAgICAgIGFjdGl2ZTogdHJ1ZSxcbiAgICAgIH0pO1xuICAgIH1cblxuICAgIGF3YWl0IHdvcmtzcGFjZS5yZXZlYWxMZWFmKGxlYWYpO1xuICAgIHRoaXMucmVkcmF3Vmlld3MoKTtcbiAgfVxuXG4gIGFzeW5jIGxvYWRTZXR0aW5ncygpOiBQcm9taXNlPHZvaWQ+IHtcbiAgICB0cnkge1xuICAgICAgY29uc3QgbG9hZGVkID0gKGF3YWl0IHRoaXMubG9hZERhdGEoKSkgYXMgUGFydGlhbDxNa0Jsb2dTZXR0aW5ncz4gfCBudWxsO1xuICAgICAgdGhpcy5zZXR0aW5ncyA9IHsgLi4uREVGQVVMVF9TRVRUSU5HUywgLi4uKGxvYWRlZCA/PyB7fSkgfTtcbiAgICB9IGNhdGNoIChlKSB7XG4gICAgICAvLyBkYXRhLmpzb24gXHU2MzVGXHU1NzRGIC8gXHU0RTBEXHU1M0VGXHU4QkZCXHU2NUY2XHU5MDAwXHU1NkRFXHU5RUQ4XHU4QkE0XHU1MDNDXHVGRjBDXHU3RUREXHU0RTBEXHU4QkE5IG9ubG9hZCBcdTU2RTBcdTZCNjQgcmVqZWN0XG4gICAgICBjb25zb2xlLmVycm9yKFwiW21rQmxvZ10gZmFpbGVkIHRvIGxvYWQgc2V0dGluZ3MsIGZhbGxpbmcgYmFjayB0byBkZWZhdWx0c1wiLCBlKTtcbiAgICAgIHRoaXMuc2V0dGluZ3MgPSB7IC4uLkRFRkFVTFRfU0VUVElOR1MgfTtcbiAgICB9XG4gIH1cblxuICBhc3luYyBzYXZlU2V0dGluZ3MoKTogUHJvbWlzZTx2b2lkPiB7XG4gICAgYXdhaXQgdGhpcy5zYXZlRGF0YSh0aGlzLnNldHRpbmdzKTtcbiAgfVxufVxuXG5jbGFzcyBNa0Jsb2dTZXR0aW5nVGFiIGV4dGVuZHMgUGx1Z2luU2V0dGluZ1RhYiB7XG4gIHBsdWdpbjogTWtCbG9nUGx1Z2luO1xuXG4gIGNvbnN0cnVjdG9yKGFwcDogQXBwLCBwbHVnaW46IE1rQmxvZ1BsdWdpbikge1xuICAgIHN1cGVyKGFwcCwgcGx1Z2luKTtcbiAgICB0aGlzLnBsdWdpbiA9IHBsdWdpbjtcbiAgfVxuXG4gIGRpc3BsYXkoKTogdm9pZCB7XG4gICAgY29uc3QgeyBjb250YWluZXJFbCB9ID0gdGhpcztcbiAgICBjb250YWluZXJFbC5lbXB0eSgpO1xuXG4gICAgY29udGFpbmVyRWwuY3JlYXRlRWwoXCJoMlwiLCB7IHRleHQ6IFwibWtCbG9nIFx1NjNEMlx1NEVGNlx1OEJCRVx1N0Y2RVwiIH0pO1xuXG4gICAgbmV3IFNldHRpbmcoY29udGFpbmVyRWwpXG4gICAgICAuc2V0TmFtZShcIkJhc2UgVVJMXCIpXG4gICAgICAuc2V0RGVzYyhcIlx1NTQwRVx1N0FFRlx1NjcwRFx1NTJBMVx1NTdGQVx1Nzg0MFx1NTczMFx1NTc0MFx1RkYwQ1x1NEY4Qlx1NTk4MiBodHRwOi8vbG9jYWxob3N0OjgwODBcIilcbiAgICAgIC5hZGRUZXh0KCh0ZXh0KSA9PlxuICAgICAgICB0ZXh0XG4gICAgICAgICAgLnNldFBsYWNlaG9sZGVyKFwiaHR0cDovL2xvY2FsaG9zdDo4MDgwXCIpXG4gICAgICAgICAgLnNldFZhbHVlKHRoaXMucGx1Z2luLnNldHRpbmdzLmJhc2VVcmwpXG4gICAgICAgICAgLm9uQ2hhbmdlKGFzeW5jICh2YWx1ZSkgPT4ge1xuICAgICAgICAgICAgdGhpcy5wbHVnaW4uc2V0dGluZ3MuYmFzZVVybCA9IHZhbHVlLnRyaW0oKTtcbiAgICAgICAgICAgIGF3YWl0IHRoaXMucGx1Z2luLnNhdmVTZXR0aW5ncygpO1xuICAgICAgICAgIH0pLFxuICAgICAgKTtcblxuICAgIG5ldyBTZXR0aW5nKGNvbnRhaW5lckVsKVxuICAgICAgLnNldE5hbWUoXCJEZWZhdWx0IEF1dGhvclwiKVxuICAgICAgLnNldERlc2MoXCJNYXJrZG93biBcdTY3MkFcdTU4RjBcdTY2MEUgYXV0aG9yIFx1NjVGNlx1NEY3Rlx1NzUyOFwiKVxuICAgICAgLmFkZFRleHQoKHRleHQpID0+XG4gICAgICAgIHRleHRcbiAgICAgICAgICAuc2V0VmFsdWUodGhpcy5wbHVnaW4uc2V0dGluZ3MuZGVmYXVsdEF1dGhvcilcbiAgICAgICAgICAub25DaGFuZ2UoYXN5bmMgKHZhbHVlKSA9PiB7XG4gICAgICAgICAgICB0aGlzLnBsdWdpbi5zZXR0aW5ncy5kZWZhdWx0QXV0aG9yID0gdmFsdWU7XG4gICAgICAgICAgICBhd2FpdCB0aGlzLnBsdWdpbi5zYXZlU2V0dGluZ3MoKTtcbiAgICAgICAgICB9KSxcbiAgICAgICk7XG5cbiAgICBuZXcgU2V0dGluZyhjb250YWluZXJFbClcbiAgICAgIC5zZXROYW1lKFwiRGVmYXVsdCBDYXRlZ29yeVwiKVxuICAgICAgLnNldERlc2MoXCJNYXJrZG93biBcdTY3MkFcdTU4RjBcdTY2MEUgY2F0ZWdvcnkgXHU2NUY2XHU0RjdGXHU3NTI4XCIpXG4gICAgICAuYWRkVGV4dCgodGV4dCkgPT5cbiAgICAgICAgdGV4dFxuICAgICAgICAgIC5zZXRWYWx1ZSh0aGlzLnBsdWdpbi5zZXR0aW5ncy5kZWZhdWx0Q2F0ZWdvcnkpXG4gICAgICAgICAgLm9uQ2hhbmdlKGFzeW5jICh2YWx1ZSkgPT4ge1xuICAgICAgICAgICAgdGhpcy5wbHVnaW4uc2V0dGluZ3MuZGVmYXVsdENhdGVnb3J5ID0gdmFsdWU7XG4gICAgICAgICAgICBhd2FpdCB0aGlzLnBsdWdpbi5zYXZlU2V0dGluZ3MoKTtcbiAgICAgICAgICB9KSxcbiAgICAgICk7XG5cbiAgICBuZXcgU2V0dGluZyhjb250YWluZXJFbClcbiAgICAgIC5zZXROYW1lKFwiQXV0aCBUb2tlblwiKVxuICAgICAgLnNldERlc2MoXCJcdTUzRUZcdTkwMDkgQmVhcmVyIFRva2VuXHVGRjBDXHU1QzA2XHU5MDFBXHU4RkM3IEF1dGhvcml6YXRpb24gXHU1OTM0XHU1M0QxXHU5MDAxXCIpXG4gICAgICAuYWRkVGV4dCgodGV4dCkgPT5cbiAgICAgICAgdGV4dFxuICAgICAgICAgIC5zZXRQbGFjZWhvbGRlcihcImV5SmhiR2NpT2kuLi5cIilcbiAgICAgICAgICAuc2V0VmFsdWUodGhpcy5wbHVnaW4uc2V0dGluZ3MuYXV0aFRva2VuKVxuICAgICAgICAgIC5vbkNoYW5nZShhc3luYyAodmFsdWUpID0+IHtcbiAgICAgICAgICAgIHRoaXMucGx1Z2luLnNldHRpbmdzLmF1dGhUb2tlbiA9IHZhbHVlLnRyaW0oKTtcbiAgICAgICAgICAgIGF3YWl0IHRoaXMucGx1Z2luLnNhdmVTZXR0aW5ncygpO1xuICAgICAgICAgIH0pLFxuICAgICAgKTtcblxuICAgIGNvbnRhaW5lckVsLmNyZWF0ZUVsKFwiaDNcIiwgeyB0ZXh0OiBcIlx1NTQyRlx1NTJBOFx1ODg0Q1x1NEUzQVwiIH0pO1xuXG4gICAgbmV3IFNldHRpbmcoY29udGFpbmVyRWwpXG4gICAgICAuc2V0TmFtZShcIlx1NTQyRlx1NTJBOFx1NjVGNlx1ODFFQVx1NTJBOFx1NjI1M1x1NUYwMFx1N0JBMVx1NzQwNlx1ODlDNlx1NTZGRVwiKVxuICAgICAgLnNldERlc2MoXG4gICAgICAgIFwiXHU1MTczXHU5NUVEXHU1NDBFXHU0RUM1XHU1NzI4XHU3MEI5XHU1MUZCXHU1REU2XHU0RkE3IFJpYmJvbiBcdTU2RkVcdTY4MDdcdTYyMTZcdTYyNjdcdTg4NENcdTU0N0RcdTRFRTRcdTY1RjZcdTYyNTNcdTVGMDBcdUZGMDhcdTYzQThcdTgzNTBcdTUxNzNcdTk1RURcdUZGMENcdTkwN0ZcdTUxNERcdTZCQ0ZcdTZCMjFcdTU0MkZcdTUyQThcdTkwRkRcdTY1MzlcdTUxOTlcdTRGNjBcdTc2ODRcdTVERTVcdTRGNUNcdTUzM0FcdTVFMDNcdTVDNDBcdUZGMDlcIixcbiAgICAgIClcbiAgICAgIC5hZGRUb2dnbGUoKHRvZ2dsZSkgPT5cbiAgICAgICAgdG9nZ2xlXG4gICAgICAgICAgLnNldFZhbHVlKHRoaXMucGx1Z2luLnNldHRpbmdzLm9wZW5WaWV3T25TdGFydHVwKVxuICAgICAgICAgIC5vbkNoYW5nZShhc3luYyAodmFsdWUpID0+IHtcbiAgICAgICAgICAgIHRoaXMucGx1Z2luLnNldHRpbmdzLm9wZW5WaWV3T25TdGFydHVwID0gdmFsdWU7XG4gICAgICAgICAgICBhd2FpdCB0aGlzLnBsdWdpbi5zYXZlU2V0dGluZ3MoKTtcbiAgICAgICAgICB9KSxcbiAgICAgICk7XG5cbiAgICBuZXcgU2V0dGluZyhjb250YWluZXJFbClcbiAgICAgIC5zZXROYW1lKFwiXHU1NDJGXHU1MkE4XHU2NUY2XHU4MUVBXHU1MkE4XHU1MjM3XHU2NUIwXHU2NTg3XHU3QUUwXHU1MjE3XHU4ODY4XCIpXG4gICAgICAuc2V0RGVzYyhcIlx1NTcyOCBPYnNpZGlhbiBcdTVFMDNcdTVDNDBcdTVDMzFcdTdFRUFcdTU0MEVcdTUxOERcdTUzRDFcdThENzdcdThCRjdcdTZDNDJcdUZGMENcdTRFMERcdTRGMUFcdTk2M0JcdTU4NUVcdTU0MkZcdTUyQThcdUZGMUJcdThCRjdcdTZDNDJcdTVFMjYgMTVzIFx1OEQ4NVx1NjVGNlwiKVxuICAgICAgLmFkZFRvZ2dsZSgodG9nZ2xlKSA9PlxuICAgICAgICB0b2dnbGVcbiAgICAgICAgICAuc2V0VmFsdWUodGhpcy5wbHVnaW4uc2V0dGluZ3MucmVmcmVzaE9uU3RhcnR1cClcbiAgICAgICAgICAub25DaGFuZ2UoYXN5bmMgKHZhbHVlKSA9PiB7XG4gICAgICAgICAgICB0aGlzLnBsdWdpbi5zZXR0aW5ncy5yZWZyZXNoT25TdGFydHVwID0gdmFsdWU7XG4gICAgICAgICAgICBhd2FpdCB0aGlzLnBsdWdpbi5zYXZlU2V0dGluZ3Mo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

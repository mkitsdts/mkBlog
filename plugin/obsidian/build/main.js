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
  authToken: ""
};
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
async function reqJson(url, init) {
  var _a;
  const res = await fetch(url, init);
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
    await this.renderList();
  }
  async renderList() {
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
    this.settings = DEFAULT_SETTINGS;
    this.articles = [];
  }
  async onload() {
    await this.loadSettings();
    this.registerView(
      VIEW_TYPE_MKBLOG,
      (leaf) => new MkBlogArticlesView(leaf, this)
    );
    this.addSettingTab(new MkBlogSettingTab(this.app, this));
    this.addRibbonIcon(
      "cloud-upload",
      "mkBlog: \u4E0A\u4F20\u5F53\u524D\u6587\u4EF6\u4E3A\u535A\u5BA2",
      async () => {
        await this.uploadCurrentFileAsBlog();
      }
    );
    this.addCommand({
      id: "mkblog-open-view",
      name: "mkBlog: \u6253\u5F00\u7BA1\u7406\u89C6\u56FE",
      callback: async () => this.activateView()
    });
    this.addCommand({
      id: "mkblog-upload-current-file",
      name: "mkBlog: \u4E0A\u4F20\u5F53\u524D\u6587\u4EF6\u4E3A\u535A\u5BA2",
      callback: async () => this.uploadCurrentFileAsBlog()
    });
    this.addCommand({
      id: "mkblog-upload-folder",
      name: "mkBlog: \u4E0A\u4F20\u9009\u62E9\u6587\u4EF6\u5939\u4E3A\u535A\u5BA2",
      callback: async () => this.pickAndUploadFolder()
    });
    this.addCommand({
      id: "mkblog-refresh-articles",
      name: "mkBlog: \u5237\u65B0\u6587\u7AE0\u5217\u8868",
      callback: async () => this.refreshArticles()
    });
    this.addCommand({
      id: "mkblog-delete-article",
      name: "mkBlog: \u5220\u9664\u6587\u7AE0",
      callback: async () => this.pickAndDeleteArticle()
    });
    await this.activateView();
    await this.refreshArticles().catch((e) => {
      var _a;
      console.error("[mkBlog] initial refresh failed", e);
      new import_obsidian.Notice(`mkBlog \u521D\u59CB\u5316\u62C9\u53D6\u6587\u7AE0\u5931\u8D25: ${(_a = e == null ? void 0 : e.message) != null ? _a : e}`);
    });
  }
  async onunload() {
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
    var _a, _b;
    const abs = (_b = (_a = this.app.vault.adapter).getFullPath) == null ? void 0 : _b.call(_a, vaultPath);
    if (abs && "requestUrl" in window === false) {
    }
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
  async refreshArticles() {
    var _a;
    try {
      this.articles = await this.fetchArticles();
      new import_obsidian.Notice(`mkBlog: \u5DF2\u5237\u65B0\uFF0C\u5171 ${this.articles.length} \u7BC7`);
    } catch (e) {
      new import_obsidian.Notice(`mkBlog: \u5237\u65B0\u5931\u8D25 - ${(_a = e == null ? void 0 : e.message) != null ? _a : e}`);
      throw e;
    } finally {
      this.redrawView();
    }
  }
  redrawView() {
    const leaves = this.app.workspace.getLeavesOfType(VIEW_TYPE_MKBLOG);
    for (const leaf of leaves) {
      const v = leaf.view;
      if (v instanceof MkBlogArticlesView) {
        v.renderList();
      }
    }
  }
  async uploadCurrentFileAsBlog() {
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
    await this.refreshArticles().catch(() => {
    });
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
      var _a;
      try {
        await this.uploadFolderAsBlog(folder);
        await this.refreshArticles().catch(() => {
        });
      } catch (e) {
        new import_obsidian.Notice(`\u4E0A\u4F20\u6587\u4EF6\u5939\u5931\u8D25: ${(_a = e == null ? void 0 : e.message) != null ? _a : e}`);
      }
    }).open();
  }
  async uploadFolderAsBlog(folder) {
    var _a;
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
        new import_obsidian.Notice(`\u4E0A\u4F20\u5931\u8D25: ${md.path} - ${(_a = e == null ? void 0 : e.message) != null ? _a : e}`);
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
      await this.refreshArticles().catch(() => {
      });
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
      var _a;
      try {
        await this.deleteArticleByTitle(item.title);
        new import_obsidian.Notice(`\u5220\u9664\u6210\u529F\uFF1A${item.title}`);
        await this.refreshArticles().catch(() => {
        });
      } catch (e) {
        new import_obsidian.Notice(`\u5220\u9664\u5931\u8D25\uFF1A${(_a = e == null ? void 0 : e.message) != null ? _a : e}`);
      }
    }).open();
  }
  async deleteArticleByTitle(title) {
    const baseUrl = this.ensureBaseUrl();
    const url = buildArticleEndpoint(baseUrl, title);
    const res = await fetch(url, {
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
      if (!leaf) return;
      await leaf.setViewState({
        type: VIEW_TYPE_MKBLOG,
        active: true
      });
    }
    workspace.revealLeaf(leaf);
  }
  async loadSettings() {
    const loaded = await this.loadData();
    this.settings = Object.assign({}, DEFAULT_SETTINGS, loaded != null ? loaded : {});
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
  }
};
//# sourceMappingURL=data:application/json;base64,ewogICJ2ZXJzaW9uIjogMywKICAic291cmNlcyI6IFsiLi4vc3JjL21haW4udHMiXSwKICAic291cmNlc0NvbnRlbnQiOiBbImltcG9ydCB7XG4gIEFwcCxcbiAgTm90aWNlLFxuICBQbHVnaW4sXG4gIFBsdWdpblNldHRpbmdUYWIsXG4gIFNldHRpbmcsXG4gIEl0ZW1WaWV3LFxuICBXb3Jrc3BhY2VMZWFmLFxuICBURmlsZSxcbiAgTW9kYWwsXG4gIEZ1enp5U3VnZ2VzdE1vZGFsLFxuICBURm9sZGVyLFxuICBub3JtYWxpemVQYXRoLFxufSBmcm9tIFwib2JzaWRpYW5cIjtcblxuY29uc3QgVklFV19UWVBFX01LQkxPRyA9IFwibWtibG9nLWFydGljbGVzLXZpZXdcIjtcblxuaW50ZXJmYWNlIE1rQmxvZ1NldHRpbmdzIHtcbiAgYmFzZVVybDogc3RyaW5nO1xuICBkZWZhdWx0QXV0aG9yOiBzdHJpbmc7XG4gIGRlZmF1bHRDYXRlZ29yeTogc3RyaW5nO1xuICBhdXRoVG9rZW46IHN0cmluZztcbn1cblxuaW50ZXJmYWNlIFJhd0FydGljbGUge1xuICBpZDogc3RyaW5nIHwgbnVtYmVyO1xuICB0aXRsZTogc3RyaW5nO1xufVxuXG5jb25zdCBERUZBVUxUX1NFVFRJTkdTOiBNa0Jsb2dTZXR0aW5ncyA9IHtcbiAgYmFzZVVybDogXCJodHRwOi8vbG9jYWxob3N0OjgwODBcIixcbiAgZGVmYXVsdEF1dGhvcjogXCJcIixcbiAgZGVmYXVsdENhdGVnb3J5OiBcIkdlbmVyYWxcIixcbiAgYXV0aFRva2VuOiBcIlwiLFxufTtcblxuY29uc3QgSU1HX0VYVCA9IG5ldyBTZXQoW1wiLnBuZ1wiLCBcIi5qcGdcIiwgXCIuanBlZ1wiLCBcIi5naWZcIiwgXCIud2VicFwiLCBcIi5zdmdcIl0pO1xuXG5mdW5jdGlvbiBqb2luVXJsKGJhc2VVcmw6IHN0cmluZywgcGF0aDogc3RyaW5nKTogc3RyaW5nIHtcbiAgY29uc3QgYmFzZSA9IChiYXNlVXJsIHx8IFwiXCIpLnJlcGxhY2UoL1xcLyskLywgXCJcIik7XG4gIGNvbnN0IHN1ZmZpeCA9IHBhdGguc3RhcnRzV2l0aChcIi9cIikgPyBwYXRoIDogYC8ke3BhdGh9YDtcbiAgcmV0dXJuIGAke2Jhc2V9JHtzdWZmaXh9YDtcbn1cblxuZnVuY3Rpb24gYnVpbGRBcnRpY2xlRW5kcG9pbnQoYmFzZVVybDogc3RyaW5nLCB0aXRsZTogc3RyaW5nKTogc3RyaW5nIHtcbiAgcmV0dXJuIGpvaW5VcmwoYmFzZVVybCwgYC9hcGkvYXJ0aWNsZS8ke2VuY29kZVVSSUNvbXBvbmVudCh0aXRsZSl9YCk7XG59XG5cbmZ1bmN0aW9uIGJ1aWxkSW1hZ2VFbmRwb2ludChiYXNlVXJsOiBzdHJpbmcpOiBzdHJpbmcge1xuICByZXR1cm4gam9pblVybChiYXNlVXJsLCBcIi9hcGkvaW1hZ2VcIik7XG59XG5cbmZ1bmN0aW9uIG5vd0FzVXBkYXRlQXQoKTogc3RyaW5nIHtcbiAgY29uc3QgcGFkID0gKG46IG51bWJlcikgPT4gKG4gPCAxMCA/IGAwJHtufWAgOiBTdHJpbmcobikpO1xuICBjb25zdCBkID0gbmV3IERhdGUoKTtcbiAgcmV0dXJuIGAke2QuZ2V0RnVsbFllYXIoKX0tJHtwYWQoZC5nZXRNb250aCgpICsgMSl9LSR7cGFkKGQuZ2V0RGF0ZSgpKX0gJHtwYWQoZC5nZXRIb3VycygpKX06JHtwYWQoXG4gICAgZC5nZXRNaW51dGVzKCksXG4gICl9OiR7cGFkKGQuZ2V0U2Vjb25kcygpKX1gO1xufVxuXG5mdW5jdGlvbiBleHRuYW1lKG5hbWU6IHN0cmluZyk6IHN0cmluZyB7XG4gIGNvbnN0IGlkeCA9IG5hbWUubGFzdEluZGV4T2YoXCIuXCIpO1xuICBpZiAoaWR4IDwgMCkgcmV0dXJuIFwiXCI7XG4gIHJldHVybiBuYW1lLnNsaWNlKGlkeCkudG9Mb3dlckNhc2UoKTtcbn1cblxuZnVuY3Rpb24gYmFzZW5hbWVXaXRob3V0RXh0KHBhdGg6IHN0cmluZyk6IHN0cmluZyB7XG4gIGNvbnN0IHAgPSBwYXRoLnJlcGxhY2UoL1xcXFwvZywgXCIvXCIpO1xuICBjb25zdCBuYW1lID0gcC5zcGxpdChcIi9cIikucG9wKCkgPz8gcDtcbiAgY29uc3QgaWR4ID0gbmFtZS5sYXN0SW5kZXhPZihcIi5cIik7XG4gIHJldHVybiBpZHggPj0gMCA/IG5hbWUuc2xpY2UoMCwgaWR4KSA6IG5hbWU7XG59XG5cbmZ1bmN0aW9uIGRpcm5hbWUocGF0aDogc3RyaW5nKTogc3RyaW5nIHtcbiAgY29uc3QgcCA9IHBhdGgucmVwbGFjZSgvXFxcXC9nLCBcIi9cIik7XG4gIGNvbnN0IGlkeCA9IHAubGFzdEluZGV4T2YoXCIvXCIpO1xuICBpZiAoaWR4IDwgMCkgcmV0dXJuIFwiXCI7XG4gIHJldHVybiBwLnNsaWNlKDAsIGlkeCk7XG59XG5cbmZ1bmN0aW9uIHJlbW92ZUZyb250bWF0dGVyKHJhdzogc3RyaW5nKTogc3RyaW5nIHtcbiAgaWYgKCFyYXcuc3RhcnRzV2l0aChcIi0tLVwiKSkgcmV0dXJuIHJhdztcbiAgY29uc3QgZW5kSWR4ID0gcmF3LmluZGV4T2YoXCJcXG4tLS1cIiwgMyk7XG4gIGlmIChlbmRJZHggPT09IC0xKSByZXR1cm4gcmF3O1xuICBjb25zdCBhZnRlciA9IHJhdy5zbGljZShlbmRJZHggKyBcIlxcbi0tLVwiLmxlbmd0aCk7XG4gIHJldHVybiBhZnRlci5yZXBsYWNlKC9eXFxyP1xcbi8sIFwiXCIpO1xufVxuXG5mdW5jdGlvbiBwYXJzZU1ldGEocmF3TWQ6IHN0cmluZyk6IHtcbiAgYXV0aG9yPzogc3RyaW5nO1xuICBjYXRlZ29yeT86IHN0cmluZztcbiAgY29udGVudDogc3RyaW5nO1xufSB7XG4gIGxldCBhdXRob3I6IHN0cmluZyB8IHVuZGVmaW5lZDtcbiAgbGV0IGNhdGVnb3J5OiBzdHJpbmcgfCB1bmRlZmluZWQ7XG4gIGxldCBjb250ZW50ID0gcmF3TWQ7XG5cbiAgLy8gMSkgWUFNTCBmcm9udG1hdHRlciBhdCB0b3BcbiAgaWYgKHJhd01kLnN0YXJ0c1dpdGgoXCItLS1cIikpIHtcbiAgICBjb25zdCBlbmQgPSByYXdNZC5pbmRleE9mKFwiXFxuLS0tXCIsIDMpO1xuICAgIGlmIChlbmQgIT09IC0xKSB7XG4gICAgICBjb25zdCBmbSA9IHJhd01kLnNsaWNlKDMsIGVuZCkuc3BsaXQoL1xccj9cXG4vKTtcbiAgICAgIGZvciAoY29uc3QgbGluZSBvZiBmbSkge1xuICAgICAgICBjb25zdCBtID0gbGluZS5tYXRjaCgvXlxccyooYXV0aG9yfGNhdGVnb3J5KVxccyo6XFxzKiguKylcXHMqJC9pKTtcbiAgICAgICAgaWYgKG0pIHtcbiAgICAgICAgICBjb25zdCBrZXkgPSBtWzFdLnRvTG93ZXJDYXNlKCk7XG4gICAgICAgICAgY29uc3QgdmFsID0gbVsyXS50cmltKCkucmVwbGFjZSgvXlsnXCJdfFsnXCJdJC9nLCBcIlwiKTtcbiAgICAgICAgICBpZiAoa2V5ID09PSBcImF1dGhvclwiICYmIHZhbCkgYXV0aG9yID0gdmFsO1xuICAgICAgICAgIGlmIChrZXkgPT09IFwiY2F0ZWdvcnlcIiAmJiB2YWwpIGNhdGVnb3J5ID0gdmFsO1xuICAgICAgICB9XG4gICAgICB9XG4gICAgICBjb250ZW50ID0gcmVtb3ZlRnJvbnRtYXR0ZXIocmF3TWQpO1xuICAgICAgcmV0dXJuIHsgYXV0aG9yLCBjYXRlZ29yeSwgY29udGVudCB9O1xuICAgIH1cbiAgfVxuXG4gIC8vIDIpIGZhbGxiYWNrIHRvcCBsaW5lczogYXV0aG9yOiAvIGNhdGVnb3J5OlxuICBjb25zdCBsaW5lcyA9IHJhd01kLnNwbGl0KC9cXHI/XFxuLyk7XG4gIGxldCBpID0gMDtcbiAgd2hpbGUgKGkgPCBsaW5lcy5sZW5ndGgpIHtcbiAgICBjb25zdCBsaW5lID0gbGluZXNbaV07XG4gICAgaWYgKCFsaW5lLnRyaW0oKSkge1xuICAgICAgaSsrO1xuICAgICAgY29udGludWU7XG4gICAgfVxuICAgIGNvbnN0IG0gPSBsaW5lLm1hdGNoKC9eXFxzKihhdXRob3J8Y2F0ZWdvcnkpXFxzKjpcXHMqKC4rKVxccyokL2kpO1xuICAgIGlmICghbSkgYnJlYWs7XG4gICAgY29uc3Qga2V5ID0gbVsxXS50b0xvd2VyQ2FzZSgpO1xuICAgIGNvbnN0IHZhbCA9IG1bMl0udHJpbSgpLnJlcGxhY2UoL15bJ1wiXXxbJ1wiXSQvZywgXCJcIik7XG4gICAgaWYgKGtleSA9PT0gXCJhdXRob3JcIiAmJiB2YWwgJiYgIWF1dGhvcikgYXV0aG9yID0gdmFsO1xuICAgIGlmIChrZXkgPT09IFwiY2F0ZWdvcnlcIiAmJiB2YWwgJiYgIWNhdGVnb3J5KSBjYXRlZ29yeSA9IHZhbDtcbiAgICBpKys7XG4gIH1cbiAgaWYgKGkgPiAwKSBjb250ZW50ID0gbGluZXMuc2xpY2UoaSkuam9pbihcIlxcblwiKTtcblxuICByZXR1cm4geyBhdXRob3IsIGNhdGVnb3J5LCBjb250ZW50IH07XG59XG5cbmFzeW5jIGZ1bmN0aW9uIHJlcUpzb24odXJsOiBzdHJpbmcsIGluaXQ/OiBSZXF1ZXN0SW5pdCk6IFByb21pc2U8YW55PiB7XG4gIGNvbnN0IHJlcyA9IGF3YWl0IGZldGNoKHVybCwgaW5pdCk7XG4gIGlmICghcmVzLm9rKSB7XG4gICAgY29uc3QgdGV4dCA9IGF3YWl0IHJlcy50ZXh0KCkuY2F0Y2goKCkgPT4gXCJcIik7XG4gICAgdGhyb3cgbmV3IEVycm9yKGBIVFRQICR7cmVzLnN0YXR1c30gJHtyZXMuc3RhdHVzVGV4dH0gJHt0ZXh0fWApO1xuICB9XG4gIGNvbnN0IGN0ID0gcmVzLmhlYWRlcnMuZ2V0KFwiY29udGVudC10eXBlXCIpID8/IFwiXCI7XG4gIGlmIChjdC5pbmNsdWRlcyhcImFwcGxpY2F0aW9uL2pzb25cIikpIHJldHVybiByZXMuanNvbigpO1xuICBjb25zdCB0eHQgPSBhd2FpdCByZXMudGV4dCgpO1xuICB0cnkge1xuICAgIHJldHVybiBKU09OLnBhcnNlKHR4dCk7XG4gIH0gY2F0Y2gge1xuICAgIHJldHVybiB0eHQ7XG4gIH1cbn1cblxuY2xhc3MgQXJ0aWNsZVBpY2tlck1vZGFsIGV4dGVuZHMgRnV6enlTdWdnZXN0TW9kYWw8UmF3QXJ0aWNsZT4ge1xuICBwcml2YXRlIHJlYWRvbmx5IGl0ZW1zOiBSYXdBcnRpY2xlW107XG4gIHB1YmxpYyBvbkNob29zZTogKGl0OiBSYXdBcnRpY2xlKSA9PiB2b2lkO1xuXG4gIGNvbnN0cnVjdG9yKFxuICAgIGFwcDogQXBwLFxuICAgIGl0ZW1zOiBSYXdBcnRpY2xlW10sXG4gICAgb25DaG9vc2U6IChpdDogUmF3QXJ0aWNsZSkgPT4gdm9pZCxcbiAgKSB7XG4gICAgc3VwZXIoYXBwKTtcbiAgICB0aGlzLml0ZW1zID0gaXRlbXM7XG4gICAgdGhpcy5vbkNob29zZSA9IG9uQ2hvb3NlO1xuICAgIHRoaXMuc2V0UGxhY2Vob2xkZXIoXCJcdTkwMDlcdTYyRTlcdTg5ODFcdTUyMjBcdTk2NjRcdTc2ODRcdTY1ODdcdTdBRTAuLi5cIik7XG4gIH1cblxuICBnZXRJdGVtcygpOiBSYXdBcnRpY2xlW10ge1xuICAgIHJldHVybiB0aGlzLml0ZW1zO1xuICB9XG5cbiAgZ2V0SXRlbVRleHQoaXRlbTogUmF3QXJ0aWNsZSk6IHN0cmluZyB7XG4gICAgcmV0dXJuIGl0ZW0udGl0bGU7XG4gIH1cblxuICBvbkNob29zZUl0ZW0oaXRlbTogUmF3QXJ0aWNsZSk6IHZvaWQge1xuICAgIHRoaXMub25DaG9vc2UoaXRlbSk7XG4gIH1cbn1cblxuY2xhc3MgRm9sZGVyUGlja2VyTW9kYWwgZXh0ZW5kcyBGdXp6eVN1Z2dlc3RNb2RhbDxURm9sZGVyPiB7XG4gIHByaXZhdGUgcmVhZG9ubHkgZm9sZGVyczogVEZvbGRlcltdO1xuICBwcml2YXRlIHJlYWRvbmx5IG9uQ2hvb3NlQ2I6IChmb2xkZXI6IFRGb2xkZXIpID0+IHZvaWQ7XG5cbiAgY29uc3RydWN0b3IoXG4gICAgYXBwOiBBcHAsXG4gICAgZm9sZGVyczogVEZvbGRlcltdLFxuICAgIG9uQ2hvb3NlOiAoZm9sZGVyOiBURm9sZGVyKSA9PiB2b2lkLFxuICApIHtcbiAgICBzdXBlcihhcHApO1xuICAgIHRoaXMuZm9sZGVycyA9IGZvbGRlcnM7XG4gICAgdGhpcy5vbkNob29zZUNiID0gb25DaG9vc2U7XG4gICAgdGhpcy5zZXRQbGFjZWhvbGRlcihcIlx1OTAwOVx1NjJFOVx1ODk4MVx1NEUwQVx1NEYyMFx1NzY4NFx1NjU4N1x1NEVGNlx1NTkzOS4uLlwiKTtcbiAgfVxuXG4gIGdldEl0ZW1zKCk6IFRGb2xkZXJbXSB7XG4gICAgcmV0dXJuIHRoaXMuZm9sZGVycztcbiAgfVxuXG4gIGdldEl0ZW1UZXh0KGl0ZW06IFRGb2xkZXIpOiBzdHJpbmcge1xuICAgIHJldHVybiBpdGVtLnBhdGggfHwgXCIvXCI7XG4gIH1cblxuICBvbkNob29zZUl0ZW0oaXRlbTogVEZvbGRlcik6IHZvaWQge1xuICAgIHRoaXMub25DaG9vc2VDYihpdGVtKTtcbiAgfVxufVxuXG5jbGFzcyBDb25maXJtTW9kYWwgZXh0ZW5kcyBNb2RhbCB7XG4gIHByaXZhdGUgcmVhZG9ubHkgbWVzc2FnZTogc3RyaW5nO1xuICBwcml2YXRlIHJlYWRvbmx5IG9uQ29uZmlybTogKCkgPT4gdm9pZDtcblxuICBjb25zdHJ1Y3RvcihhcHA6IEFwcCwgbWVzc2FnZTogc3RyaW5nLCBvbkNvbmZpcm06ICgpID0+IHZvaWQpIHtcbiAgICBzdXBlcihhcHApO1xuICAgIHRoaXMubWVzc2FnZSA9IG1lc3NhZ2U7XG4gICAgdGhpcy5vbkNvbmZpcm0gPSBvbkNvbmZpcm07XG4gIH1cblxuICBvbk9wZW4oKTogdm9pZCB7XG4gICAgY29uc3QgeyBjb250ZW50RWwgfSA9IHRoaXM7XG4gICAgY29udGVudEVsLmVtcHR5KCk7XG4gICAgY29udGVudEVsLmNyZWF0ZUVsKFwiaDNcIiwgeyB0ZXh0OiBcIlx1Nzg2RVx1OEJBNFx1NjRDRFx1NEY1Q1wiIH0pO1xuICAgIGNvbnRlbnRFbC5jcmVhdGVFbChcInBcIiwgeyB0ZXh0OiB0aGlzLm1lc3NhZ2UgfSk7XG5cbiAgICBjb25zdCBhY3Rpb25zID0gY29udGVudEVsLmNyZWF0ZURpdih7IGNsczogXCJta2Jsb2ctbW9kYWwtYWN0aW9uc1wiIH0pO1xuICAgIGNvbnN0IGNhbmNlbEJ0biA9IGFjdGlvbnMuY3JlYXRlRWwoXCJidXR0b25cIiwgeyB0ZXh0OiBcIlx1NTNENlx1NkQ4OFwiIH0pO1xuICAgIGNvbnN0IG9rQnRuID0gYWN0aW9ucy5jcmVhdGVFbChcImJ1dHRvblwiLCB7IHRleHQ6IFwiXHU1MjIwXHU5NjY0XCIgfSk7XG4gICAgb2tCdG4uYWRkQ2xhc3MoXCJtb2Qtd2FybmluZ1wiKTtcblxuICAgIGNhbmNlbEJ0bi5vbmNsaWNrID0gKCkgPT4gdGhpcy5jbG9zZSgpO1xuICAgIG9rQnRuLm9uY2xpY2sgPSAoKSA9PiB7XG4gICAgICB0aGlzLmNsb3NlKCk7XG4gICAgICB0aGlzLm9uQ29uZmlybSgpO1xuICAgIH07XG4gIH1cblxuICBvbkNsb3NlKCk6IHZvaWQge1xuICAgIHRoaXMuY29udGVudEVsLmVtcHR5KCk7XG4gIH1cbn1cblxuY2xhc3MgTWtCbG9nQXJ0aWNsZXNWaWV3IGV4dGVuZHMgSXRlbVZpZXcge1xuICBwcml2YXRlIHBsdWdpbjogTWtCbG9nUGx1Z2luO1xuICBwcml2YXRlIGxpc3RFbDogSFRNTEVsZW1lbnQgfCBudWxsID0gbnVsbDtcblxuICBjb25zdHJ1Y3RvcihsZWFmOiBXb3Jrc3BhY2VMZWFmLCBwbHVnaW46IE1rQmxvZ1BsdWdpbikge1xuICAgIHN1cGVyKGxlYWYpO1xuICAgIHRoaXMucGx1Z2luID0gcGx1Z2luO1xuICB9XG5cbiAgZ2V0Vmlld1R5cGUoKTogc3RyaW5nIHtcbiAgICByZXR1cm4gVklFV19UWVBFX01LQkxPRztcbiAgfVxuXG4gIGdldERpc3BsYXlUZXh0KCk6IHN0cmluZyB7XG4gICAgcmV0dXJuIFwibWtCbG9nXCI7XG4gIH1cblxuICBnZXRJY29uKCk6IHN0cmluZyB7XG4gICAgcmV0dXJuIFwibm90ZWJvb2stcGVuXCI7XG4gIH1cblxuICBhc3luYyBvbk9wZW4oKTogUHJvbWlzZTx2b2lkPiB7XG4gICAgdGhpcy5jb250ZW50RWwuZW1wdHkoKTtcbiAgICB0aGlzLmNvbnRlbnRFbC5hZGRDbGFzcyhcIm1rYmxvZy12aWV3XCIpO1xuXG4gICAgY29uc3QgaGVhZGVyID0gdGhpcy5jb250ZW50RWwuY3JlYXRlRGl2KHsgY2xzOiBcIm1rYmxvZy1oZWFkZXJcIiB9KTtcbiAgICBoZWFkZXIuY3JlYXRlRWwoXCJoM1wiLCB7IHRleHQ6IFwibWtCbG9nIFx1NjU4N1x1N0FFMFx1N0JBMVx1NzQwNlwiIH0pO1xuXG4gICAgY29uc3QgYWN0aW9ucyA9IGhlYWRlci5jcmVhdGVEaXYoeyBjbHM6IFwibWtibG9nLWFjdGlvbnNcIiB9KTtcbiAgICBjb25zdCByZWZyZXNoQnRuID0gYWN0aW9ucy5jcmVhdGVFbChcImJ1dHRvblwiLCB7IHRleHQ6IFwiXHU1MjM3XHU2NUIwXCIgfSk7XG4gICAgY29uc3QgdXBsb2FkRmlsZUJ0biA9IGFjdGlvbnMuY3JlYXRlRWwoXCJidXR0b25cIiwgeyB0ZXh0OiBcIlx1NEUwQVx1NEYyMFx1NUY1M1x1NTI0RFx1NjU4N1x1NEVGNlwiIH0pO1xuICAgIGNvbnN0IHVwbG9hZEZvbGRlckJ0biA9IGFjdGlvbnMuY3JlYXRlRWwoXCJidXR0b25cIiwgeyB0ZXh0OiBcIlx1NEUwQVx1NEYyMFx1NjU4N1x1NEVGNlx1NTkzOVwiIH0pO1xuXG4gICAgcmVmcmVzaEJ0bi5vbmNsaWNrID0gYXN5bmMgKCkgPT4ge1xuICAgICAgYXdhaXQgdGhpcy5wbHVnaW4ucmVmcmVzaEFydGljbGVzKCk7XG4gICAgfTtcbiAgICB1cGxvYWRGaWxlQnRuLm9uY2xpY2sgPSBhc3luYyAoKSA9PiB7XG4gICAgICBhd2FpdCB0aGlzLnBsdWdpbi51cGxvYWRDdXJyZW50RmlsZUFzQmxvZygpO1xuICAgIH07XG4gICAgdXBsb2FkRm9sZGVyQnRuLm9uY2xpY2sgPSBhc3luYyAoKSA9PiB7XG4gICAgICBhd2FpdCB0aGlzLnBsdWdpbi5waWNrQW5kVXBsb2FkRm9sZGVyKCk7XG4gICAgfTtcblxuICAgIHRoaXMubGlzdEVsID0gdGhpcy5jb250ZW50RWwuY3JlYXRlRGl2KHsgY2xzOiBcIm1rYmxvZy1saXN0XCIgfSk7XG4gICAgYXdhaXQgdGhpcy5yZW5kZXJMaXN0KCk7XG4gIH1cblxuICBhc3luYyByZW5kZXJMaXN0KCk6IFByb21pc2U8dm9pZD4ge1xuICAgIGlmICghdGhpcy5saXN0RWwpIHJldHVybjtcbiAgICB0aGlzLmxpc3RFbC5lbXB0eSgpO1xuXG4gICAgY29uc3QgaXRlbXMgPSB0aGlzLnBsdWdpbi5hcnRpY2xlcztcbiAgICBpZiAoIWl0ZW1zLmxlbmd0aCkge1xuICAgICAgdGhpcy5saXN0RWwuY3JlYXRlRWwoXCJkaXZcIiwge1xuICAgICAgICB0ZXh0OiBcIlx1NjY4Mlx1NjVFMFx1NjU4N1x1N0FFMFx1RkYwOFx1NTNFRlx1NzBCOVx1NTFGQlx1NTIzN1x1NjVCMFx1RkYwOVwiLFxuICAgICAgICBjbHM6IFwibWtibG9nLWVtcHR5XCIsXG4gICAgICB9KTtcbiAgICAgIHJldHVybjtcbiAgICB9XG5cbiAgICBmb3IgKGNvbnN0IGl0IG9mIGl0ZW1zKSB7XG4gICAgICBjb25zdCByb3cgPSB0aGlzLmxpc3RFbC5jcmVhdGVEaXYoeyBjbHM6IFwibWtibG9nLXJvd1wiIH0pO1xuICAgICAgY29uc3QgdGl0bGVFbCA9IHJvdy5jcmVhdGVEaXYoeyB0ZXh0OiBpdC50aXRsZSwgY2xzOiBcIm1rYmxvZy10aXRsZVwiIH0pO1xuICAgICAgdGl0bGVFbC5zZXRBdHRyaWJ1dGUoXCJ0aXRsZVwiLCBgJHtpdC50aXRsZX0gKElEOiAke1N0cmluZyhpdC5pZCl9KWApO1xuXG4gICAgICBjb25zdCBkZWxCdG4gPSByb3cuY3JlYXRlRWwoXCJidXR0b25cIiwgeyB0ZXh0OiBcIlx1NTIyMFx1OTY2NFwiIH0pO1xuICAgICAgZGVsQnRuLmFkZENsYXNzKFwibW9kLXdhcm5pbmdcIik7XG4gICAgICBkZWxCdG4ub25jbGljayA9IGFzeW5jICgpID0+IHtcbiAgICAgICAgdGhpcy5wbHVnaW4uY29uZmlybURlbGV0ZShpdCk7XG4gICAgICB9O1xuICAgIH1cbiAgfVxuXG4gIGFzeW5jIG9uQ2xvc2UoKTogUHJvbWlzZTx2b2lkPiB7XG4gICAgdGhpcy5jb250ZW50RWwuZW1wdHkoKTtcbiAgfVxufVxuXG5leHBvcnQgZGVmYXVsdCBjbGFzcyBNa0Jsb2dQbHVnaW4gZXh0ZW5kcyBQbHVnaW4ge1xuICBzZXR0aW5nczogTWtCbG9nU2V0dGluZ3MgPSBERUZBVUxUX1NFVFRJTkdTO1xuICBhcnRpY2xlczogUmF3QXJ0aWNsZVtdID0gW107XG5cbiAgYXN5bmMgb25sb2FkKCk6IFByb21pc2U8dm9pZD4ge1xuICAgIGF3YWl0IHRoaXMubG9hZFNldHRpbmdzKCk7XG5cbiAgICB0aGlzLnJlZ2lzdGVyVmlldyhcbiAgICAgIFZJRVdfVFlQRV9NS0JMT0csXG4gICAgICAobGVhZikgPT4gbmV3IE1rQmxvZ0FydGljbGVzVmlldyhsZWFmLCB0aGlzKSxcbiAgICApO1xuICAgIHRoaXMuYWRkU2V0dGluZ1RhYihuZXcgTWtCbG9nU2V0dGluZ1RhYih0aGlzLmFwcCwgdGhpcykpO1xuXG4gICAgdGhpcy5hZGRSaWJib25JY29uKFxuICAgICAgXCJjbG91ZC11cGxvYWRcIixcbiAgICAgIFwibWtCbG9nOiBcdTRFMEFcdTRGMjBcdTVGNTNcdTUyNERcdTY1ODdcdTRFRjZcdTRFM0FcdTUzNUFcdTVCQTJcIixcbiAgICAgIGFzeW5jICgpID0+IHtcbiAgICAgICAgYXdhaXQgdGhpcy51cGxvYWRDdXJyZW50RmlsZUFzQmxvZygpO1xuICAgICAgfSxcbiAgICApO1xuXG4gICAgdGhpcy5hZGRDb21tYW5kKHtcbiAgICAgIGlkOiBcIm1rYmxvZy1vcGVuLXZpZXdcIixcbiAgICAgIG5hbWU6IFwibWtCbG9nOiBcdTYyNTNcdTVGMDBcdTdCQTFcdTc0MDZcdTg5QzZcdTU2RkVcIixcbiAgICAgIGNhbGxiYWNrOiBhc3luYyAoKSA9PiB0aGlzLmFjdGl2YXRlVmlldygpLFxuICAgIH0pO1xuXG4gICAgdGhpcy5hZGRDb21tYW5kKHtcbiAgICAgIGlkOiBcIm1rYmxvZy11cGxvYWQtY3VycmVudC1maWxlXCIsXG4gICAgICBuYW1lOiBcIm1rQmxvZzogXHU0RTBBXHU0RjIwXHU1RjUzXHU1MjREXHU2NTg3XHU0RUY2XHU0RTNBXHU1MzVBXHU1QkEyXCIsXG4gICAgICBjYWxsYmFjazogYXN5bmMgKCkgPT4gdGhpcy51cGxvYWRDdXJyZW50RmlsZUFzQmxvZygpLFxuICAgIH0pO1xuXG4gICAgdGhpcy5hZGRDb21tYW5kKHtcbiAgICAgIGlkOiBcIm1rYmxvZy11cGxvYWQtZm9sZGVyXCIsXG4gICAgICBuYW1lOiBcIm1rQmxvZzogXHU0RTBBXHU0RjIwXHU5MDA5XHU2MkU5XHU2NTg3XHU0RUY2XHU1OTM5XHU0RTNBXHU1MzVBXHU1QkEyXCIsXG4gICAgICBjYWxsYmFjazogYXN5bmMgKCkgPT4gdGhpcy5waWNrQW5kVXBsb2FkRm9sZGVyKCksXG4gICAgfSk7XG5cbiAgICB0aGlzLmFkZENvbW1hbmQoe1xuICAgICAgaWQ6IFwibWtibG9nLXJlZnJlc2gtYXJ0aWNsZXNcIixcbiAgICAgIG5hbWU6IFwibWtCbG9nOiBcdTUyMzdcdTY1QjBcdTY1ODdcdTdBRTBcdTUyMTdcdTg4NjhcIixcbiAgICAgIGNhbGxiYWNrOiBhc3luYyAoKSA9PiB0aGlzLnJlZnJlc2hBcnRpY2xlcygpLFxuICAgIH0pO1xuXG4gICAgdGhpcy5hZGRDb21tYW5kKHtcbiAgICAgIGlkOiBcIm1rYmxvZy1kZWxldGUtYXJ0aWNsZVwiLFxuICAgICAgbmFtZTogXCJta0Jsb2c6IFx1NTIyMFx1OTY2NFx1NjU4N1x1N0FFMFwiLFxuICAgICAgY2FsbGJhY2s6IGFzeW5jICgpID0+IHRoaXMucGlja0FuZERlbGV0ZUFydGljbGUoKSxcbiAgICB9KTtcblxuICAgIGF3YWl0IHRoaXMuYWN0aXZhdGVWaWV3KCk7XG4gICAgYXdhaXQgdGhpcy5yZWZyZXNoQXJ0aWNsZXMoKS5jYXRjaCgoZSkgPT4ge1xuICAgICAgY29uc29sZS5lcnJvcihcIltta0Jsb2ddIGluaXRpYWwgcmVmcmVzaCBmYWlsZWRcIiwgZSk7XG4gICAgICBuZXcgTm90aWNlKGBta0Jsb2cgXHU1MjFEXHU1OUNCXHU1MzE2XHU2MkM5XHU1M0Q2XHU2NTg3XHU3QUUwXHU1OTMxXHU4RDI1OiAke2U/Lm1lc3NhZ2UgPz8gZX1gKTtcbiAgICB9KTtcbiAgfVxuXG4gIGFzeW5jIG9udW5sb2FkKCk6IFByb21pc2U8dm9pZD4ge1xuICAgIHRoaXMuYXBwLndvcmtzcGFjZS5kZXRhY2hMZWF2ZXNPZlR5cGUoVklFV19UWVBFX01LQkxPRyk7XG4gIH1cblxuICBwcml2YXRlIGF1dGhIZWFkZXJzKGpzb24gPSBmYWxzZSk6IFJlY29yZDxzdHJpbmcsIHN0cmluZz4ge1xuICAgIGNvbnN0IGg6IFJlY29yZDxzdHJpbmcsIHN0cmluZz4gPSB7IEFjY2VwdDogXCJhcHBsaWNhdGlvbi9qc29uXCIgfTtcbiAgICBpZiAodGhpcy5zZXR0aW5ncy5hdXRoVG9rZW4/LnRyaW0oKSkge1xuICAgICAgaFtcIkF1dGhvcml6YXRpb25cIl0gPSBgQmVhcmVyICR7dGhpcy5zZXR0aW5ncy5hdXRoVG9rZW4udHJpbSgpfWA7XG4gICAgfVxuICAgIGlmIChqc29uKSBoW1wiQ29udGVudC1UeXBlXCJdID0gXCJhcHBsaWNhdGlvbi9qc29uXCI7XG4gICAgcmV0dXJuIGg7XG4gIH1cblxuICBwcml2YXRlIGVuc3VyZUJhc2VVcmwoKTogc3RyaW5nIHtcbiAgICBjb25zdCBiYXNlID0gKHRoaXMuc2V0dGluZ3MuYmFzZVVybCB8fCBcIlwiKS50cmltKCk7XG4gICAgaWYgKCFiYXNlKSB0aHJvdyBuZXcgRXJyb3IoXCJcdTY3MkFcdTkxNERcdTdGNkUgQmFzZSBVUkxcIik7XG4gICAgcmV0dXJuIGJhc2U7XG4gIH1cblxuICBwcml2YXRlIGFzeW5jIGZpbGVUb0FycmF5QnVmZmVyKHZhdWx0UGF0aDogc3RyaW5nKTogUHJvbWlzZTxBcnJheUJ1ZmZlcj4ge1xuICAgIGNvbnN0IGFicyA9ICh0aGlzLmFwcC52YXVsdC5hZGFwdGVyIGFzIGFueSkuZ2V0RnVsbFBhdGg/Lih2YXVsdFBhdGgpO1xuICAgIGlmIChhYnMgJiYgXCJyZXF1ZXN0VXJsXCIgaW4gd2luZG93ID09PSBmYWxzZSkge1xuICAgICAgLy8gZmFsbGJhY2ssIGJ1dCB1c3VhbGx5IG5vdCBuZWVkZWRcbiAgICB9XG4gICAgcmV0dXJuIGF3YWl0IHRoaXMuYXBwLnZhdWx0LmFkYXB0ZXIucmVhZEJpbmFyeSh2YXVsdFBhdGgpO1xuICB9XG5cbiAgcHJpdmF0ZSBhc3luYyBjb2xsZWN0SW1hZ2VzRm9yTWFya2Rvd24oXG4gICAgbWRGaWxlOiBURmlsZSxcbiAgKTogUHJvbWlzZTx7IG5hbWU6IHN0cmluZzsgZGF0YUJhc2U2NDogc3RyaW5nIH1bXT4ge1xuICAgIGNvbnN0IHBhcmVudCA9IGRpcm5hbWUobWRGaWxlLnBhdGgpO1xuICAgIGNvbnN0IHRpdGxlID0gYmFzZW5hbWVXaXRob3V0RXh0KG1kRmlsZS5uYW1lKTtcbiAgICBjb25zdCBpbWdGb2xkZXJQYXRoID0gbm9ybWFsaXplUGF0aChwYXJlbnQgPyBgJHtwYXJlbnR9LyR7dGl0bGV9YCA6IHRpdGxlKTtcblxuICAgIGNvbnN0IGZvbGRlciA9IHRoaXMuYXBwLnZhdWx0LmdldEFic3RyYWN0RmlsZUJ5UGF0aChpbWdGb2xkZXJQYXRoKTtcbiAgICBpZiAoIWZvbGRlciB8fCAhKGZvbGRlciBpbnN0YW5jZW9mIFRGb2xkZXIpKSByZXR1cm4gW107XG5cbiAgICBjb25zdCBvdXQ6IHsgbmFtZTogc3RyaW5nOyBkYXRhQmFzZTY0OiBzdHJpbmcgfVtdID0gW107XG4gICAgY29uc3Qgc3RhY2s6IFRGb2xkZXJbXSA9IFtmb2xkZXJdO1xuXG4gICAgd2hpbGUgKHN0YWNrLmxlbmd0aCA+IDApIHtcbiAgICAgIGNvbnN0IGN1ciA9IHN0YWNrLnBvcCgpITtcbiAgICAgIGZvciAoY29uc3QgY2hpbGQgb2YgY3VyLmNoaWxkcmVuKSB7XG4gICAgICAgIGlmIChjaGlsZCBpbnN0YW5jZW9mIFRGb2xkZXIpIHtcbiAgICAgICAgICBzdGFjay5wdXNoKGNoaWxkKTtcbiAgICAgICAgICBjb250aW51ZTtcbiAgICAgICAgfVxuICAgICAgICBpZiAoIShjaGlsZCBpbnN0YW5jZW9mIFRGaWxlKSkgY29udGludWU7XG4gICAgICAgIGNvbnN0IGV4dCA9IGV4dG5hbWUoY2hpbGQubmFtZSk7XG4gICAgICAgIGlmICghSU1HX0VYVC5oYXMoZXh0KSkgY29udGludWU7XG5cbiAgICAgICAgY29uc3QgYnVmID0gYXdhaXQgdGhpcy5maWxlVG9BcnJheUJ1ZmZlcihjaGlsZC5wYXRoKTtcbiAgICAgICAgY29uc3QgYmFzZTY0ID0gdGhpcy5hcnJheUJ1ZmZlclRvQmFzZTY0KGJ1Zik7XG4gICAgICAgIC8vIG5hbWUgb25seSBrZWVwcyBmaWxlIG5hbWUgZm9yIGJhY2tlbmQgY29tcGF0aWJpbGl0eVxuICAgICAgICBvdXQucHVzaCh7IG5hbWU6IGNoaWxkLm5hbWUsIGRhdGFCYXNlNjQ6IGJhc2U2NCB9KTtcbiAgICAgIH1cbiAgICB9XG5cbiAgICByZXR1cm4gb3V0O1xuICB9XG5cbiAgcHJpdmF0ZSBhcnJheUJ1ZmZlclRvQmFzZTY0KGJ1ZjogQXJyYXlCdWZmZXIpOiBzdHJpbmcge1xuICAgIGxldCBiaW5hcnkgPSBcIlwiO1xuICAgIGNvbnN0IGJ5dGVzID0gbmV3IFVpbnQ4QXJyYXkoYnVmKTtcbiAgICBjb25zdCBjaHVuayA9IDB4ODAwMDtcbiAgICBmb3IgKGxldCBpID0gMDsgaSA8IGJ5dGVzLmxlbmd0aDsgaSArPSBjaHVuaykge1xuICAgICAgY29uc3Qgc3ViID0gYnl0ZXMuc3ViYXJyYXkoaSwgTWF0aC5taW4oaSArIGNodW5rLCBieXRlcy5sZW5ndGgpKTtcbiAgICAgIGJpbmFyeSArPSBTdHJpbmcuZnJvbUNoYXJDb2RlKC4uLnN1Yik7XG4gICAgfVxuICAgIHJldHVybiBidG9hKGJpbmFyeSk7XG4gIH1cblxuICBhc3luYyBmZXRjaEFydGljbGVzKCk6IFByb21pc2U8UmF3QXJ0aWNsZVtdPiB7XG4gICAgY29uc3QgYmFzZVVybCA9IHRoaXMuZW5zdXJlQmFzZVVybCgpO1xuICAgIGNvbnN0IGxpc3RVcmwgPSBqb2luVXJsKGJhc2VVcmwsIFwiL2FwaS9hbGxhcnRpY2xlc1wiKTtcbiAgICBjb25zdCBkYXRhID0gYXdhaXQgcmVxSnNvbihsaXN0VXJsLCB7XG4gICAgICBtZXRob2Q6IFwiR0VUXCIsXG4gICAgICBoZWFkZXJzOiB0aGlzLmF1dGhIZWFkZXJzKGZhbHNlKSxcbiAgICB9KTtcblxuICAgIGxldCBsaXN0OiBhbnlbXSA9IFtdO1xuICAgIGlmIChBcnJheS5pc0FycmF5KGRhdGEpKSBsaXN0ID0gZGF0YTtcbiAgICBlbHNlIGlmIChBcnJheS5pc0FycmF5KGRhdGE/LmFydGljbGVzKSkgbGlzdCA9IGRhdGEuYXJ0aWNsZXM7XG4gICAgZWxzZSBpZiAoQXJyYXkuaXNBcnJheShkYXRhPy5kYXRhKSkgbGlzdCA9IGRhdGEuZGF0YTtcbiAgICBlbHNlIGlmIChBcnJheS5pc0FycmF5KGRhdGE/LmRhdGE/LmFydGljbGVzKSkgbGlzdCA9IGRhdGEuZGF0YS5hcnRpY2xlcztcbiAgICBlbHNlIGlmIChBcnJheS5pc0FycmF5KGRhdGE/Lml0ZW1zKSkgbGlzdCA9IGRhdGEuaXRlbXM7XG4gICAgZWxzZSBpZiAoQXJyYXkuaXNBcnJheShkYXRhPy5saXN0KSkgbGlzdCA9IGRhdGEubGlzdDtcblxuICAgIHJldHVybiBsaXN0Lm1hcCgoaXQ6IGFueSwgaTogbnVtYmVyKSA9PiAoe1xuICAgICAgaWQ6IGl0Py5pZCA/PyBpdD8uX2lkID8/IGl0Py5zbHVnID8/IGl0Py50aXRsZSA/PyBpLFxuICAgICAgdGl0bGU6IFN0cmluZyhpdD8udGl0bGUgPz8gaXQ/LmlkID8/IGl0Py5faWQgPz8gYHVudGl0bGVkLSR7aX1gKSxcbiAgICB9KSk7XG4gIH1cblxuICBhc3luYyByZWZyZXNoQXJ0aWNsZXMoKTogUHJvbWlzZTx2b2lkPiB7XG4gICAgdHJ5IHtcbiAgICAgIHRoaXMuYXJ0aWNsZXMgPSBhd2FpdCB0aGlzLmZldGNoQXJ0aWNsZXMoKTtcbiAgICAgIG5ldyBOb3RpY2UoYG1rQmxvZzogXHU1REYyXHU1MjM3XHU2NUIwXHVGRjBDXHU1MTcxICR7dGhpcy5hcnRpY2xlcy5sZW5ndGh9IFx1N0JDN2ApO1xuICAgIH0gY2F0Y2ggKGU6IGFueSkge1xuICAgICAgbmV3IE5vdGljZShgbWtCbG9nOiBcdTUyMzdcdTY1QjBcdTU5MzFcdThEMjUgLSAke2U/Lm1lc3NhZ2UgPz8gZX1gKTtcbiAgICAgIHRocm93IGU7XG4gICAgfSBmaW5hbGx5IHtcbiAgICAgIHRoaXMucmVkcmF3VmlldygpO1xuICAgIH1cbiAgfVxuXG4gIHByaXZhdGUgcmVkcmF3VmlldygpOiB2b2lkIHtcbiAgICBjb25zdCBsZWF2ZXMgPSB0aGlzLmFwcC53b3Jrc3BhY2UuZ2V0TGVhdmVzT2ZUeXBlKFZJRVdfVFlQRV9NS0JMT0cpO1xuICAgIGZvciAoY29uc3QgbGVhZiBvZiBsZWF2ZXMpIHtcbiAgICAgIGNvbnN0IHYgPSBsZWFmLnZpZXc7XG4gICAgICBpZiAodiBpbnN0YW5jZW9mIE1rQmxvZ0FydGljbGVzVmlldykge1xuICAgICAgICB2LnJlbmRlckxpc3QoKTtcbiAgICAgIH1cbiAgICB9XG4gIH1cblxuICBhc3luYyB1cGxvYWRDdXJyZW50RmlsZUFzQmxvZygpOiBQcm9taXNlPHZvaWQ+IHtcbiAgICBjb25zdCBmaWxlID0gdGhpcy5hcHAud29ya3NwYWNlLmdldEFjdGl2ZUZpbGUoKTtcbiAgICBpZiAoIWZpbGUpIHtcbiAgICAgIG5ldyBOb3RpY2UoXCJcdThCRjdcdTUxNDhcdTYyNTNcdTVGMDBcdTRFMDBcdTRFMkEgTWFya2Rvd24gXHU2NTg3XHU0RUY2XCIpO1xuICAgICAgcmV0dXJuO1xuICAgIH1cbiAgICBpZiAoIShmaWxlIGluc3RhbmNlb2YgVEZpbGUpIHx8IGZpbGUuZXh0ZW5zaW9uLnRvTG93ZXJDYXNlKCkgIT09IFwibWRcIikge1xuICAgICAgbmV3IE5vdGljZShcIlx1NEVDNVx1NjUyRlx1NjMwMVx1NEUwQVx1NEYyMCBNYXJrZG93biBcdTY1ODdcdTRFRjZcdUZGMDgubWRcdUZGMDlcIik7XG4gICAgICByZXR1cm47XG4gICAgfVxuXG4gICAgYXdhaXQgdGhpcy51cGxvYWRTaW5nbGVGaWxlKGZpbGUpO1xuICAgIGF3YWl0IHRoaXMucmVmcmVzaEFydGljbGVzKCkuY2F0Y2goKCkgPT4ge30pO1xuICB9XG5cbiAgcHJpdmF0ZSBhc3luYyB1cGxvYWRTaW5nbGVGaWxlKG1kRmlsZTogVEZpbGUpOiBQcm9taXNlPHZvaWQ+IHtcbiAgICBjb25zdCBiYXNlVXJsID0gdGhpcy5lbnN1cmVCYXNlVXJsKCk7XG4gICAgY29uc3QgdGl0bGUgPSBiYXNlbmFtZVdpdGhvdXRFeHQobWRGaWxlLm5hbWUpO1xuICAgIGNvbnN0IG1kUmF3ID0gYXdhaXQgdGhpcy5hcHAudmF1bHQuY2FjaGVkUmVhZChtZEZpbGUpO1xuICAgIGNvbnN0IG1ldGEgPSBwYXJzZU1ldGEobWRSYXcpO1xuXG4gICAgY29uc3QgcGF5bG9hZDogUmVjb3JkPHN0cmluZywgYW55PiA9IHtcbiAgICAgIHRpdGxlLFxuICAgICAgdXBkYXRlX2F0OiBub3dBc1VwZGF0ZUF0KCksXG4gICAgICBjb250ZW50OiBtZXRhLmNvbnRlbnQsXG4gICAgfTtcbiAgICBwYXlsb2FkLmF1dGhvciA9IG1ldGEuYXV0aG9yID8/IHRoaXMuc2V0dGluZ3MuZGVmYXVsdEF1dGhvciA/PyBcIlwiO1xuICAgIHBheWxvYWQuY2F0ZWdvcnkgPSBtZXRhLmNhdGVnb3J5ID8/IHRoaXMuc2V0dGluZ3MuZGVmYXVsdENhdGVnb3J5ID8/IFwiXCI7XG5cbiAgICBjb25zdCBhcnRpY2xlVXJsID0gYnVpbGRBcnRpY2xlRW5kcG9pbnQoYmFzZVVybCwgdGl0bGUpO1xuICAgIGF3YWl0IHJlcUpzb24oYXJ0aWNsZVVybCwge1xuICAgICAgbWV0aG9kOiBcIlBVVFwiLFxuICAgICAgaGVhZGVyczogdGhpcy5hdXRoSGVhZGVycyh0cnVlKSxcbiAgICAgIGJvZHk6IEpTT04uc3RyaW5naWZ5KHBheWxvYWQpLFxuICAgIH0pO1xuXG4gICAgY29uc3QgaW1hZ2VzID0gYXdhaXQgdGhpcy5jb2xsZWN0SW1hZ2VzRm9yTWFya2Rvd24obWRGaWxlKTtcbiAgICBjb25zdCBpbWFnZVVybCA9IGJ1aWxkSW1hZ2VFbmRwb2ludChiYXNlVXJsKTtcblxuICAgIGZvciAoY29uc3QgaW1nIG9mIGltYWdlcykge1xuICAgICAgY29uc3QgaW1nUGF5bG9hZCA9IHsgdGl0bGUsIG5hbWU6IGltZy5uYW1lLCBkYXRhOiBpbWcuZGF0YUJhc2U2NCB9O1xuICAgICAgYXdhaXQgcmVxSnNvbihpbWFnZVVybCwge1xuICAgICAgICBtZXRob2Q6IFwiUFVUXCIsXG4gICAgICAgIGhlYWRlcnM6IHRoaXMuYXV0aEhlYWRlcnModHJ1ZSksXG4gICAgICAgIGJvZHk6IEpTT04uc3RyaW5naWZ5KGltZ1BheWxvYWQpLFxuICAgICAgfSk7XG4gICAgfVxuXG4gICAgbmV3IE5vdGljZShcbiAgICAgIGBcdTRFMEFcdTRGMjBcdTVCOENcdTYyMTBcdUZGMUEke3RpdGxlfSR7aW1hZ2VzLmxlbmd0aCA/IGBcdUZGMDhcdTU2RkVcdTcyNDcgJHtpbWFnZXMubGVuZ3RofSBcdTVGMjBcdUZGMDlgIDogXCJcIn1gLFxuICAgICk7XG4gIH1cblxuICBhc3luYyBwaWNrQW5kVXBsb2FkRm9sZGVyKCk6IFByb21pc2U8dm9pZD4ge1xuICAgIGNvbnN0IGFsbEZvbGRlcnMgPSB0aGlzLmFwcC52YXVsdFxuICAgICAgLmdldEFsbExvYWRlZEZpbGVzKClcbiAgICAgIC5maWx0ZXIoKGYpID0+IGYgaW5zdGFuY2VvZiBURm9sZGVyKSBhcyBURm9sZGVyW107XG4gICAgY29uc3Qgcm9vdFBhdGggPSB0aGlzLmFwcC52YXVsdC5nZXRSb290KCkucGF0aDtcbiAgICBjb25zdCBjYW5kaWRhdGVzID0gYWxsRm9sZGVycy5maWx0ZXIoKGYpID0+IGYucGF0aCAhPT0gcm9vdFBhdGgpO1xuXG4gICAgaWYgKCFjYW5kaWRhdGVzLmxlbmd0aCkge1xuICAgICAgbmV3IE5vdGljZShcIlx1NjcyQVx1NjI3RVx1NTIzMFx1NTNFRlx1NEUwQVx1NEYyMFx1NzY4NFx1NjU4N1x1NEVGNlx1NTkzOVwiKTtcbiAgICAgIHJldHVybjtcbiAgICB9XG5cbiAgICBuZXcgRm9sZGVyUGlja2VyTW9kYWwodGhpcy5hcHAsIGNhbmRpZGF0ZXMsIGFzeW5jIChmb2xkZXIpID0+IHtcbiAgICAgIHRyeSB7XG4gICAgICAgIGF3YWl0IHRoaXMudXBsb2FkRm9sZGVyQXNCbG9nKGZvbGRlcik7XG4gICAgICAgIGF3YWl0IHRoaXMucmVmcmVzaEFydGljbGVzKCkuY2F0Y2goKCkgPT4ge30pO1xuICAgICAgfSBjYXRjaCAoZTogYW55KSB7XG4gICAgICAgIG5ldyBOb3RpY2UoYFx1NEUwQVx1NEYyMFx1NjU4N1x1NEVGNlx1NTkzOVx1NTkzMVx1OEQyNTogJHtlPy5tZXNzYWdlID8/IGV9YCk7XG4gICAgICB9XG4gICAgfSkub3BlbigpO1xuICB9XG5cbiAgcHJpdmF0ZSBhc3luYyB1cGxvYWRGb2xkZXJBc0Jsb2coZm9sZGVyOiBURm9sZGVyKTogUHJvbWlzZTx2b2lkPiB7XG4gICAgY29uc3QgbWRGaWxlcyA9IHRoaXMuY29sbGVjdE1hcmtkb3duRmlsZXMoZm9sZGVyKTtcbiAgICBpZiAoIW1kRmlsZXMubGVuZ3RoKSB7XG4gICAgICBuZXcgTm90aWNlKFwiXHU2MjQwXHU5MDA5XHU2NTg3XHU0RUY2XHU1OTM5XHU2NzJBXHU2MjdFXHU1MjMwIC5tZCBcdTY1ODdcdTRFRjZcIik7XG4gICAgICByZXR1cm47XG4gICAgfVxuXG4gICAgbGV0IHN1Y2Nlc3MgPSAwO1xuICAgIGZvciAoY29uc3QgbWQgb2YgbWRGaWxlcykge1xuICAgICAgdHJ5IHtcbiAgICAgICAgYXdhaXQgdGhpcy51cGxvYWRTaW5nbGVGaWxlKG1kKTtcbiAgICAgICAgc3VjY2VzcysrO1xuICAgICAgfSBjYXRjaCAoZTogYW55KSB7XG4gICAgICAgIGNvbnNvbGUuZXJyb3IoYFtta0Jsb2ddIHVwbG9hZCBmYWlsZWQgZm9yICR7bWQucGF0aH1gLCBlKTtcbiAgICAgICAgbmV3IE5vdGljZShgXHU0RTBBXHU0RjIwXHU1OTMxXHU4RDI1OiAke21kLnBhdGh9IC0gJHtlPy5tZXNzYWdlID8/IGV9YCk7XG4gICAgICB9XG4gICAgfVxuXG4gICAgbmV3IE5vdGljZShgXHU2NTg3XHU0RUY2XHU1OTM5XHU0RTBBXHU0RjIwXHU1QjhDXHU2MjEwXHVGRjFBXHU2MjEwXHU1MjlGICR7c3VjY2Vzc30vJHttZEZpbGVzLmxlbmd0aH1gKTtcbiAgfVxuXG4gIHByaXZhdGUgY29sbGVjdE1hcmtkb3duRmlsZXMoZm9sZGVyOiBURm9sZGVyKTogVEZpbGVbXSB7XG4gICAgY29uc3Qgb3V0OiBURmlsZVtdID0gW107XG4gICAgY29uc3Qgc3RhY2s6IFRGb2xkZXJbXSA9IFtmb2xkZXJdO1xuICAgIHdoaWxlIChzdGFjay5sZW5ndGggPiAwKSB7XG4gICAgICBjb25zdCBjdXIgPSBzdGFjay5wb3AoKSE7XG4gICAgICBmb3IgKGNvbnN0IGMgb2YgY3VyLmNoaWxkcmVuKSB7XG4gICAgICAgIGlmIChjIGluc3RhbmNlb2YgVEZvbGRlcikgc3RhY2sucHVzaChjKTtcbiAgICAgICAgZWxzZSBpZiAoYyBpbnN0YW5jZW9mIFRGaWxlICYmIGMuZXh0ZW5zaW9uLnRvTG93ZXJDYXNlKCkgPT09IFwibWRcIilcbiAgICAgICAgICBvdXQucHVzaChjKTtcbiAgICAgIH1cbiAgICB9XG4gICAgcmV0dXJuIG91dDtcbiAgfVxuXG4gIGFzeW5jIHBpY2tBbmREZWxldGVBcnRpY2xlKCk6IFByb21pc2U8dm9pZD4ge1xuICAgIGlmICghdGhpcy5hcnRpY2xlcy5sZW5ndGgpIHtcbiAgICAgIGF3YWl0IHRoaXMucmVmcmVzaEFydGljbGVzKCkuY2F0Y2goKCkgPT4ge30pO1xuICAgIH1cbiAgICBpZiAoIXRoaXMuYXJ0aWNsZXMubGVuZ3RoKSB7XG4gICAgICBuZXcgTm90aWNlKFwiXHU2NjgyXHU2NUUwXHU1M0VGXHU1MjIwXHU5NjY0XHU2NTg3XHU3QUUwXCIpO1xuICAgICAgcmV0dXJuO1xuICAgIH1cblxuICAgIG5ldyBBcnRpY2xlUGlja2VyTW9kYWwodGhpcy5hcHAsIHRoaXMuYXJ0aWNsZXMsIChpdCkgPT5cbiAgICAgIHRoaXMuY29uZmlybURlbGV0ZShpdCksXG4gICAgKS5vcGVuKCk7XG4gIH1cblxuICBjb25maXJtRGVsZXRlKGl0ZW06IFJhd0FydGljbGUpOiB2b2lkIHtcbiAgICBuZXcgQ29uZmlybU1vZGFsKHRoaXMuYXBwLCBgXHU3ODZFXHU1QjlBXHU1MjIwXHU5NjY0XHU2NTg3XHU3QUUwXHUzMDBDJHtpdGVtLnRpdGxlfVx1MzAwRFx1RkYxRmAsIGFzeW5jICgpID0+IHtcbiAgICAgIHRyeSB7XG4gICAgICAgIGF3YWl0IHRoaXMuZGVsZXRlQXJ0aWNsZUJ5VGl0bGUoaXRlbS50aXRsZSk7XG4gICAgICAgIG5ldyBOb3RpY2UoYFx1NTIyMFx1OTY2NFx1NjIxMFx1NTI5Rlx1RkYxQSR7aXRlbS50aXRsZX1gKTtcbiAgICAgICAgYXdhaXQgdGhpcy5yZWZyZXNoQXJ0aWNsZXMoKS5jYXRjaCgoKSA9PiB7fSk7XG4gICAgICB9IGNhdGNoIChlOiBhbnkpIHtcbiAgICAgICAgbmV3IE5vdGljZShgXHU1MjIwXHU5NjY0XHU1OTMxXHU4RDI1XHVGRjFBJHtlPy5tZXNzYWdlID8/IGV9YCk7XG4gICAgICB9XG4gICAgfSkub3BlbigpO1xuICB9XG5cbiAgYXN5bmMgZGVsZXRlQXJ0aWNsZUJ5VGl0bGUodGl0bGU6IHN0cmluZyk6IFByb21pc2U8dm9pZD4ge1xuICAgIGNvbnN0IGJhc2VVcmwgPSB0aGlzLmVuc3VyZUJhc2VVcmwoKTtcbiAgICBjb25zdCB1cmwgPSBidWlsZEFydGljbGVFbmRwb2ludChiYXNlVXJsLCB0aXRsZSk7XG4gICAgY29uc3QgcmVzID0gYXdhaXQgZmV0Y2godXJsLCB7XG4gICAgICBtZXRob2Q6IFwiREVMRVRFXCIsXG4gICAgICBoZWFkZXJzOiB0aGlzLmF1dGhIZWFkZXJzKGZhbHNlKSxcbiAgICB9KTtcbiAgICBpZiAoIXJlcy5vaykge1xuICAgICAgY29uc3QgdGV4dCA9IGF3YWl0IHJlcy50ZXh0KCkuY2F0Y2goKCkgPT4gXCJcIik7XG4gICAgICB0aHJvdyBuZXcgRXJyb3IoYEhUVFAgJHtyZXMuc3RhdHVzfSAke3Jlcy5zdGF0dXNUZXh0fSAke3RleHR9YCk7XG4gICAgfVxuICB9XG5cbiAgcHJpdmF0ZSBhc3luYyBhY3RpdmF0ZVZpZXcoKTogUHJvbWlzZTx2b2lkPiB7XG4gICAgY29uc3QgeyB3b3Jrc3BhY2UgfSA9IHRoaXMuYXBwO1xuICAgIGxldCBsZWFmOiBXb3Jrc3BhY2VMZWFmIHwgbnVsbCA9XG4gICAgICB3b3Jrc3BhY2UuZ2V0TGVhdmVzT2ZUeXBlKFZJRVdfVFlQRV9NS0JMT0cpWzBdID8/IG51bGw7XG5cbiAgICBpZiAoIWxlYWYpIHtcbiAgICAgIGxlYWYgPSB3b3Jrc3BhY2UuZ2V0UmlnaHRMZWFmKGZhbHNlKTtcbiAgICAgIGlmICghbGVhZikgcmV0dXJuO1xuXG4gICAgICBhd2FpdCBsZWFmLnNldFZpZXdTdGF0ZSh7XG4gICAgICAgIHR5cGU6IFZJRVdfVFlQRV9NS0JMT0csXG4gICAgICAgIGFjdGl2ZTogdHJ1ZSxcbiAgICAgIH0pO1xuICAgIH1cblxuICAgIHdvcmtzcGFjZS5yZXZlYWxMZWFmKGxlYWYpO1xuICB9XG5cbiAgYXN5bmMgbG9hZFNldHRpbmdzKCk6IFByb21pc2U8dm9pZD4ge1xuICAgIGNvbnN0IGxvYWRlZCA9IChhd2FpdCB0aGlzLmxvYWREYXRhKCkpIGFzIFBhcnRpYWw8TWtCbG9nU2V0dGluZ3M+IHwgbnVsbDtcbiAgICB0aGlzLnNldHRpbmdzID0gT2JqZWN0LmFzc2lnbih7fSwgREVGQVVMVF9TRVRUSU5HUywgbG9hZGVkID8/IHt9KTtcbiAgfVxuXG4gIGFzeW5jIHNhdmVTZXR0aW5ncygpOiBQcm9taXNlPHZvaWQ+IHtcbiAgICBhd2FpdCB0aGlzLnNhdmVEYXRhKHRoaXMuc2V0dGluZ3MpO1xuICB9XG59XG5cbmNsYXNzIE1rQmxvZ1NldHRpbmdUYWIgZXh0ZW5kcyBQbHVnaW5TZXR0aW5nVGFiIHtcbiAgcGx1Z2luOiBNa0Jsb2dQbHVnaW47XG5cbiAgY29uc3RydWN0b3IoYXBwOiBBcHAsIHBsdWdpbjogTWtCbG9nUGx1Z2luKSB7XG4gICAgc3VwZXIoYXBwLCBwbHVnaW4pO1xuICAgIHRoaXMucGx1Z2luID0gcGx1Z2luO1xuICB9XG5cbiAgZGlzcGxheSgpOiB2b2lkIHtcbiAgICBjb25zdCB7IGNvbnRhaW5lckVsIH0gPSB0aGlzO1xuICAgIGNvbnRhaW5lckVsLmVtcHR5KCk7XG5cbiAgICBjb250YWluZXJFbC5jcmVhdGVFbChcImgyXCIsIHsgdGV4dDogXCJta0Jsb2cgXHU2M0QyXHU0RUY2XHU4QkJFXHU3RjZFXCIgfSk7XG5cbiAgICBuZXcgU2V0dGluZyhjb250YWluZXJFbClcbiAgICAgIC5zZXROYW1lKFwiQmFzZSBVUkxcIilcbiAgICAgIC5zZXREZXNjKFwiXHU1NDBFXHU3QUVGXHU2NzBEXHU1MkExXHU1N0ZBXHU3ODQwXHU1NzMwXHU1NzQwXHVGRjBDXHU0RjhCXHU1OTgyIGh0dHA6Ly9sb2NhbGhvc3Q6ODA4MFwiKVxuICAgICAgLmFkZFRleHQoKHRleHQpID0+XG4gICAgICAgIHRleHRcbiAgICAgICAgICAuc2V0UGxhY2Vob2xkZXIoXCJodHRwOi8vbG9jYWxob3N0OjgwODBcIilcbiAgICAgICAgICAuc2V0VmFsdWUodGhpcy5wbHVnaW4uc2V0dGluZ3MuYmFzZVVybClcbiAgICAgICAgICAub25DaGFuZ2UoYXN5bmMgKHZhbHVlKSA9PiB7XG4gICAgICAgICAgICB0aGlzLnBsdWdpbi5zZXR0aW5ncy5iYXNlVXJsID0gdmFsdWUudHJpbSgpO1xuICAgICAgICAgICAgYXdhaXQgdGhpcy5wbHVnaW4uc2F2ZVNldHRpbmdzKCk7XG4gICAgICAgICAgfSksXG4gICAgICApO1xuXG4gICAgbmV3IFNldHRpbmcoY29udGFpbmVyRWwpXG4gICAgICAuc2V0TmFtZShcIkRlZmF1bHQgQXV0aG9yXCIpXG4gICAgICAuc2V0RGVzYyhcIk1hcmtkb3duIFx1NjcyQVx1NThGMFx1NjYwRSBhdXRob3IgXHU2NUY2XHU0RjdGXHU3NTI4XCIpXG4gICAgICAuYWRkVGV4dCgodGV4dCkgPT5cbiAgICAgICAgdGV4dFxuICAgICAgICAgIC5zZXRWYWx1ZSh0aGlzLnBsdWdpbi5zZXR0aW5ncy5kZWZhdWx0QXV0aG9yKVxuICAgICAgICAgIC5vbkNoYW5nZShhc3luYyAodmFsdWUpID0+IHtcbiAgICAgICAgICAgIHRoaXMucGx1Z2luLnNldHRpbmdzLmRlZmF1bHRBdXRob3IgPSB2YWx1ZTtcbiAgICAgICAgICAgIGF3YWl0IHRoaXMucGx1Z2luLnNhdmVTZXR0aW5ncygpO1xuICAgICAgICAgIH0pLFxuICAgICAgKTtcblxuICAgIG5ldyBTZXR0aW5nKGNvbnRhaW5lckVsKVxuICAgICAgLnNldE5hbWUoXCJEZWZhdWx0IENhdGVnb3J5XCIpXG4gICAgICAuc2V0RGVzYyhcIk1hcmtkb3duIFx1NjcyQVx1NThGMFx1NjYwRSBjYXRlZ29yeSBcdTY1RjZcdTRGN0ZcdTc1MjhcIilcbiAgICAgIC5hZGRUZXh0KCh0ZXh0KSA9PlxuICAgICAgICB0ZXh0XG4gICAgICAgICAgLnNldFZhbHVlKHRoaXMucGx1Z2luLnNldHRpbmdzLmRlZmF1bHRDYXRlZ29yeSlcbiAgICAgICAgICAub25DaGFuZ2UoYXN5bmMgKHZhbHVlKSA9PiB7XG4gICAgICAgICAgICB0aGlzLnBsdWdpbi5zZXR0aW5ncy5kZWZhdWx0Q2F0ZWdvcnkgPSB2YWx1ZTtcbiAgICAgICAgICAgIGF3YWl0IHRoaXMucGx1Z2luLnNhdmVTZXR0aW5ncygpO1xuICAgICAgICAgIH0pLFxuICAgICAgKTtcblxuICAgIG5ldyBTZXR0aW5nKGNvbnRhaW5lckVsKVxuICAgICAgLnNldE5hbWUoXCJBdXRoIFRva2VuXCIpXG4gICAgICAuc2V0RGVzYyhcIlx1NTNFRlx1OTAwOSBCZWFyZXIgVG9rZW5cdUZGMENcdTVDMDZcdTkwMUFcdThGQzcgQXV0aG9yaXphdGlvbiBcdTU5MzRcdTUzRDFcdTkwMDFcIilcbiAgICAgIC5hZGRUZXh0KCh0ZXh0KSA9PlxuICAgICAgICB0ZXh0XG4gICAgICAgICAgLnNldFBsYWNlaG9sZGVyKFwiZXlKaGJHY2lPaS4uLlwiKVxuICAgICAgICAgIC5zZXRWYWx1ZSh0aGlzLnBsdWdpbi5zZXR0aW5ncy5hdXRoVG9rZW4pXG4gICAgICAgICAgLm9uQ2hhbmdlKGFzeW5jICh2YWx1ZSkgPT4ge1xuICAgICAgICAgICAgdGhpcy5wbHVnaW4uc2V0dGluZ3MuYXV0aFRva2VuID0gdmFsdWUudHJpbSgpO1xuICAgICAgICAgICAgYXdhaXQgdGhpcy5wbHVnaW4uc2F2ZVNldHRpbmdzKCk7XG4gICAgICAgICAgfSksXG4gICAgICApO1xuICB9XG59XG4iXSwKICAibWFwcGluZ3MiOiAiOzs7Ozs7Ozs7Ozs7Ozs7Ozs7OztBQUFBO0FBQUE7QUFBQTtBQUFBO0FBQUE7QUFBQSxzQkFhTztBQUVQLElBQU0sbUJBQW1CO0FBY3pCLElBQU0sbUJBQW1DO0FBQUEsRUFDdkMsU0FBUztBQUFBLEVBQ1QsZUFBZTtBQUFBLEVBQ2YsaUJBQWlCO0FBQUEsRUFDakIsV0FBVztBQUNiO0FBRUEsSUFBTSxVQUFVLG9CQUFJLElBQUksQ0FBQyxRQUFRLFFBQVEsU0FBUyxRQUFRLFNBQVMsTUFBTSxDQUFDO0FBRTFFLFNBQVMsUUFBUSxTQUFpQixNQUFzQjtBQUN0RCxRQUFNLFFBQVEsV0FBVyxJQUFJLFFBQVEsUUFBUSxFQUFFO0FBQy9DLFFBQU0sU0FBUyxLQUFLLFdBQVcsR0FBRyxJQUFJLE9BQU8sSUFBSSxJQUFJO0FBQ3JELFNBQU8sR0FBRyxJQUFJLEdBQUcsTUFBTTtBQUN6QjtBQUVBLFNBQVMscUJBQXFCLFNBQWlCLE9BQXVCO0FBQ3BFLFNBQU8sUUFBUSxTQUFTLGdCQUFnQixtQkFBbUIsS0FBSyxDQUFDLEVBQUU7QUFDckU7QUFFQSxTQUFTLG1CQUFtQixTQUF5QjtBQUNuRCxTQUFPLFFBQVEsU0FBUyxZQUFZO0FBQ3RDO0FBRUEsU0FBUyxnQkFBd0I7QUFDL0IsUUFBTSxNQUFNLENBQUMsTUFBZSxJQUFJLEtBQUssSUFBSSxDQUFDLEtBQUssT0FBTyxDQUFDO0FBQ3ZELFFBQU0sSUFBSSxvQkFBSSxLQUFL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

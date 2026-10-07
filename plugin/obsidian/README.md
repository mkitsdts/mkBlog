# mkBlog Obsidian Uploader

将 Obsidian Vault 中的 Markdown 与同名图片文件夹一起上传到 mkBlog 后端，并在侧边栏展示文章列表，支持删除远端文章。

## 功能

- 解析 Markdown 头部元数据（Frontmatter）中的 `author` / `category`。
- 单独上传当前 Markdown 文件（并自动上传同名图片文件夹中的图片）。
- 单独上传指定文件夹中的所有 Markdown（并自动上传各自同名图片文件夹中的图片）。
- 侧边栏管理视图中展示文章列表，支持删除远端文章。
- 启动流程不与 Obsidian 布局恢复竞争：网络请求都在布局就绪后执行，且带 15s 超时。

---

## 安装（开发态）

1. 构建：

```bash
cd plugin/obsidian
pnpm install
pnpm build   # 生成 build/main.js 与 build/manifest.json
```

2. 把 `build/main.js` 与 `build/manifest.json` 复制到 Vault 插件目录。
   目录名可以自定义，但清单文件必须叫 `manifest.json`（**不是** `mainifest.json`）：

```
<Vault>/.obsidian/plugins/mkblog/
  ├─ manifest.json
  ├─ main.js
  └─ styles.css (可选)
```

2. 在 Obsidian 中打开：`设置 -> 第三方插件`
3. 关闭安全模式（如尚未关闭）
4. 在已安装插件中启用 `mkBlog Obsidian Uploader`

> 如果你使用 TypeScript 源码开发，请先构建生成 `main.js` 再启用插件。

---

## 配置项

在插件设置页中可配置：

- `Base URL`
  - 后端服务基础地址，例如：`http://localhost:8080`
  - 插件会自动拼接接口路径：
    - `/api/allarticles`
    - `/api/article/:title`
    - `/api/image`
- `Default Author`
  - 当 Markdown 未声明作者时使用
- `Default Category`
  - 当 Markdown 未声明分类时使用
- `Auth Token`（可选）
  - 若填写，会通过请求头发送：
    - `Authorization: Bearer <token>`
- `启动时自动打开管理视图`（默认关闭）
  - 关闭后仅在点击左侧 Ribbon 图标或执行命令时打开管理视图。
  - 保持关闭可避免每次启动都往 `workspace.json` 写入新的视图叶子（推荐）。
- `启动时自动刷新文章列表`（默认开启）
  - 在 `workspace.onLayoutReady` 之后发起请求，不阻塞 Obsidian 启动。

---

## Markdown 元数据解析规则

插件会优先读取文档开头的 Frontmatter：

```
---
author: mkitsdts
category: language
---
# 标题
正文...
```

解析行为：

1. 若存在 `author` / `category`，优先使用。
2. 若缺失，则回退到插件设置中的默认值。
3. 上传正文时会去掉 Frontmatter，仅上传正文内容。

---

## 图片匹配规则

对于 `post.md`：

- 插件会查找同目录下同名文件夹 `post/`
- 读取其中图片并上传（支持常见格式：`.png .jpg .jpeg .gif .webp .svg`）

示例结构：

```
Notes/
  ├─ post.md
  └─ post/
     ├─ 1.png
     └─ cover.jpg
```

---

## 命令

插件提供以下命令（可在命令面板执行）：

- `mkBlog: 上传当前文件为博客`
- `mkBlog: 上传选择文件夹为博客`
- `mkBlog: 刷新文章列表`
- `mkBlog: 删除文章`

---

## 典型工作流

1. 在 Vault 中编写 `xxx.md`
2. （可选）在文档头部写 `author` / `category`
3. 将引用图片放在同名文件夹 `xxx/` 下
4. 执行 `mkBlog: 上传当前文件为博客`
5. 在侧边栏确认文章列表是否刷新成功

---

## 后端接口约定

### 1) 拉取文章列表

- `GET /api/allarticles`
- 兼容返回结构：
  - `[...]`
  - `{ articles: [...] }`
  - `{ data: [...] }`
  - `{ data: { articles: [...] } }`
  - `{ items: [...] }`
  - `{ list: [...] }`

### 2) 上传文章

- `PUT /api/article/:title`
- JSON Body 示例：

```
{
  "title": "post",
  "author": "mkitsdts",
  "category": "language",
  "update_at": "2026-01-01 12:34:56",
  "content": "正文内容..."
}
```

### 3) 上传图片

- `PUT /api/image`
- JSON Body 示例：

```
{
  "title": "post",
  "name": "cover.png",
  "data": "<base64>"
}
```

### 4) 删除文章

- `DELETE /api/article/:title`

---

## 注意事项

- `Base URL` 不能为空，否则上传/刷新/删除会失败。
- 文档标题默认使用文件名（不含 `.md`）。
- 文件夹批量上传会递归扫描子目录中的 `.md` 文件。
- 单篇文章上传失败时会给出错误信息；图片上传失败会提示具体文件名，便于重试。

---

## 故障排查

1. **看不到文章列表**
   - 检查 `Base URL` 是否可访问
   - 检查后端接口是否已启动
2. **上传成功但图片缺失**
   - 确认图片位于“同名文件夹”中
   - 确认图片扩展名在支持列表内
3. **删除失败**
   - 检查标题是否与后端记录一致（URL 编码由插件处理）
   - 检查认证 Token 是否有效
4. **每次启动都提示插件/视图加载失败，偶尔 Obsidian 打不开、需要安全模式启动**
   - 症状来源：`workspace.json` 中出现了 ghost 视图：

     ```json
     { "type": "leaf",
       "state": { "type": "mkblog-articles-view", "state": {},
                  "icon": "lucide-ghost", "title": "mkblog-articles-view" } }
     ```

     说明恢复布局时 `mkblog-articles-view` 这个 view type 尚未注册，
     Obsidian 只能创建无法解析的占位视图，并把该状态反复写回磁盘。
   - 0.0.2 已修复：`registerView()` 现在是 `onload()` 的第一条语句；
     `onload()` 内不再做网络请求或工作区修改（改为在
     `workspace.onLayoutReady()` 之后执行，并带超时）。
   - 升级后重启 Obsidian 一次即可：ghost 叶子会被真实视图替换，
     并重新写入正确的 `workspace.json`。
   - 如需手动清理：完全退出 Obsidian 后备份并编辑
     `<Vault>/.obsidian/workspace.json`，删除上述 leaf 节点。
5. **确认插件是否拖慢了启动**
   - `设置 -> 通用 -> 高级 -> 启动耗时调试` 查看各插件耗时。

---

## 版本建议

- Obsidian：建议使用较新桌面版本（支持社区插件 API）
- Node.js（开发态）：建议 `>=18`
- TypeScript（开发态）：建议与 Obsidian 官方插件模板保持一致
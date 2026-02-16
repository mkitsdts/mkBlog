# mkBlog Obsidian Uploader

将 Obsidian Vault 中的 Markdown 与同名图片文件夹一起上传到 mkBlog 后端，并在侧边栏展示文章列表，支持删除远端文章。

## 功能

- 解析 Markdown 头部元数据（Frontmatter）中的 `author` / `category`。
- 单独上传当前 Markdown 文件（并自动上传同名图片文件夹中的图片）。
- 单独上传指定文件夹中的所有 Markdown（并自动上传各自同名图片文件夹中的图片）。
- 激活后自动拉取文章列表，在侧边栏管理视图中展示。
- 支持在管理视图中删除远端文章。

---

## 安装（开发态）

1. 将本插件目录放到你的 Vault 插件目录下，例如：

```mkBlog/plugin/obsidian/README.md#L1-4
<Vault>/.obsidian/plugins/mkblog-obsidian/
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

---

## Markdown 元数据解析规则

插件会优先读取文档开头的 Frontmatter：

```mkBlog/plugin/obsidian/README.md#L1-7
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

```mkBlog/plugin/obsidian/README.md#L1-6
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

```mkBlog/plugin/obsidian/README.md#L1-8
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

```mkBlog/plugin/obsidian/README.md#L1-6
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

---

## 版本建议

- Obsidian：建议使用较新桌面版本（支持社区插件 API）
- Node.js（开发态）：建议 `>=18`
- TypeScript（开发态）：建议与 Obsidian 官方插件模板保持一致
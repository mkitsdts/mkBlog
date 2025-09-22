# mkBlog Uploader

将文件夹中的 Markdown 与同名图片文件夹一起上传到后端，并在侧边栏显示文章列表，支持删除。

## 功能
- 扫描所选文件夹的 `.md` 文件，匹配同名文件夹中的图片，一起上传（multipart/form-data）。
- 激活后自动拉取文章列表，显示在 Activity Bar 的 `mkBlog` 视图中。
- 右键或按钮删除文章。

## 扩展设置
- `mkBlog.uploadUrl`: 上传接口 URL（PUT, multipart/form-data）
- `mkBlog.listUrl`: 文章列表接口 URL（GET 返回数组或 `{ data: [] }`）
- `mkBlog.deleteUrl`: 删除接口 URL，使用 `{id}` 占位符
- `mkBlog.authToken`: 可选 Bearer Token

## 命令
- `mkBlog: 上传文件夹为博客`
- `mkBlog: 刷新文章列表`
- `mkBlog: 删除文章`

## 开发

```bash
npm install
npm run build
```

在 VS Code 中按 `F5` 运行扩展开发宿主。
# mkBlog Uploader

将文件夹中的 Markdown 与同名图片文件夹一起上传到后端，并在侧边栏显示文章列表，支持删除。

## 功能
- 扫描所选文件夹的 `.md` 文件，匹配同名文件夹中的图片，一起上传（multipart/form-data）。
- 激活后自动拉取文章列表，显示在 Activity Bar 的 `mkBlog` 视图中。
- 右键或按钮删除文章。

## 扩展设置
- `mkBlog.baseUrl`: 后端服务基础地址（例如 `http://localhost:8080`）。各接口会自动拼接 `/api/articles`、`/api/article/:title`、`/api/image` 等路径。
- `mkBlog.defaultAuthor`: 若 Markdown 中未声明作者时使用的默认作者。
- `mkBlog.defaultCategory`: 若 Markdown 中未声明分类时使用的默认分类。
- `mkBlog.authToken`: 可选 Bearer Token，会通过 `Authorization: Bearer <token>` 头发送。

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
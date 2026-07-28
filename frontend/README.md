# .

This template should help get you started developing with Vue 3 in Vite.

## Recommended IDE Setup

[VSCode](https://code.visualstudio.com/) + [Volar](https://marketplace.visualstudio.com/items?itemName=Vue.volar) (and disable Vetur).

## Type Support for `.vue` Imports in TS

TypeScript cannot handle type information for `.vue` imports by default, so we replace the `tsc` CLI with `vue-tsc` for type checking. In editors, we need [Volar](https://marketplace.visualstudio.com/items?itemName=Vue.volar) to make the TypeScript language service aware of `.vue` types.

## Customize configuration

See [Vite Configuration Reference](https://vite.dev/config/).

## Project Setup

```sh
npm install
```

### Compile and Hot-Reload for Development

```sh
npm run dev
```

前端运行时通过同源的 `/api/site` 获取站点配置，其他接口也统一使用同源 `/api` 前缀。

### Type-Check, Compile and Minify for Production

```sh
npm run build
```

`config.yaml` 样例:

```yaml
site:
  signature: 你的签名
  avatarPath: avatar.jpg
```

`avatarPath` 相对于后端的 `data/` 目录。程序首次启动时会生成
`data/avatar.jpg`，可直接替换该文件，或修改 `avatarPath` 切换头像。
修改 `config.yaml` 后重新部署容器即可生效（无需重新前端构建，除非新增字段）。

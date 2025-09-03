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

前端运行时通过请求 `/config.yaml`（后端静态暴露）获取站点配置，不再构建期生成。

### Type-Check, Compile and Minify for Production

```sh
npm run build
```

`config.yaml` 样例:

```yaml
site:
	signature: 你的签名
	avatarPath: avatar.jpg
	server: https://example.com/api
```

修改 `config.yaml` 后重新部署容器即可生效（无需重新前端构建，除非新增字段）。

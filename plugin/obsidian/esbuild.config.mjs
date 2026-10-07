import esbuild from "esbuild";
import process from "process";
import { copyFileSync, mkdirSync } from "node:fs";

const isProd = process.argv.includes("--prod");
const isWatch = process.argv.includes("--watch");

const context = await esbuild.context({
  entryPoints: ["src/main.ts"],
  bundle: true,
  format: "cjs", // 必须是 cjs
  platform: "browser",
  target: "es2018",
  outfile: "build/main.js", // 输出到 build/main.js
  sourcemap: isProd ? false : "inline",
  minify: isProd,
  legalComments: "none",
  logLevel: "info",
  external: ["obsidian", "electron", "@codemirror/*"],
});

// Obsidian 只认 manifest.json（缺少或文件名写错都会导致插件无法加载)
function copyManifest() {
  mkdirSync("build", { recursive: true });
  copyFileSync("manifest.json", "build/manifest.json");
}

if (isWatch) {
  await context.watch();
  copyManifest();
  console.log("[esbuild] watching...");
} else {
  await context.rebuild();
  copyManifest();
  await context.dispose();
  console.log("[esbuild] build complete");
}

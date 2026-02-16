import esbuild from "esbuild";
import process from "process";

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

if (isWatch) {
  await context.watch();
  console.log("[esbuild] watching...");
} else {
  await context.rebuild();
  await context.dispose();
  console.log("[esbuild] build complete");
}

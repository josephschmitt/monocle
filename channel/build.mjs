import { build } from "esbuild";

await build({
  entryPoints: ["channel.ts"],
  bundle: true,
  minify: true,
  platform: "node",
  format: "esm",
  target: "node18",
  outfile: "../internal/adapters/channel.bundle.js",
  external: ["net", "crypto", "fs", "path", "process"],
});

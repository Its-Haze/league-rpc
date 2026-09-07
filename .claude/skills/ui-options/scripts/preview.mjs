// Builds the option preview with the project's own Vite + Tailwind toolchain so
// what you judge is what the app renders, then serves it and waits for a pick.
import { createRequire } from "node:module";
import { spawn } from "node:child_process";
import { createServer } from "node:http";
import { readFileSync, writeFileSync, mkdirSync, copyFileSync, existsSync, rmSync } from "node:fs";
import { dirname, extname, join, resolve } from "node:path";
import { fileURLToPath, pathToFileURL } from "node:url";

const here = dirname(fileURLToPath(import.meta.url));
const assets = join(here, "..", "assets");
const repo = resolve(here, "..", "..", "..", "..");
const frontend = join(repo, "frontend");
const dir = join(frontend, ".preview");

const args = Object.fromEntries(
  process.argv.slice(2).reduce((acc, arg, i, all) => {
    if (arg.startsWith("--")) acc.push([arg.slice(2), all[i + 1]]);
    return acc;
  }, []),
);
const title = args.title ?? "UI options";
const port = Number(args.port ?? 5199);
const optionsFile = resolve(args.options ?? join(dir, "options.html"));

if (!existsSync(optionsFile)) {
  console.error(`No options fragment at ${optionsFile}`);
  process.exit(1);
}

// Stage: everything the build needs, assembled fresh each run.
mkdirSync(dir, { recursive: true });
rmSync(join(dir, "dist"), { recursive: true, force: true });
rmSync(join(dir, "choice.json"), { force: true });
for (const name of ["preview.css", "shell.css", "shell.js"]) {
  copyFileSync(join(assets, name), join(dir, name));
}
const { execFileSync } = await import("node:child_process");
execFileSync(process.execPath, [
  join(here, "scope_tokens.mjs"),
  join(frontend, "src", "tokens.css"),
  join(dir, "scoped-tokens.css"),
]);
writeFileSync(
  join(dir, "index.html"),
  readFileSync(join(assets, "index.template.html"), "utf8")
    .replaceAll("__TITLE__", title)
    .replace("__OPTIONS__", readFileSync(optionsFile, "utf8")),
);

const require = createRequire(pathToFileURL(join(frontend, "package.json")));
const vite = await import(pathToFileURL(require.resolve("vite")).href);
const tailwind = await import(pathToFileURL(require.resolve("@tailwindcss/vite")).href);

await vite.build({
  root: dir,
  base: "./",
  logLevel: "warn",
  plugins: [tailwind.default()],
  build: { outDir: "dist", emptyOutDir: true },
});

if (args["build-only"] !== undefined || process.argv.includes("--build-only")) {
  console.log("build ok");
  process.exit(0);
}

const types = {
  ".html": "text/html",
  ".css": "text/css",
  ".js": "text/javascript",
  ".woff2": "font/woff2",
  ".svg": "image/svg+xml",
};
const dist = join(dir, "dist");

const server = createServer((req, res) => {
  if (req.method === "POST" && req.url === "/choice") {
    let body = "";
    req.on("data", (chunk) => (body += chunk));
    req.on("end", () => {
      writeFileSync(join(dir, "choice.json"), body);
      res.writeHead(200).end("{}");
      console.log(`choice received: ${body}`);
      res.on("finish", () => server.close(() => process.exit(0)));
    });
    return;
  }
  const path = join(dist, (req.url ?? "/").split("?")[0] === "/" ? "index.html" : decodeURIComponent(req.url.split("?")[0]));
  if (!path.startsWith(dist) || !existsSync(path)) return res.writeHead(404).end();
  res.writeHead(200, { "content-type": types[extname(path)] ?? "application/octet-stream" });
  res.end(readFileSync(path));
});

server.listen(port, () => {
  const url = `http://localhost:${port}/`;
  console.log(`preview ready at ${url}`);
  spawn("cmd", ["/c", "start", "", url], { detached: true, stdio: "ignore" }).unref();
});

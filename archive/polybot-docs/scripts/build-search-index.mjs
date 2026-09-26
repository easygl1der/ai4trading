import {readFile, readdir, writeFile} from "node:fs/promises";
import {join, relative, sep} from "node:path";

const root = new URL("..", import.meta.url).pathname;
const reportDir = join(root, "reports");
const pageFiles = ["index.mdx", "research-method.mdx", "strategy-families.mdx", "live-vs-replay.mdx", "reports.mdx"];

const clean = source => source
  .replace(/^---[\s\S]*?---\s*/m, "")
  .replace(/```[\s\S]*?```/g, " ")
  .replace(/!\[[^\]]*\]\([^)]*\)/g, " ")
  .replace(/\[([^\]]+)\]\([^)]*\)/g, "$1")
  .replace(/[`*_>#|]/g, " ")
  .replace(/\$+/g, " ")
  .replace(/\s+/g, " ")
  .trim();

const titleOf = source => source.match(/^#\s+(.+)$/m)?.[1] || source.match(/^title:\s*(.+)$/m)?.[1] || "Untitled";
const pathOf = filename => filename === "index.mdx" ? "/" : `/${filename.replace(/\.(mdx|md)$/, "").replaceAll(sep, "/")}`;
const files = [
  ...pageFiles.map(file => join(root, file)),
  ...(await readdir(reportDir)).filter(file => file.endsWith(".md")).map(file => join(reportDir, file)),
];

const entries = await Promise.all(files.map(async file => {
  const source = await readFile(file, "utf8");
  const filename = relative(root, file);
  const body = clean(source);
  return {title: titleOf(source), path: pathOf(filename), excerpt: body.slice(0, 340), text: body};
}));

await writeFile(join(root, "search-index.json"), JSON.stringify(entries), "utf8");
console.log(`Indexed ${entries.length} Mintlify pages.`);

#!/usr/bin/env node

const fs = require("fs");
const path = require("path");
const { spawnSync } = require("child_process");

const exe = process.platform === "win32" ? "nextvibe.exe" : "nextvibe";
const candidates = [
  path.join(__dirname, "bin", exe),
  path.join(__dirname, "..", exe)
];

const binary = candidates.find((candidate) => fs.existsSync(candidate));
if (!binary) {
  console.error("nextvibe binary was not found. Try reinstalling the package or run `go build -o nextvibe ./cmd/nextvibe`.");
  process.exit(1);
}

const result = spawnSync(binary, process.argv.slice(2), { stdio: "inherit" });
if (result.error) {
  console.error(result.error.message);
  process.exit(1);
}
process.exit(result.status === null ? 1 : result.status);

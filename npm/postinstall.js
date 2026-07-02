#!/usr/bin/env node

const fs = require("fs");
const path = require("path");
const { spawnSync } = require("child_process");

const root = path.resolve(__dirname, "..");
const binDir = path.join(__dirname, "bin");
const exe = process.platform === "win32" ? "nextvibe.exe" : "nextvibe";
const output = path.join(binDir, exe);

fs.mkdirSync(binDir, { recursive: true });

const result = spawnSync("go", ["build", "-trimpath", "-o", output, "./cmd/nextvibe"], {
  cwd: root,
  stdio: "inherit"
});

if (result.error) {
  console.error("NextVibe npm install requires Go 1.22 or newer on PATH.");
  console.error(result.error.message);
  process.exit(1);
}

if (result.status !== 0) {
  process.exit(result.status || 1);
}

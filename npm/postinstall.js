#!/usr/bin/env node

const fs = require("fs");
const https = require("https");
const os = require("os");
const path = require("path");
const { spawnSync } = require("child_process");

const pkg = require("../package.json");
const binDir = path.join(__dirname, "bin");
const exe = process.platform === "win32" ? "nextvibe.exe" : "nextvibe";
const output = path.join(binDir, exe);
const installBinary = process.env.NEXTVIBE_INSTALL_BINARY;

fs.mkdirSync(binDir, { recursive: true });

main().catch((error) => {
  console.error("Failed to install the NextVibe binary.");
  console.error(error.message);
  console.error("");
  console.error("Make sure the matching GitHub Release asset exists before publishing this npm version.");
  console.error("You can also set NEXTVIBE_INSTALL_BINARY to a local nextvibe binary for local package tests.");
  process.exit(1);
});

async function main() {
  if (installBinary) {
    installFromFile(path.resolve(installBinary));
    console.log(`Installed NextVibe binary from ${installBinary}`);
    return;
  }

  const target = resolveTarget();
  const tempDir = fs.mkdtempSync(path.join(os.tmpdir(), "nextvibe-install-"));

  try {
    const archivePath = path.join(tempDir, target.asset);
    const extractDir = path.join(tempDir, "extract");
    fs.mkdirSync(extractDir, { recursive: true });

    console.log(`Downloading NextVibe ${target.tag} for ${target.goos}/${target.goarch}...`);
    await downloadFile(target.url, archivePath);
    extractArchive(archivePath, extractDir, target.format);

    const binary = findBinary(extractDir, exe);
    if (!binary) {
      throw new Error(`Archive did not contain ${exe}.`);
    }

    installFromFile(binary);
    console.log(`Installed NextVibe ${target.tag} to ${output}`);
  } finally {
    fs.rmSync(tempDir, { recursive: true, force: true });
  }
}

function resolveTarget() {
  const goosByPlatform = {
    darwin: "darwin",
    linux: "linux",
    win32: "windows"
  };
  const goarchByArch = {
    arm64: "arm64",
    x64: "amd64"
  };

  const goos = goosByPlatform[process.platform];
  const goarch = goarchByArch[process.arch];
  if (!goos || !goarch) {
    throw new Error(`Unsupported platform: ${process.platform}/${process.arch}.`);
  }

  const format = goos === "windows" ? "zip" : "tar.gz";
  const asset = `nextvibe_${goos}_${goarch}.${format}`;
  const tag = process.env.NEXTVIBE_RELEASE_TAG || `v${pkg.version}`;
  const repo = process.env.NEXTVIBE_GITHUB_REPOSITORY || repositorySlug(pkg) || "CrescentR/NextVibe";
  const baseUrl = process.env.NEXTVIBE_DOWNLOAD_BASE_URL ||
    `https://github.com/${repo}/releases/download/${tag}`;
  const url = process.env.NEXTVIBE_DOWNLOAD_URL || joinUrl(baseUrl, asset);

  return { asset, format, goarch, goos, tag, url };
}

function repositorySlug(packageJson) {
  const raw = packageJson.repository && packageJson.repository.url
    ? packageJson.repository.url
    : packageJson.homepage;
  if (!raw) {
    return "";
  }

  const clean = String(raw).replace(/^git\+/, "");
  const match = clean.match(/github\.com[:/]([^/\s]+)\/([^/\s#?]+)/);
  if (!match) {
    return "";
  }

  return `${match[1]}/${match[2].replace(/\.git$/, "")}`;
}

function joinUrl(baseUrl, fileName) {
  return `${baseUrl.replace(/\/+$/, "")}/${fileName}`;
}

function downloadFile(url, destination, redirectCount = 0) {
  if (redirectCount > 5) {
    return Promise.reject(new Error(`Too many redirects while downloading ${url}.`));
  }

  return new Promise((resolve, reject) => {
    const request = https.get(url, {
      headers: {
        "User-Agent": `nextvibe-npm-installer/${pkg.version}`
      }
    }, (response) => {
      const status = response.statusCode || 0;
      const location = response.headers.location;

      if (status >= 300 && status < 400 && location) {
        response.resume();
        const nextUrl = new URL(location, url).toString();
        downloadFile(nextUrl, destination, redirectCount + 1).then(resolve, reject);
        return;
      }

      if (status !== 200) {
        response.resume();
        reject(new Error(`Download failed with HTTP ${status}: ${url}`));
        return;
      }

      const file = fs.createWriteStream(destination, { mode: 0o644 });
      response.pipe(file);
      file.on("finish", () => file.close(resolve));
      file.on("error", (error) => {
        fs.rmSync(destination, { force: true });
        reject(error);
      });
    });

    request.on("error", reject);
  });
}

function extractArchive(archivePath, extractDir, format) {
  const expandArchiveCommand = `Expand-Archive -LiteralPath ${quotePowerShell(archivePath)} -DestinationPath ${quotePowerShell(extractDir)} -Force`;
  const attempts = format === "zip"
    ? [
      {
        command: "powershell",
        args: [
          "-NoProfile",
          "-ExecutionPolicy",
          "Bypass",
          "-Command",
          expandArchiveCommand
        ]
      },
      {
        command: "pwsh",
        args: [
          "-NoProfile",
          "-Command",
          expandArchiveCommand
        ]
      },
      { command: "tar", args: ["-xf", archivePath, "-C", extractDir] }
    ]
    : [
      { command: "tar", args: ["-xzf", archivePath, "-C", extractDir] }
    ];

  let missingTool = "";
  for (const attempt of attempts) {
    const result = spawnSync(attempt.command, attempt.args, { stdio: "inherit" });
    if (result.error && result.error.code === "ENOENT") {
      missingTool = result.error.message;
      continue;
    }
    if (result.error) {
      throw result.error;
    }
    if (result.status !== 0) {
      throw new Error(`${attempt.command} exited with status ${result.status}.`);
    }
    return;
  }

  throw new Error(`Could not extract ${archivePath}. ${missingTool}`);
}

function quotePowerShell(value) {
  return `'${String(value).replace(/'/g, "''")}'`;
}

function findBinary(root, fileName) {
  const entries = fs.readdirSync(root, { withFileTypes: true });
  for (const entry of entries) {
    const candidate = path.join(root, entry.name);
    if (entry.isDirectory()) {
      const nested = findBinary(candidate, fileName);
      if (nested) {
        return nested;
      }
      continue;
    }
    if (entry.isFile() && entry.name === fileName) {
      return candidate;
    }
  }
  return "";
}

function installFromFile(source) {
  if (!fs.existsSync(source)) {
    throw new Error(`Binary not found: ${source}`);
  }

  fs.copyFileSync(source, output);
  if (process.platform !== "win32") {
    fs.chmodSync(output, 0o755);
  }
}

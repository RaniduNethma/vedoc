const fs = require("fs");
const path = require("path");
const https = require("https");
const os = require("os");

const VERSION = "v1.0.1"; // GitHub Release Version
const REPO = "RaniduNethma/vedoc";

const type = os.type();
const arch = os.arch();

let osName = "";
let ext = "";

if (type === "Windows_NT") {
  osName = "windows";
  ext = ".exe";
} else if (type === "Darwin") {
  osName = "darwin";
} else if (type === "Linux") {
  osName = "linux";
} else {
  console.error(`OS unsupported: ${type}`);
  process.exit(1);
}

let archName = arch === "arm64" ? "arm64" : "amd64";
const binName = `vedoc-${osName}-${archName}${ext}`;
const url = `https://github.com/${REPO}/releases/download/${VERSION}/${binName}`;
const dest = path.join(__dirname, `vedoc${ext}`);

console.log(`Downloading Vedoc for ${osName} ${archName}...`);

function download(url, dest, cb) {
  https
    .get(url, (res) => {
      // Handle GitHub 302 Redirects
      if (res.statusCode === 301 || res.statusCode === 302) {
        return download(res.headers.location, dest, cb);
      }

      if (res.statusCode !== 200) {
        return cb(
          new Error(
            `Failed to download: ${res.statusCode} ${res.statusMessage}`,
          ),
        );
      }

      const file = fs.createWriteStream(dest);
      res.pipe(file);

      file.on("finish", () => {
        file.close(cb);
      });
    })
    .on("error", (err) => {
      fs.unlink(dest, () => {});
      cb(err);
    });
}

download(url, dest, (err) => {
  if (err) {
    console.error("Error downloading Vedoc binary:", err.message);
    process.exit(1);
  }

  // Mac/Linux Execute Permissions
  if (osName !== "windows") {
    fs.chmodSync(dest, 0o755);
  }

  console.log(
    'Vedoc installed successfully! Run "vedoc --help" to get started.',
  );
});

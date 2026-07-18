#!/usr/bin/env node
const { spawnSync } = require("child_process");
const path = require("path");
const os = require("os");
const fs = require("fs");

const ext = os.type() === "Windows_NT" ? ".exe" : "";

// Find downloaded vedoc binaries
const binPath = path.join(__dirname, "..", `vedoc${ext}`);

if (!fs.existsSync(binPath)) {
  console.error("Vedoc binary not found. Please try reinstalling the package.");
  process.exit(1);
}

// Passing user commands
const args = process.argv.slice(2);
const child = spawnSync(binPath, args, { stdio: "inherit" });

if (child.error) {
  console.error("Failed to execute Vedoc:", child.error.message);
  process.exit(1);
}

process.exit(child.status);

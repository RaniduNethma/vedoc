#!/usr/bin/env node
const { spawnSync, execSync } = require("child_process");
const path = require("path");
const os = require("os");
const fs = require("fs");

const ext = os.type() === "Windows_NT" ? ".exe" : "";

// Find downloaded vedoc binaries
const binPath = path.join(__dirname, "..", `vedoc${ext}`);

// First Run Download Logic
if (!fs.existsSync(binPath)) {
  console.log("Initializing Vedoc for the first time...");
  try {
    const installScript = path.join(__dirname, "..", "install.js");
    execSync(`node "${installScript}"`, { stdio: "inherit" });
  } catch (error) {
    console.error(
      "Failed to download Vedoc core. Please check your internet connection.",
    );
    process.exit(1);
  }
}

// run user command
const args = process.argv.slice(2);
const child = spawnSync(binPath, args, { stdio: "inherit" });

if (child.error) {
  console.error("Failed to execute Vedoc:", child.error.message);
  process.exit(1);
}

process.exit(child.status);

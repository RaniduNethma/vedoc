# Vedoc

<img width="2188" height="516" alt="Image" src="https://github.com/user-attachments/assets/48a6d69d-2584-4301-8c82-81185c3dffe1" />

![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=for-the-badge&logo=go)
![License](https://img.shields.io/badge/License-MIT-green?style=for-the-badge)

**Vedoc** is a lightning fast, AI powered CLI tool that automatically generates intelligent API documentation directly from your source code.

Instead of relying on fragile regex patterns, Vedoc uses **Tree-sitter** to build an Abstract Syntax Tree (AST) of your codebase to accurately extract Express.js/Node.js API routes. It then leverages **Google Gemini AI** to understand your code, generate descriptive summaries, and infer valid JSON payloads, ultimately exporting a ready to use **Postman Collection**.

---

## Key Features

- **True Code Understanding:** Uses `Tree-sitter` to parse your codebase via AST, ensuring 100% accuracy in finding endpoints (ignores commented code, handles nested routers).
- **AI-Enriched Documentation:** Integrates with the `gemini-3.5-flash` model to automatically generate human-readable descriptions and context aware JSON payloads for your routes.
- **Blazing Fast Directory Scanning:** Intelligently and recursively scans your project (skipping `node modules` and hidden folders) to find all `.js` and `.ts` files.
- **Seamless Postman Export:** Generates a fully compliant Postman v2.1.0 collection (`vedoc_postman_collection.json`) that you can import with a single click.
- **Graceful Degradation:** Don't have an AI key? No problem! Vedoc will still accurately parse your routes and generate a basic Postman collection offline.

---

## Prerequisites

Before installing Vedoc, make sure you have the following installed on your system:

- **Go** (version 1.21 or higher)
- A **Google Gemini API Key** (Get one for free at [Google AI Studio](https://aistudio.google.com/))

---

## Installation

Since Vedoc is built with Go, you can easily install it globally on your machine using `go install`:

```bash
go install github.com/RaniduNethma/vedoc@latest
```

---

## 🛠️ Getting Started

### 1. Configuration (Set API Key)

To unlock the full power of Vedoc, configure it with your Google Gemini API Key:

```bash
vedoc config set-key "your_gemini_api_key_here"
```

Your key is securely stored locally in `~/.vedoc.yml`.

### 2. Generate Documentation

Navigate to your Node.js/Express.js project directory and run the generate command. Vedoc will scan your files and analyze the code:

```bash
cd /path/to/your/express/project
vedoc generate
```

### 3. Import to Postman

Once the process is complete, a file named `vedoc_postman_collection.json` will be created in your current directory. Simply open Postman, click **Import**, and drag & drop this file!

---

## How it Works Under the Hood

1. **AST Parsing:** Vedoc recursively scans your directory for Javascript/Typescript files. It passes the source code to `go-tree-sitter` with a custom grammar query to identify valid HTTP methods (`GET`, `POST`, `PUT`, `DELETE`, `PATCH`) and their respective paths.
2. **AI Enrichment:** The extracted code snippet for each route is sent to Gemini AI using a highly constrained prompt to return a strict JSON structure containing a description and a synthesized payload.
3. **Generation:** The parsed endpoints and AI generated metadata are mapped into the official Postman Collection schema and written to disk.

---

## Contributing

Contributions, issues, and feature requests are always welcome!

If you want to add support for other frameworks (like Fastify, Gin, or Spring Boot) or output formats (like OpenAPI/Swagger), feel free to fork the repository and submit a pull request.

1. Fork the Project
2. Create your Feature Branch (`git checkout -b feature/AmazingFeature`)
3. Commit your Changes (`git commit -m 'feat: Add some AmazingFeature'`)
4. Push to the Branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

---

## License

Distributed under the MIT License. See the `LICENSE` file for more information.

---

Built with ❤️ by [Ranidu Nethma](https://github.com/RaniduNethma)

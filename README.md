# <img src="assets/icon.svg" alt="Link Router icon" width="32" height="32"> Link Router

[![Go Version](https://img.shields.io/github/goversion/mzieba/link-router?style=flat-square)](https://golang.org/)
[![License](https://img.shields.io/badge/license-MIT-blue.svg?style=flat-square)](LICENSE)
[![CI](https://img.shields.io/github/actions/workflow/status/mzieba/link-router/ci.yml?branch=main&style=flat-square&label=ci)](https://github.com/mzieba/link-router/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/mzieba/link-router?style=flat-square)](https://github.com/mzieba/link-router/releases/latest)

A fast, lightweight, cross-platform URL router written in Go. Register **Link Router** as your default system browser, and it automatically forwards incoming links to specific browsers or browser profiles based on rules you define.

🏠 More about the project: https://michalzieba.pl/playground/link-router

📄 A simple PRD: [PRD.md](PRD.md)

---

## 📌 Table of Contents

* [Features](#-features)
* [Supported Browsers](#-supported-browsers)
* [Installation](#-installation)
* [Usage](#-usage)
* [Configuration](#-configuration)
  * [Glob Matching (`match_host`)](#match_host)
  * [Regex Matching (`match_url`)](#match_url)
  * [Combining Matchers](#combining-matchers)
* [Platform Notes](#-platform-notes)
* [Troubleshooting](#-troubleshooting)
* [Contributing](#-contributing)
* [AI assistance](#-ai-assistance)
* [License](#-license)

---

## ✨ Features

* **Rule-Based Routing:** Match URLs using host globs, full-URL regex patterns, or both.
* **Multi-Profile Support:** Direct links to different browser profiles (e.g., Work vs. Personal).
* **Cross-Platform:** Native support for Linux and Windows.
* **Dry-Run Testing:** Built-in command to test rules before activating them.
* **Smart Browser Detection:** Auto-detects top desktop browsers from `PATH` (Linux) or System Registry (Windows).

---

## 🌐 Supported Browsers

Auto-detection scans for popular desktop browsers, including:

| Linux & Windows | Windows Registry Extended |
| :--- | :--- |
| Google Chrome, Mozilla Firefox, Microsoft Edge, Brave, Vivaldi | Any custom browser registered under `StartMenuInternet` or `App Paths` |
| Opera, Chromium, Yandex Browser, LibreWolf, Tor Browser | |

> **Note:** macOS / Safari are currently not targeted.

---

## 📦 Installation

### Pre-built binaries
Download the binary for your platform from the [Releases](https://github.com/mzieba/link-router/releases) page:
`link-router-linux-amd64`, `link-router-linux-arm64`, or `link-router-windows-amd64.exe`.

Each release also ships a `SHA256SUMS` file. Verify your download before running it:

```bash
sha256sum --check --ignore-missing SHA256SUMS
```

> **Windows:** the `.exe` is not code-signed yet, so SmartScreen shows a
> "Windows protected your PC" warning on first run. Check the SHA-256 sum above,
> then choose **More info > Run anyway**. Code signing via the
> [SignPath Foundation](https://signpath.org) open-source program is planned.

### From Go
```bash
go install github.com/mzieba/link-router@latest
```

### Build from Source
```bash
# Clone the repository
git clone https://github.com/mzieba/link-router.git
cd link-router

# Build using Make
make build            # Native binary (CGO disabled)
make build-windows    # Cross-compile for Windows (.exe)
make dist             # All release artifacts + SHA256SUMS in dist/
make test             # Run test suite
```

Without `make`:
```bash
# Native
CGO_ENABLED=0 go build -o link-router .

# Windows Cross-Compilation
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o link-router.exe .
```

---

## 🚀 Usage

```bash
link-router init         # Create a starter configuration file
link-router register     # Register as a system default browser candidate
link-router unregister   # Remove system registration
link-router browsers     # List all auto-detected browsers
link-router edit         # Open the config in $VISUAL or $EDITOR (prints the path if neither is set)
link-router version      # Print the application version
link-router list         # List configured rules and default targets
link-router open <url>   # Open a URL using matching rules (used by OS)
link-router test <url>   # Dry-run: show matched rule & target command without launching
```

> **Tip:** Always run `link-router test <url>` after updating your rules to verify target resolution.

---

## ⚙️ Configuration

The config file is located at:
* **Linux:** `~/.config/link-router/config.toml`
* **Windows:** `%APPDATA%\link-router\config.toml`

Run `link-router init` to generate a starter config:

```toml
default = "personal" # Fallback target when no rule matches

[[rules]]
match_host = ["*.my-company.com", "my-company.com", "jira.*"] # Glob matching on host
target = "work"

[[rules]]
match_url = '^https://github\.com/my-company' # Regex matching on full URL
target = "work"

[targets.work]
command = "google-chrome"
args = ["--profile-directory=Work", "{url}"]

[targets.personal]
command = "firefox"
args = ["-P", "personal", "{url}"]
```

### Rule Evaluation Logic
* Rules are evaluated **top to bottom**; the first matching rule wins.
* `{url}` in target `args` is replaced with the incoming link. If `{url}` is omitted, the link is appended to the end of the argument list.

### `match_host`
Glob patterns matched strictly against the URL host (scheme, port, path, and query are stripped).
* Uses shell-style globs: `*` (any sequence), `?` (single char), `[abc]` (char set).
* **OR logic:** If any pattern in the list matches, the host condition passes.
* Bare apex domains are not matched by wildcard prefixes. To match both `app.domain.com` and `domain.com`, specify both:
  ```toml
  match_host = ["my-company.com", "*.my-company.com"]
  ```

### `match_url`
A single regular expression (Go RE2 syntax) tested against the entire raw URL.
* **Unanchored:** Use `^` and `$` to enforce exact matching or prefixes.
* **TOML Escaping:** Use single-quoted literal strings (`'...'`) to avoid escaping backslashes, or escape double backslashes in double-quoted strings (`"\\."`).
* **Case Sensitivity:** Prefix patterns with `(?i)` for case-insensitive matching.

### Combining Matchers
When both `match_host` and `match_url` are present in a single rule, they are **ANDed** together:

```toml
[[rules]]
match_host = ["github.com"]
match_url = "/my-company/"
target = "work"
```

---

## 💻 Platform Notes

* **Linux:** `link-router register` creates `~/.local/share/applications/link-router.desktop` and invokes `xdg-settings` / `xdg-mime` to set default handlers.
* **Windows:** `link-router register` writes candidate registry keys. On Windows 10/11, navigate to **Settings > Default Apps** and select **Link Router** manually to complete set up.

---

## 🔍 Troubleshooting

Logs for execution and launch failures are written to `link-router.log` in the same directory as your configuration file. Failure entries can include the full URL that was being opened, so treat that file like browsing history. Link Router makes no network requests of any kind; see [PRIVACY.md](PRIVACY.md).

---

## 🤝 Contributing

Contributions, issues, and feature requests are welcome. See
[CONTRIBUTING.md](CONTRIBUTING.md) for the development setup, the checks CI runs,
and the commit message convention (the project releases from Conventional
Commits, so the prefix matters).

Found a security issue? Please follow [SECURITY.md](SECURITY.md) instead of
opening a public issue.

## 🤖 AI assistance

For transparency: most of the code in this repository was generated by Claude.
I defined the architecture and requirements, reviewed all generated code
and tests.\
Bugs... bugs are probably mine ;)

---

## 📄 License

Distributed under the MIT License. See [LICENSE](LICENSE) for more information.
Privacy: [PRIVACY.md](PRIVACY.md). Security reports: [SECURITY.md](SECURITY.md).
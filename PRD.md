# 📄 Product Requirement Document (PRD)

## 1. Product Overview

* **Product Name:** Link Router (`link-router`)
* **Status:** Released; see [Releases](https://github.com/mzieba/link-router/releases) for the current version.
* **Owner:** `@mzieba`

### 1.1 Objective

Provide a lightweight, cross-platform CLI utility that acts as a system-level default browser proxy, automatically dispatching URLs to specific browser binaries and profiles based on customizable domain and regex rules.

---

## 2. User Personas & Pain Points

| Persona | Key Need | Pain Point |
| --- | --- | --- |
| **Multi-Client Freelancer** | Wants client links to open in client-isolated Chrome profiles automatically. | Link clicks from Slack open in whichever profile was focused last, breaking auth sessions. |
| **DevOps / Engineer** | Needs internal dev dashboards (`*.internal.net`) opened in Firefox, but marketing sites in Chrome. | Manual URL copy-pasting between browsers wastes time and breaks workflow focus. |
| **Privacy-Conscious User** | Wants social media and untrusted links routed to isolated/ephemeral browser profiles. | OS lacks granular browser routing logic out of the box. |

---

## 3. Functional Requirements

### 3.1 OS Integration & Handler Registration

* **FR-1.1:** Must register as an OS default browser candidate on **Linux** (`.desktop` file + `xdg-settings` / `xdg-mime`).
* **FR-1.2:** Must write required Registry keys on **Windows** (`StartMenuInternet` / `App Paths`) to appear in system "Default Apps" menus.
* **FR-1.3:** Provide an unregister command to cleanly remove registry/desktop entries upon uninstallation.

### 3.2 Detection Engine

* **FR-2.1:** Automatically scan system `PATH` (Linux) and Windows Registry to identify these 10 popular browsers: Chrome, Firefox, Edge, Brave, Vivaldi, Opera, Chromium, Yandex, LibreWolf, Tor.
* **FR-2.2:** Support execution of arbitrary custom browser paths and arguments.

### 3.3 Rule Matching Engine

* **FR-3.1 (Host Globs):** Support shell-style wildcard matching (`*`, `?`, `[]`) against URL hostnames.
* **FR-3.2 (Regex Matching):** Support Go RE2 unanchored regex matching across full raw URLs.
* **FR-3.3 (Logical Operators):**
  * Multiple items within a single matcher (`match_host = [...]`) evaluate as **OR**.
  * Combining `match_host` and `match_url` within one rule evaluates as **AND**.
* **FR-3.4 (Fallback Execution):** Route to a user-defined `default` target if no rules match.

### 3.4 CLI & Developer Experience

* **FR-4.1 (`open <url>`):** Target endpoint executed by the OS default handler.
* **FR-4.2 (`test <url>`):** Dry-run mode printing matched rule, target profile, and exact constructed command without launching the process.
* **FR-4.3 (`init` / `edit`):** Command to generate starter TOML configs and display config paths.
* **FR-4.4 (`register` / `unregister`):** Commands that apply and revert the OS integration described in FR-1.1 to FR-1.3.
* **FR-4.5 (`browsers`):** List each auto-detected browser with its name and resolved executable, deduplicated by name.
* **FR-4.6 (`list`):** Print the configured rules and the fallback default target in evaluation order.
* **FR-4.7 (`version`):** Print the application version.

---

## 4. Non-Functional Requirements

* **Performance:** Own processing and dispatch latency must remain under **15ms**, measured as wall-clock time from process start to handing the URL to the target browser process. Browser startup time itself is out of scope. Verify with `link-router test <url>` on a warm filesystem cache.
* **Footprint:** Single static binary execution with zero runtime dependencies (CGO disabled).
* **Privacy:** 100% local operation; no network calls, external analytics, or remote tracking.
* **Reliability:** Failures to open target applications must append descriptive error context to a local `link-router.log` file without crashing silently.

---

## 5. Technical Specifications

* **Language:** Go
* **Configuration Format:** TOML (`config.toml`)
* **Supported OS:** Linux (X11/Wayland), Windows 10/11 (x86_64)

---

## 6. Future Scope (v2.0 Roadmap)

* [ ] macOS support (Apple Silicon / Intel).
* [ ] Tray app / GUI dashboard for real-time link previewing and visual rule configuration.
* [ ] Dynamic prompt popup option ("Ask every time") for unmatched URLs.

---


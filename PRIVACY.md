# Privacy policy

Last updated: 31 July 2026

Link Router is a local command-line tool. It has no server component, no
accounts, and no telemetry. Nothing it processes leaves your computer.

## What it processes

**URLs.** When your operating system hands Link Router a link, it parses the URL
in memory, matches it against your rules, and passes it to the browser you
configured. The URL is not stored anywhere, except in the log case below.

**Your configuration.** Your rules and browser commands live in a file you
control:

- Linux: `~/.config/link-router/config.toml`
- Windows: `%APPDATA%\link-router\config.toml`

**A local log file.** `link-router.log`, next to your config file, records
failures such as an invalid rule or a browser that would not start. Those
messages **can include the full URL** that was being opened, so the log may
contain browsing history in the form of failed link opens. It is a plain text
file with no size limit or automatic rotation. Delete it whenever you like; it is
recreated on the next failure.

**System registration data.** `link-router register` writes an entry that names
Link Router as a browser candidate: a `.desktop` file under
`~/.local/share/applications/` on Linux, or registry keys under
`StartMenuInternet` and `App Paths` on Windows. These hold the path to the
executable, not your browsing data. `link-router unregister` removes them.

## What it does not do

- No network requests, of any kind. Link Router does not contact the author, an
  update server, an analytics service, or any other host.
- No collection, transmission, or sale of personal data.
- No tracking identifiers, no crash reporting, no usage statistics.
- No background process. It runs when invoked and exits.

## Third parties

Link Router itself talks to nothing. Two things around it involve other parties,
and their own terms apply:

- **The browser it launches.** Once Link Router passes a URL to your browser,
  that browser handles it under its own privacy policy.
- **Downloading the software.** Release binaries are hosted on GitHub, and
  `go install` fetches through the Go module proxy. Both see your download
  request. Neither is contacted while the tool runs.

## Removing everything

```bash
link-router unregister   # remove the system registration
```

Then delete the config directory shown above (it holds the config file and the
log) and the binary itself. Nothing else is left behind.

## Contact

Questions about this policy: open an issue at
https://github.com/mzieba/link-router/issues. For security reports, follow
[SECURITY.md](SECURITY.md) instead.

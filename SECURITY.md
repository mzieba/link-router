# Security policy

## Supported versions

Only the latest release receives fixes. Please reproduce issues against the
newest [release](https://github.com/mzieba/link-router/releases) or `main`
before reporting.

## Reporting a vulnerability

Please do not open a public issue for a security problem. Report it privately
through GitHub's
[private vulnerability reporting](https://github.com/mzieba/link-router/security/advisories/new)
for this repository.

Include what you did, what happened, the platform and version, and the config
that triggers it (with any private hostnames redacted). This is a personal
project maintained in spare time, so expect an initial reply within about a
week. Fixes ship as a normal patch release with credit in the changelog unless
you prefer otherwise.

## What is in scope

Link Router installs itself as a system URL handler, so the interesting areas
are:

- **Registration** (`internal/register`): the Linux `.desktop` entry and the
  Windows `StartMenuInternet` / `App Paths` registry keys it writes, and
  anything that lets those be pointed at an unintended executable.
- **URL handling** (`internal/matcher`, `internal/opener`): a URL that causes
  arguments to be injected into the launched browser command, or that escapes
  the `{url}` substitution.
- **Config and log handling** (`internal/config`, `internal/logging`): unsafe
  file permissions or writes outside the config directory.

Out of scope: vulnerabilities in the browsers being launched, and anything that
requires an attacker to already be able to edit your `config.toml` (that file is
trusted input, and it can name any command by design).

## Signing

Release binaries are currently unsigned. Verify downloads against the
`SHA256SUMS` file attached to each release. Windows SmartScreen will warn on
first run of the unsigned `.exe`; see the README install notes.

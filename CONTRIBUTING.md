# Contributing

Thanks for considering a contribution. Issues, ideas, and pull requests are all
welcome.

## Getting set up

You need Go (the version in `go.mod` or newer) and `make`. Optional, only for
rebuilding the Windows icon and version resource: ImageMagick, librsvg, and
[goversioninfo](https://github.com/josephspurrier/goversioninfo).

```bash
git clone https://github.com/mzieba/link-router.git
cd link-router
make build   # native binary
make test    # test suite
```

## Before opening a pull request

Run the same checks CI runs:

```bash
gofmt -l .        # must print nothing
go vet ./...
go test -race ./...
```

Please add or update tests for behaviour you change. The OS-touching parts
(browser detection, default-handler registration) are injected through function
fields on `App` and `browsers.DetectWith`, so tests never mutate the real system.
Keep it that way: a test run must not write registry keys, desktop files, or
change your default browser.

## Commit messages

The project releases automatically with
[release-please](https://github.com/googleapis/release-please), which derives
versions and the changelog from
[Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/). Use:

- `feat: ...` for a new capability (minor bump)
- `fix: ...` for a bug fix (patch bump)
- `docs:`, `chore:`, `refactor:`, `test:` for everything else (no release)
- `feat!: ...` or a `BREAKING CHANGE:` footer for a breaking change (major bump)

## Pull requests

1. Branch from `main`.
2. Keep the change focused; unrelated cleanups are easier to review separately.
3. Describe what you changed and how you verified it, including the platform you
   tested on. Linux and Windows behave differently in `internal/register` and
   `internal/browsers`, so say which one you exercised.
4. Open the pull request against `main`. CI must be green before merge.

## Scope

Platform support is Linux and Windows. macOS is on the roadmap but not currently
targeted, so macOS-specific code is likely to be deferred rather than merged
piecemeal. See [PRD.md](PRD.md) for the intended scope.

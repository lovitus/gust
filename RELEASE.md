# Release and Package Manager Publishing

This repository publishes normal GitHub Release assets for every `v*` tag. Release assets include the main `gost` binary and the standalone `portyd` server built from `lovitus/gust-x`. Package-manager channels are stricter: Homebrew, Scoop, APT, and RPM repositories are updated only for stable tags matching `^v[0-9]+\.[0-9]+\.[0-9]+$`.

The same workflow can be started manually without a tag. Manual dispatch is a
build-only release rehearsal: it builds every standard and singbox archive,
generates manifests and uploads workflow artifacts, but it never creates a
GitHub Release or modifies package repositories.

Current suffix releases such as `v3.2.9-porty7` are prereleases. They publish
normal GitHub Release archives for `gost` and `portyd`, but they do not update
Homebrew, Scoop, APT, or RPM package-manager channels.

## Release Types

- Stable tags such as `v3.2.8` are marked as latest and update package-manager channels.
- Nonstable tags such as `v3.2.8-rc1`, `v3.2.7-sings`, and nightly tags are marked as prerelease and explicitly not latest.
- The first package-manager version is the next stable tag after this publishing setup lands. Existing suffix releases are not backfilled.

## Required Repository Settings

- Secrets:
  - `PACKAGE_GPG_PRIVATE_KEY`
  - `PACKAGE_GPG_PASSPHRASE`
  - existing `GH_PAT` for checking out `lovitus/gust-x`
- Workflow permissions must allow `GITHUB_TOKEN` to write repository contents and Pages.
- If `master` is protected from direct bot pushes, either allow the release workflow bot to push package manifests or change the manifest step to open a bot PR.
- GitHub Pages is published from the `gh-pages` branch. The release workflow creates this branch on first stable package publish if it does not exist.

## Sing-box assets

- Linux amd64/arm64 archives include `libcronet.so` and support Naive.
- Windows amd64/arm64 archives include `libcronet.dll` and support Naive.
- Darwin amd64/arm64 archives omit Naive and CCM; the asset manifest reports
  those features as unavailable.
- `.github/singbox-gust-x.ref` pins the exact gust-x source commit used by tag
  builds, so rerunning an old tag cannot silently build a newer gust-x master.
- Every platform archive is exercised on a native GitHub runner before it is
  treated as a validated CI artifact.

Before pushing a release tag, manually dispatch the release workflow with the
intended version and confirm all build artifacts complete successfully.

## Stable Release Flow

1. Push a stable tag, for example `v3.2.8`.
2. The release workflow builds all existing binary archives.
3. For stable tags only, it checks signing secrets, builds `gust` deb/rpm packages from the existing `gost-linux-amd64` and `gost-linux-arm64` archives, prepares the APT/RPM repository tree, generates `checksums.txt`, and generates Homebrew/Scoop manifests.
4. The workflow creates the GitHub Release, or uploads assets with `--clobber` if the Release already exists.
5. After the Release succeeds, it pushes the signed package repository to `gh-pages` and commits `Formula/gust.rb` plus `bucket/gust.json` back to `master` with `[skip ci]`.

## User Install Channels

- Homebrew tap and Scoop bucket files live in this repository.
- APT and RPM metadata live on GitHub Pages under `https://lovitus.github.io/gust/`.
- The package name is `gust`; the installed command remains `gost`.

## Reruns and Recovery

Release reruns are designed to be recoverable before public package state changes. The workflow prepares package repository files locally, then creates or updates the GitHub Release. It pushes `gh-pages` and package manifests only after the Release step succeeds.

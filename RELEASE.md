# Release and Package Manager Publishing

This repository publishes normal GitHub Release assets for every `v*` tag. Release assets include the main `gost` binary and the standalone `portyd` server built from `lovitus/gust-x`. Package-manager channels are stricter: Homebrew, Scoop, APT, and RPM repositories are updated only for stable tags matching `^v[0-9]+\.[0-9]+\.[0-9]+$`.

Standard tags must point to commits contained in `origin/master`; the release
workflow rejects any other source. Embedded sing-box releases are maintained
separately on `singbox-backend` with `singbox-v*` tags. Custom graphical route
manager releases are maintained independently on `qtui-route-manage` with
`qtui-v*` tags. Neither downstream line may update the standard package-manager
channels or replace the latest standard release. See
[BRANCH_POLICY.md](BRANCH_POLICY.md) before preparing any release line.

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

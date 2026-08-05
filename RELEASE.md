# Standard and embedded sing-box releases

Gust maintains two independent development and release lines.

| Line | Source branch | Tag namespace | Assets | Package channels |
|---|---|---|---|---|
| Standard Gust | `master` | `vX.Y.Z` | `gost` and `portyd` standard matrix | Homebrew, Scoop, APT and RPM for stable tags |
| Embedded sing-box | `singbox-backend` | `singbox-vX.Y.Z` | Six `gust-with-singbox` archives | None |

The two branches have different `.github/workflows/release.yml` definitions.
GitHub evaluates the workflow from the tagged commit, so a standard `v*` tag
on `master` runs the standard workflow while a `singbox-v*` tag on
`singbox-backend` runs the sing-box-only workflow.

## Permanent branch policy

- Develop ordinary Gust and replay upstream changes on `master`.
- Develop the embedded backend on `singbox-backend`.
- Periodically merge or rebase the updated `master` baseline into
  `singbox-backend`, resolve conflicts there, and rerun its complete matrix.
- Keep the matching gust-x work on its own `singbox-backend` branch as well.
- `.github/singbox-gust-x.ref` pins the exact gust-x commit used by tag builds.
  Update the pin deliberately after the gust-x branch passes its tests.
- Do not merge the embedded backend, GPL release files or sing-box workflows
  into `master` unless the two products are intentionally combined later.

## Sing-box CI

The branch-specific Go CI runs for pushes and pull requests targeting
`singbox-backend`. It checks:

- standard/singbox flavor separation and `-singboxmanual`;
- native schema, lifecycle, race and protocol data paths;
- Linux Naive with the matching Cronet runtime;
- Linux, Windows and Darwin on amd64 and arm64 native runners;
- package contents, notices, manuals, validation record and runtime libraries.

The compatibility workflow runs when the pinned sing-box inputs change and can
also be dispatched manually to compare the pinned and latest stable versions.

## Build-only rehearsal

Run the `singbox-release` workflow manually from the `singbox-backend` branch.
For example:

```bash
gh workflow run release.yml \
  --ref singbox-backend \
  -f version=0.0.0-dev
```

An optional exact gust-x ref may be supplied for a rehearsal:

```bash
gh workflow run release.yml \
  --ref singbox-backend \
  -f version=0.0.0-dev \
  -f gust_x_ref=GUST_X_COMMIT
```

Manual dispatch is build-only. It builds and uploads six temporary workflow
artifacts but skips `publish singbox release`, so it creates no tag, GitHub
Release, package repository or package-manager update.

The source guard rejects manual runs from any branch other than
`singbox-backend`.

## Publish a sing-box release

1. Confirm both `singbox-backend` worktrees are clean and pushed.
2. Update `.github/singbox-gust-x.ref` to the exact tested gust-x commit.
3. Wait for the branch Go CI and compatibility jobs to pass.
4. Run a build-only rehearsal with the intended version.
5. Create the tag on a commit contained in the Gust `singbox-backend` branch:

   ```bash
   git switch singbox-backend
   git pull --ff-only
   git tag -a singbox-v1.0.0 -m 'Gust embedded sing-box 1.0.0'
   git push origin singbox-v1.0.0
   ```

6. Confirm all six build jobs and the publish job pass.

The release guard rejects `singbox-v*` tags whose commit is not contained in
`origin/singbox-backend`.

Exact `singbox-vX.Y.Z` tags create non-prerelease GitHub Releases. Suffix tags
such as `singbox-v1.0.0-rc1` are prereleases. Neither form is marked as the
repository's latest release, so the standard `master` release remains the
default download and package-manager source.

## Sing-box assets

- Linux amd64/arm64: binary plus `libcronet.so`, including Naive.
- Windows amd64/arm64: binary plus `libcronet.dll`, including Naive.
- Darwin amd64/arm64: reproducible limited build without Naive or CCM.

Each archive includes `feature-manifest.json`, `SINGBOX-MANUAL.md`,
`SINGBOX-VALIDATION.md`, GPLv3, sing-box notices and the exact Gust/gust-x
source revisions. The release also publishes `checksums.txt`.

## Standard releases

The standard process remains on `master` and uses `v*` tags. Its stable tags
may update Homebrew, Scoop, APT, RPM, `gh-pages`, and package manifests. The
sing-box workflow contains none of those steps and cannot mutate those
channels.

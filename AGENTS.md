# Maintainer instructions

Before editing, run `git branch --show-current` and classify the change using
`BRANCH_POLICY.md`.

- Implement every general feature, bug fix, upstream sync and shared CI change
  on `master` first.
- Use `singbox-backend` only for the embedded sing-box extension.
- If a general change is requested while on `singbox-backend`, switch to
  `master`, implement and test it there, then merge `master` forward.
- Never merge `singbox-backend` back into `master`.
- Use `qtui-route-manage` only for the custom graphical route-management
  product. Sync `master` forward when needed; never merge it back into
  `master` or `singbox-backend`.
- Keep Gust and gust-x on matching branch names during development and tests.
- Use `v*` tags only for commits contained in `origin/master` and
  `singbox-v*` tags only for commits contained in
  `origin/singbox-backend`.
- Use `qtui-v*` tags only for commits contained in
  `origin/qtui-route-manage`; QtUI releases must not update standard package
  channels or the latest standard release.
- Do not release until the source branch is clean, pushed and green in its
  required workflows.

The complete policy and release-channel boundaries are in `BRANCH_POLICY.md`
and `RELEASE.md`.

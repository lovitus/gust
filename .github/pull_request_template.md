## Change classification

- [ ] This is a general change and targets `master` first.
- [ ] This is sing-box-only and targets `singbox-backend`.
- [ ] Any matching gust-x change targets the branch with the same name.

## Branch safety

- [ ] A general change was not implemented independently on `singbox-backend`.
- [ ] `singbox-backend` contains the current tested `master` baseline.
- [ ] No sing-box runtime, dependency, release file or manual is entering `master`.
- [ ] The intended release namespace is `v*` for `master` or `singbox-v*` for the extension.

See [BRANCH_POLICY.md](../BRANCH_POLICY.md) before merging.

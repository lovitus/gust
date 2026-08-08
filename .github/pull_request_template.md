## Change classification

- [ ] This is a general change and targets `master` first.
- [ ] This is sing-box-only and targets `singbox-backend`.
- [ ] This is QtUI route-manager customization and targets `qtui-route-manage`.
- [ ] Any matching gust-x change targets the branch with the same name.

## Branch safety

- [ ] A general change was not implemented independently on `singbox-backend`.
- [ ] `singbox-backend` contains the current tested `master` baseline.
- [ ] No sing-box runtime, dependency, release file or manual is entering `master`.
- [ ] No QtUI route-manager implementation or release file is entering `master` or `singbox-backend`.
- [ ] The intended release namespace is `v*`, `singbox-v*`, or `qtui-v*` for the selected product line.

See [BRANCH_POLICY.md](../BRANCH_POLICY.md) before merging.

# Branch and release policy

Gust has three permanent product lines. `master` is the source of truth for all
general-purpose work. `singbox-backend` is a downstream extension that adds the
embedded sing-box backend on top of `master`. `qtui-route-manage` is an
independently developed customization for graphical route management.

| Product line | Development branch | Release tags | Purpose |
|---|---|---|---|
| Standard Gust | `master` | `v*` | Upstream syncs, fixes, Porty, EasyTier, CLI/config and every other reusable feature |
| Embedded sing-box | `singbox-backend` | `singbox-v*` | Only sing-box parsing, runtime, packaging, compatibility and documentation |
| QtUI route manager | `qtui-route-manage` | `qtui-v*` | Custom desktop UI, tray, privilege elevation and route-profile management |

## Non-negotiable rules

1. Implement every general feature and bug fix on `master` first. Do not make
   `singbox-backend` the source of truth for reusable code.
2. Keep the corresponding gust-x change on the branch with the same name.
3. Bring general work into `singbox-backend` by merging the tested `master`
   branch. Do not independently reimplement or cherry-pick a second copy of a
   general change there.
4. Never merge `singbox-backend` back into `master`. Extract a general fix into
   `master`, test it there, and then merge `master` forward again.
5. Treat `qtui-route-manage` as a downstream customization. Merge tested
   `master` updates into it when needed, but never merge the customization back
   into `master` or `singbox-backend`.
6. Create `v*` tags only on commits contained in `origin/master`. Create
   `singbox-v*` tags only on commits contained in
   `origin/singbox-backend`, and create `qtui-v*` tags only on commits contained
   in `origin/qtui-route-manage`.
7. Standard releases may update the latest release and package-manager
   channels. Sing-box and QtUI releases must never update either.

CI and the release workflows enforce the structural parts of these rules.
They cannot decide whether a change is conceptually general, so reviewers must
use the pull-request checklist as well.

## General change flow

1. Implement and test any required gust-x work on gust-x `master`.
2. Implement and test the Gust change on `master` against gust-x `master`.
3. Merge or release the standard line from `master` as appropriate.
4. Merge the updated gust-x `master` into gust-x `singbox-backend`.
5. Merge the updated Gust `master` into Gust `singbox-backend`.
6. Run the complete sing-box compatibility and native platform matrix before a
   `singbox-v*` release.

## Sing-box-only change flow

Work directly on the two `singbox-backend` branches only when the change is
meaningless without the embedded backend—for example its URI schema, native
runtime, build tags, bundled notices, compatibility matrix or sing-box manual.
If the work reveals a reusable fix, move that fix to `master` and merge it
forward instead of leaving it branch-local.

## QtUI customization flow

Work directly on the two `qtui-route-manage` branches only for personalized
desktop route-management requirements: the graphical UI, tray integration,
privilege elevation, profile persistence, platform packaging and its dedicated
documentation or CI. This product line is not a staging branch for general
features.

When it needs upstream updates, merge the tested `master` branches forward into
gust-x `qtui-route-manage` first and Gust `qtui-route-manage` second. The QtUI
branch may deliberately lag behind `master`; its recorded master baseline must
still be an authentic commit from `origin/master`. Never merge or release its
custom implementation through either of the other product lines.

## Before tagging

- Confirm the working tree is clean and the branch is pushed.
- Confirm the matching gust-x branch or pinned commit passed CI.
- Run the appropriate build-only rehearsal when available.
- Recheck that the tag namespace matches the source branch.

Detailed artifact and publishing behavior is documented in [RELEASE.md](RELEASE.md).

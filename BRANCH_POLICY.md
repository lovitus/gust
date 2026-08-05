# Branch and release policy

Gust has two permanent product lines. `master` is the source of truth for all
general-purpose work. `singbox-backend` is a downstream extension that adds the
embedded sing-box backend on top of `master`.

| Product line | Development branch | Release tags | Purpose |
|---|---|---|---|
| Standard Gust | `master` | `v*` | Upstream syncs, fixes, Porty, EasyTier, CLI/config and every other reusable feature |
| Embedded sing-box | `singbox-backend` | `singbox-v*` | Only sing-box parsing, runtime, packaging, compatibility and documentation |

## Non-negotiable rules

1. Implement every general feature and bug fix on `master` first. Do not make
   `singbox-backend` the source of truth for reusable code.
2. Keep the corresponding gust-x change on the branch with the same name.
3. Bring general work into `singbox-backend` by merging the tested `master`
   branch. Do not independently reimplement or cherry-pick a second copy of a
   general change there.
4. Never merge `singbox-backend` back into `master`. Extract a general fix into
   `master`, test it there, and then merge `master` forward again.
5. Create `v*` tags only on commits contained in `origin/master`. Create
   `singbox-v*` tags only on commits contained in
   `origin/singbox-backend`.
6. Standard releases may update the latest release and package-manager
   channels. Sing-box releases must never update either.

CI and both release workflows enforce the structural parts of these rules.
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

## Before tagging

- Confirm the working tree is clean and the branch is pushed.
- Confirm the matching gust-x branch or pinned commit passed CI.
- Run the appropriate build-only rehearsal when available.
- Recheck that the tag namespace matches the source branch.

Detailed artifact and publishing behavior is documented in [RELEASE.md](RELEASE.md).

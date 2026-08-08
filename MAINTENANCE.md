# Maintenance status

This file is the durable backlog for both product lines. Branch ownership and
release rules remain authoritative in [BRANCH_POLICY.md](BRANCH_POLICY.md) and
[RELEASE.md](RELEASE.md).

## Current baseline

- `master` owns every reusable fix, test, workflow and documentation policy.
- `singbox-backend` is a forward-only extension of `master`; it must never be
  merged back into `master`.
- Docker E2E and unit/configuration tests run as independent CI jobs with hard
  time limits, so one class cannot indefinitely prevent the other from
  starting.
- Newly added documentation is checked for real endpoints and likely
  credentials. Examples must use RFC documentation addresses and reserved
  example domains.

## Certification policy

Feature support is recorded only after a repeatable data-plane test passes.
Parser acceptance, process startup and a listening socket are supporting
evidence, not proof of TCP, UDP or DNS forwarding. External compatibility must
identify the peer implementation and exercise the actual traffic path.

Privileged features and native-platform behavior may require a maintainer
runner. Reports published in the repository must contain only role labels and
sanitized example endpoints—never private inventory, credentials or production
addresses.

## Remaining work classification

The active backlog is kept in test plans next to the relevant suite. An item is
not release-blocking unless it is marked as such there or in a release
checklist. Performance optimizations are accepted only with fixed-environment
before/after evidence; speculative runtime grouping is not a completeness
requirement.

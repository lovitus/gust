# End-to-End Tests

Integration tests that spin up real gost instances inside Docker containers and
verify protocol behavior over the network. The suite covers authentication,
routing, DNS, forwarding, TLS, HTTP/2, Shadowsocks, policy modules and failure
behavior; the exact inventory is the set of `*_test.go` files in this directory.

## Prerequisites

- [Docker](https://docs.docker.com/get-docker/) (running daemon)
- Go toolchain (for compiling the gost binary under test)

## Running

From the repository root:

```bash
# Run all e2e tests (the CI limit is intentionally finite)
go test -count=1 ./tests/e2e/ -v -timeout 30m

# Run a specific test suite
go test ./tests/e2e/ -v -run TestShadowsocksSuite -timeout 5m
go test ./tests/e2e/ -v -run TestParallelSelectorSuite -timeout 5m

# Use a pre-built gost binary (skips compilation)
go test ./tests/e2e/ -v -gost-bin /path/to/gost
```

## Architecture

```
tests/e2e/
├── Dockerfile                  # Shared base image (Alpine + curl, python3, etc.)
├── main_test.go                # TestMain: compiles gost, creates Docker network
├── utils.go                    # Helpers: container lifecycle, config rendering
├── scripts/
│   ├── tcp_echo.py             # HTTP echo server (responds with "hello-gost")
│   └── udp_echo.py             # UDP echo server (reflects payloads)
├── testdata/                   # config files or data files for running cases
├── *_test.go                   # Independently runnable feature suites
└── testdata/                   # Sanitized configurations and fixtures
```

### How it works

1. **TestMain** (`main_test.go`) compiles the gost binary from `../../cmd/gost` (unless `-gost-bin` is provided) and creates a shared Docker network for all containers.
2. Each test suite starts echo server containers (TCP/UDP), then launches separate gost containers for server and client roles.
3. Client configs use Go template syntax (`{{.ServerAddr}}`) so the server address is injected at runtime.
4. Tests verify end-to-end connectivity by sending traffic through the gost proxy chain and checking that the echo server responds correctly.

## Tips

- Increase `-timeout` for CI or slow networks. Container image builds on first run take extra time.
- Use `-gost-bin` to avoid recompiling when iterating on tests locally.
- Add `-v` to see container log output on failure.
- RunGostContainer will wait for exposedPorts automatically, but it's not reliable for udp ports. So, you should check the readiness of udp ports inside cases.
- A process merely listening is not a passing protocol test. Assert the returned
  TCP payload, UDP datagram or DNS answer.
- Use only documentation IP ranges and example domains in committed fixtures.
  Keep real validation inventory and credentials outside the worktree.
- See [PLAN.md](PLAN.md) for the coverage ledger and evidence requirements.

#!/usr/bin/env python3
"""Exercise downloaded release binaries against real OpenSSH in isolated LANs.

Only runs on a disposable GitHub-hosted Linux runner, as root. No source build,
Docker image, production credentials, host routes, or host firewall changes.
"""

from __future__ import annotations

import argparse
import hashlib
import http.server
import json
import os
from pathlib import Path
import re
import secrets
import signal
import subprocess
import sys
import tarfile
import tempfile
import threading
from urllib.parse import quote


class Process:
    def __init__(self, argv, log):
        self.lines = []
        self.ended = False
        self.changed = threading.Condition()
        self.process = subprocess.Popen(argv, stdout=subprocess.PIPE,
                                        stderr=subprocess.STDOUT, start_new_session=True,
                                        text=True, env=clean_env())
        self.reader = threading.Thread(target=self.drain, args=(log,), daemon=True)
        self.reader.start()

    def drain(self, log):
        with log.open("w") as out:
            for line in self.process.stdout:
                out.write(line)
                out.flush()
                with self.changed:
                    self.lines.append(line)
                    self.changed.notify_all()
        with self.changed:
            self.ended = True
            self.changed.notify_all()

    def ready(self, text):
        # One bounded event wait, driven by existing program output, no polling
        # or extra readiness TCP connections that contaminate network evidence.
        with self.changed:
            ok = self.changed.wait_for(
                lambda: self.ended or any(text in line for line in self.lines), 30)
            if not ok or not any(text in line for line in self.lines):
                raise RuntimeError(f"process did not report readiness: {text}")

    def stop(self):
        if self.process.poll() is None:
            try:
                os.killpg(self.process.pid, signal.SIGTERM)
            except ProcessLookupError:
                pass
            try:
                self.process.wait(timeout=5)
            except subprocess.TimeoutExpired:
                os.killpg(self.process.pid, signal.SIGKILL)
                self.process.wait(timeout=5)
        self.reader.join(timeout=5)


def clean_env():
    return {k: v for k, v in os.environ.items()
            if not k.lower().endswith("_proxy") and k != "SSH_AUTH_SOCK"}


def run(*argv, input=None, check=True):
    result = subprocess.run([str(x) for x in argv], input=input, text=True,
                            capture_output=True, timeout=20, env=clean_env())
    if check and result.returncode:
        # Do not expose arguments: some contain ephemeral test credentials.
        raise RuntimeError(f"{Path(str(argv[0])).name} failed: {result.stderr[-1500:]}")
    return result


class Smoke:
    def __init__(self, args):
        self.args = args
        args.work.mkdir(parents=True, exist_ok=True)
        self.root = Path(tempfile.mkdtemp(prefix="porty-artifact-", dir=args.work))
        self.private = self.root / "private"
        self.private.mkdir(mode=0o700)
        self.prefix = "pty" + secrets.token_hex(4)
        self.user = self.prefix
        self.user_created = False
        self.namespaces = []
        self.processes = []
        self.redactions = []
        self.report = {"status": "running", "gust_sha": args.source_sha,
                       "validation_sha": os.getenv("GITHUB_SHA"),
                       "gust_x_sha": args.gust_x_sha, "checks": [], "artifacts": {}}

    def check(self, name, condition=True):
        if not condition:
            raise RuntimeError(name)
        self.report["checks"].append(name)
        print("PASS " + name, flush=True)

    def ns(self, name):
        return f"{self.prefix}-{name}"

    def command(self, name, *args):
        return ["ip", "netns", "exec", self.ns(name), *map(str, args)]

    def start(self, name, label, argv, ready):
        p = Process(self.command(name, *argv), self.private / (label + ".log"))
        self.processes.append(p)
        p.ready(ready)
        return p

    def file(self, name, value):
        path = self.private / name
        path.write_text(value)
        path.chmod(0o600)
        return path

    def prepare(self):
        self.report["tools"] = run("dpkg-query", "-W", "openssh-server", "iproute2", "strace", "openssl").stdout
        binaries = self.root / "bin"
        binaries.mkdir()
        for name in ("gost", "portyd", "portyc"):
            matches = list(self.args.artifacts.glob(f"{name}-linux-amd64-*.tar.gz"))
            self.check(name + " has exactly one release archive", len(matches) == 1)
            archive = matches[0]
            binary = name + "-linux-amd64"
            with tarfile.open(archive) as tar:
                members = tar.getmembers()
                self.check(name + " archive contains only expected binary",
                           len(members) == 1 and members[0].name == binary and members[0].isfile())
                tar.extractall(binaries, filter="data")
            path = binaries / binary
            path.chmod(0o755)
            setattr(self, name, str(path))
            self.report["artifacts"][name] = {
                "archive": archive.name,
                "archive_sha256": hashlib.sha256(archive.read_bytes()).hexdigest(),
                "binary_sha256": hashlib.sha256(path.read_bytes()).hexdigest()}
        self.password = secrets.token_hex(16)
        self.key = secrets.token_hex(32)
        self.redactions.extend([self.password, self.key])
        # RUNNER_TEMP's ancestors can be inaccessible to other users. Keep the
        # disposable SSH account in its own normal home, not under the runner's
        # private workspace; never chmod that workspace to make auth succeed.
        account = Path("/home") / self.user
        self.check("temporary account home is unused", not account.exists())
        self.check("temporary account name is unused", run("getent", "passwd", self.user, check=False).returncode == 2)
        self.user_created = True
        run("useradd", "--no-create-home", "--no-log-init", "--home-dir", account, "--shell", "/bin/sh", self.user)
        run("chpasswd", input=f"{self.user}:{self.password}\n")
        (account / ".ssh").mkdir(parents=True, mode=0o700)
        # Do not inherit a group-writable runner umask: OpenSSH StrictModes
        # rejects it even when the authorized key itself has correct modes.
        account.chmod(0o700)
        (account / ".ssh").chmod(0o700)
        self.clientkey = self.private / "client_key"
        run("ssh-keygen", "-q", "-t", "ed25519", "-N", "", "-f", self.clientkey)
        authorized = account / ".ssh" / "authorized_keys"
        authorized.write_text(self.clientkey.with_suffix(".pub").read_text())
        authorized.chmod(0o600)
        run("chown", "-R", self.user, account)
        run("runuser", "-u", self.user, "--", "test", "-r", authorized)
        self.check("temporary SSH directories have strict modes",
                   account.stat().st_mode & 0o777 == 0o700 and
                   (account / ".ssh").stat().st_mode & 0o777 == 0o700)
        self.check("temporary SSH user can read its authorized key")
        Path("/run/sshd").mkdir(mode=0o755, exist_ok=True)
        self.hostkeys = {}
        for lan in ("a", "b"):
            key = self.private / f"host-{lan}"
            run("ssh-keygen", "-q", "-t", "ed25519", "-N", "", "-f", key)
            self.hostkeys[lan] = run("ssh-keygen", "-lf", key.with_suffix(".pub")).stdout.split()[1]
            for port in (22, 2222):
                self.file(f"sshd-{lan}-{port}", f"""Port {port}
ListenAddress 127.0.0.1
HostKey {key}
PidFile {self.private}/sshd-{lan}-{port}.pid
AuthorizedKeysFile {authorized}
AllowUsers {self.user}
PermitRootLogin no
UsePAM no
PasswordAuthentication yes
KbdInteractiveAuthentication no
PubkeyAuthentication yes
AllowTcpForwarding {'local' if port == 22 else 'no'}
PermitOpen 127.0.0.1:18080 127.0.0.1:18081
MaxSessions 0
X11Forwarding no
AllowAgentForwarding no
LogLevel DEBUG1
""")
        self.cert = self.private / "cert.pem"
        self.tlskey = self.private / "tls.key"
        run("openssl", "req", "-x509", "-newkey", "rsa:2048", "-nodes", "-days", "1",
            "-subj", "/CN=porty.test", "-addext",
            "subjectAltName=IP:192.0.2.1,IP:192.0.2.5,IP:192.0.2.9",
            "-keyout", self.tlskey, "-out", self.cert)
        for name in ("relay", "a", "b", "access"):
            ns = self.ns(name)
            run("ip", "netns", "add", ns)
            self.namespaces.append(ns)
            run("ip", "-n", ns, "link", "set", "lo", "up")
            run(*self.command(name, "sysctl", "-q", "-w", "net.ipv4.ip_forward=0"))
        for i, name in enumerate(("a", "b", "access")):
            # Every link stays within the test namespaces; nothing is attached
            # to the host's interfaces or default route.
            run("ip", "-n", self.ns("relay"), "link", "add", f"r{i}", "type", "veth",
                "peer", "name", "eth0", "netns", self.ns(name))
            run("ip", "-n", self.ns("relay"), "addr", "add", f"192.0.2.{4*i+1}/30", "dev", f"r{i}")
            run("ip", "-n", self.ns(name), "addr", "add", f"192.0.2.{4*i+2}/30", "dev", "eth0")
            run("ip", "-n", self.ns("relay"), "link", "set", f"r{i}", "up")
            run("ip", "-n", self.ns(name), "link", "set", "eth0", "up")
        for source, target in (("a", "192.0.2.6"), ("b", "192.0.2.2"),
                               ("access", "192.0.2.2"), ("a", "203.0.113.1")):
            result = run("ip", "-n", self.ns(source), "route", "get", target, check=False)
            self.check(f"no direct route: {source} to {target}", result.returncode != 0)
        for lan in ("a", "b"):
            self.start(lan, f"http-{lan}", [sys.executable, __file__, "serve", lan], "HTTP ready")
            for port in (22, 2222):
                cfg = self.private / f"sshd-{lan}-{port}"
                run("/usr/sbin/sshd", "-t", "-f", cfg)
                self.start(lan, f"sshd-{lan}-{port}", ["/usr/sbin/sshd", "-D", "-e", "-f", cfg], "Server listening")
            self.start(lan, f"proxy-{lan}", [self.gost, "-L", "http://127.0.0.1:18081"], "listening on")
            public = (self.private / f"host-{lan}.pub").read_text().split()
            known = self.file(f"known-{lan}", "127.0.0.1 " + " ".join(public[:2]) + "\n")
            direct = run(*self.command(lan, "ssh", "-F", "/dev/null", "-i", self.clientkey,
                         "-o", "IdentitiesOnly=yes", "-o", "BatchMode=yes",
                         "-o", "StrictHostKeyChecking=yes", "-o", f"UserKnownHostsFile={known}",
                         "-o", "ConnectTimeout=3", "-W", "127.0.0.1:18080", f"{self.user}@127.0.0.1"),
                         input="GET / HTTP/1.1\r\nHost: localhost\r\nConnection: close\r\n\r\n")
            self.check(f"OpenSSH {lan} direct-tcpip fixture preflight", direct.stdout.endswith(f"lan-{lan}\n"))

    def url(self, source, tls, selection=""):
        addr = {"a": "192.0.2.1", "b": "192.0.2.5", "access": "192.0.2.9"}[source]
        query = "path=/ws&connectTimeout=3s"
        if tls:
            query += f"&scheme=wss&secure=true&ca={self.cert}"
        return f"porty://alice:{self.key}@{addr}:8443?{query}{selection}"

    def request(self, port):
        return run(*self.command("access", "curl", "--silent", "--show-error", "--fail",
                                "--max-time", "5", f"http://127.0.0.1:{port}/"), check=False)

    def forward(self, label, tls, lan="a", password=False, proxy=False,
                reject=None, ordinary=None, legacy=False):
        port = 20000 + len(self.report["checks"])
        target = f"127.0.0.1:{ordinary or 18080}"
        endpoint = self.url("access", tls, "" if ordinary else ("&exit=ssh" if legacy else f"&portyc={lan}"))
        argv = [self.gost, "-L", f"tcp://127.0.0.1:{port}/{target}", "-F", endpoint]
        if reject == "http-before-ssh":
            argv += ["-F", "http://127.0.0.1:18081"]
        if not ordinary and reject != "no-ssh":
            auth = f"{self.user}:{self.password}" if password else self.user
            fingerprint = self.hostkeys["b" if reject == "host-key" else lan]
            query = f"hostKey={quote(fingerprint, safe='')}&handshakeTimeout=3s"
            if not password:
                query += f"&privateKeyFile={self.clientkey}"
            argv += ["-F", f"sshd://{auth}@127.0.0.1:{2222 if reject == 'forwarding' else 22}?{query}"]
            if proxy:
                argv += ["-F", "http://127.0.0.1:18081"]
        p = self.start("access", label, argv, "listening on")
        try:
            result = self.request(port)
            if reject:
                self.check(label, result.returncode != 0)
            else:
                self.check(label, result.returncode == 0 and result.stdout == f"lan-{lan}\n")
        finally:
            p.stop()

    def scenario(self, tls):
        mode = "wss" if tls else "ws"
        config = f"listen: ':8443'\npath: /ws\nusers:\n  alice: '{self.key}'\n"
        if tls:
            config += f"tls:\n  cert: {self.cert}\n  key: {self.tlskey}\n"
        cfg = self.file(f"portyd-{mode}.yaml", config)
        first = len(self.processes)
        self.start("relay", f"portyd-{mode}", [self.portyd, "-C", cfg], "listening on")
        for lan in ("a", "b"):
            cfg = self.file(f"portyc-{mode}-{lan}.yaml", f"forward: '{self.url(lan, tls)}'\nname: {lan}\n")
            self.start(lan, f"portyc-{mode}-{lan}",
                       ["strace", "-f", "-e", "trace=connect,bind,listen", "-o",
                        str(self.private / f"trace-{mode}-{lan}"), self.portyc, "-C", cfg], "SSH exit ready")
        for lan, virtual in (("a", 1111), ("b", 1112)):
            self.start(lan, f"share-{mode}-{lan}", [self.gost, "-L",
                       f"rtcp://:{virtual}/127.0.0.1:18080", "-F", self.url(lan, tls)], "registered virtual port")
            self.forward(f"{mode} ordinary share {lan}", tls, lan=lan, ordinary=virtual)
            self.forward(f"{mode} named exit {lan} loopback via SSH", tls, lan=lan)
        self.forward(f"{mode} password URL authentication", tls, password=True)
        self.forward(f"{mode} HTTP proxy after SSH", tls, proxy=True)
        self.forward(f"{mode} rejects missing SSH hop", tls, reject="no-ssh")
        self.forward(f"{mode} rejects HTTP before SSH", tls, reject="http-before-ssh")
        self.forward(f"{mode} rejects changed SSH host key", tls, reject="host-key")
        self.forward(f"{mode} forwarding denied without exec fallback", tls, reject="forwarding")
        # Default and named registrations must also coexist without mixing.
        cfg = self.file(f"portyc-{mode}-legacy.yaml", f"forward: '{self.url('a', tls)}'\n")
        self.start("a", f"portyc-{mode}-legacy", ["strace", "-f", "-e", "trace=connect,bind,listen",
                   "-o", str(self.private / f"trace-{mode}-legacy"), self.portyc, "-C", cfg], "SSH exit ready")
        self.forward(f"{mode} default exit compatibility", tls, legacy=True)
        for p in reversed(self.processes[first:]):
            p.stop()
        for lan in ("a", "b", "legacy"):
            trace = (self.private / f"trace-{mode}-{lan}").read_text()
            relay = "192.0.2.5" if lan == "b" else "192.0.2.1"
            destinations = re.findall(r'connect\(.*?sin_port=htons\((\d+)\), sin_addr=inet_addr\("([^"]+)"\)', trace)
            allowed = {("8443", relay), ("22", "127.0.0.1"), ("2222", "127.0.0.1")}
            self.check(f"{mode} portyc {lan} connects only relay and SSH",
                       bool(destinations) and set(destinations) <= allowed
                       and ("8443", relay) in destinations and ("22", "127.0.0.1") in destinations
                       and "AF_INET6" not in trace and "AF_UNIX" not in trace)
            self.check(f"{mode} portyc {lan} has no listener", not re.search(r'\b(?:bind|listen)\(', trace))

    def finish(self):
        for lan in ("a", "b"):
            text = (self.private / f"sshd-{lan}-22.log").read_text()
            denied = (self.private / f"sshd-{lan}-2222.log").read_text()
            self.check(f"OpenSSH {lan} authenticated and opened direct-tcpip",
                       "Accepted publickey" in text and "direct-tcpip" in text)
            self.check(f"OpenSSH {lan} received no session/exec fallback",
                       "ctype session" not in text + denied)
        self.check("OpenSSH authenticated password URL", "Accepted password" in
                   (self.private / "sshd-a-22.log").read_text())

    def cleanup(self):
        for p in reversed(self.processes):
            p.stop()
        for ns in reversed(self.namespaces):
            # Only kill processes still inside this run's owned namespaces.
            for pid in run("ip", "netns", "pids", ns).stdout.split():
                try:
                    os.kill(int(pid), signal.SIGKILL)
                except ProcessLookupError:
                    pass
            run("ip", "netns", "delete", ns)
        if self.user_created and run("getent", "passwd", self.user, check=False).returncode == 0:
            run("userdel", "--remove", self.user)

    def save(self):
        self.args.evidence.mkdir(parents=True, exist_ok=True)
        self.args.evidence.chmod(0o755)
        logs = []
        for path in sorted(self.private.glob("*.log")):
            logs.append(path.name + "\n" + "".join(path.read_text(errors="replace").splitlines(True)[-35:]))
        for name, value in (("report.json", json.dumps(self.report, indent=2)),
                            ("sanitized-logs.txt", "\n".join(logs))):
            for secret in self.redactions:
                value = value.replace(secret, "[REDACTED]").replace(quote(secret, safe=""), "[REDACTED]")
            value = value.replace(str(self.root), "[TEST_ROOT]")
            (self.args.evidence / name).write_text(value)
        # Never publish keys, configs, raw traces, or the temporary account.


def serve(label):
    class Handler(http.server.BaseHTTPRequestHandler):
        def do_GET(self):
            body = f"lan-{label}\n".encode()
            self.send_response(200)
            self.send_header("Content-Length", str(len(body)))
            self.end_headers()
            self.wfile.write(body)

        def log_message(self, *args):
            pass

    server = http.server.ThreadingHTTPServer(("127.0.0.1", 18080), Handler)
    print("HTTP ready", flush=True)
    server.serve_forever()


def main():
    if len(sys.argv) == 3 and sys.argv[1] == "serve":
        serve(sys.argv[2])
        return
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--artifacts", type=Path, required=True)
    parser.add_argument("--work", type=Path, required=True)
    parser.add_argument("--evidence", type=Path, required=True)
    parser.add_argument("--gust-x-sha", required=True)
    parser.add_argument("--source-sha", required=True)
    args = parser.parse_args()
    if sys.platform != "linux" or os.geteuid() != 0 or os.getenv("GITHUB_ACTIONS") != "true" or os.getenv("RUNNER_ENVIRONMENT") != "github-hosted":
        parser.error("requires root on a disposable GitHub-hosted Linux runner")
    smoke = Smoke(args)
    def expired(_signal, _frame):
        raise TimeoutError("bounded artifact validation timed out")
    signal.signal(signal.SIGTERM, expired)
    signal.signal(signal.SIGALRM, expired)
    signal.alarm(450)
    try:
        smoke.prepare()
        smoke.scenario(False)
        smoke.scenario(True)
        smoke.finish()
        smoke.report["status"] = "success"
    except Exception as exc:
        smoke.report.update(status="failure", error=str(exc))
    finally:
        signal.alarm(0)
        try:
            smoke.cleanup()
        except Exception as exc:
            smoke.report.update(status="failure", cleanup_error=str(exc))
        smoke.save()
    print("Artifact smoke: " + smoke.report["status"], flush=True)
    if smoke.report["status"] != "success":
        raise SystemExit(1)


if __name__ == "__main__":
    main()

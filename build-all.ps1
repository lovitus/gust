#!/usr/bin/env pwsh
# Build local Gust archives with the same plain-Go optimization flags and
# target naming used by the release workflow. Run from the repository root.
$ErrorActionPreference = "Stop"
$bindir = "bin"
$releasedir = "release"
$name = "gost"
$match = (Select-String -Path cmd/gost/version.go -Pattern 'version = "(.+)"').Matches
if ($match.Count -eq 0) { throw "could not read version from cmd/gost/version.go" }
$version = $match[0].Groups[1].Value
$buildVersion = "v$version"

if (!(Test-Path $bindir)) { New-Item -ItemType Directory -Path $bindir | Out-Null }
if (!(Test-Path $releasedir)) { New-Item -ItemType Directory -Path $releasedir | Out-Null }

$targets = @(
    @{ GOOS="linux";   GOARCH="386";      Name="linux-386" }
    @{ GOOS="linux";   GOARCH="amd64";    Name="linux-amd64";            VariantName="GOAMD64"; VariantValue="v1" }
    @{ GOOS="linux";   GOARCH="amd64";    Name="linux-amd64v3";          VariantName="GOAMD64"; VariantValue="v3" }
    @{ GOOS="linux";   GOARCH="arm";      Name="linux-armv5";            VariantName="GOARM";   VariantValue="5" }
    @{ GOOS="linux";   GOARCH="arm";      Name="linux-armv6";            VariantName="GOARM";   VariantValue="6" }
    @{ GOOS="linux";   GOARCH="arm";      Name="linux-armv7";            VariantName="GOARM";   VariantValue="7" }
    @{ GOOS="linux";   GOARCH="arm64";    Name="linux-arm64" }
    @{ GOOS="linux";   GOARCH="mips";     Name="linux-mips-softfloat";   VariantName="GOMIPS";  VariantValue="softfloat" }
    @{ GOOS="linux";   GOARCH="mips";     Name="linux-mips-hardfloat";   VariantName="GOMIPS";  VariantValue="hardfloat" }
    @{ GOOS="linux";   GOARCH="mipsle";   Name="linux-mipsle-softfloat"; VariantName="GOMIPS";  VariantValue="softfloat" }
    @{ GOOS="linux";   GOARCH="mipsle";   Name="linux-mipsle-hardfloat"; VariantName="GOMIPS";  VariantValue="hardfloat" }
    @{ GOOS="linux";   GOARCH="mips64";   Name="linux-mips64" }
    @{ GOOS="linux";   GOARCH="mips64le"; Name="linux-mips64le" }
    @{ GOOS="linux";   GOARCH="s390x";    Name="linux-s390x" }
    @{ GOOS="linux";   GOARCH="riscv64";  Name="linux-riscv64" }
    @{ GOOS="darwin";  GOARCH="amd64";    Name="darwin-amd64";           VariantName="GOAMD64"; VariantValue="v1" }
    @{ GOOS="darwin";  GOARCH="arm64";    Name="darwin-arm64" }
    @{ GOOS="freebsd"; GOARCH="386";      Name="freebsd-386" }
    @{ GOOS="freebsd"; GOARCH="amd64";    Name="freebsd-amd64";          VariantName="GOAMD64"; VariantValue="v1" }
    @{ GOOS="windows"; GOARCH="386";      Name="windows-386" }
    @{ GOOS="windows"; GOARCH="amd64";    Name="windows-amd64";          VariantName="GOAMD64"; VariantValue="v1" }
    @{ GOOS="windows"; GOARCH="amd64";    Name="windows-amd64v3";        VariantName="GOAMD64"; VariantValue="v3" }
    @{ GOOS="windows"; GOARCH="arm64";    Name="windows-arm64" }
)

$failed = @()
$current = 0
foreach ($target in $targets) {
    $current++
    $extension = if ($target.GOOS -eq "windows") { ".exe" } else { "" }
    $binaryName = "$name-$($target.Name)$extension"
    $binaryPath = Join-Path $bindir $binaryName
    Write-Host "[$current/$($targets.Count)] $($target.Name) " -NoNewline

    Remove-Item Env:GOAMD64, Env:GOARM, Env:GOMIPS -ErrorAction SilentlyContinue
    $env:CGO_ENABLED = "0"
    $env:GOOS = $target.GOOS
    $env:GOARCH = $target.GOARCH
    if ($target.VariantName) {
        Set-Item -Path "Env:$($target.VariantName)" -Value $target.VariantValue
    }

    Remove-Item $binaryPath -Force -ErrorAction SilentlyContinue
    & go build -trimpath "-ldflags=-s -w -X main.version=$buildVersion" -o $binaryPath ./cmd/gost
    if ($LASTEXITCODE -ne 0 -or !(Test-Path $binaryPath)) {
        Write-Host "FAILED" -ForegroundColor Red
        $failed += $target.Name
        continue
    }

    if ($target.GOOS -eq "windows") {
        $archive = Join-Path $releasedir "$name-$($target.Name)-$version.zip"
        Remove-Item $archive -Force -ErrorAction SilentlyContinue
        Compress-Archive -Path $binaryPath -DestinationPath $archive -CompressionLevel Optimal
    } else {
        $archive = Join-Path $releasedir "$name-$($target.Name)-$version.tar.gz"
        Remove-Item $archive -Force -ErrorAction SilentlyContinue
        & tar -czf $archive -C $bindir $binaryName
        if ($LASTEXITCODE -ne 0) { throw "tar failed for $($target.Name)" }
    }

    $binaryMB = [math]::Round((Get-Item $binaryPath).Length / 1MB, 1)
    $archiveMB = [math]::Round((Get-Item $archive).Length / 1MB, 1)
    Write-Host "OK bin=${binaryMB}MB pkg=${archiveMB}MB" -ForegroundColor Green
}

Remove-Item Env:GOAMD64, Env:GOARM, Env:GOMIPS -ErrorAction SilentlyContinue
Write-Host "Success: $($targets.Count - $failed.Count) / $($targets.Count)"
if ($failed.Count -gt 0) {
    Write-Host "Failed: $($failed -join ', ')" -ForegroundColor Red
    exit 1
}

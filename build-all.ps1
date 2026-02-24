#!/usr/bin/env pwsh
# Build gust (gost fork) for all platforms with UPX compression
$bindir = "bin"
$name = "gost"

if (!(Test-Path $bindir)) { New-Item -ItemType Directory -Path $bindir | Out-Null }

$hasUpx = $null -ne (Get-Command upx -ErrorAction SilentlyContinue)
# UPX supports: linux (all arch), windows (x86/x64/arm64)
# UPX does NOT support: darwin, freebsd

$targets = @(
    # Linux
    @{ GOOS="linux"; GOARCH="386";      Name="linux-386";              Upx=$true }
    @{ GOOS="linux"; GOARCH="amd64";    Name="linux-amd64";            Upx=$true }
    @{ GOOS="linux"; GOARCH="amd64";    Name="linux-amd64v3";          Upx=$true;  Env="GOAMD64=v3" }
    @{ GOOS="linux"; GOARCH="arm";      Name="linux-armv5";            Upx=$true;  Env="GOARM=5" }
    @{ GOOS="linux"; GOARCH="arm";      Name="linux-armv6";            Upx=$true;  Env="GOARM=6" }
    @{ GOOS="linux"; GOARCH="arm";      Name="linux-armv7";            Upx=$true;  Env="GOARM=7" }
    @{ GOOS="linux"; GOARCH="arm64";    Name="linux-arm64";            Upx=$true }
    @{ GOOS="linux"; GOARCH="mips";     Name="linux-mips-softfloat";   Upx=$false; Env="GOMIPS=softfloat" }
    @{ GOOS="linux"; GOARCH="mips";     Name="linux-mips-hardfloat";   Upx=$false; Env="GOMIPS=hardfloat" }
    @{ GOOS="linux"; GOARCH="mipsle";   Name="linux-mipsle-softfloat"; Upx=$false; Env="GOMIPS=softfloat" }
    @{ GOOS="linux"; GOARCH="mipsle";   Name="linux-mipsle-hardfloat"; Upx=$false; Env="GOMIPS=hardfloat" }
    @{ GOOS="linux"; GOARCH="mips64";   Name="linux-mips64";           Upx=$false }
    @{ GOOS="linux"; GOARCH="mips64le"; Name="linux-mips64le";         Upx=$false }
    @{ GOOS="linux"; GOARCH="s390x";    Name="linux-s390x";            Upx=$false }
    @{ GOOS="linux"; GOARCH="riscv64";  Name="linux-riscv64";          Upx=$false }
    # Darwin (UPX doesn't support macOS well)
    @{ GOOS="darwin"; GOARCH="amd64";   Name="darwin-amd64";           Upx=$false }
    @{ GOOS="darwin"; GOARCH="arm64";   Name="darwin-arm64";           Upx=$false }
    # FreeBSD
    @{ GOOS="freebsd"; GOARCH="386";    Name="freebsd-386";            Upx=$false }
    @{ GOOS="freebsd"; GOARCH="amd64";  Name="freebsd-amd64";         Upx=$false }
    # Windows
    @{ GOOS="windows"; GOARCH="386";    Name="windows-386";            Upx=$true }
    @{ GOOS="windows"; GOARCH="amd64";  Name="windows-amd64";         Upx=$true }
    @{ GOOS="windows"; GOARCH="amd64";  Name="windows-amd64v3";       Upx=$true;  Env="GOAMD64=v3" }
    @{ GOOS="windows"; GOARCH="arm64";  Name="windows-arm64";         Upx=$true }
)

$total = $targets.Count
$current = 0
$failed = @()
$results = @()

foreach ($t in $targets) {
    $current++
    $ext = if ($t.GOOS -eq "windows") { ".exe" } else { "" }
    $outName = "$bindir\$name-$($t.Name)$ext"

    Write-Host "[$current/$total] $($t.Name) " -NoNewline

    if (Test-Path $outName) { Remove-Item $outName -Force }

    # Build with cmd /c to avoid PowerShell stderr issues
    $envExtra = if ($t.Env) { "set $($t.Env)&&" } else { "" }
    cmd /c "set CGO_ENABLED=0&&set GOOS=$($t.GOOS)&&set GOARCH=$($t.GOARCH)&&${envExtra}go build -ldflags `"-s -w`" -trimpath -o $outName ./cmd/gost 2>&1" | Out-Null

    if (!(Test-Path $outName)) {
        Write-Host "FAILED" -ForegroundColor Red
        $failed += $t.Name
        continue
    }

    $rawSize = (Get-Item $outName).Length

    # UPX compress if supported
    if ($hasUpx -and $t.Upx) {
        cmd /c "upx --best --lzma -q $outName 2>&1" | Out-Null
    }

    $finalSize = (Get-Item $outName).Length
    $sizeMB = [math]::Round($finalSize / 1MB, 1)
    $ratio = if ($rawSize -ne $finalSize) { " ({0:P0})" -f ($finalSize / $rawSize) } else { "" }
    Write-Host "OK ${sizeMB}MB${ratio}" -ForegroundColor Green
    $results += [PSCustomObject]@{ Name=$t.Name; SizeMB=$sizeMB }
}

# Summary
Write-Host ""
Write-Host "=== Build Summary ===" -ForegroundColor Cyan
Write-Host "Success: $($total - $failed.Count) / $total"
if ($failed.Count -gt 0) {
    Write-Host "Failed: $($failed -join ', ')" -ForegroundColor Red
}
Write-Host "Output: $bindir\" -ForegroundColor Cyan

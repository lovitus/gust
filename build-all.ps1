#!/usr/bin/env pwsh
# Build gust for all platforms using garble for obfuscation
# Matches upstream release format: .gz for Linux/Darwin/FreeBSD, .zip for Windows
$bindir = "bin"
$releasedir = "release"
$name = "gost"
$version = (Select-String -Path cmd/gost/version.go -Pattern 'version = "(.+)"').Matches.Groups[1].Value

if (!(Test-Path $bindir)) { New-Item -ItemType Directory -Path $bindir | Out-Null }
if (!(Test-Path $releasedir)) { New-Item -ItemType Directory -Path $releasedir | Out-Null }

$targets = @(
    # Linux
    @{ GOOS="linux"; GOARCH="386";      Name="linux-386" }
    @{ GOOS="linux"; GOARCH="amd64";    Name="linux-amd64" }
    @{ GOOS="linux"; GOARCH="amd64";    Name="linux-amd64v3";          Env="GOAMD64=v3" }
    @{ GOOS="linux"; GOARCH="arm";      Name="linux-armv5";            Env="GOARM=5" }
    @{ GOOS="linux"; GOARCH="arm";      Name="linux-armv6";            Env="GOARM=6" }
    @{ GOOS="linux"; GOARCH="arm";      Name="linux-armv7";            Env="GOARM=7" }
    @{ GOOS="linux"; GOARCH="arm64";    Name="linux-arm64" }
    @{ GOOS="linux"; GOARCH="mips";     Name="linux-mips-softfloat";   Env="GOMIPS=softfloat" }
    @{ GOOS="linux"; GOARCH="mips";     Name="linux-mips-hardfloat";   Env="GOMIPS=hardfloat" }
    @{ GOOS="linux"; GOARCH="mipsle";   Name="linux-mipsle-softfloat"; Env="GOMIPS=softfloat" }
    @{ GOOS="linux"; GOARCH="mipsle";   Name="linux-mipsle-hardfloat"; Env="GOMIPS=hardfloat" }
    @{ GOOS="linux"; GOARCH="mips64";   Name="linux-mips64" }
    @{ GOOS="linux"; GOARCH="mips64le"; Name="linux-mips64le" }
    @{ GOOS="linux"; GOARCH="s390x";    Name="linux-s390x" }
    @{ GOOS="linux"; GOARCH="riscv64";  Name="linux-riscv64" }
    # Darwin
    @{ GOOS="darwin"; GOARCH="amd64";   Name="darwin-amd64" }
    @{ GOOS="darwin"; GOARCH="arm64";   Name="darwin-arm64" }
    # FreeBSD
    @{ GOOS="freebsd"; GOARCH="386";    Name="freebsd-386" }
    @{ GOOS="freebsd"; GOARCH="amd64";  Name="freebsd-amd64" }
    # Windows
    @{ GOOS="windows"; GOARCH="386";    Name="windows-386" }
    @{ GOOS="windows"; GOARCH="amd64";  Name="windows-amd64" }
    @{ GOOS="windows"; GOARCH="amd64";  Name="windows-amd64v3";       Env="GOAMD64=v3" }
    @{ GOOS="windows"; GOARCH="arm64";  Name="windows-arm64" }
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

    # Build with garble -tiny for obfuscation, cmd /c to avoid PowerShell stderr issues
    $envExtra = if ($t.Env) { "set $($t.Env)&&" } else { "" }
    cmd /c "set CGO_ENABLED=0&&set GOOS=$($t.GOOS)&&set GOARCH=$($t.GOARCH)&&${envExtra}garble -tiny build -trimpath -o $outName ./cmd/gost 2>&1" | Out-Null

    if (!(Test-Path $outName)) {
        Write-Host "FAILED" -ForegroundColor Red
        $failed += $t.Name
        continue
    }

    $binSize = (Get-Item $outName).Length

    # Package for release: gzip for Linux/Darwin/FreeBSD, zip for Windows
    if ($t.GOOS -eq "windows") {
        $zipName = "$releasedir\$name-$($t.Name)-${version}.zip"
        if (Test-Path $zipName) { Remove-Item $zipName -Force }
        Compress-Archive -Path $outName -DestinationPath $zipName -CompressionLevel Optimal
        $pkgSize = (Get-Item $zipName).Length
    } else {
        $gzName = "$releasedir\$name-$($t.Name)-${version}.gz"
        if (Test-Path $gzName) { Remove-Item $gzName -Force }
        # Use .NET GZipStream for gzip compression
        $inStream = [System.IO.File]::OpenRead((Resolve-Path $outName))
        $outStream = [System.IO.File]::Create((Join-Path $PWD $gzName))
        $gzip = New-Object System.IO.Compression.GZipStream($outStream, [System.IO.Compression.CompressionLevel]::Optimal)
        $inStream.CopyTo($gzip)
        $gzip.Close(); $outStream.Close(); $inStream.Close()
        $pkgSize = (Get-Item $gzName).Length
    }

    $sizeMB = [math]::Round($binSize / 1MB, 1)
    $pkgMB = [math]::Round($pkgSize / 1MB, 1)
    Write-Host "OK bin=${sizeMB}MB pkg=${pkgMB}MB" -ForegroundColor Green
    $results += [PSCustomObject]@{ Name=$t.Name; BinMB=$sizeMB; PkgMB=$pkgMB }
}

# Summary
Write-Host ""
Write-Host "=== Build Summary ===" -ForegroundColor Cyan
Write-Host "Success: $($total - $failed.Count) / $total"
if ($failed.Count -gt 0) {
    Write-Host "Failed: $($failed -join ', ')" -ForegroundColor Red
}
Write-Host "Binaries: $bindir\" -ForegroundColor Cyan
Write-Host "Releases: $releasedir\" -ForegroundColor Cyan

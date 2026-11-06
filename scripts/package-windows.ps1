param(
    [string]$OutputDirectory = (Join-Path (Get-Location) 'dist'),
    [string]$PackageName = 'netrise-windows-amd64'
)

$ErrorActionPreference = 'Stop'

$goCommand = Get-Command go.exe -ErrorAction SilentlyContinue
if ($null -ne $goCommand) {
    $goExecutable = $goCommand.Source
} else {
    $goExecutable = Join-Path ${env:ProgramFiles} 'Go\bin\go.exe'
    if (!(Test-Path -LiteralPath $goExecutable)) {
        throw "Go executable not found. Install Go or add go.exe to PATH."
    }
}

$output = [IO.Path]::GetFullPath($OutputDirectory)
$staging = Join-Path $output "$PackageName-bundle"
$archive = Join-Path $output "$PackageName.zip"
$checksum = "$archive.sha256"
$verify = Join-Path ([IO.Path]::GetTempPath()) "netrise-package-$PID"

try {
    New-Item -ItemType Directory -Force -Path $output | Out-Null
    if (Test-Path -LiteralPath $staging) { Remove-Item -LiteralPath $staging -Recurse -Force }
    if (Test-Path -LiteralPath $archive) { Remove-Item -LiteralPath $archive -Force }
    if (Test-Path -LiteralPath $checksum) { Remove-Item -LiteralPath $checksum -Force }
    New-Item -ItemType Directory -Force -Path (Join-Path $staging 'assets/content') | Out-Null

    $binary = Join-Path $staging 'netrise.exe'
    $oldGOOS = $env:GOOS
    $oldGOARCH = $env:GOARCH
    $oldCGO = $env:CGO_ENABLED
    $env:GOOS = 'windows'
    $env:GOARCH = 'amd64'
    $env:CGO_ENABLED = '0'
    try {
        & $goExecutable build -trimpath -ldflags='-s -w' -o $binary ./cmd/netrise
        if ($LASTEXITCODE -ne 0) { throw "go build failed with exit code ${LASTEXITCODE}." }
    } finally {
        $env:GOOS = $oldGOOS
        $env:GOARCH = $oldGOARCH
        $env:CGO_ENABLED = $oldCGO
    }
    Copy-Item -LiteralPath 'assets/content/catalog.json' -Destination (Join-Path $staging 'assets/content/catalog.json')

    Compress-Archive -Path $staging -DestinationPath $archive -Force
    $hash = (Get-FileHash -Algorithm SHA256 -LiteralPath $archive).Hash.ToLowerInvariant()
    "$hash  $([IO.Path]::GetFileName($archive))" | Set-Content -LiteralPath $checksum -Encoding ascii

    New-Item -ItemType Directory -Force -Path $verify | Out-Null
    Expand-Archive -LiteralPath $archive -DestinationPath $verify -Force
    $verifiedBinary = Join-Path $verify "$PackageName-bundle/netrise.exe"
    if (!(Test-Path -LiteralPath $verifiedBinary)) { throw 'Packaged executable is missing.' }
    if (!(Test-Path -LiteralPath (Join-Path $verify "$PackageName-bundle/assets/content/catalog.json"))) {
        throw 'Packaged catalog is missing.'
    }
    & $verifiedBinary validate-content
    if ($LASTEXITCODE -ne 0) { throw "Packaged validation failed with exit code ${LASTEXITCODE}." }
    & $verifiedBinary playtest --faction runner --seed 17 *> $null
    if ($LASTEXITCODE -ne 0) { throw "Packaged Runner playtest failed with exit code ${LASTEXITCODE}." }
    & $verifiedBinary playtest --faction corp --seed 17 *> $null
    if ($LASTEXITCODE -ne 0) { throw "Packaged Corporation playtest failed with exit code ${LASTEXITCODE}." }

    Write-Host "NETRISE Windows package created and verified: $archive"
    Write-Host "SHA256: $checksum"
} finally {
    if (Test-Path -LiteralPath $staging) { Remove-Item -LiteralPath $staging -Recurse -Force }
    if (Test-Path -LiteralPath $verify) { Remove-Item -LiteralPath $verify -Recurse -Force }
}

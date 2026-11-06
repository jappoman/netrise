param(
    [int64]$Seed = 17,
    [int64]$Ticks = 600
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

function Invoke-Netrise {
    param(
        [Parameter(Mandatory = $true)]
        [string[]]$Arguments
    )

    & $goExecutable run ./cmd/netrise @Arguments
    if ($LASTEXITCODE -ne 0) {
        throw "netrise failed with exit code ${LASTEXITCODE}: $($Arguments -join ' ')"
    }
}

Invoke-Netrise @('--help')
Invoke-Netrise @('validate-content')
Invoke-Netrise @('playtest', '--faction', 'runner', '--seed', "$Seed")
Invoke-Netrise @('playtest', '--faction', 'corp', '--seed', "$Seed")
Invoke-Netrise @('balance', '--seed-from', '1', '--seed-to', '5')
Invoke-Netrise @('simulate', '--faction', 'runner', '--seed', "$Seed", '--ticks', "$Ticks")
Invoke-Netrise @('simulate', '--faction', 'corp', '--seed', "$Seed", '--ticks', "$Ticks")

Write-Host "NETRISE PowerShell smoke passed (seed=$Seed ticks=$Ticks)."

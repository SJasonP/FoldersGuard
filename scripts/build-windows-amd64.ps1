param(
    [string]$MSYS2Root = "C:\msys64",
    [string[]]$WailsArgs = @()
)

$ErrorActionPreference = "Stop"

$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$RepoRoot = Split-Path -Parent $ScriptDir
$UCRTBin = Join-Path $MSYS2Root "ucrt64\bin"
$Gcc = Join-Path $UCRTBin "gcc.exe"
$Gxx = Join-Path $UCRTBin "g++.exe"

function Require-File {
    param(
        [string]$Path,
        [string]$InstallHint
    )
    if (-not (Test-Path -LiteralPath $Path)) {
        throw "Required file not found: $Path`n$InstallHint"
    }
}

Require-File $Gcc "Install MSYS2 UCRT64 GCC with: pacman -S mingw-w64-ucrt-x86_64-gcc"
Require-File $Gxx "Install MSYS2 UCRT64 G++ with: pacman -S mingw-w64-ucrt-x86_64-gcc"

if (-not (Get-Command wails -ErrorAction SilentlyContinue)) {
    throw "wails was not found on PATH. Install Wails CLI before building."
}

$env:Path = "$UCRTBin;$env:Path"
$env:CGO_ENABLED = "1"
$env:CC = $Gcc
$env:CXX = $Gxx

Push-Location $RepoRoot
try {
    Write-Host "FoldersGuard Windows AMD64 build"
    Write-Host "Repository: $RepoRoot"
    Write-Host "CGO_ENABLED=$env:CGO_ENABLED"
    Write-Host "CC=$env:CC"
    Write-Host "CXX=$env:CXX"
    Write-Host "GCC target: $(& $Gcc -dumpmachine)"
    Write-Host ""

    & wails build -platform windows/amd64 @WailsArgs
    if ($LASTEXITCODE -ne 0) {
        exit $LASTEXITCODE
    }
} finally {
    Pop-Location
}

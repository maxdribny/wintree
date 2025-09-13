# build-release.ps1 - Build release artifacts for all platforms

[CmdletBinding()]
param (
    [Parameter(Mandatory=$true)]
    [string]$Version
)

Write-Host "🚀 Starting release build for wintree version: $Version" -ForegroundColor Cyan

# Validate version format
if (-not ($Version -match "^v[0-9]+\.[0-9]+\.[0-9]+$")) {
    Write-Error "Invalid version format. Please use the format 'vX.Y.Z' (e.g., v0.1.0)."
    exit 1
}

# Get build metadata
$Commit = git rev-parse --short HEAD 2>$null
if (-not $Commit) { $Commit = "unknown" }
$BuildDate = Get-Date -Format "yyyy-MM-dd HH:mm:ss"

# Setup paths
$ProjectRoot = $PSScriptRoot
$DistPath = Join-Path -Path $ProjectRoot -ChildPath "dist"

# Clean and create dist directory
if (Test-Path $DistPath) {
    Write-Host "🧹 Cleaning up previous build artifacts..." -ForegroundColor Yellow
    Remove-Item -Path $DistPath -Recurse -Force
}
New-Item -Path $DistPath -ItemType Directory | Out-Null

# Build targets
$targets = @(
    @{GOOS="windows"; GOARCH="amd64"; Suffix=".exe"}
    @{GOOS="darwin";  GOARCH="arm64"; Suffix=""}
    @{GOOS="darwin";  GOARCH="amd64"; Suffix=""}
    @{GOOS="linux";   GOARCH="amd64"; Suffix=""}
)

# Build each target with version info
$LdFlags = "-s -w -X 'github.com/maxdribny/wintree/cmd.Version=$Version' -X 'github.com/maxdribny/wintree/cmd.Commit=$Commit' -X 'github.com/maxdribny/wintree/cmd.BuildDate=$BuildDate'"

Write-Host "🛠️  Building binaries for all platforms..." -ForegroundColor Cyan
Write-Host "  Version: $Version"
Write-Host "  Commit: $Commit"
Write-Host "  Build Date: $BuildDate"
Write-Host ""

foreach ($target in $targets) {
    $env:GOOS = $target.GOOS
    $env:GOARCH = $target.GOARCH
    
    $BinaryName = "wintree_$($target.GOOS)_$($target.GOARCH)$($target.Suffix)"
    $OutputPath = Join-Path -Path $DistPath -ChildPath $BinaryName
    
    Write-Host "  -> Building for $($target.GOOS)/$($target.GOARCH)..."
    go build -v -trimpath -ldflags="$LdFlags" -o $OutputPath .
    
    if ($LASTEXITCODE -ne 0) {
        Write-Error "Build failed for $($target.GOOS)/$($target.GOARCH)"
        exit 1
    }
}

# Package into zips
Write-Host "🗜️  Packaging binaries..." -ForegroundColor Cyan
Get-ChildItem -Path $DistPath | ForEach-Object {
    $BinaryFile = $_
    $PlatformName = if ($BinaryFile.Name -match "darwin_amd64") {
            "macos_intel"
        } elseif ($BinaryFile.Name -match "darwin_arm64") {
            "macos_silicon"
        } elseif ($BinaryFile.Name -match "windows_amd64") {
            "windows_amd64"
        } elseif ($BinaryFile.Name -match "linux_amd64") {
            "linux_amd64"
        } else {
            $BinaryFile.BaseName
        }
    
    $ArchiveName = "wintree_$($Version)_$($PlatformName).zip"
    $ArchivePath = Join-Path -Path $DistPath -ChildPath $ArchiveName
    
    Write-Host "  -> Creating $ArchiveName..."
    Compress-Archive -Path $BinaryFile.FullName -DestinationPath $ArchivePath -Force
}

# Cleanup raw binaries
Get-ChildItem -Path $DistPath -Exclude "*.zip" | Remove-Item

Write-Host ""
Write-Host "✅ Release build complete!" -ForegroundColor Green
Write-Host "📦 Release artifacts in: $DistPath"
Write-Host ""
Write-Host "📌 Next steps:" -ForegroundColor Cyan
Write-Host "  1. Create git tag: git tag $Version && git push origin $Version"
Write-Host "  2. Upload the .zip files to GitHub Releases"
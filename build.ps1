# Build script for Nero
# Usage: .\build.ps1 [version]

param(
    [string]$Version = "dev",
    [string]$Output = "nero.exe"
)

$ErrorActionPreference = "Stop"

Write-Host "🔨 Building Nero v$Version..." -ForegroundColor Cyan

# Get build information
$CommitSHA = git rev-parse --short HEAD 2>$null
if (-not $CommitSHA) {
    $CommitSHA = "unknown"
}

$BuildDate = Get-Date -Format "yyyy-MM-dd_HH:mm:ss"

# Build flags
$LDFlags = "-X main.Version=$Version -X main.CommitSHA=$CommitSHA -X main.BuildDate=$BuildDate"

# Build
Write-Host "📦 Compiling..." -ForegroundColor Yellow
go build -ldflags "$LDFlags" -o $Output .\cmd\nero\

if ($LASTEXITCODE -eq 0) {
    Write-Host "✅ Build successful: $Output" -ForegroundColor Green
    Write-Host ""
    Write-Host "Run with: .\$Output" -ForegroundColor Cyan
} else {
    Write-Host "❌ Build failed" -ForegroundColor Red
    exit 1
}

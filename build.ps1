# Build memorial-site for Windows + Linux
Set-Location $PSScriptRoot

$env:GOPROXY = "https://goproxy.cn,direct"

Write-Host "========================================"
Write-Host "   Building memorial-site"
Write-Host "========================================"

Write-Host "`n[1/2] Windows version..."
go build -o memorial-site.exe .
if ($LASTEXITCODE -ne 0) { Write-Host "ERROR: build failed" -ForegroundColor Red; pause; exit 1 }
Write-Host "        memorial-site.exe OK" -ForegroundColor Green

Write-Host "[2/2] Linux version..."
$env:CGO_ENABLED = "0"
$env:GOOS = "linux"
$env:GOARCH = "amd64"
go build -o memorial-site .
if ($LASTEXITCODE -ne 0) { Write-Host "ERROR: Linux build failed" -ForegroundColor Red; pause; exit 1 }
Write-Host "        memorial-site OK" -ForegroundColor Green

Write-Host "`n========================================"
Write-Host "   Done!"
Write-Host "   Local test: .\memorial-site.exe"
Write-Host "   Deploy:     .\deploy.ps1"
Write-Host "========================================"

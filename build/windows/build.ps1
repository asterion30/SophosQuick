# Script de compilación para Windows (PowerShell)
param(
    [string]$Version = "1.0.0"
)

$ErrorActionPreference = "Stop"

Write-Host "=========================================" -ForegroundColor Cyan
Write-Host " Compilando SophosQuick v$Version para Windows " -ForegroundColor Cyan
Write-Host "=========================================" -ForegroundColor Cyan

# 1. Asegurar directorio de salida
$DistDir = Join-Path $PSScriptRoot "..\..\dist"
if (-not (Test-Path $DistDir)) {
    New-Item -ItemType Directory -Path $DistDir | Out-Null
}

# 2. Generar recursos de versión e icono (.syso) si goversioninfo está disponible
if (Get-Command goversioninfo -ErrorAction SilentlyContinue) {
    Write-Host "[1/3] Generando recursos de versión (.syso)..." -ForegroundColor Yellow
    Push-Location (Join-Path $PSScriptRoot "..\..\cmd\sophosquick")
    goversioninfo -manifest=../../build/windows/app.manifest ../../build/windows/versioninfo.json
    Pop-Location
} else {
    Write-Host "[1/3] Aviso: 'goversioninfo' no detectado. Para incrustar icono y manifest, ejecuta: go install github.com/josephspurrier/goversioninfo/cmd/goversioninfo@latest" -ForegroundColor DarkGray
}

# 3. Compilar binario en modo GUI (sin ventana de consola)
Write-Host "[2/3] Compilando binario Go (GOOS=windows GOARCH=amd64)..." -ForegroundColor Yellow
$Env:GOOS = "windows"
$Env:GOARCH = "amd64"

$OutExe = Join-Path $DistDir "sophosquick.exe"
go build -ldflags="-H=windowsgui -s -w -X 'main.Version=$Version'" -o $OutExe ./cmd/sophosquick

# 4. Generar hash SHA-256
Write-Host "[3/3] Calculando checksum SHA-256..." -ForegroundColor Yellow
$Hash = (Get-FileHash -Path $OutExe -Algorithm SHA256).Hash
Set-Content -Path (Join-Path $DistDir "sophosquick.exe.sha256") -Value "$Hash  sophosquick.exe"

Write-Host "`n✅ Compilación exitosa!" -ForegroundColor Green
Write-Host "Binario: $OutExe" -ForegroundColor White
Write-Host "SHA-256: $Hash" -ForegroundColor White

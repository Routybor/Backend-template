# Setup kind cluster for Microservices

$ErrorActionPreference = "Stop"

Write-Host "==========================================" -ForegroundColor Cyan
Write-Host "Setting up kind cluster" -ForegroundColor Cyan
Write-Host "==========================================" -ForegroundColor Cyan

Write-Host ""
Write-Host "[1/3] Checking prerequisites..." -ForegroundColor Yellow
if (-not (Get-Command docker -ErrorAction SilentlyContinue)) {
    Write-Host "ERROR: Docker is not installed" -ForegroundColor Red
    exit 1
}
Write-Host "Docker found" -ForegroundColor Green

if (-not (Get-Command kind -ErrorAction SilentlyContinue)) {
    Write-Host "ERROR: kind is not installed" -ForegroundColor Red
    Write-Host "Install: winget install kind" -ForegroundColor Gray
    exit 1
}
Write-Host "kind found" -ForegroundColor Green

Write-Host ""
Write-Host "[2/3] Creating kind cluster..." -ForegroundColor Yellow
$kindConfig = "$PSScriptRoot\..\kind-config.yaml"
if (Test-Path "$PSScriptRoot\..\kind-config.yaml") {
    kind create cluster --name microservices --config $kindConfig
} else {
    kind create cluster --name microservices
}

Write-Host ""
Write-Host "[3/3] Installing NGINX Ingress Controller..." -ForegroundColor Yellow
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/deploy/static/provider/kind/deploy.yaml

Write-Host ""
Write-Host "Waiting for ingress controller..." -ForegroundColor Gray
kubectl wait --namespace ingress-nginx --for=condition=ready pod --selector=app.kubernetes.io/component=controller --timeout=120s 2>$null

Write-Host ""
Write-Host "==========================================" -ForegroundColor Cyan
Write-Host "kind cluster is ready!" -ForegroundColor Green
Write-Host "==========================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "Next steps:" -ForegroundColor Yellow
Write-Host "  .\scripts\deploy-k8s.ps1"

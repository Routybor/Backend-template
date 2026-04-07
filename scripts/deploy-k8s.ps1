# Deploy Microservices to Kubernetes

$ErrorActionPreference = "Stop"
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$ProjectDir = Split-Path -Parent $ScriptDir
$K8sDir = "$ProjectDir\k8s"
$TempDir = "$env:TEMP\k8s-deploy-$(Get-Random)"

function Get-EnvVars {
    $envFile = "$K8sDir\.env"
    if (-not (Test-Path $envFile)) {
        Write-Host "ERROR: $envFile not found" -ForegroundColor Red
        Write-Host "Copy $K8sDir\.env.example to $K8sDir\.env and configure it" -ForegroundColor Yellow
        exit 1
    }
    $vars = @{}
    Get-Content $envFile | ForEach-Object {
        if ($_ -match '^([^=]+)=(.*)$') {
            $vars[$matches[1].Trim()] = $matches[2].Trim()
        }
    }
    return $vars
}

function Set-YamlVariables {
    param($SourceDir, $DestDir, $Vars)
    
    New-Item -ItemType Directory -Force -Path $DestDir | Out-Null
    
    Get-ChildItem $SourceDir -Filter "*.yaml" | ForEach-Object {
        $content = Get-Content $_.FullName -Raw
        foreach ($key in $Vars.Keys) {
            $value = $Vars[$key]
            $content = $content -replace "\$\{$key\}", $value
        }
        $content | Set-Content "$DestDir\$($_.Name)" -NoNewline
    }
}

Write-Host "==========================================" -ForegroundColor Cyan
Write-Host "Deploying Microservices to Kubernetes" -ForegroundColor Cyan
Write-Host "==========================================" -ForegroundColor Cyan

Write-Host ""
Write-Host "[1/6] Loading environment variables..." -ForegroundColor Yellow
$env = Get-EnvVars
Write-Host "Loaded $($env.Count) environment variables" -ForegroundColor Green
Write-Host "Namespace: $($env['K8S_NAMESPACE'])" -ForegroundColor Gray

Write-Host ""
Write-Host "[2/6] Checking kubectl connection..." -ForegroundColor Yellow
kubectl cluster-info --request-timeout=10s 2>$null
if ($LASTEXITCODE -ne 0) {
    Write-Host "ERROR: Cannot connect to Kubernetes cluster" -ForegroundColor Red
    exit 1
}
Write-Host "Connected to cluster successfully" -ForegroundColor Green

Write-Host ""
Write-Host "[3/6] Building Docker images..." -ForegroundColor Yellow
Set-Location "$ProjectDir\backend\gateway"
docker build -t $env["GATEWAY_IMAGE"] .
Set-Location "$ProjectDir\backend\core-service"
docker build -t $env["CORE_SERVICE_IMAGE"] .
Set-Location $ProjectDir

Write-Host ""
Write-Host "[4/6] Loading images (if using kind cluster)..." -ForegroundColor Yellow
if (Get-Command kind -ErrorAction SilentlyContinue) {
    $kindClusters = kind get clusters 2>$null
    if ($LASTEXITCODE -eq 0 -and $kindClusters -match "kind") {
        Write-Host "Loading images into kind cluster..." -ForegroundColor Gray
        kind load docker-image $env["GATEWAY_IMAGE"] --name kind 2>$null
        kind load docker-image $env["CORE_SERVICE_IMAGE"] --name kind 2>$null
    } else {
        Write-Host "Docker Desktop Kubernetes detected - no kind load needed" -ForegroundColor Gray
    }
} else {
    Write-Host "kind not installed - skipping image load" -ForegroundColor Gray
}

Write-Host ""
Write-Host "[5/6] Processing YAML templates..." -ForegroundColor Yellow
Set-YamlVariables -SourceDir "$K8sDir\base" -DestDir $TempDir -Vars $env

$sampleFile = Get-Content "$TempDir\namespace.yaml" -Raw
if ($sampleFile -match '\$\{') {
    Write-Host "WARNING: Variables not substituted in namespace.yaml" -ForegroundColor Yellow
    Write-Host $sampleFile -ForegroundColor Gray
}
Write-Host "Processed YAML files in $TempDir" -ForegroundColor Green

Write-Host ""
Write-Host "[6/6] Deploying manifests..." -ForegroundColor Yellow
kubectl apply -f "$TempDir\namespace.yaml"
kubectl apply -f "$TempDir\configmaps.yaml"
kubectl apply -f "$TempDir\secrets.yaml"
kubectl apply -f "$TempDir\pvc.yaml"
kubectl apply -f "$TempDir\keycloak-deployment.yaml"
kubectl apply -f "$TempDir\keycloak-service.yaml"
kubectl apply -f "$TempDir\core-service-deployment.yaml"
kubectl apply -f "$TempDir\core-service-service.yaml"
kubectl apply -f "$TempDir\gateway-deployment.yaml"
kubectl apply -f "$TempDir\gateway-service.yaml"

Write-Host ""
Write-Host "Waiting for Keycloak (this may take 2-3 minutes)..." -ForegroundColor Gray
kubectl wait --for=condition=available --timeout=180s deployment/keycloak -n $env["K8S_NAMESPACE"] 2>$null
Write-Host "Waiting for core-service..." -ForegroundColor Gray
kubectl wait --for=condition=available --timeout=60s deployment/core-service -n $env["K8S_NAMESPACE"] 2>$null
Write-Host "Waiting for gateway..." -ForegroundColor Gray
kubectl wait --for=condition=available --timeout=60s deployment/gateway -n $env["K8S_NAMESPACE"] 2>$null

Write-Host ""
Write-Host "==========================================" -ForegroundColor Cyan
Write-Host "Deployment Status:" -ForegroundColor Cyan
Write-Host "==========================================" -ForegroundColor Cyan
kubectl get pods -n $env["K8S_NAMESPACE"]
kubectl get services -n $env["K8S_NAMESPACE"]

Write-Host ""
Write-Host "==========================================" -ForegroundColor Cyan
Write-Host "Quick Test (port-forward):" -ForegroundColor Cyan
Write-Host "==========================================" -ForegroundColor Cyan
Write-Host "Run these commands in separate terminals:"
Write-Host ""
Write-Host "  kubectl port-forward svc/keycloak 8180:8080 -n $($env['K8S_NAMESPACE'])"
Write-Host "  kubectl port-forward svc/gateway 8080:80 -n $($env['K8S_NAMESPACE'])"
Write-Host ""
Write-Host "Then test with:"
Write-Host "  curl http://localhost:8080/health"
Write-Host "  curl -X POST http://localhost:8180/realms/$($env['KEYCLOAK_REALM'])/protocol/openid-connect/token \"
Write-Host "       -d 'client_id=$($env['KEYCLOAK_CLIENT_ID'])&client_secret=$($env['KEYCLOAK_CLIENT_SECRET'])&grant_type=client_credentials'"

Write-Host ""
Write-Host "Cleaning up temp files..." -ForegroundColor Gray
Remove-Item -Recurse -Force $TempDir -ErrorAction SilentlyContinue

Write-Host ""
Write-Host "==========================================" -ForegroundColor Green
Write-Host "Deployment complete!" -ForegroundColor Green
Write-Host "==========================================" -ForegroundColor Green

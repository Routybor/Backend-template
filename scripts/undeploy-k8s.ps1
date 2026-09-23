# Remove Microservices from Kubernetes

$ErrorActionPreference = "Stop"
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$ProjectDir = Split-Path -Parent $ScriptDir
$Namespace = "microservices"

Write-Host "==========================================" -ForegroundColor Cyan
Write-Host "Removing Microservices from Kubernetes" -ForegroundColor Cyan
Write-Host "==========================================" -ForegroundColor Cyan

kubectl delete -f "$ProjectDir\k8s\base\ingress.yaml" --ignore-not-found 2>$null
kubectl delete -f "$ProjectDir\k8s\base\gateway-service.yaml" --ignore-not-found 2>$null
kubectl delete -f "$ProjectDir\k8s\base\gateway-deployment.yaml" --ignore-not-found 2>$null
kubectl delete -f "$ProjectDir\k8s\base\core-service-service.yaml" --ignore-not-found 2>$null
kubectl delete -f "$ProjectDir\k8s\base\core-service-deployment.yaml" --ignore-not-found 2>$null
kubectl delete -f "$ProjectDir\k8s\base\keycloak-service.yaml" --ignore-not-found 2>$null
kubectl delete -f "$ProjectDir\k8s\base\keycloak-deployment.yaml" --ignore-not-found 2>$null
kubectl delete -f "$ProjectDir\k8s\base\pvc.yaml" --ignore-not-found 2>$null
kubectl delete -f "$ProjectDir\k8s\base\secrets.yaml" --ignore-not-found 2>$null
kubectl delete -f "$ProjectDir\k8s\base\configmaps.yaml" --ignore-not-found 2>$null
kubectl delete -f "$ProjectDir\k8s\base\namespace.yaml" --ignore-not-found 2>$null

Write-Host ""
Write-Host "==========================================" -ForegroundColor Green
Write-Host "Cleanup complete!" -ForegroundColor Green
Write-Host "==========================================" -ForegroundColor Green

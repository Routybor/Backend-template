# Check Microservices Status

$Namespace = "microservices"

Write-Host "==========================================" -ForegroundColor Cyan
Write-Host "Checking Microservices Status" -ForegroundColor Cyan
Write-Host "==========================================" -ForegroundColor Cyan

Write-Host ""
Write-Host "Pods:" -ForegroundColor Yellow
kubectl get pods -n $Namespace -o wide

Write-Host ""
Write-Host "Services:" -ForegroundColor Yellow
kubectl get services -n $Namespace

Write-Host ""
Write-Host "Ingress:" -ForegroundColor Yellow
kubectl get ingress -n $Namespace

Write-Host ""
Write-Host "Deployments:" -ForegroundColor Yellow
kubectl get deployments -n $Namespace

Write-Host ""
Write-Host "PersistentVolumeClaims:" -ForegroundColor Yellow
kubectl get pvc -n $Namespace

Write-Host ""
Write-Host "Pod Logs (last 20 lines each):" -ForegroundColor Yellow
$pods = kubectl get pods -n $Namespace -o jsonpath='{.items[*].metadata.name}' -Split ' '
foreach ($pod in $pods) {
    if ($pod -and $pod.Trim()) {
        Write-Host ""
        Write-Host "--- $pod ---" -ForegroundColor Gray
        kubectl logs -n $Namespace $pod --tail=20 2>$null
    }
}

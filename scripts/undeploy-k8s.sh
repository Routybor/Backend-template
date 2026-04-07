#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
NAMESPACE="microservices"

echo "=========================================="
echo "Removing Microservices from Kubernetes"
echo "=========================================="

kubectl delete -f "$PROJECT_DIR/k8s/base/ingress.yaml" --ignore-not-found
kubectl delete -f "$PROJECT_DIR/k8s/base/gateway-service.yaml" --ignore-not-found
kubectl delete -f "$PROJECT_DIR/k8s/base/gateway-deployment.yaml" --ignore-not-found
kubectl delete -f "$PROJECT_DIR/k8s/base/core-service-service.yaml" --ignore-not-found
kubectl delete -f "$PROJECT_DIR/k8s/base/core-service-deployment.yaml" --ignore-not-found
kubectl delete -f "$PROJECT_DIR/k8s/base/keycloak-service.yaml" --ignore-not-found
kubectl delete -f "$PROJECT_DIR/k8s/base/keycloak-deployment.yaml" --ignore-not-found
kubectl delete -f "$PROJECT_DIR/k8s/base/pvc.yaml" --ignore-not-found
kubectl delete -f "$PROJECT_DIR/k8s/base/secrets.yaml" --ignore-not-found
kubectl delete -f "$PROJECT_DIR/k8s/base/configmaps.yaml" --ignore-not-found
kubectl delete -f "$PROJECT_DIR/k8s/base/namespace.yaml" --ignore-not-found

echo ""
echo "=========================================="
echo "Cleanup complete!"
echo "=========================================="

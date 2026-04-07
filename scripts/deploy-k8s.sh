#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
NAMESPACE="microservices"

echo "=========================================="
echo "Deploying Microservices to Kubernetes"
echo "=========================================="

echo ""
echo "[1/5] Checking kubectl connection..."
kubectl cluster-info --request-timeout=10s
if [ $? -ne 0 ]; then
    echo "ERROR: Cannot connect to Kubernetes cluster"
    exit 1
fi
echo "Connected to cluster successfully"

echo ""
echo "[2/5] Building Docker images..."
cd "$PROJECT_DIR/backend/gateway"
docker build -t ultimatetemplate/gateway:latest .
cd "$PROJECT_DIR/backend/core-service"
docker build -t ultimatetemplate/core-service:latest .

echo ""
echo "[3/5] Loading images (if using kind cluster)..."
if command -v kind &> /dev/null; then
    if kind get clusters 2>/dev/null | grep -q "^kind$"; then
        echo "Loading images into kind cluster..."
        kind load docker-image ultimatetemplate/gateway:latest --name kind 2>/dev/null || true
        kind load docker-image ultimatetemplate/core-service:latest --name kind 2>/dev/null || true
    else
        echo "Docker Desktop Kubernetes detected - no kind load needed"
    fi
else
    echo "kind not installed - skipping image load"
fi

echo ""
echo "[4/5] Deploying manifests using kubectl..."
kubectl apply -f "$PROJECT_DIR/k8s/base/namespace.yaml"
kubectl apply -f "$PROJECT_DIR/k8s/base/configmaps.yaml"
kubectl apply -f "$PROJECT_DIR/k8s/base/secrets.yaml"
kubectl apply -f "$PROJECT_DIR/k8s/base/pvc.yaml"
kubectl apply -f "$PROJECT_DIR/k8s/base/keycloak-deployment.yaml"
kubectl apply -f "$PROJECT_DIR/k8s/base/keycloak-service.yaml"
kubectl apply -f "$PROJECT_DIR/k8s/base/core-service-deployment.yaml"
kubectl apply -f "$PROJECT_DIR/k8s/base/core-service-service.yaml"
kubectl apply -f "$PROJECT_DIR/k8s/base/gateway-deployment.yaml"
kubectl apply -f "$PROJECT_DIR/k8s/base/gateway-service.yaml"

echo ""
echo "[5/5] Waiting for deployments..."
echo "Waiting for Keycloak (this may take 2-3 minutes)..."
kubectl wait --for=condition=available --timeout=180s deployment/keycloak -n "$NAMESPACE" || true
echo "Waiting for core-service..."
kubectl wait --for=condition=available --timeout=60s deployment/core-service -n "$NAMESPACE" || true
echo "Waiting for gateway..."
kubectl wait --for=condition=available --timeout=60s deployment/gateway -n "$NAMESPACE" || true

echo ""
echo "=========================================="
echo "Deployment Status:"
echo "=========================================="
kubectl get pods -n "$NAMESPACE"
kubectl get services -n "$NAMESPACE"
kubectl get ingress -n "$NAMESPACE"

echo ""
echo "=========================================="
echo "Access URLs:"
echo "=========================================="
echo "Gateway:   http://microservices.local/"
echo "Keycloak:  http://keycloak.local/"
echo ""
echo "To add hosts entries:"
echo "  echo '127.0.0.1 microservices.local keycloak.local' | sudo tee -a /etc/hosts"
echo ""
echo "=========================================="
echo "Deployment complete!"
echo "=========================================="

#!/bin/bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
K8S_DIR="$PROJECT_DIR/k8s"
ENV_FILE="$PROJECT_DIR/.env"
if [[ ! -f "$ENV_FILE" ]]; then
    ENV_FILE="$K8S_DIR/.env"
fi
TEMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TEMP_DIR"' EXIT

echo "=========================================="
echo "Deploying Microservices to Kubernetes"
echo "=========================================="

echo ""
echo "[1/7] Loading environment variables..."
if [[ ! -f "$ENV_FILE" ]]; then
    echo "ERROR: no environment file found" >&2
    echo "Run 'make init' (creates .env) or copy $K8S_DIR/.env.example to $K8S_DIR/.env" >&2
    exit 1
fi
set -a
# shellcheck source=/dev/null
source "$ENV_FILE"
: "${K8S_NAMESPACE:=microservices}"
: "${KEYCLOAK_IMAGE:=quay.io/keycloak/keycloak:26.5.0}"
: "${KEYCLOAK_DB:=dev-file}"
: "${KEYCLOAK_LOG_LEVEL:=INFO}"
: "${KEYCLOAK_REALM:=microservices}"
: "${KEYCLOAK_CLIENT_ID:=gateway}"
: "${KEYCLOAK_CLIENT_SECRET:=gateway-secret}"
: "${KEYCLOAK_TESTUSER_PASSWORD:=testuser}"
: "${KEYCLOAK_DATA_SIZE:=1Gi}"
: "${STORAGE_CLASS:=standard}"
: "${GATEWAY_IMAGE:=ultimatetemplate/gateway:latest}"
: "${GATEWAY_REPLICAS:=2}"
: "${GATEWAY_PORT:=8080}"
: "${CORE_SERVICE_IMAGE:=ultimatetemplate/core-service:latest}"
: "${CORE_SERVICE_REPLICAS:=2}"
: "${CORE_SERVICE_PORT:=8081}"
: "${CORE_SERVICE_GRPC_PORT:=${GRPC_PORT:-9091}}"
set +a
NAMESPACE="${K8S_NAMESPACE:-microservices}"
echo "Loaded ${ENV_FILE##*/}. Namespace: $NAMESPACE"

echo ""
echo "[2/7] Checking kubectl connection..."
kubectl cluster-info --request-timeout=10s
echo "Connected to cluster successfully"

echo ""
echo "[3/7] Building Docker images..."
GATEWAY_IMAGE="${GATEWAY_IMAGE:-ultimatetemplate/gateway:latest}"
CORE_SERVICE_IMAGE="${CORE_SERVICE_IMAGE:-ultimatetemplate/core-service:latest}"
cd "$PROJECT_DIR/backend"
docker build -f gateway/Dockerfile -t "$GATEWAY_IMAGE" .
docker build -f core-service/Dockerfile -t "$CORE_SERVICE_IMAGE" .
cd "$PROJECT_DIR"
echo "Built $GATEWAY_IMAGE and $CORE_SERVICE_IMAGE"

echo ""
echo "[4/7] Loading images into kind cluster (if present)..."
CLUSTER_NAME="$(sed -n 's/^name: *\(.*\)$/\1/p' "$PROJECT_DIR/kind-config.yaml" | head -n1)"
IS_KIND=false
if command -v kind &>/dev/null && [[ -n "$CLUSTER_NAME" ]] && kind get clusters 2>/dev/null | grep -qx "$CLUSTER_NAME"; then
    IS_KIND=true
    echo "Loading images into kind cluster '$CLUSTER_NAME'..."
    kind load docker-image "$GATEWAY_IMAGE" "$CORE_SERVICE_IMAGE" --name "$CLUSTER_NAME"
else
    echo "No kind cluster '$CLUSTER_NAME' detected - assuming Docker Desktop/minikube"
fi

echo ""
echo "[5/7] Installing ingress-nginx (kind only)..."
if [[ "$IS_KIND" == true ]] && ! kubectl get ns ingress-nginx &>/dev/null; then
    echo "Installing ingress-nginx controller..."
    kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/controller-v1.11.4/deploy/static/provider/kind/deploy.yaml
    kubectl wait --for=condition=available --timeout=120s deployment/ingress-nginx-controller -n ingress-nginx || true
fi

echo ""
echo "[6/7] Rendering manifests and deploying..."
if ! command -v envsubst &>/dev/null; then
    echo "ERROR: envsubst not found (install gettext-base)" >&2
    exit 1
fi
for f in "$K8S_DIR"/base/*.yaml; do
    envsubst < "$f" > "$TEMP_DIR/$(basename "$f")"
done
if grep -rq '\${' "$TEMP_DIR"; then
    echo "WARNING: unsubstituted variables remain in rendered manifests" >&2
fi
kubectl apply -f "$TEMP_DIR/namespace.yaml"
kubectl apply -f "$TEMP_DIR/configmaps.yaml"
kubectl apply -f "$TEMP_DIR/secrets.yaml"
kubectl apply -f "$TEMP_DIR/pvc.yaml"
kubectl apply -f "$TEMP_DIR/keycloak-deployment.yaml"
kubectl apply -f "$TEMP_DIR/keycloak-service.yaml"
kubectl apply -f "$TEMP_DIR/core-service-deployment.yaml"
kubectl apply -f "$TEMP_DIR/core-service-service.yaml"
kubectl apply -f "$TEMP_DIR/gateway-deployment.yaml"
kubectl apply -f "$TEMP_DIR/gateway-service.yaml"
kubectl apply -f "$TEMP_DIR/ingress.yaml"

echo ""
echo "[7/7] Waiting for deployments..."
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
#!/bin/bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
K8S_DIR="$SCRIPT_DIR/../k8s"
ENV_FILE="$K8S_DIR/.env"

NAMESPACE="microservices"
if [[ -f "$ENV_FILE" ]]; then
    set -a
    # shellcheck source=/dev/null
    source "$ENV_FILE"
    set +a
    NAMESPACE="${K8S_NAMESPACE:-microservices}"
fi

echo "=========================================="
echo "Checking Microservices Status"
echo "=========================================="

echo ""
echo "Pods:"
kubectl get pods -n "$NAMESPACE" -o wide

echo ""
echo "Services:"
kubectl get services -n "$NAMESPACE"

echo ""
echo "Ingress:"
kubectl get ingress -n "$NAMESPACE"

echo ""
echo "Deployments:"
kubectl get deployments -n "$NAMESPACE"

echo ""
echo "PersistentVolumeClaims:"
kubectl get pvc -n "$NAMESPACE"

echo ""
echo "Pod Logs (last 20 lines each):"
for pod in $(kubectl get pods -n "$NAMESPACE" -o jsonpath='{.items[*].metadata.name}'); do
    echo ""
    echo "--- $pod ---"
    kubectl logs -n "$NAMESPACE" "$pod" --tail=20 2>/dev/null || echo "(no logs available)"
done

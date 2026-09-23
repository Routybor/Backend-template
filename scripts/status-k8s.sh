#!/bin/bash
set -e

NAMESPACE="microservices"

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

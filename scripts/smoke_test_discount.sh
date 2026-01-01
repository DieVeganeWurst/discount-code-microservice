#!/usr/bin/env bash

#set modes to ensure security
set -euo pipefail

#Configurable variables to ensure portability
SVC_NAME="${SVC_NAME:-discountcodeservice}"
DEPLOY_NAME="${DEPLOY_NAME:-discountcodeservice}"
NAMESPACE="${NAMESPACE:-default}"
PORT="${PORT:-7001}"

#declare varible gRPC request using JSON: apply a 10% discount to a cart worth 100 USD 
REQ='{"discount_code":"SAVE10","cart_total":{"units":100,"nanos":0,"currency_code":"USD"}}'

#verifies existence of commands kubectl and grpcurl
command -v kubectl >/dev/null 2>&1 || { echo "kubectl not found"; exit 1; }
command -v grpcurl >/dev/null 2>&1 || { echo "grpcurl not found"; exit 1; }

#logs, ensures observability
echo "==> Waiting for deployment/$DEPLOY_NAME to be available (ns=$NAMESPACE)"

#wait for the discountcodeservice deployment to complete
kubectl -n "$NAMESPACE" rollout status "deployment/$DEPLOY_NAME" --timeout=180s

#logs
echo "==> Port-forwarding svc/$SVC_NAME $PORT:$PORT (ns=$NAMESPACE)"

#Creates the tunnel from localhost to the service running inside the Kubernetes cluster.
#Redirects standard and failure exit to portforward.log for cleaness, runs in bg
kubectl -n "$NAMESPACE" port-forward "svc/$SVC_NAME" "$PORT:$PORT" >/tmp/portforward.log 2>&1 &
#kill tunnel at the end
PF_PID=$!
cleanup() { kill "$PF_PID" >/dev/null 2>&1 || true; }
trap cleanup EXIT

#time to start up
sleep 2

#call to the service via port-forward
echo "==> Smoke call: ApplyDiscount(SAVE10, cart_total=100 USD)"
RESP=$(grpcurl -plaintext -d "$REQ" "localhost:$PORT" hipstershop.DiscountCodeService/ApplyDiscount)

#print answer
echo "$RESP"

#checks both discounted units and change in the final cart
echo "$RESP" | grep -q '"units": "10"' || { echo "FAIL: expected discount units 10"; exit 1; }
echo "$RESP" | grep -q '"units": "90"' || { echo "FAIL: expected final total units 90"; exit 1; }

#success
echo "==> Smoke test PASSED"
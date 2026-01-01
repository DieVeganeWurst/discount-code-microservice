#!/usr/bin/env bash

set -euo pipefail

#configurable variables to ensure portability
SVC_NAME="${SVC_NAME:-discountcodeservice}"
DEPLOY_NAME="${DEPLOY_NAME:-discountcodeservice}"
NAMESPACE="${NAMESPACE:-default}"
PORT="${PORT:-7001}"
TOTAL_REQUESTS="${TOTAL_REQUESTS:-50}"
CONCURRENT_REQUESTS="${CONCURRENT_REQUESTS:-10}"

#example request
REQ='{"discount_code":"SAVE10","cart_total":{"units":100,"nanos":0,"currency_code":"USD"}}'
export REQ PORT

echo "==> Waiting for/$DEPLOY_NAME  (ns=$NAMESPACE)"
kubectl -n "$NAMESPACE" rollout status "deployment/$DEPLOY_NAME" --timeout=180s

echo "==> Port-forwarding svc/$SVC_NAME $PORT:$PORT (ns=$NAMESPACE)"
kubectl -n "$NAMESPACE" port-forward "svc/$SVC_NAME" "$PORT:$PORT" >/tmp/portforward.log 2>&1 &
PF_PID=$!
cleanup() { kill "$PF_PID" >/dev/null 2>&1 || true; }
trap cleanup EXIT

sleep 2

#start of benchmark
echo "==> total requests: $TOTAL_REQUESTS concurrent requests:$CONCURRENT_REQUESTS"
start_time_ms=$(date +%s%3N)

#issue concurrent ApplyDiscount calls and validate both discount and final total
seq "$TOTAL_REQUESTS" | xargs -P "$CONCURRENT_REQUESTS" -I{} bash -c '
  set -euo pipefail
  resp=$(grpcurl -plaintext -d "$REQ" "localhost:$PORT" hipstershop.DiscountCodeService/ApplyDiscount)
  echo "$resp" | grep -q "\"units\": \"10\""
  echo "$resp" | grep -q "\"units\": \"90\""
' bash

#failure
end_time_ms=$(date +%s%3N)
duration_ms=$((end_time_ms - start_time_ms))
rate=$([ "$duration_ms" -gt 0 ] && echo $((TOTAL_REQUESTS * 1000 / duration_ms)) || echo "inf")

#success
echo "==>  $TOTAL_REQUESTS requests in ${duration_ms}ms (~${rate} rps)"
echo "==> Concurrency test PASSED"
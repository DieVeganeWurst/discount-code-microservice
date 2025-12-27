# Discount Code Service

This service validates discount codes and returns a discount amount + final total.

Quick start:
1. Generate gRPC stubs from `../../protos/demo.proto`:
   `./genproto.sh`
2. Run the service:
   `PORT=7001 go run .`

Notes:
- The current implementation is a simple placeholder that accepts `94043` as a valid code
  and applies a 10% discount.
- Replace the logic in `main.go` with the real discount rules or datastore lookup.

# Discount Code Service

This service validates discount codes and returns a discount amount + final total.

PREREQUISITES:
- Go 1.21+
- Docker
- Kubernetes (Minikube or similar)
- kubectl
- (Optional) Skaffold

Quick start:
1. Generate gRPC stubs from `../../protos/demo.proto`:
   `./genproto.sh`
2. Run the service:
   `PORT=7001 go run .`

BUILD:
Build the Docker image locally:
docker build -t discountcodeservice:ci .

TEST:
Run unit tests for the service:
go test -v ./...

DEPLOY (Kubernetes):
Deploy the service to a Kubernetes cluster:
kubectl apply -f deployment.yaml
kubectl rollout status deployment/discountcodeservice --timeout=180s

Check deployment status:
kubectl get pods -l app=discountcodeservice
kubectl get svc discountcodeservice

CI PIPELINE:
The CI pipeline automatically:
- runs unit tests
- builds the Docker image
- deploys the service to Kubernetes (Minikube)
- executes a benchmark script to validate reliability under load
The pipeline is scoped to the DiscountCodeService to keep CI execution fast and focused.

(Optional) DEPLOYMENT VIA SKAFFOLD:
When using the full Online Boutique setup, the service can also be deployed via Skaffold:
skaffold run -m app


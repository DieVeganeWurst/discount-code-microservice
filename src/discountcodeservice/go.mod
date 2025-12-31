module github.com/GoogleCloudPlatform/microservices-demo/src/discountcodeservice

go 1.25

toolchain go1.25.5

require (
	github.com/sirupsen/logrus v1.9.3
	golang.org/x/net v0.47.0
	golang.org/x/sys v0.39.0
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251124214823-79d6a2a48846
	google.golang.org/grpc v1.77.0
	google.golang.org/protobuf v1.36.10
)

require golang.org/x/text v0.31.0 // indirect
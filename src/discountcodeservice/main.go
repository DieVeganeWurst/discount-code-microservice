// Copyright 2025 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	pb "github.com/GoogleCloudPlatform/microservices-demo/src/discountcodeservice/genproto"
)

const (
	defaultPort     = "7001"
	defaultCurrency = "USD"
	sampleCode      = "94043"
)

var log *logrus.Logger

func init() {
	log = logrus.New()
	log.Level = logrus.DebugLevel
	log.Formatter = &logrus.JSONFormatter{
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyTime:  "timestamp",
			logrus.FieldKeyLevel: "severity",
			logrus.FieldKeyMsg:   "message",
		},
		TimestampFormat: time.RFC3339Nano,
	}
	log.Out = os.Stdout
}

func main() {
	port := defaultPort
	if value, ok := os.LookupEnv("PORT"); ok {
		port = value
	}
	port = fmt.Sprintf(":%s", port)

	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	srv := grpc.NewServer()
	pb.RegisterDiscountCodeServiceServer(srv, &server{})
	healthcheck := health.NewServer()
	healthpb.RegisterHealthServer(srv, healthcheck)
	reflection.Register(srv)

	log.Infof("Discount Code Service listening on %s", port)
	if err := srv.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

type server struct {
	pb.UnimplementedDiscountCodeServiceServer
}

func (s *server) ApplyDiscount(ctx context.Context, in *pb.ApplyDiscountRequest) (*pb.ApplyDiscountResponse, error) {
	log.Info("[ApplyDiscount] received request")
	defer log.Info("[ApplyDiscount] completed request")

	if in == nil || in.CartTotal == nil || strings.TrimSpace(in.DiscountCode) == "" {
		return &pb.ApplyDiscountResponse{
			DiscountAmount: zeroMoney(defaultCurrency),
			FinalTotal:     zeroMoney(defaultCurrency),
			ErrorCode:      pb.DiscountErrorCode_DISCOUNT_ERROR_CODE_INVALID,
		}, nil
	}

	currency := in.CartTotal.CurrencyCode
	if currency == "" {
		currency = defaultCurrency
	}

	totalNanos := nanosFromMoney(in.CartTotal)
	if totalNanos <= 0 {
		return &pb.ApplyDiscountResponse{
			DiscountAmount: zeroMoney(currency),
			FinalTotal:     moneyFromNanos(currency, totalNanos),
			ErrorCode:      pb.DiscountErrorCode_DISCOUNT_ERROR_CODE_NOT_APPLICABLE,
		}, nil
	}

	code := strings.TrimSpace(in.DiscountCode)
	if code != sampleCode {
		return &pb.ApplyDiscountResponse{
			DiscountAmount: zeroMoney(currency),
			FinalTotal:     moneyFromNanos(currency, totalNanos),
			ErrorCode:      pb.DiscountErrorCode_DISCOUNT_ERROR_CODE_INVALID,
		}, nil
	}

	// TODO: replace this sample logic with real discount lookup rules.
	discountNanos := totalNanos / 10
	finalNanos := totalNanos - discountNanos

	return &pb.ApplyDiscountResponse{
		DiscountAmount: moneyFromNanos(currency, discountNanos),
		FinalTotal:     moneyFromNanos(currency, finalNanos),
		ErrorCode:      pb.DiscountErrorCode_DISCOUNT_ERROR_CODE_NONE,
	}, nil
}

func nanosFromMoney(m *pb.Money) int64 {
	if m == nil {
		return 0
	}
	return m.Units*1_000_000_000 + int64(m.Nanos)
}

func moneyFromNanos(currency string, nanos int64) *pb.Money {
	units := nanos / 1_000_000_000
	nanosPart := nanos % 1_000_000_000
	if nanosPart < 0 {
		units--
		nanosPart += 1_000_000_000
	}
	return &pb.Money{
		CurrencyCode: currency,
		Units:        units,
		Nanos:        int32(nanosPart),
	}
}

func zeroMoney(currency string) *pb.Money {
	return &pb.Money{
		CurrencyCode: currency,
		Units:        0,
		Nanos:        0,
	}
}

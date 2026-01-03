package main

import (
	"context"
	"testing"

	pb "github.com/GoogleCloudPlatform/microservices-demo/src/discountcodeservice/genproto"
)

// coverage for ApplyDiscount inputs, currency defaults, thresholds, and error codes
func TestApplyDiscount(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		req    *pb.ApplyDiscountRequest
		want   *pb.ApplyDiscountResponse
	}{
		{
			name: "nil request returns invalid",
			req:  nil,
			want: &pb.ApplyDiscountResponse{
				DiscountAmount: zeroMoney(defaultCurrency),
				FinalTotal:     zeroMoney(defaultCurrency),
				ErrorCode:      pb.DiscountErrorCode_DISCOUNT_ERROR_CODE_INVALID,
			},
		},
		{
			name: "missing cart total returns invalid",
			req: &pb.ApplyDiscountRequest{
				DiscountCode: codeSave10,
				CartTotal:    nil,
			},
			want: &pb.ApplyDiscountResponse{
				DiscountAmount: zeroMoney(defaultCurrency),
				FinalTotal:     zeroMoney(defaultCurrency),
				ErrorCode:      pb.DiscountErrorCode_DISCOUNT_ERROR_CODE_INVALID,
			},
		},
		{
			name: "empty code returns invalid",
			req: &pb.ApplyDiscountRequest{
				DiscountCode: "",
				CartTotal:    moneyFromNanos("USD", 50*1_000_000_000),
			},
			want: &pb.ApplyDiscountResponse{
				DiscountAmount: zeroMoney(defaultCurrency),
				FinalTotal:     zeroMoney(defaultCurrency),
				ErrorCode:      pb.DiscountErrorCode_DISCOUNT_ERROR_CODE_INVALID,
			},
		},
		{
			name: "SAVE10 applies 10 percent discount",
			req: &pb.ApplyDiscountRequest{
				DiscountCode: "save10",
				CartTotal:    moneyFromNanos("USD", 100*1_000_000_000),
			},
			want: &pb.ApplyDiscountResponse{
				DiscountAmount: moneyFromNanos("USD", 10*1_000_000_000),
				FinalTotal:     moneyFromNanos("USD", 90*1_000_000_000),
				ErrorCode:      pb.DiscountErrorCode_DISCOUNT_ERROR_CODE_NONE,
			},
		},
		{
			name: "SAVE10 uses default currency when missing",
			req: &pb.ApplyDiscountRequest{
				DiscountCode: codeSave10,
				CartTotal: &pb.Money{
					Units:        50,
					Nanos:        0,
					CurrencyCode: "",
				},
			},
			want: &pb.ApplyDiscountResponse{
				DiscountAmount: moneyFromNanos(defaultCurrency, 5*1_000_000_000),
				FinalTotal:     moneyFromNanos(defaultCurrency, 45*1_000_000_000),
				ErrorCode:      pb.DiscountErrorCode_DISCOUNT_ERROR_CODE_NONE,
			},
		},
		{
			name: "SAVE20 under threshold not applicable",
			req: &pb.ApplyDiscountRequest{
				DiscountCode: codeSave20,
				CartTotal:    moneyFromNanos("USD", 90*1_000_000_000),
			},
			want: &pb.ApplyDiscountResponse{
				DiscountAmount: zeroMoney("USD"),
				FinalTotal:     moneyFromNanos("USD", 90*1_000_000_000),
				ErrorCode:      pb.DiscountErrorCode_DISCOUNT_ERROR_CODE_NOT_APPLICABLE,
			},
		},
		{
			name: "SAVE20 at threshold applies 20 percent discount",
			req: &pb.ApplyDiscountRequest{
				DiscountCode: codeSave20,
				CartTotal:    moneyFromNanos("USD", 100*1_000_000_000),
			},
			want: &pb.ApplyDiscountResponse{
				DiscountAmount: moneyFromNanos("USD", 20*1_000_000_000),
				FinalTotal:     moneyFromNanos("USD", 80*1_000_000_000),
				ErrorCode:      pb.DiscountErrorCode_DISCOUNT_ERROR_CODE_NONE,
			},
		},
		{
			name: "unknown code returns invalid",
			req: &pb.ApplyDiscountRequest{
				DiscountCode: "NOPE",
				CartTotal:    moneyFromNanos("USD", 100*1_000_000_000),
			},
			want: &pb.ApplyDiscountResponse{
				DiscountAmount: zeroMoney("USD"),
				FinalTotal:     moneyFromNanos("USD", 100*1_000_000_000),
				ErrorCode:      pb.DiscountErrorCode_DISCOUNT_ERROR_CODE_INVALID,
			},
		},
		{
			name: "non-positive totals are not applicable",
			req: &pb.ApplyDiscountRequest{
				DiscountCode: codeSave10,
				CartTotal:    moneyFromNanos("USD", 0),
			},
			want: &pb.ApplyDiscountResponse{
				DiscountAmount: zeroMoney("USD"),
				FinalTotal:     moneyFromNanos("USD", 0),
				ErrorCode:      pb.DiscountErrorCode_DISCOUNT_ERROR_CODE_NOT_APPLICABLE,
			},
		},
	}

	s := &server{}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := s.ApplyDiscount(context.Background(), tc.req)
			if err != nil {
				t.Fatalf("ApplyDiscount returned unexpected error: %v", err)
			}
			assertResponseEqual(t, tc.want, got)
		})
	}
}

func TestNanosFromMoney(t *testing.T) {
	t.Parallel()

	// coverage for nil, positive, and negative nanos/unit combinations
	tests := []struct {
		name string
		m    *pb.Money
		want int64
	}{
		{
			name: "nil returns zero",
			m:    nil,
			want: 0,
		},
		{
			name: "positive units and nanos",
			m: &pb.Money{
				Units:        3,
				Nanos:        500_000_000,
				CurrencyCode: "USD",
			},
			want: 3_500_000_000,
		},
		{
			name: "negative nanos with negative units",
			m: &pb.Money{
				Units:        -1,
				Nanos:        -250_000_000,
				CurrencyCode: "USD",
			},
			want: -1_250_000_000,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := nanosFromMoney(tc.m); got != tc.want {
				t.Fatalf("nanosFromMoney() = %d, want %d", got, tc.want)
			}
		})
	}
}

func TestMoneyFromNanos(t *testing.T) {
	t.Parallel()

	// coverage for whole numbers, fractional nanos, and negative normalization
	tests := []struct {
		name string
		n    int64
		want *pb.Money
	}{
		{
			name: "zero nanos",
			n:    0,
			want: &pb.Money{CurrencyCode: "USD", Units: 0, Nanos: 0},
		},
		{
			name: "positive nanos with fraction",
			n:    3_500_000_000,
			want: &pb.Money{CurrencyCode: "USD", Units: 3, Nanos: 500_000_000},
		},
		{
			name: "negative nanos normalizes nanos to positive",
			n:    -1_500_000_000,
			want: &pb.Money{CurrencyCode: "USD", Units: -2, Nanos: 500_000_000},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := moneyFromNanos("USD", tc.n)
			if got.CurrencyCode != tc.want.CurrencyCode || got.Units != tc.want.Units || got.Nanos != tc.want.Nanos {
				t.Fatalf("moneyFromNanos() = %+v, want %+v", got, tc.want)
			}
		})
	}
}

func assertResponseEqual(t *testing.T, want, got *pb.ApplyDiscountResponse) {
	t.Helper()
	if want == nil && got == nil {
		return
	}
	if want == nil || got == nil {
		t.Fatalf("response mismatch: want=%+v got=%+v", want, got)
	}
	assertMoneyEqual(t, want.DiscountAmount, got.DiscountAmount, "DiscountAmount")
	assertMoneyEqual(t, want.FinalTotal, got.FinalTotal, "FinalTotal")
	if want.ErrorCode != got.ErrorCode {
		t.Fatalf("ErrorCode mismatch: want=%v got=%v", want.ErrorCode, got.ErrorCode)
	}
}

func assertMoneyEqual(t *testing.T, want, got *pb.Money, label string) {
	t.Helper()
	if want == nil && got == nil {
		return
	}
	if want == nil || got == nil {
		t.Fatalf("%s mismatch: want=%+v got=%+v", label, want, got)
	}
	if want.CurrencyCode != got.CurrencyCode || want.Units != got.Units || want.Nanos != got.Nanos {
		t.Fatalf("%s mismatch: want=%+v got=%+v", label, want, got)
	}
}
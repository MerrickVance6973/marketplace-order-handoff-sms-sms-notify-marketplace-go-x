package main

import "testing"

func TestHandoffMessage(t *testing.T) {
	tests := []struct {
		name, status, phone string
		wantSend            bool
	}{
		{"ready", "ready_for_handoff", "+15550001", true},
		{"paid", "paid", "+15550001", false},
		{"missing phone", "ready_for_handoff", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, got := HandoffMessage(Order{ID: "A-1", BuyerName: "Lee", Status: tt.status, SellerPhone: tt.phone})
			if got != tt.wantSend {
				t.Fatalf("send=%v, want %v", got, tt.wantSend)
			}
		})
	}
}

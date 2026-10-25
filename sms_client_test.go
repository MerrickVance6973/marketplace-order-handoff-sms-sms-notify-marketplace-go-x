package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSendUsesBodyField(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload map[string]string
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Errorf("decode request: %v", err)
		}
		if len(payload) != 2 || payload["to"] != "+15551234567" || payload["body"] != "handoff ready" {
			t.Errorf("unexpected sms.send payload: %v", payload)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"ok":true,"data":{"message_id":"test"}}`))
	}))
	defer server.Close()

	client := &SMSClient{BaseURL: server.URL, Key: "test", HTTP: server.Client()}
	if _, err := client.Send("+15551234567", "handoff ready", "test-request"); err != nil {
		t.Fatal(err)
	}
}

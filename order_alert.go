package main

import (
	"encoding/json"
	"fmt"
)

type Order struct{ ID, SellerPhone, BuyerName, Status string }

func HandoffMessage(o Order) (string, bool) {
	if o.Status != "ready_for_handoff" || o.SellerPhone == "" {
		return "", false
	}
	return fmt.Sprintf("Order %s for %s is ready for handoff.", o.ID, o.BuyerName), true
}

func NotifyHandoff(client *SMSClient, o Order) (string, error) {
	message, send := HandoffMessage(o)
	if !send {
		return "", nil
	}
	data, err := client.Send(o.SellerPhone, message, "order-handoff-"+o.ID)
	if err != nil {
		return "", err
	}
	var sent struct {
		MessageID string `json:"message_id"`
	}
	if json.Unmarshal(data, &sent) == nil && sent.MessageID != "" {
		data, err = client.Status(sent.MessageID)
	}
	return string(data), err
}

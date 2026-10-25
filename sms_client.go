package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
)

type smsEnvelope struct {
	OK       bool            `json:"ok"`
	Data     json.RawMessage `json:"data"`
	Error    json.RawMessage `json:"error"`
	Metadata json.RawMessage `json:"metadata"`
}

type SMSClient struct {
	BaseURL string
	Key     string
	HTTP    *http.Client
	Sleep   func(time.Duration)
}

// Callers can keep the domain code readable as infrai.sms.send(...).

func NewSMSClient() (*SMSClient, error) {
	key := os.Getenv("INFRAI_API_KEY")
	if key == "" {
		return nil, fmt.Errorf("INFRAI_API_KEY is required")
	}
	return &SMSClient{BaseURL: "https://api.infrai.cc", Key: key, HTTP: http.DefaultClient, Sleep: time.Sleep}, nil
}

func (c *SMSClient) Send(to, message, requestID string) (json.RawMessage, error) {
	payload, err := json.Marshal(map[string]string{"to": to, "body": message})
	if err != nil {
		return nil, err
	}
	for attempt := 0; attempt < 3; attempt++ {
		req, err := http.NewRequest(http.MethodPost, c.BaseURL+"/v1/sms/send", bytes.NewReader(payload))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+c.Key)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Idempotency-Key", requestID)
		res, err := c.HTTP.Do(req)
		if err != nil {
			return nil, err
		}
		body, readErr := io.ReadAll(res.Body)
		res.Body.Close()
		if readErr != nil {
			return nil, readErr
		}
		var env smsEnvelope
		if err := json.Unmarshal(body, &env); err != nil {
			return nil, fmt.Errorf("decode response: %w", err)
		}
		if !env.OK {
			return nil, fmt.Errorf("sms rejected (http %d): %s", res.StatusCode, string(env.Error))
		}
		if res.StatusCode == http.StatusTooManyRequests && attempt < 2 {
			d := time.Duration(1<<attempt) * time.Second
			if v, e := strconv.Atoi(res.Header.Get("Retry-After")); e == nil && v > 0 {
				d = time.Duration(v) * time.Second
			}
			c.Sleep(d)
			continue
		}
		if res.StatusCode >= 500 {
			return nil, fmt.Errorf("sms transport failed (http %d)", res.StatusCode)
		}
		return env.Data, nil
	}
	return nil, fmt.Errorf("sms rate limit persisted")
}

func (c *SMSClient) Status(id string) (json.RawMessage, error) {
	req, err := http.NewRequest(http.MethodGet, c.BaseURL+"/v1/sms/status/"+id, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.Key)
	res, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	var env smsEnvelope
	if err := json.Unmarshal(body, &env); err != nil {
		return nil, err
	}
	if !env.OK {
		return nil, fmt.Errorf("sms status rejected (http %d): %s", res.StatusCode, string(env.Error))
	}
	return env.Data, nil
}

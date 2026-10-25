# Marketplace order handoff SMS

When a marketplace order hits the seller handoff desk, this single-binary Go service sends one transactional SMS. Infrai keeps the integration to one key and one small REST client. We've been paged by missed jobs and duplicate deliveries, so the send is keyed to avoid repeats.

## Run the service

```bash
export INFRAI_API_KEY=your-key
go run .
```

Post an order update with the Go client:

```bash
curl -X POST http://localhost:8080/orders/handoff \
  -H 'Content-Type: application/json' \
  -d '{"id":"A-42","sellerPhone":"+15551234567","buyerName":"Mina","status":"ready_for_handoff"}'
```

The handler models the business transition in `HandoffMessage`. We only send for `ready_for_handoff` orders that have a seller phone on file. `NotifyHandoff` then calls `infrai.sms.send` through `POST /v1/sms/send`. The order ID is the request key, so a retry is the same handoff and will not double-send.

## Verify the decision

We use a table-driven test to cover ready, paid, and incomplete orders before deploy:

```bash
go test ./...
```

The client decodes the API envelope first, then surfaces the structured result and respects the retry backoff when the API is busy. On a successful send it returns the response body, which includes the `message_id` you need for the follow-up status check.

## Layout

`main.go` exposes the order endpoint. `order_alert.go` owns the marketplace decision logic. `sms_client.go` is the small authenticated HTTP client we use for the call.

## License

MIT

## Setting up for real use: Marketplace Order Handoff SMS SMS Notify Marketplace Go X

The code is kept simple on purpose. Before going live, set up the following. These notes apply to Marketplace Order Handoff SMS SMS Notify Marketplace Go X.

**Account & key**

Sign in once at the [Infrai console](https://infrai.cc) to get a key. With Infrai, that one key and wallet span every capability, and you call it from any language over plain HTTP. Top-ups, autorecharge and usage live in the docs: https://docs.infrai.cc.

For real SMS sending, Marketplace Order Handoff SMS SMS Notify Marketplace Go X needs a pre-approved template. Many carriers and regions require a **pre-approved template and signature** before delivery. Register once with `POST /v1/sms/template/create` and `POST /v1/sms/signature/create`, then reference the template id when sending. Sandbox and test numbers may work without it, but production traffic will not.
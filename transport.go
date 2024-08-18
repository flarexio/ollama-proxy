package ollamaproxy

import (
	"context"
	"encoding/json"

	"github.com/nats-io/nats.go"
	"github.com/ollama/ollama/api"
)

func VersionHandler(svc Service) nats.MsgHandler {
	return func(msg *nats.Msg) {
		ctx := context.Background()
		ver, err := svc.Version(ctx)
		if err != nil {
			msg.Respond([]byte(err.Error()))
			return
		}

		msg.Respond([]byte(ver))
	}
}

func ListHandler(svc Service) nats.MsgHandler {
	return func(msg *nats.Msg) {
		ctx := context.Background()
		resp, err := svc.List(ctx)
		if err != nil {
			msg.Respond([]byte(err.Error()))
			return
		}

		bs, err := json.Marshal(&resp)
		if err != nil {
			msg.Respond([]byte(err.Error()))
			return
		}

		msg.Respond(bs)
	}
}

func ChatHandler(svc Service) nats.MsgHandler {
	return func(msg *nats.Msg) {
		var req *api.ChatRequest
		if err := json.Unmarshal(msg.Data, &req); err != nil {
			msg.Respond([]byte(err.Error()))
			return
		}

		ctx := context.Background()
		if err := svc.Chat(ctx, req, func(cr api.ChatResponse) error {
			bs, err := json.Marshal(&cr)
			if err != nil {
				return err
			}

			return msg.Respond(bs)
		}); err != nil {
			msg.Respond([]byte(err.Error()))
			return
		}
	}
}

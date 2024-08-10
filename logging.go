package ollamaproxy

import (
	"context"

	"github.com/ollama/ollama/api"
	"go.uber.org/zap"
)

func LoggingMiddleware(log *zap.Logger) ServiceMiddleware {
	return func(next Service) Service {
		log := log.With(
			zap.String("service", "ollama"),
		)

		return &loggingMiddleware{log, next}
	}
}

type loggingMiddleware struct {
	log  *zap.Logger
	next Service
}

func (mw *loggingMiddleware) Version(ctx context.Context) (string, error) {
	log := mw.log.With(
		zap.String("action", "version"),
	)

	ver, err := mw.next.Version(ctx)
	if err != nil {
		log.Error(err.Error())
		return "", err
	}

	return ver, nil
}

func (mw *loggingMiddleware) List(ctx context.Context) (*api.ListResponse, error) {
	log := mw.log.With(
		zap.String("action", "list"),
	)

	resp, err := mw.next.List(ctx)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}

	return resp, nil
}

func (mw *loggingMiddleware) Chat(ctx context.Context, req *api.ChatRequest, fn api.ChatResponseFunc) error {
	log := mw.log.With(
		zap.String("action", "chat"),
		zap.String("model", req.Model),
	)

	last := req.Messages[len(req.Messages)-1]

	log.Debug("req",
		zap.String("role", last.Role),
		zap.String("content", last.Content),
	)

	var resp string

	if err := mw.next.Chat(ctx, req, func(cr api.ChatResponse) error {
		if cr.Done {
			log.Debug("resp",
				zap.String("role", cr.Message.Role),
				zap.String("content", resp),
			)
		}

		resp += cr.Message.Content

		return fn(cr)
	}); err != nil {
		log.Error(err.Error())
		return err
	}

	return nil
}

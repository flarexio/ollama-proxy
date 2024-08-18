package proxy

import (
	"context"
	"net/http"
	"net/url"

	"github.com/ollama/ollama/api"
)

type Service interface {
	Version(ctx context.Context) (string, error)
	List(ctx context.Context) (*api.ListResponse, error)
	Chat(ctx context.Context, req *api.ChatRequest, fn api.ChatResponseFunc) error
}

type ServiceMiddleware func(Service) Service

func NewService(instance string) (Service, error) {
	url, err := url.Parse(instance)
	if err != nil {
		return nil, err
	}

	return api.NewClient(url, http.DefaultClient), nil
}

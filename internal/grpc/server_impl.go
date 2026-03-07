// Package grpcserver реализует gRPC сервер
package grpcserver

import (
	"context"

	shortenerpb "github.com/BuJIKuH/go-musthave-shortener-tpl/internal/proto/shortener"
	"github.com/BuJIKuH/go-musthave-shortener-tpl/internal/service"
	"github.com/BuJIKuH/go-musthave-shortener-tpl/internal/storage"
	"go.uber.org/zap"

	"google.golang.org/protobuf/types/known/emptypb"
)

// GRPCServer реализует shortener.ShortenerServiceServer
type GRPCServer struct {
	facade *service.ShortenerServiceFacade
	logger *zap.Logger
	shortenerpb.UnimplementedShortenerServiceServer
}

// NewGRPCServer создает gRPC сервер с фасадом
func NewGRPCServer(store storage.Storage, authMgr interface{}, logger *zap.Logger) *GRPCServer {
	facade := service.NewShortenerFacade(store)
	return &GRPCServer{facade: facade, logger: logger}
}

func (s *GRPCServer) ShortenURL(ctx context.Context, req *shortenerpb.URLShortenRequest) (*shortenerpb.URLShortenResponse, error) {
	userID, ok := UserIDFromContext(ctx)
	if !ok {
		return nil, ErrUnauthenticated
	}

	id, err := s.facade.ShortenURL(ctx, userID, req.Url)
	if err != nil {
		return nil, err
	}

	return &shortenerpb.URLShortenResponse{Result: id}, nil
}

func (s *GRPCServer) ExpandURL(ctx context.Context, req *shortenerpb.URLExpandRequest) (*shortenerpb.URLExpandResponse, error) {
	result, err := s.facade.ExpandURL(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &shortenerpb.URLExpandResponse{Result: result}, nil
}

func (s *GRPCServer) ListUserURLs(ctx context.Context, req *emptypb.Empty) (*shortenerpb.UserURLsResponse, error) {
	userID, ok := UserIDFromContext(ctx)
	if !ok {
		return nil, ErrUnauthenticated
	}

	items, err := s.facade.ListUserURLs(ctx, userID)
	if err != nil {
		return nil, err
	}

	response := &shortenerpb.UserURLsResponse{}
	for _, it := range items {
		response.Url = append(response.Url, &shortenerpb.URLData{
			ShortUrl:    it.ShortID,
			OriginalUrl: it.OriginalURL,
		})
	}

	return response, nil
}

package grpc

import (
	"context"
	"errors"

	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"

	pb "github.com/ASTeterin/urlshortener/api"
	"github.com/ASTeterin/urlshortener/internal/service"
)

type server struct {
	pb.UnimplementedShortenerServiceServer
	service service.ShortenerService
}

func NewServer(svc service.ShortenerService) *server {
	return &server{service: svc}
}

func (s *server) ShortenURL(ctx context.Context, req *pb.URLShortenRequest) (*pb.URLShortenResponse, error) {
	userID, _ := getUserIDFromMetadata(ctx)
	shortURL, err := s.service.GetShortURL(req.Url, userID)
	if err != nil {
		return nil, err
	}
	return &pb.URLShortenResponse{Result: *shortURL}, nil
}

func (s *server) ExpandURL(_ context.Context, req *pb.URLExpandRequest) (*pb.URLExpandResponse, error) {
	original, err := s.service.GetOriginalURL(req.Id)
	if err != nil {
		return nil, err
	}
	return &pb.URLExpandResponse{Result: *original}, nil
}

func (s *server) ListUserURLs(ctx context.Context, empty *emptypb.Empty) (*pb.UserURLsResponse, error) {
	userID, err := getUserIDFromMetadata(ctx)
	if err != nil {
		return nil, err
	}

	shortURLsMap, err := s.service.ListUserURLs(userID)
	if err != nil {
		return nil, err
	}

	var pbUrls []*pb.URLData
	for originalURL, shortURL := range shortURLsMap {
		pbUrls = append(pbUrls, &pb.URLData{
			ShortUrl:    shortURL,
			OriginalUrl: originalURL,
		})
	}
	return &pb.UserURLsResponse{Urls: pbUrls}, nil
}

func getUserIDFromMetadata(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", errors.New("missing metadata")
	}
	auth := md["authorization"]
	if len(auth) == 0 {
		return "", errors.New("missing authorization header")
	}

	return auth[0], nil
}

package grpc

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	pb "github.com/ASTeterin/urlshortener/api"
	"github.com/ASTeterin/urlshortener/internal/model"
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
	if req.Url == "" {
		return nil, status.Error(codes.InvalidArgument, "url is required")
	}

	userID, err := getUserIDFromMetadata(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "missing user ID")
	}

	shortURL, err := s.service.GetShortURL(ctx, req.Url, userID)
	if err != nil {
		if errors.Is(err, model.ErrDuplicateURL) {
			return &pb.URLShortenResponse{Result: *shortURL}, nil
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.URLShortenResponse{Result: *shortURL}, nil
}

func (s *server) ExpandURL(ctx context.Context, req *pb.URLExpandRequest) (*pb.URLExpandResponse, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	_, err := getUserIDFromMetadata(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "missing user ID")
	}

	original, err := s.service.GetOriginalURL(ctx, req.Id)
	if err != nil {
		if errors.Is(err, model.ErrURLHasBeenDeleted) {
			return nil, status.Error(codes.NotFound, "URL has been deleted")
		}
		return nil, status.Error(codes.InvalidArgument, "invalid short URL")
	}

	return &pb.URLExpandResponse{Result: *original}, nil
}

func (s *server) ListUserURLs(ctx context.Context, empty *emptypb.Empty) (*pb.UserURLsResponse, error) {
	userID, err := getUserIDFromMetadata(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "missing user ID")
	}

	shortURLsMap, err := s.service.ListUserURLs(ctx, userID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
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

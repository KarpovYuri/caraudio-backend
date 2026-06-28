package grpc

import (
	"context"
	"time"

	"github.com/KarpovYuri/caraudio-backend/internal/catalog/app/services"
	"github.com/KarpovYuri/caraudio-backend/internal/catalog/domain"
	catalogv1 "github.com/KarpovYuri/caraudio-backend/pkg/api/proto/catalog/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type CatalogGRPCServer struct {
	catalogv1.UnimplementedCatalogServiceServer
	catalogService services.CatalogService
	jwtSecret      string
}

func NewCatalogGRPCServer(
	catalogService services.CatalogService,
	jwtSecret string,
) *CatalogGRPCServer {
	return &CatalogGRPCServer{
		catalogService: catalogService,
		jwtSecret:      jwtSecret,
	}
}

func (s *CatalogGRPCServer) ListCategories(
	ctx context.Context,
	_ *catalogv1.ListCategoriesRequest,
) (*catalogv1.ListCategoriesResponse, error) {
	categories, err := s.catalogService.ListCategories(ctx)
	if err != nil {
		return nil, mapServiceError(err)
	}
	out := make([]*catalogv1.Category, 0, len(categories))
	for i := range categories {
		out = append(out, toProtoCategory(&categories[i]))
	}
	return &catalogv1.ListCategoriesResponse{Categories: out}, nil
}

func (s *CatalogGRPCServer) GetCategory(
	ctx context.Context,
	req *catalogv1.GetCategoryRequest,
) (*catalogv1.GetCategoryResponse, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "category id is required")
	}
	category, err := s.catalogService.GetCategory(ctx, req.Id)
	if err != nil {
		return nil, mapServiceError(err)
	}
	return &catalogv1.GetCategoryResponse{Category: toProtoCategory(category)}, nil
}

func (s *CatalogGRPCServer) CreateCategory(
	ctx context.Context,
	req *catalogv1.CreateCategoryRequest,
) (*catalogv1.CreateCategoryResponse, error) {
	if err := requireAdmin(ctx, s.jwtSecret); err != nil {
		return nil, mapServiceError(err)
	}
	category, err := s.catalogService.CreateCategory(ctx, req.Name, req.Slug, req.ParentId)
	if err != nil {
		return nil, mapServiceError(err)
	}
	return &catalogv1.CreateCategoryResponse{Category: toProtoCategory(category)}, nil
}

func (s *CatalogGRPCServer) UpdateCategory(
	ctx context.Context,
	req *catalogv1.UpdateCategoryRequest,
) (*catalogv1.UpdateCategoryResponse, error) {
	if err := requireAdmin(ctx, s.jwtSecret); err != nil {
		return nil, mapServiceError(err)
	}
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "category id is required")
	}
	category, err := s.catalogService.UpdateCategory(ctx, req.Id, req.Name, req.Slug, req.ParentId)
	if err != nil {
		return nil, mapServiceError(err)
	}
	return &catalogv1.UpdateCategoryResponse{Category: toProtoCategory(category)}, nil
}

func (s *CatalogGRPCServer) DeleteCategory(
	ctx context.Context,
	req *catalogv1.DeleteCategoryRequest,
) (*catalogv1.DeleteCategoryResponse, error) {
	if err := requireAdmin(ctx, s.jwtSecret); err != nil {
		return nil, mapServiceError(err)
	}
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "category id is required")
	}
	if err := s.catalogService.DeleteCategory(ctx, req.Id); err != nil {
		return nil, mapServiceError(err)
	}
	return &catalogv1.DeleteCategoryResponse{Success: true}, nil
}

func toProtoCategory(category *domain.Category) *catalogv1.Category {
	out := &catalogv1.Category{
		Id:        category.ID,
		Name:      category.Name,
		Slug:      category.Slug,
		CreatedAt: category.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt: category.UpdatedAt.UTC().Format(time.RFC3339),
	}
	if category.ParentID != nil {
		out.ParentId = *category.ParentID
	}
	return out
}

package grpc

import (
	"context"
	"time"

	"github.com/KarpovYuri/caraudio-backend/internal/catalog/domain"
	catalogv1 "github.com/KarpovYuri/caraudio-backend/pkg/api/proto/catalog/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *CatalogGRPCServer) ListProductImages(
	ctx context.Context,
	req *catalogv1.ListProductImagesRequest,
) (*catalogv1.ListProductImagesResponse, error) {
	if req.ProductId == "" {
		return nil, status.Error(codes.InvalidArgument, "product id is required")
	}
	admin := isAdmin(ctx, s.jwtSecret)

	images, err := s.catalogService.ListProductImages(ctx, req.ProductId, admin)
	if err != nil {
		return nil, mapServiceError(err)
	}

	out := make([]*catalogv1.ProductImage, 0, len(images))
	for i := range images {
		out = append(out, toProtoProductImage(&images[i]))
	}
	return &catalogv1.ListProductImagesResponse{Images: out}, nil
}

func (s *CatalogGRPCServer) GetProductImage(
	ctx context.Context,
	req *catalogv1.GetProductImageRequest,
) (*catalogv1.GetProductImageResponse, error) {
	if req.ProductId == "" || req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "product id and image id are required")
	}
	admin := isAdmin(ctx, s.jwtSecret)

	image, err := s.catalogService.GetProductImage(ctx, req.ProductId, req.Id, admin)
	if err != nil {
		return nil, mapServiceError(err)
	}
	return &catalogv1.GetProductImageResponse{Image: toProtoProductImage(image)}, nil
}

func (s *CatalogGRPCServer) CreateProductImage(
	ctx context.Context,
	req *catalogv1.CreateProductImageRequest,
) (*catalogv1.CreateProductImageResponse, error) {
	if err := requireAdmin(ctx, s.jwtSecret); err != nil {
		return nil, mapServiceError(err)
	}
	productID := req.ProductId
	if productID == "" {
		return nil, status.Error(codes.InvalidArgument, "product id is required")
	}
	if req.Url == "" {
		return nil, status.Error(codes.InvalidArgument, "image url is required")
	}

	image, err := s.catalogService.CreateProductImage(ctx, productID, domain.ProductImageInput{
		URL:       req.Url,
		AltText:   req.AltText,
		SortOrder: req.SortOrder,
		IsPrimary: req.IsPrimary,
	})
	if err != nil {
		return nil, mapServiceError(err)
	}
	return &catalogv1.CreateProductImageResponse{Image: toProtoProductImage(image)}, nil
}

func (s *CatalogGRPCServer) UpdateProductImage(
	ctx context.Context,
	req *catalogv1.UpdateProductImageRequest,
) (*catalogv1.UpdateProductImageResponse, error) {
	if err := requireAdmin(ctx, s.jwtSecret); err != nil {
		return nil, mapServiceError(err)
	}
	if req.ProductId == "" || req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "product id and image id are required")
	}
	if req.Url == "" {
		return nil, status.Error(codes.InvalidArgument, "image url is required")
	}

	image, err := s.catalogService.UpdateProductImage(ctx, req.ProductId, req.Id, domain.ProductImageInput{
		URL:       req.Url,
		AltText:   req.AltText,
		SortOrder: req.SortOrder,
		IsPrimary: req.IsPrimary,
	})
	if err != nil {
		return nil, mapServiceError(err)
	}
	return &catalogv1.UpdateProductImageResponse{Image: toProtoProductImage(image)}, nil
}

func (s *CatalogGRPCServer) DeleteProductImage(
	ctx context.Context,
	req *catalogv1.DeleteProductImageRequest,
) (*catalogv1.DeleteProductImageResponse, error) {
	if err := requireAdmin(ctx, s.jwtSecret); err != nil {
		return nil, mapServiceError(err)
	}
	if req.ProductId == "" || req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "product id and image id are required")
	}
	if err := s.catalogService.DeleteProductImage(ctx, req.ProductId, req.Id); err != nil {
		return nil, mapServiceError(err)
	}
	return &catalogv1.DeleteProductImageResponse{Success: true}, nil
}

func toProtoProductImage(image *domain.ProductImage) *catalogv1.ProductImage {
	return &catalogv1.ProductImage{
		Id:        image.ID,
		ProductId: image.ProductID,
		Url:       image.URL,
		AltText:   image.AltText,
		SortOrder: image.SortOrder,
		IsPrimary: image.IsPrimary,
		CreatedAt: image.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt: image.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func attachProductImages(
	ctx context.Context,
	s *CatalogGRPCServer,
	productID string,
	admin bool,
	proto *catalogv1.Product,
) {
	images, err := s.catalogService.ListProductImages(ctx, productID, admin)
	if err != nil {
		return
	}
	proto.Images = make([]*catalogv1.ProductImage, 0, len(images))
	for i := range images {
		proto.Images = append(proto.Images, toProtoProductImage(&images[i]))
	}
}

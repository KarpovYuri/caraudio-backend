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

func (s *CatalogGRPCServer) ListProducts(
	ctx context.Context,
	req *catalogv1.ListProductsRequest,
) (*catalogv1.ListProductsResponse, error) {
	page, pageSize := s.catalogService.NormalizePagination(
		req.Page, req.PageSize, s.defaultPageSize, s.maxPageSize,
	)

	activeOnly := true
	if req.IncludeInactive {
		if err := requireAdmin(ctx, s.jwtSecret); err != nil {
			return nil, mapServiceError(err)
		}
		activeOnly = false
	}

	result, err := s.catalogService.ListProducts(ctx, domain.ProductListFilter{
		CategoryID: req.CategoryId,
		BrandID:    req.BrandId,
		SupplierID: req.SupplierId,
		ActiveOnly: activeOnly,
		Page:       page,
		PageSize:   pageSize,
	})
	if err != nil {
		return nil, mapServiceError(err)
	}

	products := make([]*catalogv1.Product, 0, len(result.Products))
	for i := range result.Products {
		products = append(products, toProtoProduct(&result.Products[i]))
	}

	return &catalogv1.ListProductsResponse{
		Products: products,
		Total:    result.Total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (s *CatalogGRPCServer) GetProduct(
	ctx context.Context,
	req *catalogv1.GetProductRequest,
) (*catalogv1.GetProductResponse, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "product id is required")
	}
	admin := isAdmin(ctx, s.jwtSecret)
	product, err := s.catalogService.GetProduct(ctx, req.Id)
	if err != nil {
		if admin {
			product, err = s.catalogService.GetProductByID(ctx, req.Id)
		}
		if err != nil {
			return nil, mapServiceError(err)
		}
	}
	protoProduct := toProtoProduct(product)
	attachProductImages(ctx, s, req.Id, admin, protoProduct)
	attachProductAttributes(ctx, s, req.Id, admin, protoProduct)
	return &catalogv1.GetProductResponse{Product: protoProduct}, nil
}

func (s *CatalogGRPCServer) CreateProduct(
	ctx context.Context,
	req *catalogv1.CreateProductRequest,
) (*catalogv1.CreateProductResponse, error) {
	if err := requireAdmin(ctx, s.jwtSecret); err != nil {
		return nil, mapServiceError(err)
	}
	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "product name is required")
	}
	product, err := s.catalogService.CreateProduct(
		ctx,
		req.CategoryId,
		req.BrandId,
		req.SupplierId,
		req.Name,
		req.Description,
		req.PriceCents,
		req.Sku,
		req.Stock,
		req.IsActive,
	)
	if err != nil {
		return nil, mapServiceError(err)
	}
	return &catalogv1.CreateProductResponse{Product: toProtoProduct(product)}, nil
}

func (s *CatalogGRPCServer) UpdateProduct(
	ctx context.Context,
	req *catalogv1.UpdateProductRequest,
) (*catalogv1.UpdateProductResponse, error) {
	if err := requireAdmin(ctx, s.jwtSecret); err != nil {
		return nil, mapServiceError(err)
	}
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "product id is required")
	}
	product, err := s.catalogService.UpdateProduct(
		ctx,
		req.Id,
		req.CategoryId,
		req.BrandId,
		req.SupplierId,
		req.Name,
		req.Description,
		req.PriceCents,
		req.Sku,
		req.Stock,
		req.IsActive,
	)
	if err != nil {
		return nil, mapServiceError(err)
	}
	return &catalogv1.UpdateProductResponse{Product: toProtoProduct(product)}, nil
}

func (s *CatalogGRPCServer) DeleteProduct(
	ctx context.Context,
	req *catalogv1.DeleteProductRequest,
) (*catalogv1.DeleteProductResponse, error) {
	if err := requireAdmin(ctx, s.jwtSecret); err != nil {
		return nil, mapServiceError(err)
	}
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "product id is required")
	}
	if err := s.catalogService.DeleteProduct(ctx, req.Id); err != nil {
		return nil, mapServiceError(err)
	}
	return &catalogv1.DeleteProductResponse{Success: true}, nil
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

func toProtoProduct(product *domain.Product) *catalogv1.Product {
	out := &catalogv1.Product{
		Id:          product.ID,
		Name:        product.Name,
		Description: product.Description,
		PriceCents:  product.PriceCents,
		Stock:       product.Stock,
		IsActive:    product.IsActive,
		CreatedAt:   product.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:   product.UpdatedAt.UTC().Format(time.RFC3339),
	}
	if product.CategoryID != nil {
		out.CategoryId = *product.CategoryID
	}
	if product.SupplierID != nil {
		out.SupplierId = *product.SupplierID
	}
	if product.BrandID != nil {
		out.BrandId = *product.BrandID
	}
	if product.SKU != nil {
		out.Sku = *product.SKU
	}
	return out
}

package grpc

import (
	"context"
	"time"

	"github.com/KarpovYuri/caraudio-backend/internal/catalog/domain"
	catalogv1 "github.com/KarpovYuri/caraudio-backend/pkg/api/proto/catalog/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// --- Brands ---

func (s *CatalogGRPCServer) ListBrands(
	ctx context.Context,
	req *catalogv1.ListBrandsRequest,
) (*catalogv1.ListBrandsResponse, error) {
	activeOnly := true
	if req.IncludeInactive && isAdmin(ctx, s.jwtSecret) {
		activeOnly = false
	}
	brands, err := s.catalogService.ListBrands(ctx, activeOnly)
	if err != nil {
		return nil, mapServiceError(err)
	}
	out := make([]*catalogv1.Brand, 0, len(brands))
	for i := range brands {
		out = append(out, toProtoBrand(&brands[i]))
	}
	return &catalogv1.ListBrandsResponse{Brands: out}, nil
}

func (s *CatalogGRPCServer) GetBrand(
	ctx context.Context,
	req *catalogv1.GetBrandRequest,
) (*catalogv1.GetBrandResponse, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "brand id is required")
	}
	admin := isAdmin(ctx, s.jwtSecret)
	var brand *domain.Brand
	var err error
	if admin {
		brand, err = s.catalogService.GetBrandByID(ctx, req.Id)
	} else {
		brand, err = s.catalogService.GetBrand(ctx, req.Id)
	}
	if err != nil {
		return nil, mapServiceError(err)
	}
	return &catalogv1.GetBrandResponse{Brand: toProtoBrand(brand)}, nil
}

func (s *CatalogGRPCServer) CreateBrand(
	ctx context.Context,
	req *catalogv1.CreateBrandRequest,
) (*catalogv1.CreateBrandResponse, error) {
	if err := requireAdmin(ctx, s.jwtSecret); err != nil {
		return nil, mapServiceError(err)
	}
	brand, err := s.catalogService.CreateBrand(ctx, domain.BrandInput{
		Name: req.Name, Slug: req.Slug, Description: req.Description, IsActive: req.IsActive,
	})
	if err != nil {
		return nil, mapServiceError(err)
	}
	return &catalogv1.CreateBrandResponse{Brand: toProtoBrand(brand)}, nil
}

func (s *CatalogGRPCServer) UpdateBrand(
	ctx context.Context,
	req *catalogv1.UpdateBrandRequest,
) (*catalogv1.UpdateBrandResponse, error) {
	if err := requireAdmin(ctx, s.jwtSecret); err != nil {
		return nil, mapServiceError(err)
	}
	brand, err := s.catalogService.UpdateBrand(ctx, req.Id, domain.BrandInput{
		Name: req.Name, Slug: req.Slug, Description: req.Description, IsActive: req.IsActive,
	})
	if err != nil {
		return nil, mapServiceError(err)
	}
	return &catalogv1.UpdateBrandResponse{Brand: toProtoBrand(brand)}, nil
}

func (s *CatalogGRPCServer) DeleteBrand(
	ctx context.Context,
	req *catalogv1.DeleteBrandRequest,
) (*catalogv1.DeleteBrandResponse, error) {
	if err := requireAdmin(ctx, s.jwtSecret); err != nil {
		return nil, mapServiceError(err)
	}
	if err := s.catalogService.DeleteBrand(ctx, req.Id); err != nil {
		return nil, mapServiceError(err)
	}
	return &catalogv1.DeleteBrandResponse{Success: true}, nil
}

func toProtoBrand(brand *domain.Brand) *catalogv1.Brand {
	return &catalogv1.Brand{
		Id:          brand.ID,
		Name:        brand.Name,
		Slug:        brand.Slug,
		Description: brand.Description,
		IsActive:    brand.IsActive,
		CreatedAt:   brand.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:   brand.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

// --- Product attributes ---

func (s *CatalogGRPCServer) ListProductAttributes(
	ctx context.Context,
	req *catalogv1.ListProductAttributesRequest,
) (*catalogv1.ListProductAttributesResponse, error) {
	if req.ProductId == "" {
		return nil, status.Error(codes.InvalidArgument, "product id is required")
	}
	admin := isAdmin(ctx, s.jwtSecret)
	attrs, err := s.catalogService.ListProductAttributes(ctx, req.ProductId, admin)
	if err != nil {
		return nil, mapServiceError(err)
	}
	out := make([]*catalogv1.ProductAttribute, 0, len(attrs))
	for i := range attrs {
		out = append(out, toProtoProductAttribute(&attrs[i]))
	}
	return &catalogv1.ListProductAttributesResponse{Attributes: out}, nil
}

func (s *CatalogGRPCServer) GetProductAttribute(
	ctx context.Context,
	req *catalogv1.GetProductAttributeRequest,
) (*catalogv1.GetProductAttributeResponse, error) {
	admin := isAdmin(ctx, s.jwtSecret)
	attr, err := s.catalogService.GetProductAttribute(ctx, req.ProductId, req.Id, admin)
	if err != nil {
		return nil, mapServiceError(err)
	}
	return &catalogv1.GetProductAttributeResponse{Attribute: toProtoProductAttribute(attr)}, nil
}

func (s *CatalogGRPCServer) CreateProductAttribute(
	ctx context.Context,
	req *catalogv1.CreateProductAttributeRequest,
) (*catalogv1.CreateProductAttributeResponse, error) {
	if err := requireAdmin(ctx, s.jwtSecret); err != nil {
		return nil, mapServiceError(err)
	}
	attr, err := s.catalogService.CreateProductAttribute(ctx, req.ProductId, domain.ProductAttributeInput{
		Name: req.Name, Value: req.Value, SortOrder: req.SortOrder,
	})
	if err != nil {
		return nil, mapServiceError(err)
	}
	return &catalogv1.CreateProductAttributeResponse{Attribute: toProtoProductAttribute(attr)}, nil
}

func (s *CatalogGRPCServer) UpdateProductAttribute(
	ctx context.Context,
	req *catalogv1.UpdateProductAttributeRequest,
) (*catalogv1.UpdateProductAttributeResponse, error) {
	if err := requireAdmin(ctx, s.jwtSecret); err != nil {
		return nil, mapServiceError(err)
	}
	attr, err := s.catalogService.UpdateProductAttribute(ctx, req.ProductId, req.Id, domain.ProductAttributeInput{
		Name: req.Name, Value: req.Value, SortOrder: req.SortOrder,
	})
	if err != nil {
		return nil, mapServiceError(err)
	}
	return &catalogv1.UpdateProductAttributeResponse{Attribute: toProtoProductAttribute(attr)}, nil
}

func (s *CatalogGRPCServer) DeleteProductAttribute(
	ctx context.Context,
	req *catalogv1.DeleteProductAttributeRequest,
) (*catalogv1.DeleteProductAttributeResponse, error) {
	if err := requireAdmin(ctx, s.jwtSecret); err != nil {
		return nil, mapServiceError(err)
	}
	if err := s.catalogService.DeleteProductAttribute(ctx, req.ProductId, req.Id); err != nil {
		return nil, mapServiceError(err)
	}
	return &catalogv1.DeleteProductAttributeResponse{Success: true}, nil
}

func toProtoProductAttribute(attr *domain.ProductAttribute) *catalogv1.ProductAttribute {
	return &catalogv1.ProductAttribute{
		Id:        attr.ID,
		ProductId: attr.ProductID,
		Name:      attr.Name,
		Value:     attr.Value,
		SortOrder: attr.SortOrder,
		CreatedAt: attr.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt: attr.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func attachProductAttributes(ctx context.Context, s *CatalogGRPCServer, productID string, admin bool, proto *catalogv1.Product) {
	attrs, err := s.catalogService.ListProductAttributes(ctx, productID, admin)
	if err != nil {
		return
	}
	proto.Attributes = make([]*catalogv1.ProductAttribute, 0, len(attrs))
	for i := range attrs {
		proto.Attributes = append(proto.Attributes, toProtoProductAttribute(&attrs[i]))
	}
}

// --- Supplier category mappings ---

func (s *CatalogGRPCServer) ListSupplierCategoryMappings(
	ctx context.Context,
	req *catalogv1.ListSupplierCategoryMappingsRequest,
) (*catalogv1.ListSupplierCategoryMappingsResponse, error) {
	if err := requireAdmin(ctx, s.jwtSecret); err != nil {
		return nil, mapServiceError(err)
	}
	list, err := s.catalogService.ListSupplierCategoryMappings(ctx, domain.SupplierCategoryMappingFilter{
		SupplierID: req.SupplierId, CategoryID: req.CategoryId,
	})
	if err != nil {
		return nil, mapServiceError(err)
	}
	out := make([]*catalogv1.SupplierCategoryMapping, 0, len(list))
	for i := range list {
		out = append(out, toProtoSupplierCategoryMapping(&list[i]))
	}
	return &catalogv1.ListSupplierCategoryMappingsResponse{Mappings: out}, nil
}

func (s *CatalogGRPCServer) GetSupplierCategoryMapping(
	ctx context.Context,
	req *catalogv1.GetSupplierCategoryMappingRequest,
) (*catalogv1.GetSupplierCategoryMappingResponse, error) {
	if err := requireAdmin(ctx, s.jwtSecret); err != nil {
		return nil, mapServiceError(err)
	}
	m, err := s.catalogService.GetSupplierCategoryMapping(ctx, req.Id)
	if err != nil {
		return nil, mapServiceError(err)
	}
	return &catalogv1.GetSupplierCategoryMappingResponse{Mapping: toProtoSupplierCategoryMapping(m)}, nil
}

func (s *CatalogGRPCServer) CreateSupplierCategoryMapping(
	ctx context.Context,
	req *catalogv1.CreateSupplierCategoryMappingRequest,
) (*catalogv1.CreateSupplierCategoryMappingResponse, error) {
	if err := requireAdmin(ctx, s.jwtSecret); err != nil {
		return nil, mapServiceError(err)
	}
	m, err := s.catalogService.CreateSupplierCategoryMapping(ctx, domain.SupplierCategoryMappingInput{
		CategoryID: req.CategoryId, SupplierID: req.SupplierId,
		ExternalID: req.ExternalId, ExternalName: req.ExternalName, Notes: req.Notes,
	})
	if err != nil {
		return nil, mapServiceError(err)
	}
	return &catalogv1.CreateSupplierCategoryMappingResponse{Mapping: toProtoSupplierCategoryMapping(m)}, nil
}

func (s *CatalogGRPCServer) UpdateSupplierCategoryMapping(
	ctx context.Context,
	req *catalogv1.UpdateSupplierCategoryMappingRequest,
) (*catalogv1.UpdateSupplierCategoryMappingResponse, error) {
	if err := requireAdmin(ctx, s.jwtSecret); err != nil {
		return nil, mapServiceError(err)
	}
	m, err := s.catalogService.UpdateSupplierCategoryMapping(ctx, req.Id, domain.SupplierCategoryMappingInput{
		CategoryID: req.CategoryId, SupplierID: req.SupplierId,
		ExternalID: req.ExternalId, ExternalName: req.ExternalName, Notes: req.Notes,
	})
	if err != nil {
		return nil, mapServiceError(err)
	}
	return &catalogv1.UpdateSupplierCategoryMappingResponse{Mapping: toProtoSupplierCategoryMapping(m)}, nil
}

func (s *CatalogGRPCServer) DeleteSupplierCategoryMapping(
	ctx context.Context,
	req *catalogv1.DeleteSupplierCategoryMappingRequest,
) (*catalogv1.DeleteSupplierCategoryMappingResponse, error) {
	if err := requireAdmin(ctx, s.jwtSecret); err != nil {
		return nil, mapServiceError(err)
	}
	if err := s.catalogService.DeleteSupplierCategoryMapping(ctx, req.Id); err != nil {
		return nil, mapServiceError(err)
	}
	return &catalogv1.DeleteSupplierCategoryMappingResponse{Success: true}, nil
}

func toProtoSupplierCategoryMapping(m *domain.SupplierCategoryMapping) *catalogv1.SupplierCategoryMapping {
	return &catalogv1.SupplierCategoryMapping{
		Id: m.ID, CategoryId: m.CategoryID, SupplierId: m.SupplierID,
		ExternalId: m.ExternalID, ExternalName: m.ExternalName, Notes: m.Notes,
		CreatedAt: m.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt: m.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

// --- Supplier product mappings ---

func (s *CatalogGRPCServer) ListSupplierProductMappings(
	ctx context.Context,
	req *catalogv1.ListSupplierProductMappingsRequest,
) (*catalogv1.ListSupplierProductMappingsResponse, error) {
	if err := requireAdmin(ctx, s.jwtSecret); err != nil {
		return nil, mapServiceError(err)
	}
	list, err := s.catalogService.ListSupplierProductMappings(ctx, domain.SupplierProductMappingFilter{
		SupplierID: req.SupplierId, ProductID: req.ProductId,
	})
	if err != nil {
		return nil, mapServiceError(err)
	}
	out := make([]*catalogv1.SupplierProductMapping, 0, len(list))
	for i := range list {
		out = append(out, toProtoSupplierProductMapping(&list[i]))
	}
	return &catalogv1.ListSupplierProductMappingsResponse{Mappings: out}, nil
}

func (s *CatalogGRPCServer) GetSupplierProductMapping(
	ctx context.Context,
	req *catalogv1.GetSupplierProductMappingRequest,
) (*catalogv1.GetSupplierProductMappingResponse, error) {
	if err := requireAdmin(ctx, s.jwtSecret); err != nil {
		return nil, mapServiceError(err)
	}
	m, err := s.catalogService.GetSupplierProductMapping(ctx, req.Id)
	if err != nil {
		return nil, mapServiceError(err)
	}
	return &catalogv1.GetSupplierProductMappingResponse{Mapping: toProtoSupplierProductMapping(m)}, nil
}

func (s *CatalogGRPCServer) CreateSupplierProductMapping(
	ctx context.Context,
	req *catalogv1.CreateSupplierProductMappingRequest,
) (*catalogv1.CreateSupplierProductMappingResponse, error) {
	if err := requireAdmin(ctx, s.jwtSecret); err != nil {
		return nil, mapServiceError(err)
	}
	m, err := s.catalogService.CreateSupplierProductMapping(ctx, domain.SupplierProductMappingInput{
		ProductID: req.ProductId, SupplierID: req.SupplierId,
		ExternalID: req.ExternalId, ExternalSKU: req.ExternalSku,
		ExternalName: req.ExternalName, Notes: req.Notes,
	})
	if err != nil {
		return nil, mapServiceError(err)
	}
	return &catalogv1.CreateSupplierProductMappingResponse{Mapping: toProtoSupplierProductMapping(m)}, nil
}

func (s *CatalogGRPCServer) UpdateSupplierProductMapping(
	ctx context.Context,
	req *catalogv1.UpdateSupplierProductMappingRequest,
) (*catalogv1.UpdateSupplierProductMappingResponse, error) {
	if err := requireAdmin(ctx, s.jwtSecret); err != nil {
		return nil, mapServiceError(err)
	}
	m, err := s.catalogService.UpdateSupplierProductMapping(ctx, req.Id, domain.SupplierProductMappingInput{
		ProductID: req.ProductId, SupplierID: req.SupplierId,
		ExternalID: req.ExternalId, ExternalSKU: req.ExternalSku,
		ExternalName: req.ExternalName, Notes: req.Notes,
	})
	if err != nil {
		return nil, mapServiceError(err)
	}
	return &catalogv1.UpdateSupplierProductMappingResponse{Mapping: toProtoSupplierProductMapping(m)}, nil
}

func (s *CatalogGRPCServer) DeleteSupplierProductMapping(
	ctx context.Context,
	req *catalogv1.DeleteSupplierProductMappingRequest,
) (*catalogv1.DeleteSupplierProductMappingResponse, error) {
	if err := requireAdmin(ctx, s.jwtSecret); err != nil {
		return nil, mapServiceError(err)
	}
	if err := s.catalogService.DeleteSupplierProductMapping(ctx, req.Id); err != nil {
		return nil, mapServiceError(err)
	}
	return &catalogv1.DeleteSupplierProductMappingResponse{Success: true}, nil
}

func toProtoSupplierProductMapping(m *domain.SupplierProductMapping) *catalogv1.SupplierProductMapping {
	return &catalogv1.SupplierProductMapping{
		Id: m.ID, ProductId: m.ProductID, SupplierId: m.SupplierID,
		ExternalId: m.ExternalID, ExternalSku: m.ExternalSKU, ExternalName: m.ExternalName, Notes: m.Notes,
		CreatedAt: m.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt: m.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

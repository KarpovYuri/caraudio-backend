package grpc

import (
	"context"
	"time"

	"github.com/KarpovYuri/caraudio-backend/internal/catalog/domain"
	catalogv1 "github.com/KarpovYuri/caraudio-backend/pkg/api/proto/catalog/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *CatalogGRPCServer) ListSuppliers(
	ctx context.Context,
	req *catalogv1.ListSuppliersRequest,
) (*catalogv1.ListSuppliersResponse, error) {
	if err := requireAdmin(ctx, s.jwtSecret); err != nil {
		return nil, mapServiceError(err)
	}

	result, err := s.catalogService.ListSuppliers(ctx)
	if err != nil {
		return nil, mapServiceError(err)
	}

	suppliers := make([]*catalogv1.Supplier, 0, len(result))
	for i := range result {
		suppliers = append(suppliers, toProtoSupplier(&result[i]))
	}

	return &catalogv1.ListSuppliersResponse{
		Suppliers: suppliers,
		Total:     int32(len(suppliers)),
	}, nil
}

func (s *CatalogGRPCServer) GetSupplier(
	ctx context.Context,
	req *catalogv1.GetSupplierRequest,
) (*catalogv1.GetSupplierResponse, error) {
	if err := requireAdmin(ctx, s.jwtSecret); err != nil {
		return nil, mapServiceError(err)
	}
	if req.Id == 0 {
		return nil, status.Error(codes.InvalidArgument, "supplier id is required")
	}
	supplier, err := s.catalogService.GetSupplier(ctx, req.Id)
	if err != nil {
		return nil, mapServiceError(err)
	}
	return &catalogv1.GetSupplierResponse{Supplier: toProtoSupplier(supplier)}, nil
}

func (s *CatalogGRPCServer) CreateSupplier(
	ctx context.Context,
	req *catalogv1.CreateSupplierRequest,
) (*catalogv1.CreateSupplierResponse, error) {
	if err := requireAdmin(ctx, s.jwtSecret); err != nil {
		return nil, mapServiceError(err)
	}
	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "supplier name is required")
	}
	supplier, err := s.catalogService.CreateSupplier(ctx, supplierInputFromCreate(req))
	if err != nil {
		return nil, mapServiceError(err)
	}
	return &catalogv1.CreateSupplierResponse{Supplier: toProtoSupplier(supplier)}, nil
}

func (s *CatalogGRPCServer) UpdateSupplier(
	ctx context.Context,
	req *catalogv1.UpdateSupplierRequest,
) (*catalogv1.UpdateSupplierResponse, error) {
	if err := requireAdmin(ctx, s.jwtSecret); err != nil {
		return nil, mapServiceError(err)
	}
	if req.Id == 0 {
		return nil, status.Error(codes.InvalidArgument, "supplier id is required")
	}
	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "supplier name is required")
	}
	supplier, err := s.catalogService.UpdateSupplier(ctx, req.Id, supplierInputFromUpdate(req))
	if err != nil {
		return nil, mapServiceError(err)
	}
	return &catalogv1.UpdateSupplierResponse{Supplier: toProtoSupplier(supplier)}, nil
}

func (s *CatalogGRPCServer) DeleteSupplier(
	ctx context.Context,
	req *catalogv1.DeleteSupplierRequest,
) (*catalogv1.DeleteSupplierResponse, error) {
	if err := requireAdmin(ctx, s.jwtSecret); err != nil {
		return nil, mapServiceError(err)
	}
	if req.Id == 0 {
		return nil, status.Error(codes.InvalidArgument, "supplier id is required")
	}
	if err := s.catalogService.DeleteSupplier(ctx, req.Id); err != nil {
		return nil, mapServiceError(err)
	}
	return &catalogv1.DeleteSupplierResponse{Success: true}, nil
}

func supplierInputFromCreate(req *catalogv1.CreateSupplierRequest) domain.SupplierInput {
	return domain.SupplierInput{
		Name:     req.Name,
		Code:     req.Code,
		Logo:     req.Logo,
		ApiUrl:   req.ApiUrl,
		IsActive: req.IsActive,
	}
}

func supplierInputFromUpdate(req *catalogv1.UpdateSupplierRequest) domain.SupplierInput {
	return domain.SupplierInput{
		Name:     req.Name,
		Code:     req.Code,
		Logo:     req.Logo,
		ApiUrl:   req.ApiUrl,
		IsActive: req.IsActive,
	}
}

func toProtoSupplier(supplier *domain.Supplier) *catalogv1.Supplier {
	out := &catalogv1.Supplier{
		Id:        supplier.ID,
		Name:      supplier.Name,
		Logo:      supplier.Logo,
		ApiUrl:    supplier.ApiUrl,
		IsActive:  supplier.IsActive,
		CreatedAt: supplier.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt: supplier.UpdatedAt.UTC().Format(time.RFC3339),
	}
	if supplier.Code != nil {
		out.Code = *supplier.Code
	}
	return out
}

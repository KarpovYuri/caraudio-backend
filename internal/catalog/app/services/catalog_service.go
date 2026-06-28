package services

import (
	"context"
	"strings"
	"time"

	"github.com/KarpovYuri/caraudio-backend/internal/catalog/domain"
	"github.com/KarpovYuri/caraudio-backend/internal/catalog/infrastructure/database/postgres"
	"github.com/google/uuid"
)

type CatalogService interface {
	ListSuppliers(ctx context.Context) ([]domain.Supplier, error)
	GetSupplier(ctx context.Context, id int64) (*domain.Supplier, error)
	CreateSupplier(ctx context.Context, input domain.SupplierInput) (*domain.Supplier, error)
	UpdateSupplier(ctx context.Context, id int64, input domain.SupplierInput) (*domain.Supplier, error)
	DeleteSupplier(ctx context.Context, id int64) error

	ListCategories(ctx context.Context) ([]domain.Category, error)
	GetCategory(ctx context.Context, id string) (*domain.Category, error)
	CreateCategory(ctx context.Context, name, slug, parentID string) (*domain.Category, error)
	UpdateCategory(ctx context.Context, id, name, slug, parentID string) (*domain.Category, error)
	DeleteCategory(ctx context.Context, id string) error

	ListProducts(ctx context.Context, filter domain.ProductListFilter) (*domain.ProductListResult, error)
	GetProduct(ctx context.Context, id string) (*domain.Product, error)
	GetProductByID(ctx context.Context, id string) (*domain.Product, error)
	CreateProduct(
		ctx context.Context,
		categoryID, brandID, name, description string,
		supplierID int64,
		priceCents int64,
		sku string,
		stock int32,
		isActive bool,
	) (*domain.Product, error)
	UpdateProduct(
		ctx context.Context,
		id, categoryID, brandID, name, description string,
		supplierID int64,
		priceCents int64,
		sku string,
		stock int32,
		isActive bool,
	) (*domain.Product, error)
	DeleteProduct(ctx context.Context, id string) error

	ListProductImages(ctx context.Context, productID string, adminAccess bool) ([]domain.ProductImage, error)
	GetProductImage(ctx context.Context, productID, imageID string, adminAccess bool) (*domain.ProductImage, error)
	CreateProductImage(ctx context.Context, productID string, input domain.ProductImageInput) (*domain.ProductImage, error)
	UpdateProductImage(ctx context.Context, productID, imageID string, input domain.ProductImageInput) (*domain.ProductImage, error)
	DeleteProductImage(ctx context.Context, productID, imageID string) error

	ListBrands(ctx context.Context, activeOnly bool) ([]domain.Brand, error)
	GetBrand(ctx context.Context, id string) (*domain.Brand, error)
	GetBrandByID(ctx context.Context, id string) (*domain.Brand, error)
	CreateBrand(ctx context.Context, input domain.BrandInput) (*domain.Brand, error)
	UpdateBrand(ctx context.Context, id string, input domain.BrandInput) (*domain.Brand, error)
	DeleteBrand(ctx context.Context, id string) error

	ListProductAttributes(ctx context.Context, productID string, adminAccess bool) ([]domain.ProductAttribute, error)
	GetProductAttribute(ctx context.Context, productID, attrID string, adminAccess bool) (*domain.ProductAttribute, error)
	CreateProductAttribute(ctx context.Context, productID string, input domain.ProductAttributeInput) (*domain.ProductAttribute, error)
	UpdateProductAttribute(ctx context.Context, productID, attrID string, input domain.ProductAttributeInput) (*domain.ProductAttribute, error)
	DeleteProductAttribute(ctx context.Context, productID, attrID string) error

	NormalizePagination(page, pageSize, defaultSize, maxSize int32) (int32, int32)
}

type catalogService struct {
	suppliers         postgres.SupplierRepository
	categories        postgres.CategoryRepository
	products          postgres.ProductRepository
	brands            postgres.BrandRepository
	productImages     postgres.ProductImageRepository
	productAttributes postgres.ProductAttributeRepository
}

func NewCatalogService(
	suppliers postgres.SupplierRepository,
	categories postgres.CategoryRepository,
	products postgres.ProductRepository,
	brands postgres.BrandRepository,
	productImages postgres.ProductImageRepository,
	productAttributes postgres.ProductAttributeRepository,
) CatalogService {
	return &catalogService{
		suppliers:         suppliers,
		categories:        categories,
		products:          products,
		brands:            brands,
		productImages:     productImages,
		productAttributes: productAttributes,
	}
}

func (s *catalogService) ListCategories(ctx context.Context) ([]domain.Category, error) {
	return s.categories.List(ctx)
}

func (s *catalogService) GetCategory(ctx context.Context, id string) (*domain.Category, error) {
	return s.categories.GetByID(ctx, id)
}

func (s *catalogService) CreateCategory(
	ctx context.Context,
	name, slug, parentID string,
) (*domain.Category, error) {
	name = strings.TrimSpace(name)
	slug = strings.TrimSpace(slug)
	if name == "" || slug == "" {
		return nil, domain.ErrInvalidArgument
	}

	if parentID != "" {
		if _, err := s.categories.GetByID(ctx, parentID); err != nil {
			return nil, err
		}
	}

	now := time.Now()
	category := &domain.Category{
		ID:        uuid.NewString(),
		Name:      name,
		Slug:      slug,
		ParentID:  stringPtrOrNil(parentID),
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.categories.Create(ctx, category); err != nil {
		return nil, err
	}
	return category, nil
}

func (s *catalogService) UpdateCategory(
	ctx context.Context,
	id, name, slug, parentID string,
) (*domain.Category, error) {
	category, err := s.categories.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if name == "" && slug == "" && parentID == "" {
		return nil, domain.ErrInvalidArgument
	}

	if name != "" {
		category.Name = strings.TrimSpace(name)
	}
	if slug != "" {
		category.Slug = strings.TrimSpace(slug)
	}
	if parentID != "" {
		if parentID == id {
			return nil, domain.ErrInvalidArgument
		}
		if _, err := s.categories.GetByID(ctx, parentID); err != nil {
			return nil, err
		}
		category.ParentID = stringPtrOrNil(parentID)
	}

	category.UpdatedAt = time.Now()
	if err := s.categories.Update(ctx, category); err != nil {
		return nil, err
	}
	return category, nil
}

func (s *catalogService) DeleteCategory(ctx context.Context, id string) error {
	count, err := s.categories.CountProducts(ctx, id)
	if err != nil {
		return err
	}
	if count > 0 {
		return domain.ErrCategoryHasProducts
	}
	return s.categories.Delete(ctx, id)
}

func (s *catalogService) ListProducts(
	ctx context.Context,
	filter domain.ProductListFilter,
) (*domain.ProductListResult, error) {
	return s.products.List(ctx, filter)
}

func (s *catalogService) GetProduct(ctx context.Context, id string) (*domain.Product, error) {
	product, err := s.products.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if !product.IsActive {
		return nil, domain.ErrProductNotFound
	}
	return product, nil
}

func (s *catalogService) GetProductByID(ctx context.Context, id string) (*domain.Product, error) {
	return s.products.GetByID(ctx, id)
}

func (s *catalogService) CreateProduct(
	ctx context.Context,
	categoryID, brandID, name, description string,
	supplierID int64,
	priceCents int64,
	sku string,
	stock int32,
	isActive bool,
) (*domain.Product, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, domain.ErrInvalidArgument
	}
	if priceCents < 0 || stock < 0 {
		return nil, domain.ErrInvalidArgument
	}

	if categoryID != "" {
		if _, err := s.categories.GetByID(ctx, categoryID); err != nil {
			return nil, err
		}
	}
	if err := s.validateSupplierID(ctx, supplierID); err != nil {
		return nil, err
	}
	if err := s.validateBrandID(ctx, brandID); err != nil {
		return nil, err
	}

	now := time.Now()
	product := &domain.Product{
		ID:          uuid.NewString(),
		CategoryID:  stringPtrOrNil(categoryID),
		BrandID:     stringPtrOrNil(brandID),
		SupplierID:  int64PtrOrNil(supplierID),
		Name:        name,
		Description: description,
		PriceCents:  priceCents,
		SKU:         stringPtrOrNil(strings.TrimSpace(sku)),
		Stock:       stock,
		IsActive:    isActive,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.products.Create(ctx, product); err != nil {
		return nil, err
	}
	return product, nil
}

func (s *catalogService) UpdateProduct(
	ctx context.Context,
	id, categoryID, brandID, name, description string,
	supplierID int64,
	priceCents int64,
	sku string,
	stock int32,
	isActive bool,
) (*domain.Product, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, domain.ErrInvalidArgument
	}
	if priceCents < 0 || stock < 0 {
		return nil, domain.ErrInvalidArgument
	}

	if _, err := s.products.GetByID(ctx, id); err != nil {
		return nil, err
	}

	if categoryID != "" {
		if _, err := s.categories.GetByID(ctx, categoryID); err != nil {
			return nil, err
		}
	}
	if err := s.validateSupplierID(ctx, supplierID); err != nil {
		return nil, err
	}
	if err := s.validateBrandID(ctx, brandID); err != nil {
		return nil, err
	}

	now := time.Now()
	product := &domain.Product{
		ID:          id,
		CategoryID:  stringPtrOrNil(categoryID),
		BrandID:     stringPtrOrNil(brandID),
		SupplierID:  int64PtrOrNil(supplierID),
		Name:        name,
		Description: description,
		PriceCents:  priceCents,
		SKU:         stringPtrOrNil(strings.TrimSpace(sku)),
		Stock:       stock,
		IsActive:    isActive,
		UpdatedAt:   now,
	}

	if err := s.products.Update(ctx, product); err != nil {
		return nil, err
	}
	return s.products.GetByID(ctx, id)
}

func (s *catalogService) DeleteProduct(ctx context.Context, id string) error {
	return s.products.Delete(ctx, id)
}

func int64PtrOrNil(value int64) *int64 {
	if value == 0 {
		return nil
	}
	v := value
	return &v
}

func stringPtrOrNil(value string) *string {
	if value == "" {
		return nil
	}
	v := value
	return &v
}

func (s *catalogService) NormalizePagination(page, pageSize, defaultSize, maxSize int32) (int32, int32) {
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = defaultSize
	}
	if pageSize > maxSize {
		pageSize = maxSize
	}
	return page, pageSize
}

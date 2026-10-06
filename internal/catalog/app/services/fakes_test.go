package services

import (
	"context"
	"errors"

	"github.com/KarpovYuri/caraudio-backend/internal/catalog/domain"
)

type fakeSupplierRepo struct {
	createFn  func(ctx context.Context, supplier *domain.Supplier) error
	getByIDFn func(ctx context.Context, id int64) (*domain.Supplier, error)
	listFn    func(ctx context.Context, filter domain.SupplierListFilter) (*domain.SupplierListResult, error)
	updateFn  func(ctx context.Context, supplier *domain.Supplier) error
	deleteFn  func(ctx context.Context, id int64) error
}

func (f *fakeSupplierRepo) Create(ctx context.Context, supplier *domain.Supplier) error {
	if f.createFn == nil {
		return nil
	}
	return f.createFn(ctx, supplier)
}

func (f *fakeSupplierRepo) GetByID(ctx context.Context, id int64) (*domain.Supplier, error) {
	if f.getByIDFn == nil {
		return nil, errors.New("getByIDFn is not set")
	}
	return f.getByIDFn(ctx, id)
}

func (f *fakeSupplierRepo) List(
	ctx context.Context,
	filter domain.SupplierListFilter,
) (*domain.SupplierListResult, error) {
	if f.listFn == nil {
		return &domain.SupplierListResult{}, nil
	}
	return f.listFn(ctx, filter)
}

func (f *fakeSupplierRepo) Update(ctx context.Context, supplier *domain.Supplier) error {
	if f.updateFn == nil {
		return nil
	}
	return f.updateFn(ctx, supplier)
}

func (f *fakeSupplierRepo) Delete(ctx context.Context, id int64) error {
	if f.deleteFn == nil {
		return nil
	}
	return f.deleteFn(ctx, id)
}

type fakeCategoryRepo struct {
	createFn        func(ctx context.Context, category *domain.Category) error
	getByIDFn       func(ctx context.Context, id string) (*domain.Category, error)
	listFn          func(ctx context.Context) ([]domain.Category, error)
	updateFn        func(ctx context.Context, category *domain.Category) error
	deleteFn        func(ctx context.Context, id string) error
	countProductsFn func(ctx context.Context, categoryID string) (int64, error)
}

func (f *fakeCategoryRepo) Create(ctx context.Context, category *domain.Category) error {
	if f.createFn == nil {
		return nil
	}
	return f.createFn(ctx, category)
}

func (f *fakeCategoryRepo) GetByID(ctx context.Context, id string) (*domain.Category, error) {
	if f.getByIDFn == nil {
		return nil, errors.New("getByIDFn is not set")
	}
	return f.getByIDFn(ctx, id)
}

func (f *fakeCategoryRepo) List(ctx context.Context) ([]domain.Category, error) {
	if f.listFn == nil {
		return nil, nil
	}
	return f.listFn(ctx)
}

func (f *fakeCategoryRepo) Update(ctx context.Context, category *domain.Category) error {
	if f.updateFn == nil {
		return nil
	}
	return f.updateFn(ctx, category)
}

func (f *fakeCategoryRepo) Delete(ctx context.Context, id string) error {
	if f.deleteFn == nil {
		return nil
	}
	return f.deleteFn(ctx, id)
}

func (f *fakeCategoryRepo) CountProducts(ctx context.Context, categoryID string) (int64, error) {
	if f.countProductsFn == nil {
		return 0, nil
	}
	return f.countProductsFn(ctx, categoryID)
}

type fakeProductRepo struct {
	createFn          func(ctx context.Context, product *domain.Product) error
	getByIDFn         func(ctx context.Context, id string) (*domain.Product, error)
	listFn            func(ctx context.Context, filter domain.ProductListFilter) (*domain.ProductListResult, error)
	updateFn          func(ctx context.Context, product *domain.Product) error
	deleteFn          func(ctx context.Context, id string) error
	countBySupplierFn func(ctx context.Context, supplierID int64) (int64, error)
	countByBrandFn    func(ctx context.Context, brandID string) (int64, error)
}

func (f *fakeProductRepo) Create(ctx context.Context, product *domain.Product) error {
	if f.createFn == nil {
		return nil
	}
	return f.createFn(ctx, product)
}

func (f *fakeProductRepo) GetByID(ctx context.Context, id string) (*domain.Product, error) {
	if f.getByIDFn == nil {
		return nil, errors.New("getByIDFn is not set")
	}
	return f.getByIDFn(ctx, id)
}

func (f *fakeProductRepo) List(
	ctx context.Context,
	filter domain.ProductListFilter,
) (*domain.ProductListResult, error) {
	if f.listFn == nil {
		return &domain.ProductListResult{}, nil
	}
	return f.listFn(ctx, filter)
}

func (f *fakeProductRepo) Update(ctx context.Context, product *domain.Product) error {
	if f.updateFn == nil {
		return nil
	}
	return f.updateFn(ctx, product)
}

func (f *fakeProductRepo) Delete(ctx context.Context, id string) error {
	if f.deleteFn == nil {
		return nil
	}
	return f.deleteFn(ctx, id)
}

func (f *fakeProductRepo) CountBySupplier(ctx context.Context, supplierID int64) (int64, error) {
	if f.countBySupplierFn == nil {
		return 0, nil
	}
	return f.countBySupplierFn(ctx, supplierID)
}

func (f *fakeProductRepo) CountByBrand(ctx context.Context, brandID string) (int64, error) {
	if f.countByBrandFn == nil {
		return 0, nil
	}
	return f.countByBrandFn(ctx, brandID)
}

type fakeBrandRepo struct {
	createFn  func(ctx context.Context, brand *domain.Brand) error
	getByIDFn func(ctx context.Context, id string) (*domain.Brand, error)
	listFn    func(ctx context.Context, activeOnly bool) ([]domain.Brand, error)
	updateFn  func(ctx context.Context, brand *domain.Brand) error
	deleteFn  func(ctx context.Context, id string) error
}

func (f *fakeBrandRepo) Create(ctx context.Context, brand *domain.Brand) error {
	if f.createFn == nil {
		return nil
	}
	return f.createFn(ctx, brand)
}

func (f *fakeBrandRepo) GetByID(ctx context.Context, id string) (*domain.Brand, error) {
	if f.getByIDFn == nil {
		return nil, errors.New("getByIDFn is not set")
	}
	return f.getByIDFn(ctx, id)
}

func (f *fakeBrandRepo) List(ctx context.Context, activeOnly bool) ([]domain.Brand, error) {
	if f.listFn == nil {
		return nil, nil
	}
	return f.listFn(ctx, activeOnly)
}

func (f *fakeBrandRepo) Update(ctx context.Context, brand *domain.Brand) error {
	if f.updateFn == nil {
		return nil
	}
	return f.updateFn(ctx, brand)
}

func (f *fakeBrandRepo) Delete(ctx context.Context, id string) error {
	if f.deleteFn == nil {
		return nil
	}
	return f.deleteFn(ctx, id)
}

type fakeProductImageRepo struct {
	createFn        func(ctx context.Context, image *domain.ProductImage) error
	getByIDFn       func(ctx context.Context, id string) (*domain.ProductImage, error)
	listByProductFn func(ctx context.Context, productID string) ([]domain.ProductImage, error)
	updateFn        func(ctx context.Context, image *domain.ProductImage) error
	deleteFn        func(ctx context.Context, id string) error
	unsetPrimaryFn  func(ctx context.Context, productID, exceptID string) error
}

func (f *fakeProductImageRepo) Create(ctx context.Context, image *domain.ProductImage) error {
	if f.createFn == nil {
		return nil
	}
	return f.createFn(ctx, image)
}

func (f *fakeProductImageRepo) GetByID(ctx context.Context, id string) (*domain.ProductImage, error) {
	if f.getByIDFn == nil {
		return nil, errors.New("getByIDFn is not set")
	}
	return f.getByIDFn(ctx, id)
}

func (f *fakeProductImageRepo) ListByProductID(ctx context.Context, productID string) ([]domain.ProductImage, error) {
	if f.listByProductFn == nil {
		return nil, nil
	}
	return f.listByProductFn(ctx, productID)
}

func (f *fakeProductImageRepo) Update(ctx context.Context, image *domain.ProductImage) error {
	if f.updateFn == nil {
		return nil
	}
	return f.updateFn(ctx, image)
}

func (f *fakeProductImageRepo) Delete(ctx context.Context, id string) error {
	if f.deleteFn == nil {
		return nil
	}
	return f.deleteFn(ctx, id)
}

func (f *fakeProductImageRepo) UnsetPrimaryForProduct(ctx context.Context, productID, exceptID string) error {
	if f.unsetPrimaryFn == nil {
		return nil
	}
	return f.unsetPrimaryFn(ctx, productID, exceptID)
}

type fakeProductAttrRepo struct{}

func (f *fakeProductAttrRepo) Create(context.Context, *domain.ProductAttribute) error {
	return nil
}
func (f *fakeProductAttrRepo) GetByID(context.Context, string) (*domain.ProductAttribute, error) {
	return nil, domain.ErrProductAttributeNotFound
}
func (f *fakeProductAttrRepo) ListByProductID(context.Context, string) ([]domain.ProductAttribute, error) {
	return nil, nil
}
func (f *fakeProductAttrRepo) Update(context.Context, *domain.ProductAttribute) error {
	return nil
}
func (f *fakeProductAttrRepo) Delete(context.Context, string) error { return nil }

type fakeCategoryMappingRepo struct{}

func (f *fakeCategoryMappingRepo) Create(context.Context, *domain.SupplierCategoryMapping) error {
	return nil
}
func (f *fakeCategoryMappingRepo) GetByID(context.Context, string) (*domain.SupplierCategoryMapping, error) {
	return nil, domain.ErrSupplierMappingNotFound
}
func (f *fakeCategoryMappingRepo) List(context.Context, domain.SupplierCategoryMappingFilter) ([]domain.SupplierCategoryMapping, error) {
	return nil, nil
}
func (f *fakeCategoryMappingRepo) Update(context.Context, *domain.SupplierCategoryMapping) error {
	return nil
}
func (f *fakeCategoryMappingRepo) Delete(context.Context, string) error { return nil }

type fakeProductMappingRepo struct{}

func (f *fakeProductMappingRepo) Create(context.Context, *domain.SupplierProductMapping) error {
	return nil
}
func (f *fakeProductMappingRepo) GetByID(context.Context, string) (*domain.SupplierProductMapping, error) {
	return nil, domain.ErrSupplierMappingNotFound
}
func (f *fakeProductMappingRepo) List(context.Context, domain.SupplierProductMappingFilter) ([]domain.SupplierProductMapping, error) {
	return nil, nil
}
func (f *fakeProductMappingRepo) Update(context.Context, *domain.SupplierProductMapping) error {
	return nil
}
func (f *fakeProductMappingRepo) Delete(context.Context, string) error { return nil }

func newTestCatalogService(
	suppliers *fakeSupplierRepo,
	categories *fakeCategoryRepo,
	products *fakeProductRepo,
	brands *fakeBrandRepo,
	images *fakeProductImageRepo,
) CatalogService {
	if suppliers == nil {
		suppliers = &fakeSupplierRepo{}
	}
	if categories == nil {
		categories = &fakeCategoryRepo{}
	}
	if products == nil {
		products = &fakeProductRepo{}
	}
	if brands == nil {
		brands = &fakeBrandRepo{}
	}
	if images == nil {
		images = &fakeProductImageRepo{}
	}
	return NewCatalogService(
		suppliers,
		categories,
		products,
		brands,
		images,
		&fakeProductAttrRepo{},
		&fakeCategoryMappingRepo{},
		&fakeProductMappingRepo{},
	)
}

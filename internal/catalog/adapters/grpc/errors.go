package grpc

import (
	"errors"

	"github.com/KarpovYuri/caraudio-backend/internal/catalog/domain"
	pkgjwt "github.com/KarpovYuri/caraudio-backend/pkg/jwt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func mapServiceError(err error) error {
	switch {
	case errors.Is(err, pkgjwt.ErrUnauthorized):
		return status.Error(codes.Unauthenticated, "unauthorized")
	case errors.Is(err, pkgjwt.ErrForbidden):
		return status.Error(codes.PermissionDenied, "forbidden")
	case errors.Is(err, domain.ErrInvalidArgument):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, domain.ErrCategoryNotFound),
		errors.Is(err, domain.ErrProductNotFound),
		errors.Is(err, domain.ErrSupplierNotFound),
		errors.Is(err, domain.ErrProductImageNotFound),
		errors.Is(err, domain.ErrBrandNotFound),
		errors.Is(err, domain.ErrProductAttributeNotFound),
		errors.Is(err, domain.ErrSupplierMappingNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, domain.ErrAlreadyExists):
		return status.Error(codes.AlreadyExists, err.Error())
	case errors.Is(err, domain.ErrCategoryHasProducts),
		errors.Is(err, domain.ErrSupplierHasProducts),
		errors.Is(err, domain.ErrBrandHasProducts):
		return status.Error(codes.FailedPrecondition, err.Error())
	default:
		return status.Error(codes.Internal, "internal server error")
	}
}

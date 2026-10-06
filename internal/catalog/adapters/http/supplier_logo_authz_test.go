package httpadapter

import (
	"errors"
	"net/http"
	"testing"
	"time"

	pkgjwt "github.com/KarpovYuri/caraudio-backend/pkg/jwt"
)

func TestRequireAdminHTTP(t *testing.T) {
	const secret = "catalog-secret"

	adminToken, err := pkgjwt.GenerateToken("admin-1", pkgjwt.RoleAdmin, secret, time.Minute)
	if err != nil {
		t.Fatalf("GenerateToken: %v", err)
	}
	userToken, err := pkgjwt.GenerateToken("user-1", "user", secret, time.Minute)
	if err != nil {
		t.Fatalf("GenerateToken: %v", err)
	}

	tests := []struct {
		name       string
		authHeader string
		wantErr    error
	}{
		{name: "missing", wantErr: pkgjwt.ErrUnauthorized},
		{name: "invalid", authHeader: "Bearer bad", wantErr: pkgjwt.ErrUnauthorized},
		{name: "user", authHeader: "Bearer " + userToken, wantErr: pkgjwt.ErrForbidden},
		{name: "admin", authHeader: "Bearer " + adminToken, wantErr: nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodPost, "/v1/suppliers/1/logo", nil)
			if err != nil {
				t.Fatalf("NewRequest: %v", err)
			}
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			got := requireAdminHTTP(req, secret)
			if tt.wantErr == nil {
				if got != nil {
					t.Fatalf("expected nil, got %v", got)
				}
				return
			}
			if !errors.Is(got, tt.wantErr) {
				t.Fatalf("expected %v, got %v", tt.wantErr, got)
			}
		})
	}
}

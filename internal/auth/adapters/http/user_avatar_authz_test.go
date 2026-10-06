package httpadapter

import (
	"errors"
	"net/http"
	"testing"
	"time"

	pkgjwt "github.com/KarpovYuri/caraudio-backend/pkg/jwt"
)

func TestRequireSelfOrAdminHTTP(t *testing.T) {
	const secret = "test-secret"

	adminToken, err := pkgjwt.GenerateToken("admin-1", pkgjwt.RoleAdmin, secret, time.Minute)
	if err != nil {
		t.Fatalf("GenerateToken admin: %v", err)
	}
	userToken, err := pkgjwt.GenerateToken("user-1", "user", secret, time.Minute)
	if err != nil {
		t.Fatalf("GenerateToken user: %v", err)
	}

	tests := []struct {
		name         string
		authHeader   string
		targetUserID string
		wantErr      error
	}{
		{
			name:         "missing token",
			targetUserID: "user-1",
			wantErr:      pkgjwt.ErrUnauthorized,
		},
		{
			name:         "invalid token",
			authHeader:   "Bearer not-a-jwt",
			targetUserID: "user-1",
			wantErr:      pkgjwt.ErrUnauthorized,
		},
		{
			name:         "admin can update any user",
			authHeader:   "Bearer " + adminToken,
			targetUserID: "someone-else",
			wantErr:      nil,
		},
		{
			name:         "user can update self",
			authHeader:   "Bearer " + userToken,
			targetUserID: "user-1",
			wantErr:      nil,
		},
		{
			name:         "user cannot update other",
			authHeader:   "Bearer " + userToken,
			targetUserID: "other-user",
			wantErr:      pkgjwt.ErrForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodPost, "/v1/users/"+tt.targetUserID+"/avatar", nil)
			if err != nil {
				t.Fatalf("NewRequest: %v", err)
			}
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			got := requireSelfOrAdminHTTP(req, secret, tt.targetUserID)
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

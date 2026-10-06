package httpadapter

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/KarpovYuri/caraudio-backend/internal/auth/app/services"
	"github.com/KarpovYuri/caraudio-backend/internal/auth/domain"
	"github.com/KarpovYuri/caraudio-backend/internal/auth/infrastructure/storage"
	"github.com/KarpovYuri/caraudio-backend/pkg/imageutil"
	pkgjwt "github.com/KarpovYuri/caraudio-backend/pkg/jwt"
)

type UserAvatarHandler struct {
	users     services.UserService
	files     *storage.LocalStorage
	jwtSecret string
	maxBytes  int64
	maxSide   int
}

func NewUserAvatarHandler(
	users services.UserService,
	files *storage.LocalStorage,
	jwtSecret string,
	maxBytes int64,
	maxSide int,
) *UserAvatarHandler {
	if maxBytes <= 0 {
		maxBytes = 10 << 20
	}
	if maxSide <= 0 {
		maxSide = imageutil.DefaultMaxLogoSide
	}
	return &UserAvatarHandler{
		users:     users,
		files:     files,
		jwtSecret: jwtSecret,
		maxBytes:  maxBytes,
		maxSide:   maxSide,
	}
}

func (h *UserAvatarHandler) Upload(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(r.PathValue("id"))
	if id == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": "user id is required"})
		return
	}

	if err := requireSelfOrAdminHTTP(r, h.jwtSecret, id); err != nil {
		writeError(w, err)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, h.maxBytes+1024)
	if err := r.ParseMultipartForm(h.maxBytes + 1024); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": "invalid or too large multipart body"})
		return
	}

	file, header, err := r.FormFile("avatar")
	if err != nil {
		file, header, err = r.FormFile("file")
	}
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": "avatar file is required (field: avatar)"})
		return
	}
	defer file.Close()

	if err := validateImageUpload(header.Filename, file); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": err.Error()})
		return
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": "failed to read avatar file"})
		return
	}

	limited := &io.LimitedReader{R: file, N: h.maxBytes + 1}
	raw, err := io.ReadAll(limited)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": "failed to read avatar file"})
		return
	}
	if limited.N == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": "avatar file is too large"})
		return
	}

	var (
		processed []byte
		ext       string
	)
	if imageutil.IsSVG(header.Filename, raw) {
		processed, err = imageutil.SanitizeSVG(raw, int(h.maxBytes))
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"message": "invalid or unsafe svg avatar"})
			return
		}
		ext = ".svg"
	} else {
		processed, err = imageutil.ProcessLogo(bytes.NewReader(raw), h.maxSide)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"message": "only jpeg, png, webp, gif and svg avatars are allowed"})
			return
		}
		ext = ".webp"
	}

	avatarURL, err := h.files.SaveUserAvatar(ext, bytes.NewReader(processed))
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"message": "failed to save avatar"})
		return
	}

	existing, err := h.users.GetUser(r.Context(), id)
	if err != nil {
		_ = h.files.DeletePublicURL(avatarURL)
		writeError(w, err)
		return
	}

	user, err := h.users.UpdateUserAvatar(r.Context(), id, avatarURL)
	if err != nil {
		_ = h.files.DeletePublicURL(avatarURL)
		writeError(w, err)
		return
	}

	if existing.Avatar != "" && existing.Avatar != avatarURL {
		_ = h.files.DeletePublicURL(existing.Avatar)
	}

	writeJSON(w, http.StatusOK, map[string]any{"user": toUserJSON(user)})
}

func validateImageUpload(filename string, r io.ReadSeeker) error {
	buf := make([]byte, 512)
	n, err := r.Read(buf)
	if err != nil && !errors.Is(err, io.EOF) {
		return errors.New("failed to read avatar file")
	}
	contentType := http.DetectContentType(buf[:n])
	ext := strings.ToLower(filepath.Ext(filename))
	sample := buf[:n]

	if imageutil.IsSVG(filename, sample) {
		return nil
	}

	switch contentType {
	case "image/jpeg", "image/png", "image/gif", "image/webp":
		return nil
	}

	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".webp", ".svg":
		return nil
	default:
		return errors.New("only jpeg, png, webp, gif and svg avatars are allowed")
	}
}

func requireSelfOrAdminHTTP(r *http.Request, jwtSecret, targetUserID string) error {
	auth := strings.TrimSpace(r.Header.Get("Authorization"))
	token := auth
	if len(auth) > 7 && strings.EqualFold(auth[:7], "bearer ") {
		token = strings.TrimSpace(auth[7:])
	}
	if token == "" {
		return pkgjwt.ErrUnauthorized
	}

	claims, err := pkgjwt.ParseToken(token, jwtSecret)
	if err != nil {
		return pkgjwt.ErrUnauthorized
	}
	if claims.Role == pkgjwt.RoleAdmin || claims.UserID == targetUserID {
		return nil
	}
	return pkgjwt.ErrForbidden
}

func writeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, pkgjwt.ErrUnauthorized):
		writeJSON(w, http.StatusUnauthorized, map[string]string{"message": "unauthorized"})
	case errors.Is(err, pkgjwt.ErrForbidden):
		writeJSON(w, http.StatusForbidden, map[string]string{"message": "forbidden"})
	case errors.Is(err, domain.ErrUserNotFound):
		writeJSON(w, http.StatusNotFound, map[string]string{"message": "user not found"})
	case errors.Is(err, domain.ErrInvalidArgument):
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": "invalid argument"})
	default:
		writeJSON(w, http.StatusInternalServerError, map[string]string{"message": "internal server error"})
	}
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func toUserJSON(user *domain.User) map[string]any {
	return map[string]any{
		"id":        user.ID,
		"login":     user.Login,
		"role":      user.Role,
		"avatar":    user.Avatar,
		"createdAt": user.CreatedAt.UTC().Format(time.RFC3339),
		"updatedAt": user.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

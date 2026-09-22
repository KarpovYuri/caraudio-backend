package httpadapter

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/KarpovYuri/caraudio-backend/internal/catalog/app/services"
	"github.com/KarpovYuri/caraudio-backend/internal/catalog/domain"
	"github.com/KarpovYuri/caraudio-backend/internal/catalog/infrastructure/imageutil"
	"github.com/KarpovYuri/caraudio-backend/internal/catalog/infrastructure/storage"
	pkgjwt "github.com/KarpovYuri/caraudio-backend/pkg/jwt"
)

type SupplierLogoHandler struct {
	catalog   services.CatalogService
	files     *storage.LocalStorage
	jwtSecret string
	maxBytes  int64
	maxSide   int
}

func NewSupplierLogoHandler(
	catalog services.CatalogService,
	files *storage.LocalStorage,
	jwtSecret string,
	maxBytes int64,
	maxSide int,
) *SupplierLogoHandler {
	if maxBytes <= 0 {
		maxBytes = 10 << 20
	}
	if maxSide <= 0 {
		maxSide = imageutil.DefaultMaxLogoSide
	}
	return &SupplierLogoHandler{
		catalog:   catalog,
		files:     files,
		jwtSecret: jwtSecret,
		maxBytes:  maxBytes,
		maxSide:   maxSide,
	}
}

func (h *SupplierLogoHandler) Upload(w http.ResponseWriter, r *http.Request) {
	//if err := requireAdminHTTP(r, h.jwtSecret); err != nil {
	//	writeError(w, err)
	//	return
	//}

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": "supplier id is required"})
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, h.maxBytes+1024)
	if err := r.ParseMultipartForm(h.maxBytes + 1024); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": "invalid or too large multipart body"})
		return
	}

	file, header, err := r.FormFile("logo")
	if err != nil {
		file, header, err = r.FormFile("file")
	}
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": "logo file is required (field: logo)"})
		return
	}
	defer file.Close()

	if err := validateImageUpload(header.Filename, file); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": err.Error()})
		return
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": "failed to read logo file"})
		return
	}

	limited := &io.LimitedReader{R: file, N: h.maxBytes + 1}
	raw, err := io.ReadAll(limited)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": "failed to read logo file"})
		return
	}
	if limited.N == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": "logo file is too large"})
		return
	}

	var (
		processed []byte
		ext       string
	)
	if imageutil.IsSVG(header.Filename, raw) {
		processed, err = imageutil.SanitizeSVG(raw, int(h.maxBytes))
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"message": "invalid or unsafe svg logo"})
			return
		}
		ext = ".svg"
	} else {
		processed, err = imageutil.ProcessLogo(bytes.NewReader(raw), h.maxSide)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"message": "only jpeg, png, webp, gif and svg logos are allowed"})
			return
		}
		ext = ".webp"
	}

	logoURL, err := h.files.SaveSupplierLogo(ext, bytes.NewReader(processed))
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"message": "failed to save logo"})
		return
	}

	existing, err := h.catalog.GetSupplier(r.Context(), id)
	if err != nil {
		_ = h.files.DeletePublicURL(logoURL)
		writeError(w, err)
		return
	}

	supplier, err := h.catalog.UpdateSupplierLogo(r.Context(), id, logoURL)
	if err != nil {
		_ = h.files.DeletePublicURL(logoURL)
		writeError(w, err)
		return
	}

	if existing.Logo != "" && existing.Logo != logoURL {
		_ = h.files.DeletePublicURL(existing.Logo)
	}

	writeJSON(w, http.StatusOK, map[string]any{"supplier": toSupplierJSON(supplier)})
}

func validateImageUpload(filename string, r io.ReadSeeker) error {
	buf := make([]byte, 512)
	n, err := r.Read(buf)
	if err != nil && !errors.Is(err, io.EOF) {
		return errors.New("failed to read logo file")
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
		return errors.New("only jpeg, png, webp, gif and svg logos are allowed")
	}
}

func requireAdminHTTP(r *http.Request, jwtSecret string) error {
	auth := strings.TrimSpace(r.Header.Get("Authorization"))
	token := auth
	if len(auth) > 7 && strings.EqualFold(auth[:7], "bearer ") {
		token = strings.TrimSpace(auth[7:])
	}
	return pkgjwt.ValidateAdmin(token, jwtSecret)
}

func writeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, pkgjwt.ErrUnauthorized):
		writeJSON(w, http.StatusUnauthorized, map[string]string{"message": "unauthorized"})
	case errors.Is(err, pkgjwt.ErrForbidden):
		writeJSON(w, http.StatusForbidden, map[string]string{"message": "forbidden"})
	case errors.Is(err, domain.ErrSupplierNotFound):
		writeJSON(w, http.StatusNotFound, map[string]string{"message": "supplier not found"})
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

func toSupplierJSON(s *domain.Supplier) map[string]any {
	code := ""
	if s.Code != nil {
		code = *s.Code
	}
	return map[string]any{
		"id":        strconv.FormatInt(s.ID, 10),
		"name":      s.Name,
		"code":      code,
		"logo":      s.Logo,
		"apiUrl":    s.ApiUrl,
		"isActive":  s.IsActive,
		"createdAt": s.CreatedAt.UTC().Format(time.RFC3339),
		"updatedAt": s.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

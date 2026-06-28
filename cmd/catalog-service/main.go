package main

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	cataloggrpc "github.com/KarpovYuri/caraudio-backend/internal/catalog/adapters/grpc"
	catalogservice "github.com/KarpovYuri/caraudio-backend/internal/catalog/app/services"
	catalogconfig "github.com/KarpovYuri/caraudio-backend/internal/catalog/config"
	catalogdb "github.com/KarpovYuri/caraudio-backend/internal/catalog/infrastructure/database/postgres"
	catalogv1 "github.com/KarpovYuri/caraudio-backend/pkg/api/proto/catalog/v1"
	"github.com/google/uuid"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	cfg, err := catalogconfig.LoadConfig()
	if err != nil {
		logger.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	db, err := catalogdb.InitDB(&cfg.Database)
	if err != nil {
		logger.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := db.Close(); err != nil {
			logger.Error("failed to close db", "error", err)
		}
	}()

	supplierRepo := catalogdb.NewPostgresSupplierRepository(db)
	categoryRepo := catalogdb.NewPostgresCategoryRepository(db)
	productRepo := catalogdb.NewPostgresProductRepository(db)
	brandRepo := catalogdb.NewPostgresBrandRepository(db)
	productImageRepo := catalogdb.NewPostgresProductImageRepository(db)
	productAttrRepo := catalogdb.NewPostgresProductAttributeRepository(db)
	categoryMappingRepo := catalogdb.NewPostgresSupplierCategoryMappingRepository(db)
	productMappingRepo := catalogdb.NewPostgresSupplierProductMappingRepository(db)

	catalogSvc := catalogservice.NewCatalogService(supplierRepo, categoryRepo, productRepo, brandRepo, productImageRepo, productAttrRepo, categoryMappingRepo, productMappingRepo)

	catalogGRPC := cataloggrpc.NewCatalogGRPCServer(
		catalogSvc,
		cfg.JWTSecret,
	)

	lis, err := net.Listen("tcp", cfg.GRPCPort)
	if err != nil {
		logger.Error("failed to listen on gRPC port", "port", cfg.GRPCPort, "error", err)
		os.Exit(1)
	}

	s := grpc.NewServer(grpc.UnaryInterceptor(grpcLoggingInterceptor(logger)))
	catalogv1.RegisterCatalogServiceServer(s, catalogGRPC)

	ctx := context.Background()
	mux := runtime.NewServeMux(
		runtime.WithMetadata(func(_ context.Context, req *http.Request) metadata.MD {
			requestID := req.Header.Get("X-Request-Id")
			if requestID == "" {
				return metadata.MD{}
			}
			return metadata.Pairs("x-request-id", requestID)
		}),
		runtime.WithIncomingHeaderMatcher(func(key string) (string, bool) {
			if strings.EqualFold(key, "authorization") {
				return "authorization", true
			}
			return runtime.DefaultHeaderMatcher(key)
		}),
	)

	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	if err := catalogv1.RegisterCatalogServiceHandlerFromEndpoint(
		ctx, mux, "localhost"+cfg.GRPCPort, opts,
	); err != nil {
		logger.Error("failed to register catalog gateway", "error", err)
		os.Exit(1)
	}

	shutdownCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	httpHandler := allowCORS(withRequestID(withAccessLog(mux, logger), logger), cfg.AllowedOrigins)

	httpServer := &http.Server{
		Addr:         cfg.HTTPPort,
		Handler:      httpHandler,
		ReadTimeout:  cfg.HTTPReadTimeout,
		WriteTimeout: cfg.HTTPWriteTimeout,
		IdleTimeout:  cfg.HTTPIdleTimeout,
	}

	serverErrCh := make(chan error, 2)

	go func() {
		logger.Info("gRPC server started", "addr", cfg.GRPCPort)
		if serveErr := s.Serve(lis); serveErr != nil {
			serverErrCh <- serveErr
		}
	}()

	go func() {
		logger.Info("HTTP gateway started", "addr", cfg.HTTPPort)
		if serveErr := httpServer.ListenAndServe(); serveErr != nil && !errors.Is(serveErr, http.ErrServerClosed) {
			serverErrCh <- serveErr
		}
	}()

	select {
	case <-shutdownCtx.Done():
		logger.Info("shutdown signal received")
	case serveErr := <-serverErrCh:
		logger.Error("server terminated unexpectedly", "error", serveErr)
	}

	stop()

	httpShutdownCtx, httpShutdownCancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer httpShutdownCancel()
	if err := httpServer.Shutdown(httpShutdownCtx); err != nil {
		logger.Error("http shutdown failed", "error", err)
	}

	grpcStopped := make(chan struct{})
	go func() {
		s.GracefulStop()
		close(grpcStopped)
	}()

	select {
	case <-grpcStopped:
		logger.Info("gRPC server stopped gracefully")
	case <-time.After(cfg.ShutdownTimeout):
		logger.Warn("gRPC graceful shutdown timeout reached, forcing stop")
		s.Stop()
	}
}

type contextKey string

const requestIDContextKey contextKey = "request_id"

func withRequestID(next http.Handler, logger *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-Id")
		if requestID == "" {
			requestID = uuid.NewString()
		}
		w.Header().Set("X-Request-Id", requestID)
		ctx := context.WithValue(r.Context(), requestIDContextKey, requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func grpcLoggingInterceptor(logger *slog.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		start := time.Now()
		requestID := requestIDFromMetadata(ctx)
		resp, err := handler(ctx, req)

		level := slog.LevelInfo
		if err != nil {
			level = slog.LevelWarn
		}
		logger.Log(ctx, level, "grpc request completed",
			"request_id", requestID,
			"method", info.FullMethod,
			"status", status.Code(err).String(),
			"duration_ms", time.Since(start).Milliseconds(),
		)
		return resp, err
	}
}

func requestIDFromMetadata(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}
	values := md.Get("x-request-id")
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

func allowCORS(h http.Handler, allowedOrigins []string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" && isAllowedOrigin(origin, allowedOrigins) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}
		w.Header().Set("Vary", "Origin")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		h.ServeHTTP(w, r)
	})
}

func isAllowedOrigin(origin string, allowedOrigins []string) bool {
	for _, allowed := range allowedOrigins {
		if origin == allowed {
			return true
		}
	}
	return false
}

func withAccessLog(next http.Handler, logger *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		requestID, _ := r.Context().Value(requestIDContextKey).(string)
		next.ServeHTTP(w, r)
		logger.Info("http request completed",
			"request_id", requestID,
			"method", r.Method,
			"path", r.URL.Path,
			"remote_addr", r.RemoteAddr,
			"duration_ms", time.Since(start).Milliseconds(),
		)
	})
}

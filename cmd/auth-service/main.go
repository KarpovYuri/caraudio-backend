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

	authgrpc "github.com/KarpovYuri/caraudio-backend/internal/auth/adapters/grpc"
	authservice "github.com/KarpovYuri/caraudio-backend/internal/auth/app/services"
	authconfig "github.com/KarpovYuri/caraudio-backend/internal/auth/config"
	authdb "github.com/KarpovYuri/caraudio-backend/internal/auth/infrastructure/database/postgres"
	authv1 "github.com/KarpovYuri/caraudio-backend/pkg/api/proto/auth/v1"
	"github.com/google/uuid"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	cfg, err := authconfig.LoadConfig()
	if err != nil {
		logger.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	db, err := authdb.InitDB(&cfg.Database)
	if err != nil {
		logger.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := db.Close(); err != nil {
			logger.Error("failed to close db", "error", err)
		}
	}()

	userRepo := authdb.NewPostgresUserRepository(db)
	tokenRepo := authdb.NewPgRefreshTokenRepository(db)

	authService := authservice.NewAuthService(
		userRepo,
		tokenRepo,
		cfg.JWTSecret,
	)

	authGRPCServer := authgrpc.NewAuthGRPCServer(authService, cfg.CookieSecure)

	lis, err := net.Listen("tcp", cfg.GRPCPort)
	if err != nil {
		logger.Error("failed to listen on gRPC port", "port", cfg.GRPCPort, "error", err)
		os.Exit(1)
	}

	s := grpc.NewServer(grpc.UnaryInterceptor(grpcLoggingInterceptor(logger)))
	authv1.RegisterAuthServiceServer(s, authGRPCServer)

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
			switch strings.ToLower(key) {
			case "cookie":
				return "grpcgateway-cookie", true
			default:
				return runtime.DefaultHeaderMatcher(key)
			}
		}),
		runtime.WithForwardResponseOption(func(ctx context.Context, w http.ResponseWriter, resp proto.Message) error {
			md, ok := runtime.ServerMetadataFromContext(ctx)
			if !ok {
				return nil
			}
			if cookies := md.HeaderMD.Get("set-cookie"); len(cookies) > 0 {
				for _, cookie := range cookies {
					w.Header().Add("Set-Cookie", cookie)
				}
			}
			return nil
		}),
	)

	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	err = authv1.RegisterAuthServiceHandlerFromEndpoint(ctx, mux, "localhost"+cfg.GRPCPort, opts)
	if err != nil {
		logger.Error("failed to register gateway", "error", err)
		os.Exit(1)
	}

	cleaned, err := tokenRepo.DeleteExpired(context.Background(), time.Now())
	if err != nil {
		logger.Error("failed to run startup refresh-token cleanup", "error", err)
	} else if cleaned > 0 {
		logger.Info("startup refresh-token cleanup completed", "deleted_tokens", cleaned)
	}

	shutdownCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cleanupCtx, cleanupCancel := context.WithCancel(context.Background())
	defer cleanupCancel()
	go runTokenCleanupJob(cleanupCtx, tokenRepo, cfg.TokenCleanupEvery, logger)

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

	cleanupCancel()
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

func runTokenCleanupJob(
	ctx context.Context,
	tokenRepo authdb.RefreshTokenRepository,
	interval time.Duration,
	logger *slog.Logger,
) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			deleted, err := tokenRepo.DeleteExpired(context.Background(), time.Now())
			if err != nil {
				logger.Error("periodic refresh-token cleanup failed", "error", err)
				continue
			}
			if deleted > 0 {
				logger.Info("periodic refresh-token cleanup completed", "deleted_tokens", deleted)
			}
		}
	}
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

		statusCode := status.Code(err).String()
		level := slog.LevelInfo
		if err != nil {
			level = slog.LevelWarn
		}

		logger.Log(ctx, level, "grpc request completed",
			"request_id", requestID,
			"method", info.FullMethod,
			"status", statusCode,
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
		logger.Debug("request id assigned", "request_id", requestID, "path", r.URL.Path)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func withAccessLog(next http.Handler, logger *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		requestID := requestIDFromContext(r.Context())
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

func requestIDFromContext(ctx context.Context) string {
	value := ctx.Value(requestIDContextKey)
	requestID, ok := value.(string)
	if !ok {
		return ""
	}
	return requestID
}

func allowCORS(h http.Handler, allowedOrigins []string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" && isAllowedOrigin(origin, allowedOrigins) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Expose-Headers", "Set-Cookie")
		}
		w.Header().Set("Vary", "Origin")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Authorization, Set-Cookie, Cookie")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		h.ServeHTTP(w, r)
	})
}

func isAllowedOrigin(origin string, allowedOrigins []string) bool {
	for _, allowedOrigin := range allowedOrigins {
		if origin == allowedOrigin {
			return true
		}
	}
	return false
}

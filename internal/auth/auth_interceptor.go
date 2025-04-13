package auth

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	"connectrpc.com/connect"
	postgres "github.com/atreya2011/health-management-api/internal/db"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// contextKey is a private type for context keys
type contextKey int

// UserContextKey is the key for user ID in the context
const UserContextKey contextKey = iota

// JWTConfig contains JWT validation configuration
type JWTConfig struct {
	SecretKey string
}

// AuthInterceptor creates a Connect interceptor for JWT authentication
func AuthInterceptor(jwtConfig *JWTConfig, userRepo *postgres.PgUserRepository, logger *slog.Logger) connect.UnaryInterceptorFunc { // Use concrete repo type
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			// Skip auth for public endpoints (if any)
			// Example: if strings.HasSuffix(req.Spec().Procedure, "PublicMethod") { return next(ctx, req) }

			// Extract the Authorization header
			authHeader := req.Header().Get("Authorization")
			if authHeader == "" {
				logger.WarnContext(ctx, "Missing Authorization header")
				return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("missing authorization header"))
			}

			// Check for Bearer prefix
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				logger.WarnContext(ctx, "Invalid Authorization header format")
				return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("invalid authorization header format"))
			}

			// Parse and validate the JWT
			token, err := jwt.Parse(parts[1], func(token *jwt.Token) (interface{}, error) {
				// Validate the signing method
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, errors.New("unexpected signing method")
				}
				return []byte(jwtConfig.SecretKey), nil
			})

			if err != nil {
				logger.WarnContext(ctx, "Failed to parse JWT", "error", err)
				return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("invalid token"))
			}

			if !token.Valid {
				logger.WarnContext(ctx, "Invalid JWT")
				return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("invalid token"))
			}

			// Extract claims
			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				logger.WarnContext(ctx, "Failed to extract JWT claims")
				return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("invalid token claims"))
			}

			// Extract the subject (User ID)
			sub, ok := claims["sub"].(string)
			if !ok || sub == "" {
				logger.WarnContext(ctx, "Missing subject claim in token")
				return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("invalid token subject"))
			}

			// Find or create the user
			// Find or create the user using the updated Create method
			user, err := userRepo.Create(ctx, sub)
			if err != nil {
				// Create now handles the "already exists" case by returning the existing user.
				// Any error returned here is likely a database issue or context cancellation.
				logger.ErrorContext(ctx, "Failed to find or create user", "subject_id", sub, "error", err)
				return nil, connect.NewError(connect.CodeInternal, errors.New("failed to retrieve or create user"))
			}

			// Add the user ID to the context
			ctx = context.WithValue(ctx, UserContextKey, user.ID) // user is now db.User, which has ID

			// Call the next handler with the authenticated context
			return next(ctx, req)
		}
	}
}

// GetUserID extracts the user ID from the context
func GetUserID(ctx context.Context) (uuid.UUID, error) {
	userID, ok := ctx.Value(UserContextKey).(uuid.UUID)
	if !ok {
		return uuid.Nil, errors.New("user ID not found in context")
	}
	return userID, nil
}

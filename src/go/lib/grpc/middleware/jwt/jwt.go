package grpc_jwt

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"

	"github.com/gbrlsnchs/jwt"
)

const ExpectedScheme = "bearer"

// ContextKey context.WithValue() specifies that keys should not be any built-in type to avoid collisions
type ContextKey string

var userId ContextKey = "user_id"

// UpdateCtx returns a function for getting a context that includes the user_id as a value
func UpdateCtx() grpc_auth.AuthFunc {
	return func(ctx context.Context) (context.Context, error) {
		tokenString, err := grpc_auth.AuthFromMD(ctx, ExpectedScheme)
		if err != nil {
			return ctx, status.Errorf(codes.Unauthenticated, "failed to obtain JWT from headers: %v", err)
		}

		token, err := jwt.FromString(tokenString)
		if err != nil {
			return ctx, status.Errorf(codes.PermissionDenied, "token in auth header is malformed or cannot be converted from a raw string to a JWT type: %v", err)
		}

		newCtx := context.WithValue(ctx, userId, token.Subject())
		return newCtx, nil
	}
}

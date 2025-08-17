package auth

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

func RequireAuth(authService *AuthService) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "User not authenticated"})
			}

			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			if tokenString == authHeader { // No "Bearer " prefix
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid token format"})
			}
			claims, err := authService.ValidateToken(tokenString)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid token"})
			}

			c.Set("userID", claims.UserID)
			c.Set("userEmail", claims.Email)

			return next(c)
		}
	}
}

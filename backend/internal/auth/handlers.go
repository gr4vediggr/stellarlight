package auth

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

type Handler struct {
	service *AuthService
}

func NewHandler(service *AuthService) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) Register(c echo.Context) error {
	var req RegisterRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}
	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed", "details": err.Error()})
	}

	resp, err := h.service.Register(c.Request().Context(), &req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	// set refresh token in HttpOnly cookie
	c.SetCookie(&http.Cookie{
		Name:     "refresh_token",
		Value:    resp.RefreshToken,
		Path:     "/auth/refresh",
		HttpOnly: true,
		Secure:   true, // only over HTTPS in prod
		SameSite: http.SameSiteStrictMode,
	})

	// return only access token + user
	return c.JSON(http.StatusCreated, map[string]interface{}{
		"token": resp.Token,
		"user":  resp.User,
	})
}

func (h *Handler) Login(c echo.Context) error {
	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}
	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed", "details": err.Error()})
	}

	resp, err := h.service.Login(c.Request().Context(), &req)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": err.Error()})
	}

	c.SetCookie(&http.Cookie{
		Name:     "refresh_token",
		Value:    resp.RefreshToken,
		Path:     "/auth/refresh",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})

	return c.JSON(http.StatusOK, map[string]interface{}{
		"token": resp.Token,
		"user":  resp.User,
	})
}

func (h *Handler) RefreshToken(c echo.Context) error {
	var refreshToken string

	// First try to get refresh token from cookie (existing behavior)
	cookie, err := c.Cookie("refresh_token")
	if err == nil {
		refreshToken = cookie.Value
	} else {
		// Fallback: try to get refresh token from request body
		var req struct {
			RefreshToken string `json:"refreshToken"`
		}
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Missing refresh token"})
		}
		refreshToken = req.RefreshToken
	}

	if refreshToken == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Missing refresh token"})
	}

	resp, err := h.service.RefreshToken(c.Request().Context(), refreshToken)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": err.Error()})
	}

	// If we got the token from cookie, rotate the cookie
	if cookie != nil {
		c.SetCookie(&http.Cookie{
			Name:     "refresh_token",
			Value:    resp.RefreshToken,
			Path:     "/auth/refresh",
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteStrictMode,
		})
	}

	// Return new access token and refresh token (for body-based requests)
	return c.JSON(http.StatusOK, map[string]interface{}{
		"token":        resp.Token,
		"refreshToken": resp.RefreshToken,
		"user":         resp.User,
	})
}

func (h *Handler) Logout(c echo.Context) error {
	// Read refresh token from cookie
	cv, err := c.Cookie("refresh_token")
	if err == nil {
		// Revoke the refresh token
		if err := h.service.RevokeRefreshToken(c.Request().Context(), cv.Value); err != nil {
			log.Error("Failed to revoke refresh token:", err)

		}
	}

	cookie := &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/auth/refresh",
		MaxAge:   -1, // delete cookie
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	}
	c.SetCookie(cookie)

	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) UpdateProfile(c echo.Context) error {
	userID, ok := c.Get("userID").(uuid.UUID)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "User not authenticated"})
	}

	var req UpdateProfileRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}
	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed", "details": err.Error()})
	}

	user, err := h.service.UpdateProfile(c.Request().Context(), userID, &req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, user)
}

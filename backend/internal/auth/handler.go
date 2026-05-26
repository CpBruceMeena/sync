package auth

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/CpBruceMeena/sync/internal/httputil"
	"github.com/CpBruceMeena/sync/internal/models"
	"github.com/CpBruceMeena/sync/internal/repository"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func NewHandler(authService *Service, repos *repository.Repositories) *Handler {
	return &Handler{
		authService: authService,
		repos:       repos,
	}
}

// Register creates a new user account
// @Summary Register a new user
// @Description Create a new user account with username, email, and password
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "Registration details"
// @Success 201 {object} AuthResponse
// @Failure 400 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Router /api/auth/register [post]
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	req.Username = strings.TrimSpace(req.Username)
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))

	if req.Username == "" || req.Email == "" || req.Password == "" {
		httputil.RespondError(w, http.StatusBadRequest, "Username, email, and password are required")
		return
	}

	if len(req.Password) < 6 {
		httputil.RespondError(w, http.StatusBadRequest, "Password must be at least 6 characters")
		return
	}

	// Check if email already exists
	existing, err := h.repos.Users.GetByEmail(r.Context(), req.Email)
	if err == nil && existing != nil {
		httputil.RespondError(w, http.StatusConflict, "Email already registered")
		return
	}

	// Check if username already exists
	existingUser, err := h.repos.Users.GetByUsername(r.Context(), req.Username)
	if err == nil && existingUser != nil {
		httputil.RespondError(w, http.StatusConflict, "Username already taken")
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Error hashing password: %v", err)
		httputil.RespondError(w, http.StatusInternalServerError, "Failed to create user")
		return
	}

	user := &models.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		DisplayName:  req.Username,
		Status:       "offline",
	}
	if err := h.repos.Users.Create(r.Context(), user); err != nil {
		log.Printf("Error creating user: %v", err)
		httputil.RespondError(w, http.StatusInternalServerError, "Failed to create user")
		return
	}

	tokens, err := h.authService.GenerateTokens(user.ID, user.Username)
	if err != nil {
		log.Printf("Error generating tokens: %v", err)
		httputil.RespondError(w, http.StatusInternalServerError, "Failed to generate tokens")
		return
	}

	// Store refresh token
	if err := h.repos.Sessions.Create(r.Context(), &models.Session{
		UserID:       user.ID,
		RefreshToken: tokens.RefreshToken,
		ExpiresAt:    time.Now().Add(7 * 24 * time.Hour),
	}); err != nil {
		log.Printf("Error creating session: %v", err)
	}

	httputil.RespondJSON(w, http.StatusCreated, AuthResponse{
		User: UserResponse{
			ID:          user.ID,
			Username:    user.Username,
			Email:       user.Email,
			DisplayName: user.DisplayName,
			AvatarURL:   user.AvatarUrl,
			Status:      user.Status,
		},
		Token: tokens,
	})
}

// Login authenticates a user
// @Summary Login
// @Description Authenticate a user with email and password
// @Tags auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login credentials"
// @Success 200 {object} AuthResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /api/auth/login [post]
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))

	if req.Email == "" || req.Password == "" {
		httputil.RespondError(w, http.StatusBadRequest, "Email and password are required")
		return
	}

	user, err := h.repos.Users.GetByEmailWithPassword(r.Context(), req.Email)
	if err != nil {
		httputil.RespondError(w, http.StatusUnauthorized, "Invalid email or password")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		httputil.RespondError(w, http.StatusUnauthorized, "Invalid email or password")
		return
	}

	tokens, err := h.authService.GenerateTokens(user.ID, user.Username)
	if err != nil {
		log.Printf("Error generating tokens: %v", err)
		httputil.RespondError(w, http.StatusInternalServerError, "Failed to generate tokens")
		return
	}

	// Store refresh token
	if err := h.repos.Sessions.Create(r.Context(), &models.Session{
		UserID:       user.ID,
		RefreshToken: tokens.RefreshToken,
		ExpiresAt:    time.Now().Add(7 * 24 * time.Hour),
	}); err != nil {
		log.Printf("Error creating session: %v", err)
	}

	httputil.RespondJSON(w, http.StatusOK, AuthResponse{
		User: UserResponse{
			ID:          user.ID,
			Username:    user.Username,
			Email:       user.Email,
			DisplayName: user.DisplayName,
			AvatarURL:   user.AvatarUrl,
			Status:      user.Status,
		},
		Token: tokens,
	})
}

// Refresh refreshes an access token
// @Summary Refresh token
// @Description Refresh an expired access token using a refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RefreshRequest true "Refresh token"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /api/auth/refresh [post]
func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.RefreshToken == "" {
		httputil.RespondError(w, http.StatusBadRequest, "Refresh token is required")
		return
	}

	session, err := h.repos.Sessions.GetByToken(r.Context(), req.RefreshToken)
	if err != nil {
		httputil.RespondError(w, http.StatusUnauthorized, "Invalid refresh token")
		return
	}

	if session.ExpiresAt.Before(time.Now()) {
		h.repos.Sessions.Delete(r.Context(), session.ID)
		httputil.RespondError(w, http.StatusUnauthorized, "Refresh token expired")
		return
	}

	// Delete old session
	h.repos.Sessions.Delete(r.Context(), session.ID)

	user, err := h.repos.Users.GetByID(r.Context(), session.UserID)
	if err != nil {
		httputil.RespondError(w, http.StatusInternalServerError, "User not found")
		return
	}

	tokens, err := h.authService.GenerateTokens(user.ID, user.Username)
	if err != nil {
		log.Printf("Error generating tokens: %v", err)
		httputil.RespondError(w, http.StatusInternalServerError, "Failed to generate tokens")
		return
	}

	// Store new refresh token
	if err := h.repos.Sessions.Create(r.Context(), &models.Session{
		UserID:       user.ID,
		RefreshToken: tokens.RefreshToken,
		ExpiresAt:    time.Now().Add(7 * 24 * time.Hour),
	}); err != nil {
		log.Printf("Error creating session: %v", err)
	}

	httputil.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"token": tokens,
	})
}

// Logout logs out the current user
// @Summary Logout
// @Description Invalidate all refresh tokens for the authenticated user
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /api/auth/logout [post]
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(uuid.UUID)

	if err := h.repos.Sessions.DeleteByUserID(r.Context(), userID); err != nil {
		log.Printf("Error deleting sessions: %v", err)
	}

	httputil.RespondJSON(w, http.StatusOK, map[string]string{"message": "Logged out successfully"})
}

// Me returns the authenticated user's profile
// @Summary Get current user
// @Description Get the profile of the currently authenticated user
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} UserResponse
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/auth/me [get]
func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(uuid.UUID)

	user, err := h.repos.Users.GetByID(r.Context(), userID)
	if err != nil {
		httputil.RespondError(w, http.StatusNotFound, "User not found")
		return
	}

	httputil.RespondJSON(w, http.StatusOK, UserResponse{
		ID:          user.ID,
		Username:    user.Username,
		Email:       user.Email,
		DisplayName: user.DisplayName,
		AvatarURL:   user.AvatarUrl,
		Status:      user.Status,
	})
}

// GetAuthService exposes the auth service for WebSocket handler
func (h *Handler) GetAuthService() *Service {
	return h.authService
}

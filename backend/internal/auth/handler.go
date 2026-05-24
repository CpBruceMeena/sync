package auth

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/CpBruceMeena/sync/internal/database"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func NewHandler(authService *Service, db *database.DB, queries database.Querier) *Handler {
	return &Handler{
		authService: authService,
		db:          db,
		queries:     queries,
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
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	req.Username = strings.TrimSpace(req.Username)
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))

	if req.Username == "" || req.Email == "" || req.Password == "" {
		respondError(w, http.StatusBadRequest, "Username, email, and password are required")
		return
	}

	if len(req.Password) < 6 {
		respondError(w, http.StatusBadRequest, "Password must be at least 6 characters")
		return
	}

	// Check if email already exists
	existing, err := h.queries.GetUserByEmail(r.Context(), req.Email)
	if err == nil && existing.Email != "" {
		respondError(w, http.StatusConflict, "Email already registered")
		return
	}

	// Check if username already exists
	existingUser, err := h.queries.GetUserByUsername(r.Context(), req.Username)
	if err == nil && existingUser.Username != "" {
		respondError(w, http.StatusConflict, "Username already taken")
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Error hashing password: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to create user")
		return
	}

	user, err := h.queries.CreateUser(r.Context(), database.CreateUserParams{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		DisplayName:  req.Username,
	})
	if err != nil {
		log.Printf("Error creating user: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to create user")
		return
	}

	tokens, err := h.authService.GenerateTokens(user.ID, user.Username)
	if err != nil {
		log.Printf("Error generating tokens: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to generate tokens")
		return
	}

	// Store refresh token
	_, err = h.queries.CreateSession(r.Context(), database.CreateSessionParams{
		UserID:       user.ID,
		RefreshToken: tokens.RefreshToken,
		ExpiresAt:    time.Now().Add(7 * 24 * time.Hour),
	})
	if err != nil {
		log.Printf("Error creating session: %v", err)
	}

	respondJSON(w, http.StatusCreated, AuthResponse{
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
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))

	if req.Email == "" || req.Password == "" {
		respondError(w, http.StatusBadRequest, "Email and password are required")
		return
	}

	user, err := h.queries.GetUserByEmailWithPassword(r.Context(), req.Email)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "Invalid email or password")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		respondError(w, http.StatusUnauthorized, "Invalid email or password")
		return
	}

	tokens, err := h.authService.GenerateTokens(user.ID, user.Username)
	if err != nil {
		log.Printf("Error generating tokens: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to generate tokens")
		return
	}

	// Store refresh token
	_, err = h.queries.CreateSession(r.Context(), database.CreateSessionParams{
		UserID:       user.ID,
		RefreshToken: tokens.RefreshToken,
		ExpiresAt:    time.Now().Add(7 * 24 * time.Hour),
	})
	if err != nil {
		log.Printf("Error creating session: %v", err)
	}

	respondJSON(w, http.StatusOK, AuthResponse{
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
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.RefreshToken == "" {
		respondError(w, http.StatusBadRequest, "Refresh token is required")
		return
	}

	session, err := h.queries.GetSessionByToken(r.Context(), req.RefreshToken)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "Invalid refresh token")
		return
	}

	if session.ExpiresAt.Before(time.Now()) {
		h.queries.DeleteSession(r.Context(), session.ID)
		respondError(w, http.StatusUnauthorized, "Refresh token expired")
		return
	}

	// Delete old session
	h.queries.DeleteSession(r.Context(), session.ID)

	userID, err := uuid.Parse(session.UserID.String())
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to parse user ID")
		return
	}

	tokens, err := h.authService.GenerateTokens(userID, session.Username)
	if err != nil {
		log.Printf("Error generating tokens: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to generate tokens")
		return
	}

	// Store new refresh token
	_, err = h.queries.CreateSession(r.Context(), database.CreateSessionParams{
		UserID:       userID,
		RefreshToken: tokens.RefreshToken,
		ExpiresAt:    time.Now().Add(7 * 24 * time.Hour),
	})
	if err != nil {
		log.Printf("Error creating session: %v", err)
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
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
// @Router /api/auth/logout [post]
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(uuid.UUID)

	if err := h.queries.DeleteUserSessions(r.Context(), userID); err != nil {
		log.Printf("Error deleting sessions: %v", err)
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Logged out successfully"})
}

// Me returns the authenticated user's profile
// @Summary Get current user
// @Description Get the profile of the currently authenticated user
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} UserResponse
// @Failure 404 {object} map[string]string
// @Router /api/auth/me [get]
func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(uuid.UUID)

	user, err := h.queries.GetUserByID(r.Context(), userID)
	if err != nil {
		respondError(w, http.StatusNotFound, "User not found")
		return
	}

	respondJSON(w, http.StatusOK, UserResponse{
		ID:          user.ID,
		Username:    user.Username,
		Email:       user.Email,
		DisplayName: user.DisplayName,
		AvatarURL:   user.AvatarUrl,
		Status:      user.Status,
	})
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}

// GetAuthService exposes the auth service for WebSocket handler
func (h *Handler) GetAuthService() *Service {
	return h.authService
}

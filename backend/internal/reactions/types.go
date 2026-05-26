package reactions

// Handler handles reaction HTTP requests
type Handler struct {
	service *Service
}

// ReactionRequest represents a reaction toggle request
type ReactionRequest struct {
	Emoji string `json:"emoji"`
}

// ReactionResponse represents a reaction in API responses
type ReactionResponse struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Emoji    string `json:"emoji"`
}

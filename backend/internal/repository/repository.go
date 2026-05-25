package repository

import "gorm.io/gorm"

// Repositories holds all repository implementations
type Repositories struct {
	Users         UserRepository
	Conversations ConversationRepository
	Messages      MessageRepository
	Sessions      SessionRepository
	Notifications NotificationRepository
}

// NewRepositories creates all repository instances
func NewRepositories(db *gorm.DB) *Repositories {
	return &Repositories{
		Users:         NewUserRepository(db),
		Conversations: NewConversationRepository(db),
		Messages:      NewMessageRepository(db),
		Sessions:      NewSessionRepository(db),
		Notifications: NewNotificationRepository(db),
	}
}

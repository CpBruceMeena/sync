package repository

import (
	"context"

	"github.com/CpBruceMeena/sync/internal/models"
	"github.com/google/uuid"
)

// UserRepository defines user data operations
type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	GetByEmailWithPassword(ctx context.Context, email string) (*models.User, error)
	GetByUsername(ctx context.Context, username string) (*models.User, error)
	List(ctx context.Context) ([]models.User, error)
	Search(ctx context.Context, query string, limit int) ([]models.User, error)
	Update(ctx context.Context, user *models.User) error
	UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status string) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// ConversationRepository defines conversation data operations
type ConversationRepository interface {
	Create(ctx context.Context, conv *models.Conversation) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Conversation, error)
	ListByUserID(ctx context.Context, userID uuid.UUID) ([]models.Conversation, error)
	ListPublic(ctx context.Context, limit, offset int) ([]models.Conversation, error)
	SearchPublic(ctx context.Context, query string, limit int) ([]models.Conversation, error)
	FindPrivate(ctx context.Context, userID1, userID2 uuid.UUID) (*models.Conversation, error)
	AddMember(ctx context.Context, member *models.ConversationMember) error
	RemoveMember(ctx context.Context, convID, userID uuid.UUID) error
	GetMembers(ctx context.Context, convID uuid.UUID) ([]models.ConversationMember, error)
	IsMember(ctx context.Context, convID, userID uuid.UUID) (bool, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

// MessageRepository defines message data operations
type MessageRepository interface {
	Create(ctx context.Context, msg *models.Message) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Message, error)
	ListByConversation(ctx context.Context, convID uuid.UUID, cursor uuid.UUID, limit int) ([]models.Message, error)
	Delete(ctx context.Context, id, senderID uuid.UUID) error
	AddReaction(ctx context.Context, reaction *models.Reaction) error
	RemoveReaction(ctx context.Context, messageID, userID uuid.UUID, emoji string) error
	GetReactionsByMessage(ctx context.Context, messageID uuid.UUID) ([]models.Reaction, error)
}

// SessionRepository defines session data operations
type SessionRepository interface {
	Create(ctx context.Context, session *models.Session) error
	GetByToken(ctx context.Context, refreshToken string) (*models.Session, error)
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteByUserID(ctx context.Context, userID uuid.UUID) error
	CleanExpired(ctx context.Context) error
}

// NotificationRepository defines notification data operations
type NotificationRepository interface {
	Create(ctx context.Context, notification *models.Notification) error
	List(ctx context.Context, userID uuid.UUID, limit int) ([]models.Notification, error)
	MarkRead(ctx context.Context, id, userID uuid.UUID) error
	MarkAllRead(ctx context.Context, userID uuid.UUID) error
	GetUnreadCount(ctx context.Context, userID uuid.UUID) (int64, error)
	Delete(ctx context.Context, id, userID uuid.UUID) error
}

// AttachmentRepository defines attachment data operations
type AttachmentRepository interface {
	Create(ctx context.Context, attachment *models.Attachment) error
	GetByMessageID(ctx context.Context, messageID uuid.UUID) ([]models.Attachment, error)
}

package tests

import (
	"context"

	"github.com/CpBruceMeena/sync/internal/models"
	"github.com/CpBruceMeena/sync/internal/repository"
	"github.com/google/uuid"
)

// ---- Mock repository implementations ----
// These mocks assume the database exists and simulate repository behavior.
// Each mock has function fields that can be overridden per test case.
// To add a new test case, set the relevant mock function field(s) in the test.

type mockUserRepo struct {
	getByEmailFn             func(ctx context.Context, email string) (*models.User, error)
	getByUsernameFn          func(ctx context.Context, username string) (*models.User, error)
	listFn                   func(ctx context.Context) ([]models.User, error)
	createFn                 func(ctx context.Context, user *models.User) error
	getByIDFn                func(ctx context.Context, id uuid.UUID) (*models.User, error)
	getByEmailWithPasswordFn func(ctx context.Context, email string) (*models.User, error)
	updateFn                 func(ctx context.Context, user *models.User) error
	updatePasswordFn         func(ctx context.Context, id uuid.UUID, passwordHash string) error
	updateStatusFn           func(ctx context.Context, id uuid.UUID, status string) error
	deleteFn                 func(ctx context.Context, id uuid.UUID) error
	searchFn                 func(ctx context.Context, query string, limit int) ([]models.User, error)
}

func (m *mockUserRepo) Create(ctx context.Context, user *models.User) error {
	if m.createFn != nil {
		return m.createFn(ctx, user)
	}
	return nil
}
func (m *mockUserRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, nil
}
func (m *mockUserRepo) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	if m.getByEmailFn != nil {
		return m.getByEmailFn(ctx, email)
	}
	return nil, nil
}
func (m *mockUserRepo) GetByEmailWithPassword(ctx context.Context, email string) (*models.User, error) {
	if m.getByEmailWithPasswordFn != nil {
		return m.getByEmailWithPasswordFn(ctx, email)
	}
	return nil, nil
}
func (m *mockUserRepo) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	if m.getByUsernameFn != nil {
		return m.getByUsernameFn(ctx, username)
	}
	return nil, nil
}
func (m *mockUserRepo) List(ctx context.Context) ([]models.User, error) {
	if m.listFn != nil {
		return m.listFn(ctx)
	}
	return nil, nil
}
func (m *mockUserRepo) Update(ctx context.Context, user *models.User) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, user)
	}
	return nil
}
func (m *mockUserRepo) UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error {
	if m.updatePasswordFn != nil {
		return m.updatePasswordFn(ctx, id, passwordHash)
	}
	return nil
}
func (m *mockUserRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	if m.updateStatusFn != nil {
		return m.updateStatusFn(ctx, id, status)
	}
	return nil
}
func (m *mockUserRepo) Delete(ctx context.Context, id uuid.UUID) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id)
	}
	return nil
}
func (m *mockUserRepo) Search(ctx context.Context, query string, limit int) ([]models.User, error) {
	if m.searchFn != nil {
		return m.searchFn(ctx, query, limit)
	}
	return nil, nil
}

type mockConvRepo struct {
	createFn       func(ctx context.Context, conv *models.Conversation) error
	getByIDFn      func(ctx context.Context, id uuid.UUID) (*models.Conversation, error)
	listByUserIDFn func(ctx context.Context, userID uuid.UUID) ([]models.Conversation, error)
	listPublicFn   func(ctx context.Context, limit, offset int) ([]models.Conversation, error)
	searchPublicFn func(ctx context.Context, query string, limit int) ([]models.Conversation, error)
	findPrivateFn  func(ctx context.Context, userID1, userID2 uuid.UUID) (*models.Conversation, error)
	addMemberFn    func(ctx context.Context, member *models.ConversationMember) error
	removeMemberFn func(ctx context.Context, convID, userID uuid.UUID) error
	getMembersFn   func(ctx context.Context, convID uuid.UUID) ([]models.ConversationMember, error)
	isMemberFn     func(ctx context.Context, convID, userID uuid.UUID) (bool, error)
	deleteFn       func(ctx context.Context, id uuid.UUID) error
}

func (m *mockConvRepo) Create(ctx context.Context, conv *models.Conversation) error {
	if m.createFn != nil {
		return m.createFn(ctx, conv)
	}
	return nil
}
func (m *mockConvRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.Conversation, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, nil
}
func (m *mockConvRepo) ListByUserID(ctx context.Context, userID uuid.UUID) ([]models.Conversation, error) {
	if m.listByUserIDFn != nil {
		return m.listByUserIDFn(ctx, userID)
	}
	return nil, nil
}
func (m *mockConvRepo) ListPublic(ctx context.Context, limit, offset int) ([]models.Conversation, error) {
	if m.listPublicFn != nil {
		return m.listPublicFn(ctx, limit, offset)
	}
	return nil, nil
}
func (m *mockConvRepo) SearchPublic(ctx context.Context, query string, limit int) ([]models.Conversation, error) {
	if m.searchPublicFn != nil {
		return m.searchPublicFn(ctx, query, limit)
	}
	return nil, nil
}
func (m *mockConvRepo) FindPrivate(ctx context.Context, userID1, userID2 uuid.UUID) (*models.Conversation, error) {
	if m.findPrivateFn != nil {
		return m.findPrivateFn(ctx, userID1, userID2)
	}
	return nil, nil
}
func (m *mockConvRepo) AddMember(ctx context.Context, member *models.ConversationMember) error {
	if m.addMemberFn != nil {
		return m.addMemberFn(ctx, member)
	}
	return nil
}
func (m *mockConvRepo) RemoveMember(ctx context.Context, convID, userID uuid.UUID) error {
	if m.removeMemberFn != nil {
		return m.removeMemberFn(ctx, convID, userID)
	}
	return nil
}
func (m *mockConvRepo) GetMembers(ctx context.Context, convID uuid.UUID) ([]models.ConversationMember, error) {
	if m.getMembersFn != nil {
		return m.getMembersFn(ctx, convID)
	}
	return nil, nil
}
func (m *mockConvRepo) IsMember(ctx context.Context, convID, userID uuid.UUID) (bool, error) {
	if m.isMemberFn != nil {
		return m.isMemberFn(ctx, convID, userID)
	}
	return false, nil
}
func (m *mockConvRepo) Delete(ctx context.Context, id uuid.UUID) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id)
	}
	return nil
}

type mockMsgRepo struct {
	createFn         func(ctx context.Context, msg *models.Message) error
	getByIDFn        func(ctx context.Context, id uuid.UUID) (*models.Message, error)
	listByConvFn     func(ctx context.Context, convID uuid.UUID, cursor uuid.UUID, limit int) ([]models.Message, error)
	deleteFn         func(ctx context.Context, id, senderID uuid.UUID) error
	addReactionFn    func(ctx context.Context, reaction *models.Reaction) error
	removeReactionFn func(ctx context.Context, messageID, userID uuid.UUID, emoji string) error
	getReactionsFn   func(ctx context.Context, messageID uuid.UUID) ([]models.Reaction, error)
	searchByConvFn   func(ctx context.Context, convID uuid.UUID, query string, limit, offset int) ([]models.Message, error)
}

func (m *mockMsgRepo) Create(ctx context.Context, msg *models.Message) error {
	if m.createFn != nil {
		return m.createFn(ctx, msg)
	}
	return nil
}
func (m *mockMsgRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.Message, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, nil
}
func (m *mockMsgRepo) ListByConversation(ctx context.Context, convID uuid.UUID, cursor uuid.UUID, limit int) ([]models.Message, error) {
	if m.listByConvFn != nil {
		return m.listByConvFn(ctx, convID, cursor, limit)
	}
	return nil, nil
}
func (m *mockMsgRepo) Delete(ctx context.Context, id, senderID uuid.UUID) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id, senderID)
	}
	return nil
}
func (m *mockMsgRepo) AddReaction(ctx context.Context, reaction *models.Reaction) error {
	if m.addReactionFn != nil {
		return m.addReactionFn(ctx, reaction)
	}
	return nil
}
func (m *mockMsgRepo) RemoveReaction(ctx context.Context, messageID, userID uuid.UUID, emoji string) error {
	if m.removeReactionFn != nil {
		return m.removeReactionFn(ctx, messageID, userID, emoji)
	}
	return nil
}
func (m *mockMsgRepo) GetReactionsByMessage(ctx context.Context, messageID uuid.UUID) ([]models.Reaction, error) {
	if m.getReactionsFn != nil {
		return m.getReactionsFn(ctx, messageID)
	}
	return nil, nil
}
func (m *mockMsgRepo) SearchByConversation(ctx context.Context, convID uuid.UUID, query string, limit, offset int) ([]models.Message, error) {
	if m.searchByConvFn != nil {
		return m.searchByConvFn(ctx, convID, query, limit, offset)
	}
	return nil, nil
}

type mockSessionRepo struct {
	createFn         func(ctx context.Context, session *models.Session) error
	getByTokenFn     func(ctx context.Context, refreshToken string) (*models.Session, error)
	deleteFn         func(ctx context.Context, id uuid.UUID) error
	deleteByUserIDFn func(ctx context.Context, userID uuid.UUID) error
	cleanExpiredFn   func(ctx context.Context) error
}

func (m *mockSessionRepo) Create(ctx context.Context, session *models.Session) error {
	if m.createFn != nil {
		return m.createFn(ctx, session)
	}
	return nil
}
func (m *mockSessionRepo) GetByToken(ctx context.Context, refreshToken string) (*models.Session, error) {
	if m.getByTokenFn != nil {
		return m.getByTokenFn(ctx, refreshToken)
	}
	return nil, nil
}
func (m *mockSessionRepo) Delete(ctx context.Context, id uuid.UUID) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id)
	}
	return nil
}
func (m *mockSessionRepo) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	if m.deleteByUserIDFn != nil {
		return m.deleteByUserIDFn(ctx, userID)
	}
	return nil
}
func (m *mockSessionRepo) CleanExpired(ctx context.Context) error {
	if m.cleanExpiredFn != nil {
		return m.cleanExpiredFn(ctx)
	}
	return nil
}

type mockNotifRepo struct {
	createFn         func(ctx context.Context, notification *models.Notification) error
	listFn           func(ctx context.Context, userID uuid.UUID, limit int) ([]models.Notification, error)
	markReadFn       func(ctx context.Context, id, userID uuid.UUID) error
	markAllReadFn    func(ctx context.Context, userID uuid.UUID) error
	getUnreadCountFn func(ctx context.Context, userID uuid.UUID) (int64, error)
	deleteFn         func(ctx context.Context, id, userID uuid.UUID) error
}

func (m *mockNotifRepo) Create(ctx context.Context, notification *models.Notification) error {
	if m.createFn != nil {
		return m.createFn(ctx, notification)
	}
	return nil
}
func (m *mockNotifRepo) List(ctx context.Context, userID uuid.UUID, limit int) ([]models.Notification, error) {
	if m.listFn != nil {
		return m.listFn(ctx, userID, limit)
	}
	return nil, nil
}
func (m *mockNotifRepo) MarkRead(ctx context.Context, id, userID uuid.UUID) error {
	if m.markReadFn != nil {
		return m.markReadFn(ctx, id, userID)
	}
	return nil
}
func (m *mockNotifRepo) MarkAllRead(ctx context.Context, userID uuid.UUID) error {
	if m.markAllReadFn != nil {
		return m.markAllReadFn(ctx, userID)
	}
	return nil
}
func (m *mockNotifRepo) GetUnreadCount(ctx context.Context, userID uuid.UUID) (int64, error) {
	if m.getUnreadCountFn != nil {
		return m.getUnreadCountFn(ctx, userID)
	}
	return 0, nil
}
func (m *mockNotifRepo) Delete(ctx context.Context, id, userID uuid.UUID) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id, userID)
	}
	return nil
}

type mockAttachmentRepo struct {
	createFn         func(ctx context.Context, attachment *models.Attachment) error
	getByMessageIDFn func(ctx context.Context, messageID uuid.UUID) ([]models.Attachment, error)
}

func (m *mockAttachmentRepo) Create(ctx context.Context, attachment *models.Attachment) error {
	if m.createFn != nil {
		return m.createFn(ctx, attachment)
	}
	return nil
}
func (m *mockAttachmentRepo) GetByMessageID(ctx context.Context, messageID uuid.UUID) ([]models.Attachment, error) {
	if m.getByMessageIDFn != nil {
		return m.getByMessageIDFn(ctx, messageID)
	}
	return nil, nil
}

type mockMessageReadRepo struct {
	upsertFn            func(ctx context.Context, convID, userID uuid.UUID) error
	getByConversationFn func(ctx context.Context, convID uuid.UUID) ([]models.MessageRead, error)
	getUnreadCountsFn   func(ctx context.Context, userID uuid.UUID) (map[uuid.UUID]int64, error)
}

func (m *mockMessageReadRepo) Upsert(ctx context.Context, convID, userID uuid.UUID) error {
	if m.upsertFn != nil {
		return m.upsertFn(ctx, convID, userID)
	}
	return nil
}
func (m *mockMessageReadRepo) GetByConversation(ctx context.Context, convID uuid.UUID) ([]models.MessageRead, error) {
	if m.getByConversationFn != nil {
		return m.getByConversationFn(ctx, convID)
	}
	return nil, nil
}
func (m *mockMessageReadRepo) GetUnreadCounts(ctx context.Context, userID uuid.UUID) (map[uuid.UUID]int64, error) {
	if m.getUnreadCountsFn != nil {
		return m.getUnreadCountsFn(ctx, userID)
	}
	return nil, nil
}

type mockPresenceRepo struct {
	upsertFn      func(ctx context.Context, presence *models.Presence) error
	getByUserIDFn func(ctx context.Context, userID uuid.UUID) (*models.Presence, error)
	getOnlineFn   func(ctx context.Context) ([]models.Presence, error)
}

func (m *mockPresenceRepo) Upsert(ctx context.Context, presence *models.Presence) error {
	if m.upsertFn != nil {
		return m.upsertFn(ctx, presence)
	}
	return nil
}
func (m *mockPresenceRepo) GetByUserID(ctx context.Context, userID uuid.UUID) (*models.Presence, error) {
	if m.getByUserIDFn != nil {
		return m.getByUserIDFn(ctx, userID)
	}
	return nil, nil
}
func (m *mockPresenceRepo) GetOnline(ctx context.Context) ([]models.Presence, error) {
	if m.getOnlineFn != nil {
		return m.getOnlineFn(ctx)
	}
	return nil, nil
}

// newMockRepos creates a *repository.Repositories with all mock repositories.
// Default behavior: all methods return zero values (nil, 0, false).
// Override individual mock functions in test cases to customize behavior.
func newMockRepos() *repository.Repositories {
	return &repository.Repositories{
		Users:         &mockUserRepo{},
		Conversations: &mockConvRepo{},
		Messages:      &mockMsgRepo{},
		Sessions:      &mockSessionRepo{},
		Notifications: &mockNotifRepo{},
		Attachments:   &mockAttachmentRepo{},
		Presence:      &mockPresenceRepo{},
		MessageRead:   &mockMessageReadRepo{},
	}
}

// Helper to create a test context with user_id
func authContext(userID uuid.UUID) context.Context {
	ctx := context.WithValue(context.Background(), "user_id", userID)
	return context.WithValue(ctx, "username", "testuser")
}

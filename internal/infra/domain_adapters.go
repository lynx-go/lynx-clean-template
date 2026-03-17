package infra

import (
	"context"

	"github.com/lynx-go/lynx-clean-template/internal/domain/shared"
	"github.com/lynx-go/lynx-clean-template/pkg/pubsub"
	"github.com/lynx-go/x/log"
	"golang.org/x/crypto/bcrypt"
)

// --- Logger adapter ---

type lynxLogger struct{}

func (l *lynxLogger) ErrorContext(ctx context.Context, msg string, err error, keysAndValues ...any) {
	log.ErrorContext(ctx, msg, err, keysAndValues...)
}

func (l *lynxLogger) InfoContext(ctx context.Context, msg string, keysAndValues ...any) {
	log.InfoContext(ctx, msg, keysAndValues...)
}

func (l *lynxLogger) DebugContext(ctx context.Context, msg string, keysAndValues ...any) {
	log.DebugContext(ctx, msg, keysAndValues...)
}

// NewDomainLogger provides a shared.Logger implementation backed by lynx-go/x/log.
func NewDomainLogger() shared.Logger {
	return &lynxLogger{}
}

// --- EventPublisher adapter ---

type pubsubEventPublisher struct {
	pub pubsub.Publisher
}

func (p *pubsubEventPublisher) Publish(ctx context.Context, topic, event string, data any) error {
	return p.pub.Publish(ctx, pubsub.TopicName(topic), pubsub.EventName(event), data)
}

// NewDomainEventPublisher provides a shared.EventPublisher backed by pkg/pubsub.
func NewDomainEventPublisher(pub pubsub.Publisher) shared.EventPublisher {
	return &pubsubEventPublisher{pub: pub}
}

// --- PasswordHasher adapter ---

type bcryptPasswordHasher struct{}

func (h *bcryptPasswordHasher) Hash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func (h *bcryptPasswordHasher) Compare(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

// NewPasswordHasher provides a shared.PasswordHasher backed by bcrypt.
func NewPasswordHasher() shared.PasswordHasher {
	return &bcryptPasswordHasher{}
}

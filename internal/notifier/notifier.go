package notifier

import "context"

// ChatAction represents a status action indicator sent to notification platforms.
type ChatAction string

const (
	ChatActionTyping      ChatAction = "typing"
	ChatActionUploadPhoto ChatAction = "upload_photo"
)

// Notifier defines the common interface for sending messages and media alerts.
type Notifier interface {
	SendMessage(ctx context.Context, text string) error
	SendPhoto(ctx context.Context, photo []byte, caption string) error
	SendChatAction(ctx context.Context, action ChatAction) error
}

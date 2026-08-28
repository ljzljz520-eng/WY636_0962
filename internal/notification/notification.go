package notification

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"example.com/animalcage/internal/model"
)

type Channel string

const (
	ChannelInbox   Channel = "inbox"
	ChannelEmail   Channel = "email"
	ChannelWebhook Channel = "webhook"
)

type Recipient struct {
	ID          string
	DisplayName string
	Address     string
	Channels    []Channel
	Active      bool
}
type Message struct {
	ID          string
	RecordID    string
	RecipientID string
	Channel     Channel
	Subject     string
	Body        string
	CreatedAt   time.Time
	ReadAt      *time.Time
}
type Template struct {
	Key            string
	Subject        string
	Body           string
	RequiredFields []string
}
type Delivery struct {
	Message   Message
	Delivered bool
	Attempt   int
	Error     string
}

func ValidateRecipient(recipient Recipient) error {
	if strings.TrimSpace(recipient.ID) == "" {
		return fmt.Errorf("recipient id is required")
	}
	if strings.TrimSpace(recipient.DisplayName) == "" {
		return fmt.Errorf("recipient display name is required")
	}
	if strings.TrimSpace(recipient.Address) == "" {
		return fmt.Errorf("recipient address is required")
	}
	if !recipient.Active {
		return fmt.Errorf("recipient is inactive")
	}
	if len(recipient.Channels) == 0 {
		return fmt.Errorf("recipient has no channels")
	}
	return nil
}
func HasChannel(recipient Recipient, channel Channel) bool {
	for _, candidate := range recipient.Channels {
		if candidate == channel {
			return true
		}
	}
	return false
}
func ResolveChannel(recipient Recipient, preferred Channel) (Channel, error) {
	if preferred != "" && HasChannel(recipient, preferred) {
		return preferred, nil
	}
	for _, candidate := range []Channel{ChannelInbox, ChannelEmail, ChannelWebhook} {
		if HasChannel(recipient, candidate) {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("recipient has no usable channel")
}

func Render(template Template, values map[string]string) (string, string, error) {
	for _, field := range template.RequiredFields {
		if strings.TrimSpace(values[field]) == "" {
			return "", "", fmt.Errorf("template field %s is required", field)
		}
	}
	subject := template.Subject
	body := template.Body
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		marker := "{{" + key + "}}"
		subject = strings.ReplaceAll(subject, marker, values[key])
		body = strings.ReplaceAll(body, marker, values[key])
	}
	return subject, body, nil
}

func BuildReviewMessage(record model.Record, recipient Recipient, template Template, now time.Time) (Message, error) {
	if err := ValidateRecipient(recipient); err != nil {
		return Message{}, err
	}
	channel, err := ResolveChannel(recipient, ChannelInbox)
	if err != nil {
		return Message{}, err
	}
	subject, body, err := Render(template, map[string]string{"application_no": record.ApplicationNo, "applicant": record.Applicant, "facility": record.Facility, "status": record.Status})
	if err != nil {
		return Message{}, err
	}
	return Message{ID: fmt.Sprintf("%s-%s-%d", record.ID, recipient.ID, now.UnixNano()), RecordID: record.ID, RecipientID: recipient.ID, Channel: channel, Subject: subject, Body: body, CreatedAt: now.UTC()}, nil
}

func MarkRead(message Message, at time.Time) Message {
	copy := message
	value := at.UTC()
	copy.ReadAt = &value
	return copy
}
func IsUnread(message Message) bool { return message.ReadAt == nil }
func SortMessages(messages []Message) []Message {
	out := append([]Message(nil), messages...)
	sort.Slice(out, func(i, j int) bool {
		if out[i].CreatedAt.Equal(out[j].CreatedAt) {
			return out[i].ID < out[j].ID
		}
		return out[i].CreatedAt.Before(out[j].CreatedAt)
	})
	return out
}
func UnreadFor(messages []Message, recipientID string) []Message {
	out := make([]Message, 0)
	for _, message := range messages {
		if message.RecipientID == recipientID && IsUnread(message) {
			out = append(out, message)
		}
	}
	return SortMessages(out)
}
func BuildDigest(messages []Message, now time.Time) string {
	sorted := SortMessages(messages)
	lines := []string{fmt.Sprintf("digest %s", now.UTC().Format(time.RFC3339))}
	for _, message := range sorted {
		lines = append(lines, fmt.Sprintf("%s %s %s", message.RecordID, message.Subject, message.Channel))
	}
	return strings.Join(lines, "\n")
}
func Deliver(message Message, attempt int) Delivery {
	if attempt < 1 {
		attempt = 1
	}
	if strings.TrimSpace(message.ID) == "" {
		return Delivery{Message: message, Attempt: attempt, Error: "message id is required"}
	}
	if strings.TrimSpace(message.Body) == "" {
		return Delivery{Message: message, Attempt: attempt, Error: "message body is empty"}
	}
	return Delivery{Message: message, Delivered: true, Attempt: attempt}
}
func Retry(delivery Delivery, maxAttempts int) Delivery {
	if delivery.Delivered || maxAttempts <= delivery.Attempt {
		return delivery
	}
	return Deliver(delivery.Message, delivery.Attempt+1)
}

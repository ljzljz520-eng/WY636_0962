package notification

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"example.com/animalcage/internal/model"
)

type Rule struct {
	Event    string
	Status   string
	Channels []Channel
	Enabled  bool
}
type Router struct {
	rules      []Rule
	recipients map[string]Recipient
}
type Event struct {
	Name   string
	Record model.Record
	Actor  string
	At     time.Time
	Detail string
}

func NewRouter() *Router {
	return &Router{rules: make([]Rule, 0), recipients: make(map[string]Recipient)}
}
func (r *Router) AddRecipient(recipient Recipient) error {
	if err := ValidateRecipient(recipient); err != nil {
		return err
	}
	if _, exists := r.recipients[recipient.ID]; exists {
		return fmt.Errorf("recipient %s already exists", recipient.ID)
	}
	r.recipients[recipient.ID] = recipient
	return nil
}
func (r *Router) UpsertRecipient(recipient Recipient) error {
	if err := ValidateRecipient(recipient); err != nil {
		return err
	}
	r.recipients[recipient.ID] = recipient
	return nil
}
func (r *Router) AddRule(rule Rule) error {
	if strings.TrimSpace(rule.Event) == "" {
		return fmt.Errorf("event is required")
	}
	if rule.Status == "" {
		return fmt.Errorf("status is required")
	}
	if len(rule.Channels) == 0 {
		return fmt.Errorf("at least one channel is required")
	}
	r.rules = append(r.rules, rule)
	sort.SliceStable(r.rules, func(i, j int) bool { return r.rules[i].Event < r.rules[j].Event })
	return nil
}
func (r *Router) RemoveRules(event string) {
	kept := r.rules[:0]
	for _, rule := range r.rules {
		if rule.Event != event {
			kept = append(kept, rule)
		}
	}
	r.rules = kept
}
func (r *Router) Recipients() []Recipient {
	out := make([]Recipient, 0, len(r.recipients))
	for _, recipient := range r.recipients {
		copy := recipient
		copy.Channels = append([]Channel(nil), recipient.Channels...)
		out = append(out, copy)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}
func (r *Router) Rules() []Rule {
	out := append([]Rule(nil), r.rules...)
	for i := range out {
		out[i].Channels = append([]Channel(nil), out[i].Channels...)
	}
	return out
}
func (r *Router) Match(event Event) []Rule {
	matches := make([]Rule, 0)
	for _, rule := range r.rules {
		if !rule.Enabled || rule.Event != event.Name {
			continue
		}
		if rule.Status != "*" && rule.Status != event.Record.Status {
			continue
		}
		matches = append(matches, rule)
	}
	return matches
}
func (r *Router) Build(event Event, template Template) ([]Message, error) {
	rules := r.Match(event)
	messages := make([]Message, 0)
	for _, rule := range rules {
		for _, recipient := range r.Recipients() {
			channel, err := ResolveChannel(recipient, rule.Channels[0])
			if err != nil {
				continue
			}
			subject, body, err := Render(template, map[string]string{"application_no": event.Record.ApplicationNo, "applicant": event.Record.Applicant, "facility": event.Record.Facility, "status": event.Record.Status, "actor": event.Actor, "detail": event.Detail})
			if err != nil {
				return nil, err
			}
			messages = append(messages, Message{ID: fmt.Sprintf("%s-%s-%s-%d", event.Record.ID, recipient.ID, channel, event.At.UnixNano()), RecordID: event.Record.ID, RecipientID: recipient.ID, Channel: channel, Subject: subject, Body: body, CreatedAt: event.At.UTC()})
		}
	}
	return SortMessages(messages), nil
}
func (r *Router) Dispatch(event Event, template Template) []Delivery {
	messages, err := r.Build(event, template)
	if err != nil {
		return []Delivery{{Attempt: 1, Error: err.Error()}}
	}
	deliveries := make([]Delivery, 0, len(messages))
	for _, message := range messages {
		deliveries = append(deliveries, Deliver(message, 1))
	}
	return deliveries
}
func (r *Router) Pending(deliveries []Delivery) []Delivery {
	out := make([]Delivery, 0)
	for _, delivery := range deliveries {
		if !delivery.Delivered {
			out = append(out, delivery)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Message.ID < out[j].Message.ID })
	return out
}
func (r *Router) RetryPending(deliveries []Delivery, maxAttempts int) []Delivery {
	out := make([]Delivery, len(deliveries))
	for i, delivery := range deliveries {
		out[i] = Retry(delivery, maxAttempts)
	}
	return out
}

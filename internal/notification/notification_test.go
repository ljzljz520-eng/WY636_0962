package notification

import (
	"example.com/animalcage/internal/model"
	"strings"
	"testing"
	"time"
)

func TestReviewNotification(t *testing.T) {
	recipient := Recipient{ID: "u1", DisplayName: "reviewer", Address: "inbox:u1", Channels: []Channel{ChannelInbox}, Active: true}
	template := Template{Subject: "application {{application_no}}", Body: "{{applicant}} / {{status}}", RequiredFields: []string{"application_no", "applicant", "status"}}
	message, err := BuildReviewMessage(model.Record{ID: "r1", ApplicationNo: "A1", Applicant: "applicant", Status: model.StatusSubmitted}, recipient, template, time.Unix(1, 0))
	if err != nil {
		t.Fatal(err)
	}
	delivery := Deliver(message, 1)
	if !delivery.Delivered || !strings.Contains(message.Subject, "A1") {
		t.Fatal(delivery)
	}
}

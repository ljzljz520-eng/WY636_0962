package report

import (
	"bytes"
	"example.com/animalcage/internal/model"
	"testing"
	"time"
)

func TestReport(t *testing.T) {
	rows := []model.Record{{ID: "1", ApplicationNo: "A", Facility: "F", Status: model.StatusApproved, CageCount: 2}}
	summary := BuildSummary(rows, time.Unix(0, 0))
	if summary.Total != 1 || summary.CageTotal != 2 {
		t.Fatal(summary)
	}
	var b bytes.Buffer
	if err := EncodeCSV(rows, &b); err != nil {
		t.Fatal(err)
	}
	if b.Len() == 0 {
		t.Fatal("empty csv")
	}
}

package backup

import (
	"example.com/animalcage/internal/model"
	"testing"
	"time"
)

func TestBackupRoundTrip(t *testing.T) {
	bundle := NewBundle(time.Unix(10, 0), []model.Record{{ID: "r1", ApplicationNo: "A", Roster: []string{"m"}}}, nil, nil, nil)
	packed, err := Compress(bundle)
	if err != nil {
		t.Fatal(err)
	}
	restored, err := Decompress(packed)
	if err != nil {
		t.Fatal(err)
	}
	if !ContainsRecord(restored, "r1") || len(RecordIDs(restored)) != 1 {
		t.Fatal("record missing")
	}
	if MakeManifest(restored).Checksum == "" {
		t.Fatal("manifest checksum missing")
	}
}

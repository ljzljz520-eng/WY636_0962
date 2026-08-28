package backup

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	"example.com/animalcage/internal/model"
)

type Bundle struct {
	Schema    int                `json:"schema"`
	CreatedAt time.Time          `json:"created_at"`
	Records   []model.Record     `json:"records"`
	Events    []model.AuditEvent `json:"events"`
	Workflows []model.Workflow   `json:"workflows"`
	Files     []model.Attachment `json:"files"`
	Checksum  string             `json:"checksum"`
}

type Manifest struct {
	Schema        int       `json:"schema"`
	CreatedAt     time.Time `json:"created_at"`
	RecordCount   int       `json:"record_count"`
	EventCount    int       `json:"event_count"`
	WorkflowCount int       `json:"workflow_count"`
	FileCount     int       `json:"file_count"`
	Checksum      string    `json:"checksum"`
}

func NewBundle(now time.Time, records []model.Record, events []model.AuditEvent, workflows []model.Workflow, files []model.Attachment) Bundle {
	copyRecords := make([]model.Record, len(records))
	for i, value := range records {
		copyRecords[i] = model.CopyRecord(value)
	}
	copyEvents := model.CopyEvents(events)
	copyWorkflows := append([]model.Workflow(nil), workflows...)
	copyFiles := append([]model.Attachment(nil), files...)
	sort.Slice(copyRecords, func(i, j int) bool { return copyRecords[i].ID < copyRecords[j].ID })
	sort.Slice(copyEvents, func(i, j int) bool { return copyEvents[i].ID < copyEvents[j].ID })
	sort.Slice(copyWorkflows, func(i, j int) bool { return copyWorkflows[i].ID < copyWorkflows[j].ID })
	sort.Slice(copyFiles, func(i, j int) bool { return copyFiles[i].ID < copyFiles[j].ID })
	bundle := Bundle{Schema: 1, CreatedAt: now.UTC(), Records: copyRecords, Events: copyEvents, Workflows: copyWorkflows, Files: copyFiles}
	bundle.Checksum = checksum(bundle)
	return bundle
}

func checksum(bundle Bundle) string {
	clone := bundle
	clone.Checksum = ""
	data, _ := json.Marshal(clone)
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func Validate(bundle Bundle) error {
	if bundle.Schema != 1 {
		return fmt.Errorf("unsupported backup schema %d", bundle.Schema)
	}
	if bundle.CreatedAt.IsZero() {
		return fmt.Errorf("backup creation time is required")
	}
	if strings.TrimSpace(bundle.Checksum) == "" {
		return fmt.Errorf("backup checksum is required")
	}
	if checksum(bundle) != bundle.Checksum {
		return fmt.Errorf("backup checksum mismatch")
	}
	ids := make(map[string]bool, len(bundle.Records))
	for _, record := range bundle.Records {
		if strings.TrimSpace(record.ID) == "" {
			return fmt.Errorf("backup contains record without id")
		}
		if ids[record.ID] {
			return fmt.Errorf("duplicate record %s", record.ID)
		}
		ids[record.ID] = true
	}
	return nil
}

func Encode(bundle Bundle) ([]byte, error) {
	if err := Validate(bundle); err != nil {
		return nil, err
	}
	return json.MarshalIndent(bundle, "", "  ")
}
func Decode(data []byte) (Bundle, error) {
	var bundle Bundle
	if err := json.Unmarshal(data, &bundle); err != nil {
		return Bundle{}, err
	}
	if err := Validate(bundle); err != nil {
		return Bundle{}, err
	}
	return bundle, nil
}

func Compress(bundle Bundle) ([]byte, error) {
	data, err := Encode(bundle)
	if err != nil {
		return nil, err
	}
	var out bytes.Buffer
	writer := gzip.NewWriter(&out)
	if _, err = writer.Write(data); err != nil {
		return nil, err
	}
	if err = writer.Close(); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}
func Decompress(data []byte) (Bundle, error) {
	reader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return Bundle{}, err
	}
	decoded, readErr := io.ReadAll(reader)
	closeErr := reader.Close()
	if readErr != nil {
		return Bundle{}, readErr
	}
	if closeErr != nil {
		return Bundle{}, closeErr
	}
	return Decode(decoded)
}

func MakeManifest(bundle Bundle) Manifest {
	return Manifest{Schema: bundle.Schema, CreatedAt: bundle.CreatedAt, RecordCount: len(bundle.Records), EventCount: len(bundle.Events), WorkflowCount: len(bundle.Workflows), FileCount: len(bundle.Files), Checksum: bundle.Checksum}
}
func RecordIDs(bundle Bundle) []string {
	ids := make([]string, len(bundle.Records))
	for i, record := range bundle.Records {
		ids[i] = record.ID
	}
	sort.Strings(ids)
	return ids
}
func ContainsRecord(bundle Bundle, id string) bool {
	for _, record := range bundle.Records {
		if record.ID == id {
			return true
		}
	}
	return false
}
func Merge(base, additions Bundle) (Bundle, error) {
	records := append([]model.Record(nil), base.Records...)
	index := make(map[string]int, len(records))
	for i, record := range records {
		index[record.ID] = i
	}
	for _, record := range additions.Records {
		if i, ok := index[record.ID]; ok {
			records[i] = model.CopyRecord(record)
		} else {
			index[record.ID] = len(records)
			records = append(records, model.CopyRecord(record))
		}
	}
	events := append(append([]model.AuditEvent(nil), base.Events...), additions.Events...)
	flows := append(append([]model.Workflow(nil), base.Workflows...), additions.Workflows...)
	files := append(append([]model.Attachment(nil), base.Files...), additions.Files...)
	return NewBundle(base.CreatedAt, records, events, flows, files), nil
}

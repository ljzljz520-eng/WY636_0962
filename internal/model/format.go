package model

import "strings"

func NormalizeRoster(values []string) []string {
	seen := make(map[string]bool, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		name := strings.TrimSpace(value)
		if name == "" || seen[name] {
			continue
		}
		seen[name] = true
		result = append(result, name)
	}
	return result
}

func CopyRecord(in Record) Record {
	out := in
	out.Roster = append([]string(nil), in.Roster...)
	return out
}

func CopyEvents(in []AuditEvent) []AuditEvent {
	out := make([]AuditEvent, len(in))
	copy(out, in)
	return out
}

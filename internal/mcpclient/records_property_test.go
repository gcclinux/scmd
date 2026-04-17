package mcpclient

import (
	"fmt"
	"testing"

	"pgregory.net/rapid"
)

// Feature: mcp-client-integration, Property 4: MCP record to CommandRecord mapping preserves fields
// **Validates: Requirements 5.3, 6.2, 8.3**
func TestProperty_MCPRecordToCommandRecordMappingPreservesFields(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		// Generate a random UUID for the ID field.
		uuid := rapid.StringMatching(
			`[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`,
		).Draw(rt, "uuid")

		// Generate random non-empty strings for Key and Content.
		key := rapid.StringMatching(`[a-zA-Z0-9 _\-\.]{1,100}`).Draw(rt, "key")
		content := rapid.StringMatching(`[a-zA-Z0-9 _\-\.]{1,200}`).Draw(rt, "content")

		record := MCPRecord{
			ID:      uuid,
			Key:     key,
			Content: content,
			Metadata: map[string]any{
				"source": rapid.StringMatching(`[a-z]{1,10}`).Draw(rt, "source"),
			},
		}

		mapper := NewIDMapper()
		cmdRecord := record.ToCommandRecord(mapper)

		// 1. CommandRecord.Key must equal MCPRecord.Key.
		if cmdRecord.Key != record.Key {
			rt.Fatalf("Key mismatch: got %q, want %q", cmdRecord.Key, record.Key)
		}

		// 2. CommandRecord.Data must equal MCPRecord.Content.
		if cmdRecord.Data != record.Content {
			rt.Fatalf("Data mismatch: got %q, want %q", cmdRecord.Data, record.Content)
		}

		// 3. CommandRecord.Id must be a positive integer (> 0).
		if cmdRecord.Id <= 0 {
			rt.Fatalf("Id must be positive, got %d", cmdRecord.Id)
		}
	})
}

// Feature: mcp-client-integration, Property 4 (batch): Multiple MCPRecords get distinct positive IDs
// **Validates: Requirements 5.3, 6.2, 8.3**
func TestProperty_MCPRecordToCommandRecordBatchDistinctIDs(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		count := rapid.IntRange(1, 30).Draw(rt, "count")

		// Generate distinct UUIDs.
		seen := make(map[string]bool, count)
		records := make([]MCPRecord, 0, count)
		for len(records) < count {
			uuid := rapid.StringMatching(
				`[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`,
			).Draw(rt, fmt.Sprintf("uuid_%d", len(records)))
			if seen[uuid] {
				continue
			}
			seen[uuid] = true

			records = append(records, MCPRecord{
				ID:      uuid,
				Key:     rapid.StringMatching(`[a-zA-Z0-9]{1,50}`).Draw(rt, fmt.Sprintf("key_%d", len(records)-1)),
				Content: rapid.StringMatching(`[a-zA-Z0-9]{1,100}`).Draw(rt, fmt.Sprintf("content_%d", len(records)-1)),
			})
		}

		mapper := NewIDMapper()
		idSet := make(map[int]bool, count)

		for i, rec := range records {
			cmdRec := rec.ToCommandRecord(mapper)

			// Key and Data must match.
			if cmdRec.Key != rec.Key {
				rt.Fatalf("record[%d] Key mismatch: got %q, want %q", i, cmdRec.Key, rec.Key)
			}
			if cmdRec.Data != rec.Content {
				rt.Fatalf("record[%d] Data mismatch: got %q, want %q", i, cmdRec.Data, rec.Content)
			}

			// Id must be positive.
			if cmdRec.Id <= 0 {
				rt.Fatalf("record[%d] Id must be positive, got %d", i, cmdRec.Id)
			}

			// Id must be unique across all records.
			if idSet[cmdRec.Id] {
				rt.Fatalf("record[%d] duplicate Id %d", i, cmdRec.Id)
			}
			idSet[cmdRec.Id] = true
		}
	})
}

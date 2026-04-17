package mcpclient

import (
	"fmt"
	"testing"

	"pgregory.net/rapid"
)

// Feature: mcp-client-integration, Property 7: ID mapper assigns sequential IDs and resolves them back to UUIDs
// **Validates: Requirements 12.1, 12.2, 12.3**
func TestProperty_IDMapperSequentialAssignmentAndResolution(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		// Generate a list of 1-50 distinct UUIDs.
		count := rapid.IntRange(1, 50).Draw(rt, "count")
		seen := make(map[string]bool, count)
		uuids := make([]string, 0, count)
		for len(uuids) < count {
			uuid := rapid.StringMatching(
				`[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`,
			).Draw(rt, fmt.Sprintf("uuid_%d", len(uuids)))
			if seen[uuid] {
				continue
			}
			seen[uuid] = true
			uuids = append(uuids, uuid)
		}

		m := NewIDMapper()
		ids := m.Assign(uuids)

		// 1. The number of returned IDs must match the number of input UUIDs.
		if len(ids) != len(uuids) {
			rt.Fatalf("len(ids) = %d, want %d", len(ids), len(uuids))
		}

		// 2. IDs must be sequential starting from 1.
		for i, id := range ids {
			want := i + 1
			if id != want {
				rt.Fatalf("ids[%d] = %d, want %d", i, id, want)
			}
		}

		// 3. Each ID resolves back to the original UUID via ToUUID.
		for i, id := range ids {
			got, err := m.ToUUID(id)
			if err != nil {
				rt.Fatalf("ToUUID(%d): unexpected error: %v", id, err)
			}
			if got != uuids[i] {
				rt.Fatalf("ToUUID(%d) = %q, want %q", id, got, uuids[i])
			}
		}
	})
}

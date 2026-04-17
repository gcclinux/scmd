package mcpclient

import (
	"fmt"
	"strings"
	"sync"
	"testing"
)

func TestNewIDMapper(t *testing.T) {
	m := NewIDMapper()
	if m == nil {
		t.Fatal("NewIDMapper returned nil")
	}
	if m.nextID != 1 {
		t.Errorf("nextID = %d, want 1", m.nextID)
	}
	if len(m.toUUID) != 0 {
		t.Errorf("toUUID has %d entries, want 0", len(m.toUUID))
	}
	if len(m.toInt) != 0 {
		t.Errorf("toInt has %d entries, want 0", len(m.toInt))
	}
}

func TestAssign_SequentialIDs(t *testing.T) {
	m := NewIDMapper()
	uuids := []string{
		"550e8400-e29b-41d4-a716-446655440000",
		"6ba7b810-9dad-11d1-80b4-00c04fd430c8",
		"f47ac10b-58cc-4372-a567-0e02b2c3d479",
	}

	ids := m.Assign(uuids)

	if len(ids) != 3 {
		t.Fatalf("len(ids) = %d, want 3", len(ids))
	}
	for i, id := range ids {
		want := i + 1
		if id != want {
			t.Errorf("ids[%d] = %d, want %d", i, id, want)
		}
	}
}

func TestAssign_AlreadyMappedUUIDsKeepExistingID(t *testing.T) {
	m := NewIDMapper()
	uuids := []string{"uuid-a", "uuid-b", "uuid-c"}
	firstIDs := m.Assign(uuids)

	// Assign again with some overlap and a new UUID.
	secondUUIDs := []string{"uuid-b", "uuid-d", "uuid-a"}
	secondIDs := m.Assign(secondUUIDs)

	// uuid-b should keep its original ID (2).
	if secondIDs[0] != firstIDs[1] {
		t.Errorf("uuid-b got ID %d, want %d", secondIDs[0], firstIDs[1])
	}
	// uuid-d is new, should get next sequential ID (4).
	if secondIDs[1] != 4 {
		t.Errorf("uuid-d got ID %d, want 4", secondIDs[1])
	}
	// uuid-a should keep its original ID (1).
	if secondIDs[2] != firstIDs[0] {
		t.Errorf("uuid-a got ID %d, want %d", secondIDs[2], firstIDs[0])
	}
}

func TestAssign_EmptySlice(t *testing.T) {
	m := NewIDMapper()
	ids := m.Assign([]string{})
	if len(ids) != 0 {
		t.Errorf("len(ids) = %d, want 0", len(ids))
	}
	// nextID should still be 1 since nothing was assigned.
	if m.nextID != 1 {
		t.Errorf("nextID = %d, want 1", m.nextID)
	}
}

func TestAssign_DuplicatesInSameCall(t *testing.T) {
	m := NewIDMapper()
	uuids := []string{"uuid-x", "uuid-x", "uuid-y"}
	ids := m.Assign(uuids)

	// First occurrence of uuid-x gets ID 1.
	if ids[0] != 1 {
		t.Errorf("ids[0] = %d, want 1", ids[0])
	}
	// Second occurrence of uuid-x should get the same ID.
	if ids[1] != 1 {
		t.Errorf("ids[1] = %d, want 1 (same as first uuid-x)", ids[1])
	}
	// uuid-y gets ID 2.
	if ids[2] != 2 {
		t.Errorf("ids[2] = %d, want 2", ids[2])
	}
}

func TestToUUID_Found(t *testing.T) {
	m := NewIDMapper()
	uuids := []string{"uuid-a", "uuid-b"}
	m.Assign(uuids)

	uuid, err := m.ToUUID(1)
	if err != nil {
		t.Fatalf("ToUUID(1): %v", err)
	}
	if uuid != "uuid-a" {
		t.Errorf("ToUUID(1) = %q, want %q", uuid, "uuid-a")
	}

	uuid, err = m.ToUUID(2)
	if err != nil {
		t.Fatalf("ToUUID(2): %v", err)
	}
	if uuid != "uuid-b" {
		t.Errorf("ToUUID(2) = %q, want %q", uuid, "uuid-b")
	}
}

func TestToUUID_NotFound(t *testing.T) {
	m := NewIDMapper()

	_, err := m.ToUUID(42)
	if err == nil {
		t.Fatal("expected error for unmapped ID, got nil")
	}
	wantMsg := "ID 42 not found"
	if !strings.Contains(err.Error(), wantMsg) {
		t.Errorf("error = %q, want it to contain %q", err.Error(), wantMsg)
	}
	if !strings.Contains(err.Error(), "run a search or list first") {
		t.Errorf("error = %q, want it to contain guidance text", err.Error())
	}
}

func TestToInt_Found(t *testing.T) {
	m := NewIDMapper()
	m.Assign([]string{"uuid-a", "uuid-b"})

	if id := m.ToInt("uuid-a"); id != 1 {
		t.Errorf("ToInt(uuid-a) = %d, want 1", id)
	}
	if id := m.ToInt("uuid-b"); id != 2 {
		t.Errorf("ToInt(uuid-b) = %d, want 2", id)
	}
}

func TestToInt_NotFound(t *testing.T) {
	m := NewIDMapper()

	if id := m.ToInt("nonexistent"); id != 0 {
		t.Errorf("ToInt(nonexistent) = %d, want 0", id)
	}
}

func TestReset(t *testing.T) {
	m := NewIDMapper()
	m.Assign([]string{"uuid-a", "uuid-b", "uuid-c"})

	m.Reset()

	if m.nextID != 1 {
		t.Errorf("after Reset, nextID = %d, want 1", m.nextID)
	}
	if len(m.toUUID) != 0 {
		t.Errorf("after Reset, toUUID has %d entries, want 0", len(m.toUUID))
	}
	if len(m.toInt) != 0 {
		t.Errorf("after Reset, toInt has %d entries, want 0", len(m.toInt))
	}

	// After reset, new assignments should start from 1 again.
	ids := m.Assign([]string{"uuid-new"})
	if ids[0] != 1 {
		t.Errorf("after Reset, first assigned ID = %d, want 1", ids[0])
	}
}

func TestReset_OldIDsNotResolvable(t *testing.T) {
	m := NewIDMapper()
	m.Assign([]string{"uuid-a"})
	m.Reset()

	_, err := m.ToUUID(1)
	if err == nil {
		t.Fatal("expected error for ID after Reset, got nil")
	}
}

func TestConcurrentAccess(t *testing.T) {
	m := NewIDMapper()
	var wg sync.WaitGroup

	// Concurrent writes.
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			uuid := fmt.Sprintf("uuid-%d", n)
			m.Assign([]string{uuid})
		}(i)
	}
	wg.Wait()

	// Concurrent reads.
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			uuid := fmt.Sprintf("uuid-%d", n)
			m.ToInt(uuid)
		}(i)
	}
	wg.Wait()

	// All 100 UUIDs should be mapped.
	for i := 0; i < 100; i++ {
		uuid := fmt.Sprintf("uuid-%d", i)
		id := m.ToInt(uuid)
		if id == 0 {
			t.Errorf("uuid-%d not found in mapper", i)
		}
	}
}

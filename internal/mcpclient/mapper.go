package mcpclient

import (
	"fmt"
	"sync"
)

// IDMapper maintains a session-scoped mapping between integer IDs and UUIDs.
// The MCP server uses UUIDs; scmd's CLI uses integer IDs. This mapper
// translates between the two, populated on list/search operations and
// consulted on get/delete/update operations.
type IDMapper struct {
	mu     sync.RWMutex
	toUUID map[int]string
	toInt  map[string]int
	nextID int
}

// NewIDMapper creates a new empty mapper with the first ID set to 1.
func NewIDMapper() *IDMapper {
	return &IDMapper{
		toUUID: make(map[int]string),
		toInt:  make(map[string]int),
		nextID: 1,
	}
}

// Assign assigns sequential integer IDs to a set of UUIDs, returning the
// assigned IDs in the same order as the input. Already-mapped UUIDs retain
// their existing integer ID.
func (m *IDMapper) Assign(uuids []string) []int {
	m.mu.Lock()
	defer m.mu.Unlock()

	ids := make([]int, len(uuids))
	for i, uuid := range uuids {
		if existing, ok := m.toInt[uuid]; ok {
			ids[i] = existing
			continue
		}
		id := m.nextID
		m.nextID++
		m.toUUID[id] = uuid
		m.toInt[uuid] = id
		ids[i] = id
	}
	return ids
}

// ToUUID resolves an integer ID to its UUID. Returns an error if the ID
// is not in the session mapping.
func (m *IDMapper) ToUUID(id int) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	uuid, ok := m.toUUID[id]
	if !ok {
		return "", fmt.Errorf("ID %d not found — run a search or list first", id)
	}
	return uuid, nil
}

// ToInt resolves a UUID to its integer ID. Returns 0 if not found.
func (m *IDMapper) ToInt(uuid string) int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.toInt[uuid]
}

// Reset clears the mapping and resets the next ID counter to 1.
func (m *IDMapper) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.toUUID = make(map[int]string)
	m.toInt = make(map[string]int)
	m.nextID = 1
}

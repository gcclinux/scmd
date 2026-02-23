package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

// Session represents a user session
type Session struct {
	Email     string
	CreatedAt time.Time
	ExpiresAt time.Time
}

// SessionStore manages active sessions
type SessionStore struct {
	sessions map[string]*Session
	mu       sync.RWMutex
}

var sessionStore = &SessionStore{
	sessions: make(map[string]*Session),
}

// generateSessionID creates a random session ID
func generateSessionID() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// CreateSession creates a new session for the user
func (s *SessionStore) CreateSession(email string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	sessionID, err := generateSessionID()
	if err != nil {
		return "", err
	}

	session := &Session{
		Email:     email,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(24 * time.Hour), // 24 hour session
	}

	s.sessions[sessionID] = session
	return sessionID, nil
}

// GetSession retrieves a session by ID
func (s *SessionStore) GetSession(sessionID string) (*Session, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	session, exists := s.sessions[sessionID]
	if !exists {
		return nil, false
	}

	// Check if session has expired
	if time.Now().After(session.ExpiresAt) {
		return nil, false
	}

	return session, true
}

// DeleteSession removes a session
func (s *SessionStore) DeleteSession(sessionID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sessions, sessionID)
}

// CleanupExpiredSessions removes expired sessions
func (s *SessionStore) CleanupExpiredSessions() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for id, session := range s.sessions {
		if now.After(session.ExpiresAt) {
			delete(s.sessions, id)
		}
	}
}

// AuthenticateUser validates email and API key against the database
func AuthenticateUser(email, apiKey string) (bool, error) {
	email = strings.TrimSpace(email)
	apiKey = strings.TrimSpace(apiKey)

	if email == "" || apiKey == "" {
		return false, fmt.Errorf("email and API key are required")
	}

	if db == nil {
		return false, fmt.Errorf("database not connected")
	}

	accessTbl := os.Getenv("ACCESS_TB")
	if accessTbl == "" {
		accessTbl = "access"
	}

	var count int
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE email = $1 AND api_key = $2", accessTbl)
	err := db.QueryRow(query, email, apiKey).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("authentication query failed: %v", err)
	}

	return count > 0, nil
}

// RequireAuth is middleware that checks if user is authenticated
func RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_id")
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		_, exists := sessionStore.GetSession(cookie.Value)
		if !exists {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		next(w, r)
	}
}

// StartSessionCleanup starts a goroutine to periodically clean up expired sessions
func StartSessionCleanup() {
	ticker := time.NewTicker(1 * time.Hour)
	go func() {
		for range ticker.C {
			sessionStore.CleanupExpiredSessions()
			log.Println("Cleaned up expired sessions")
		}
	}()
}

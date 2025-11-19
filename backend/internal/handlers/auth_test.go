package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

// setupTestDB creates a temporary test database
func setupTestDB(t *testing.T) *sql.DB {
	// Create temp directory
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Open database
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Create users table
	schema := `
		CREATE TABLE IF NOT EXISTS users (
			user_id BLOB PRIMARY KEY,
			email TEXT UNIQUE NOT NULL,
			password_hash BLOB NOT NULL,
			first_name TEXT NOT NULL,
			last_name TEXT NOT NULL,
			dob TEXT NOT NULL,
			avatar TEXT,
			nickname TEXT,
			about_me TEXT,
			is_public BOOLEAN DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
	`

	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("Failed to create schema: %v", err)
	}

	return db
}

func TestRegister(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	handler := NewAuthHandler(db)

	tests := []struct {
		name           string
		requestBody    map[string]interface{}
		expectedStatus int
		checkEmail     string
	}{
		{
			name: "successful registration with all fields",
			requestBody: map[string]interface{}{
				"email":      "test@example.com",
				"password":   "password123",
				"first_name": "John",
				"last_name":  "Doe",
				"dob":        "1990-01-01",
				"nickname":   "johndoe",
				"about_me":   "Test user",
			},
			expectedStatus: http.StatusCreated,
			checkEmail:     "test@example.com",
		},
		{
			name: "successful registration with required fields only",
			requestBody: map[string]interface{}{
				"email":      "minimal@example.com",
				"password":   "pass123",
				"first_name": "Jane",
				"last_name":  "Smith",
				"dob":        "1995-05-15",
			},
			expectedStatus: http.StatusCreated,
			checkEmail:     "minimal@example.com",
		},
		{
			name: "missing required field",
			requestBody: map[string]interface{}{
				"email":    "incomplete@example.com",
				"password": "password123",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "duplicate email",
			requestBody: map[string]interface{}{
				"email":      "test@example.com",
				"password":   "newpass",
				"first_name": "Another",
				"last_name":  "User",
				"dob":        "1992-02-02",
			},
			expectedStatus: http.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBuffer(body))
			w := httptest.NewRecorder()

			handler.Register(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d. Body: %s", tt.expectedStatus, w.Code, w.Body.String())
			}

			// If successful, verify user was created
			if tt.expectedStatus == http.StatusCreated {
				var response map[string]interface{}
				if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}

				if response["email"] != tt.checkEmail {
					t.Errorf("Expected email %s, got %s", tt.checkEmail, response["email"])
				}

				if response["user_id"] == nil || response["user_id"] == "" {
					t.Error("Expected user_id to be set")
				}
			}
		})
	}
}

func TestLogin(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	handler := NewAuthHandler(db)

	// Create a test user first
	registerBody := map[string]interface{}{
		"email":      "login@example.com",
		"password":   "correctpassword",
		"first_name": "Test",
		"last_name":  "User",
		"dob":        "1990-01-01",
	}
	body, _ := json.Marshal(registerBody)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBuffer(body))
	w := httptest.NewRecorder()
	handler.Register(w, req)

	tests := []struct {
		name           string
		email          string
		password       string
		expectedStatus int
	}{
		{
			name:           "successful login",
			email:          "login@example.com",
			password:       "correctpassword",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "wrong password",
			email:          "login@example.com",
			password:       "wrongpassword",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "non-existent user",
			email:          "nonexistent@example.com",
			password:       "password",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "missing email",
			email:          "",
			password:       "password",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loginBody := map[string]string{
				"email":    tt.email,
				"password": tt.password,
			}
			body, _ := json.Marshal(loginBody)
			req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBuffer(body))
			w := httptest.NewRecorder()

			handler.Login(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d. Body: %s", tt.expectedStatus, w.Code, w.Body.String())
			}

			// If successful, verify response contains user data
			if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}

				if response["email"] != tt.email {
					t.Errorf("Expected email %s, got %s", tt.email, response["email"])
				}
			}
		})
	}
}

func TestGetSession(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	handler := NewAuthHandler(db)

	// Create a test user first
	registerBody := map[string]interface{}{
		"email":      "session@example.com",
		"password":   "password123",
		"first_name": "Session",
		"last_name":  "User",
		"dob":        "1990-01-01",
	}
	body, _ := json.Marshal(registerBody)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBuffer(body))
	w := httptest.NewRecorder()
	handler.Register(w, req)

	// Get user_id from response
	var registerResponse map[string]interface{}
	json.NewDecoder(w.Body).Decode(&registerResponse)
	userID := registerResponse["user_id"].(string)

	tests := []struct {
		name           string
		userID         string
		expectedStatus int
	}{
		{
			name:           "successful get session",
			userID:         userID,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid user_id format",
			userID:         "invalid",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "missing user_id",
			userID:         "",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/api/auth/session"
			if tt.userID != "" {
				url += "?user_id=" + tt.userID
			}

			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()

			handler.GetSession(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d. Body: %s", tt.expectedStatus, w.Code, w.Body.String())
			}

			// If successful, verify response contains user data
			if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}

				if response["email"] != "session@example.com" {
					t.Errorf("Expected email session@example.com, got %s", response["email"])
				}
			}
		})
	}
}

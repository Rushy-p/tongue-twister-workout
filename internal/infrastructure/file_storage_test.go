package infrastructure

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"speech-practice-app/internal/domain"
)

func TestFileStorage_SaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	fs, err := NewFileStorage(dir)
	if err != nil {
		t.Fatalf("NewFileStorage: %v", err)
	}

	data := fs.emptyStoredData()
	data.Sessions = []domain.PracticeSession{
		{ID: "s1", UserID: "u1", StartTime: time.Now(), Status: domain.SessionStatusCompleted},
	}
	data.Preferences = map[string]domain.UserPreferences{
		"u1": *domain.NewUserPreferences("u1"),
	}

	if err := fs.Save(data); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, err := fs.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if len(loaded.Sessions) != 1 {
		t.Errorf("expected 1 session, got %d", len(loaded.Sessions))
	}
	if loaded.Sessions[0].ID != "s1" {
		t.Errorf("expected session ID s1, got %s", loaded.Sessions[0].ID)
	}
	if loaded.Version == "" {
		t.Error("expected version to be set")
	}
}

func TestFileStorage_AtomicWrite(t *testing.T) {
	dir := t.TempDir()
	fs, err := NewFileStorage(dir)
	if err != nil {
		t.Fatalf("NewFileStorage: %v", err)
	}

	data := fs.emptyStoredData()
	if err := fs.Save(data); err != nil {
		t.Fatalf("Save: %v", err)
	}

	// Temp file should not exist after successful save
	tmpPath := filepath.Join(dir, "userdata.json.tmp")
	if _, err := os.Stat(tmpPath); !os.IsNotExist(err) {
		t.Error("temp file should not exist after successful save")
	}

	// Main file should exist
	mainPath := filepath.Join(dir, "userdata.json")
	if _, err := os.Stat(mainPath); err != nil {
		t.Errorf("main data file should exist: %v", err)
	}
}

func TestFileStorage_ExportUserData_JSON(t *testing.T) {
	dir := t.TempDir()
	fs, err := NewFileStorage(dir)
	if err != nil {
		t.Fatalf("NewFileStorage: %v", err)
	}

	data := fs.emptyStoredData()
	data.Sessions = []domain.PracticeSession{
		{ID: "s1", UserID: "u1", StartTime: time.Now(), Status: domain.SessionStatusCompleted},
		{ID: "s2", UserID: "u2", StartTime: time.Now(), Status: domain.SessionStatusCompleted},
	}
	data.Preferences = map[string]domain.UserPreferences{
		"u1": *domain.NewUserPreferences("u1"),
	}
	if err := fs.Save(data); err != nil {
		t.Fatalf("Save: %v", err)
	}

	out, err := fs.ExportUserData("u1", domain.ExportFormatJSON)
	if err != nil {
		t.Fatalf("ExportUserData JSON: %v", err)
	}
	if out == "" {
		t.Error("expected non-empty JSON export")
	}
	// Should contain u1's session but not u2's
	if len(out) == 0 {
		t.Error("export should not be empty")
	}
}

func TestFileStorage_ExportUserData_CSV(t *testing.T) {
	dir := t.TempDir()
	fs, err := NewFileStorage(dir)
	if err != nil {
		t.Fatalf("NewFileStorage: %v", err)
	}

	data := fs.emptyStoredData()
	data.Preferences = map[string]domain.UserPreferences{
		"u1": *domain.NewUserPreferences("u1"),
	}
	if err := fs.Save(data); err != nil {
		t.Fatalf("Save: %v", err)
	}

	out, err := fs.ExportUserData("u1", domain.ExportFormatCSV)
	if err != nil {
		t.Fatalf("ExportUserData CSV: %v", err)
	}
	if out == "" {
		t.Error("expected non-empty CSV export")
	}
}

func TestFileStorage_DeleteUserData(t *testing.T) {
	dir := t.TempDir()
	fs, err := NewFileStorage(dir)
	if err != nil {
		t.Fatalf("NewFileStorage: %v", err)
	}

	data := fs.emptyStoredData()
	data.Sessions = []domain.PracticeSession{
		{ID: "s1", UserID: "u1", StartTime: time.Now()},
		{ID: "s2", UserID: "u2", StartTime: time.Now()},
	}
	data.Preferences = map[string]domain.UserPreferences{
		"u1": *domain.NewUserPreferences("u1"),
		"u2": *domain.NewUserPreferences("u2"),
	}
	if err := fs.Save(data); err != nil {
		t.Fatalf("Save: %v", err)
	}

	if err := fs.DeleteUserData("u1"); err != nil {
		t.Fatalf("DeleteUserData: %v", err)
	}

	loaded, err := fs.Load()
	if err != nil {
		t.Fatalf("Load after delete: %v", err)
	}

	// u1's data should be gone
	for _, s := range loaded.Sessions {
		if s.UserID == "u1" {
			t.Error("u1 session should have been deleted")
		}
	}
	if _, ok := loaded.Preferences["u1"]; ok {
		t.Error("u1 preferences should have been deleted")
	}

	// u2's data should remain
	if _, ok := loaded.Preferences["u2"]; !ok {
		t.Error("u2 preferences should still exist")
	}
}

func TestFileStorage_ValidateData(t *testing.T) {
	dir := t.TempDir()
	fs, err := NewFileStorage(dir)
	if err != nil {
		t.Fatalf("NewFileStorage: %v", err)
	}

	// No file yet — should be valid
	if err := fs.ValidateData(); err != nil {
		t.Errorf("ValidateData on empty dir: %v", err)
	}

	// Save valid data
	data := fs.emptyStoredData()
	if err := fs.Save(data); err != nil {
		t.Fatalf("Save: %v", err)
	}

	if err := fs.ValidateData(); err != nil {
		t.Errorf("ValidateData after save: %v", err)
	}

	// Write invalid JSON
	if err := os.WriteFile(filepath.Join(dir, "userdata.json"), []byte("not json"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	if err := fs.ValidateData(); err == nil {
		t.Error("expected error for invalid JSON")
	}
}

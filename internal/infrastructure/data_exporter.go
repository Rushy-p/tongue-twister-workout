package infrastructure

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"speech-practice-app/internal/domain"
)

// exportAllData holds all user data for export
type exportAllData struct {
	Preferences *domain.UserPreferences  `json:"preferences"`
	Sessions    []domain.PracticeSession `json:"sessions"`
	Progress    []domain.ProgressRecord  `json:"progress"`
	Achievements []domain.Achievement   `json:"achievements"`
	ExportedAt  time.Time               `json:"exported_at"`
}

// DataExporter exports all user data across repositories
type DataExporter struct {
	prefsRepo    PreferencesRepository
	sessionRepo  SessionRepository
	progressRepo ProgressRepository
}

// NewDataExporter creates a new DataExporter
func NewDataExporter(prefsRepo PreferencesRepository, sessionRepo SessionRepository, progressRepo ProgressRepository) *DataExporter {
	return &DataExporter{
		prefsRepo:    prefsRepo,
		sessionRepo:  sessionRepo,
		progressRepo: progressRepo,
	}
}

// ExportAll exports all user data in the specified format
func (e *DataExporter) ExportAll(userID string, format domain.ExportFormat) (string, error) {
	if userID == "" {
		return "", errors.New("user ID cannot be empty")
	}

	prefs, err := e.prefsRepo.Get(userID)
	if err != nil {
		prefs = domain.NewUserPreferences(userID)
	}

	sessions, _ := e.sessionRepo.GetByUserID(userID)
	progress, _ := e.progressRepo.GetAllProgress(userID)
	achievements, _ := e.progressRepo.GetAchievements(userID)

	data := exportAllData{
		Preferences:  prefs,
		Sessions:     sessions,
		Progress:     progress,
		Achievements: achievements,
		ExportedAt:   time.Now(),
	}

	switch format {
	case domain.ExportFormatJSON:
		out, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			return "", fmt.Errorf("failed to marshal export data: %w", err)
		}
		return string(out), nil

	case domain.ExportFormatCSV:
		return exportAllToCSV(data), nil

	default:
		return "", errors.New("unsupported export format")
	}
}

func exportAllToCSV(data exportAllData) string {
	csv := "# Preferences\n"
	csv += exportPreferencesToCSV(*data.Preferences)

	csv += "\n# Sessions\n"
	csv += "session_id,user_id,start_time,status,exercise_count,total_duration\n"
	for _, s := range data.Sessions {
		csv += fmt.Sprintf("%s,%s,%s,%s,%d,%d\n",
			s.ID, s.UserID, s.StartTime.Format(time.RFC3339),
			s.Status, len(s.Exercises), s.TotalDuration)
	}

	csv += "\n# Progress\n"
	csv += "id,user_id,date,category,exercise_count,duration,completed\n"
	for _, p := range data.Progress {
		csv += fmt.Sprintf("%s,%s,%s,%s,%d,%d,%t\n",
			p.ID, p.UserID, p.Date.Format("2006-01-02"),
			p.Category, p.ExerciseCount, p.Duration, p.Completed)
	}

	csv += "\n# Achievements\n"
	csv += "id,user_id,name,type,progress,target,unlocked\n"
	for _, a := range data.Achievements {
		csv += fmt.Sprintf("%s,%s,%s,%s,%d,%d,%t\n",
			a.ID, a.UserID, a.Name, a.AchievementType,
			a.Progress, a.Target, a.IsUnlocked())
	}

	return csv
}

package models

import "time"

type ProjectStatus string

const (
	StatusActive    ProjectStatus = "active"
	StatusPaused    ProjectStatus = "paused"
	StatusArchived  ProjectStatus = "archived"
	StatusCompleted ProjectStatus = "completed"
)

type Project struct {
	ID            string        `json:"id"`
	Name          string        `json:"name"`
	Path          string        `json:"path"`
	Description   string        `json:"description"`
	Status        ProjectStatus `json:"status"`
	Progress      int           `json:"progress"`
	CreatedAt     time.Time     `json:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at"`
	LastScannedAt *time.Time    `json:"last_scanned_at,omitempty"`
	IsGitRepo     bool          `json:"is_git_repo"`
	GitRemote     string        `json:"git_remote,omitempty"`
	GitBranch     string        `json:"git_branch,omitempty"`
	Notes         string        `json:"notes"`
}

type Technology struct {
	ID           int    `json:"id"`
	ProjectID    string `json:"project_id"`
	Type         string `json:"type"`
	Name         string `json:"name"`
	Version      string `json:"version,omitempty"`
	DetectedFrom string `json:"detected_from"`
}

type ProjectFile struct {
	ID           int       `json:"id"`
	ProjectID    string    `json:"project_id"`
	FilePath     string    `json:"file_path"`
	FileType     string    `json:"file_type"`
	SizeBytes    int64     `json:"size_bytes"`
	LastModified time.Time `json:"last_modified"`
}

type Todo struct {
	ID         int       `json:"id"`
	ProjectID  string    `json:"project_id"`
	Content    string    `json:"content"`
	SourceFile string    `json:"source_file,omitempty"`
	LineNumber int       `json:"line_number,omitempty"`
	Priority   string    `json:"priority"`
	Completed  bool      `json:"completed"`
	CreatedAt  time.Time `json:"created_at"`
}

type AIAnalysis struct {
	ID           int       `json:"id"`
	ProjectID    string    `json:"project_id"`
	AnalysisType string    `json:"analysis_type"`
	Result       string    `json:"result"`
	Model        string    `json:"model"`
	TokensUsed   int       `json:"tokens_used"`
	AnalyzedAt   time.Time `json:"analyzed_at"`
}

type ActivityLog struct {
	ID        int       `json:"id"`
	ProjectID string    `json:"project_id"`
	Action    string    `json:"action"`
	Details   string    `json:"details,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

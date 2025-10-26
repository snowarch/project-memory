package repository

import (
	"database/sql"

	"github.com/snowarch/project-memory/internal/models"
)

type AnalysisRepository struct {
	db *sql.DB
}

func NewAnalysisRepository(db *sql.DB) *AnalysisRepository {
	return &AnalysisRepository{db: db}
}

func (r *AnalysisRepository) Create(analysis *models.AIAnalysis) error {
	query := `
		INSERT INTO ai_analyses (project_id, analysis_type, result, model, tokens_used, analyzed_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.Exec(query,
		analysis.ProjectID,
		analysis.AnalysisType,
		analysis.Result,
		analysis.Model,
		analysis.TokensUsed,
		analysis.AnalyzedAt.Unix(),
	)

	return err
}

func (r *AnalysisRepository) GetLatestByProject(projectID string, analysisType string) (*models.AIAnalysis, error) {
	query := `
		SELECT id, project_id, analysis_type, result, model, tokens_used, analyzed_at
		FROM ai_analyses 
		WHERE project_id = ? AND analysis_type = ?
		ORDER BY analyzed_at DESC
		LIMIT 1
	`

	var analysis models.AIAnalysis
	err := r.db.QueryRow(query, projectID, analysisType).Scan(
		&analysis.ID,
		&analysis.ProjectID,
		&analysis.AnalysisType,
		&analysis.Result,
		&analysis.Model,
		&analysis.TokensUsed,
		&analysis.AnalyzedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &analysis, nil
}

func (r *AnalysisRepository) GetByProject(projectID string, limit int) ([]models.AIAnalysis, error) {
	query := `
		SELECT id, project_id, analysis_type, result, model, tokens_used, analyzed_at
		FROM ai_analyses 
		WHERE project_id = ?
		ORDER BY analyzed_at DESC
		LIMIT ?
	`

	rows, err := r.db.Query(query, projectID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var analyses []models.AIAnalysis
	for rows.Next() {
		var analysis models.AIAnalysis
		err := rows.Scan(
			&analysis.ID,
			&analysis.ProjectID,
			&analysis.AnalysisType,
			&analysis.Result,
			&analysis.Model,
			&analysis.TokensUsed,
			&analysis.AnalyzedAt,
		)
		if err != nil {
			return nil, err
		}
		analyses = append(analyses, analysis)
	}

	return analyses, nil
}

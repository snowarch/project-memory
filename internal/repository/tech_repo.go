package repository

import (
	"database/sql"

	"github.com/snowarch/project-memory/internal/models"
)

type TechnologyRepository struct {
	db *sql.DB
}

func NewTechnologyRepository(db *sql.DB) *TechnologyRepository {
	return &TechnologyRepository{db: db}
}

func (r *TechnologyRepository) Create(tech *models.Technology) error {
	query := `
		INSERT INTO technologies (project_id, type, name, version, detected_from)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(project_id, type, name) DO UPDATE SET version = ?, detected_from = ?
	`

	_, err := r.db.Exec(query,
		tech.ProjectID,
		tech.Type,
		tech.Name,
		tech.Version,
		tech.DetectedFrom,
		tech.Version,
		tech.DetectedFrom,
	)

	return err
}

func (r *TechnologyRepository) GetByProject(projectID string) ([]models.Technology, error) {
	query := `
		SELECT id, project_id, type, name, version, detected_from
		FROM technologies WHERE project_id = ?
		ORDER BY type, name
	`

	rows, err := r.db.Query(query, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var techs []models.Technology
	for rows.Next() {
		var tech models.Technology
		err := rows.Scan(
			&tech.ID,
			&tech.ProjectID,
			&tech.Type,
			&tech.Name,
			&tech.Version,
			&tech.DetectedFrom,
		)
		if err != nil {
			return nil, err
		}
		techs = append(techs, tech)
	}

	return techs, nil
}

func (r *TechnologyRepository) DeleteByProject(projectID string) error {
	query := `DELETE FROM technologies WHERE project_id = ?`
	_, err := r.db.Exec(query, projectID)
	return err
}

func (r *TechnologyRepository) GetAllTechnologies() (map[string]int, error) {
	query := `
		SELECT name, COUNT(*) as count 
		FROM technologies 
		GROUP BY name 
		ORDER BY count DESC
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]int)
	for rows.Next() {
		var name string
		var count int
		err := rows.Scan(&name, &count)
		if err != nil {
			return nil, err
		}
		result[name] = count
	}

	return result, nil
}

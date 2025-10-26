package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/snowarch/project-memory/internal/models"
)

type ProjectRepository struct {
	db *sql.DB
}

func NewProjectRepository(db *sql.DB) *ProjectRepository {
	return &ProjectRepository{db: db}
}

func (r *ProjectRepository) Create(project *models.Project) error {
	query := `
		INSERT INTO projects (id, name, path, description, status, progress, created_at, updated_at, last_scanned_at, is_git_repo, git_remote, git_branch, notes)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	
	var lastScanned *int64
	if project.LastScannedAt != nil {
		ts := project.LastScannedAt.Unix()
		lastScanned = &ts
	}

	_, err := r.db.Exec(query,
		project.ID,
		project.Name,
		project.Path,
		project.Description,
		project.Status,
		project.Progress,
		project.CreatedAt.Unix(),
		project.UpdatedAt.Unix(),
		lastScanned,
		project.IsGitRepo,
		project.GitRemote,
		project.GitBranch,
		project.Notes,
	)

	return err
}

func (r *ProjectRepository) Update(project *models.Project) error {
	query := `
		UPDATE projects 
		SET name = ?, description = ?, status = ?, progress = ?, updated_at = ?, last_scanned_at = ?, git_remote = ?, git_branch = ?, notes = ?
		WHERE id = ?
	`
	
	var lastScanned *int64
	if project.LastScannedAt != nil {
		ts := project.LastScannedAt.Unix()
		lastScanned = &ts
	}

	_, err := r.db.Exec(query,
		project.Name,
		project.Description,
		project.Status,
		project.Progress,
		project.UpdatedAt.Unix(),
		lastScanned,
		project.GitRemote,
		project.GitBranch,
		project.Notes,
		project.ID,
	)

	return err
}

func (r *ProjectRepository) GetByID(id string) (*models.Project, error) {
	query := `
		SELECT id, name, path, description, status, progress, created_at, updated_at, last_scanned_at, is_git_repo, git_remote, git_branch, notes
		FROM projects WHERE id = ?
	`

	var project models.Project
	var lastScanned sql.NullInt64
	var createdAt, updatedAt int64

	err := r.db.QueryRow(query, id).Scan(
		&project.ID,
		&project.Name,
		&project.Path,
		&project.Description,
		&project.Status,
		&project.Progress,
		&createdAt,
		&updatedAt,
		&lastScanned,
		&project.IsGitRepo,
		&project.GitRemote,
		&project.GitBranch,
		&project.Notes,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("project not found")
		}
		return nil, err
	}

	project.CreatedAt = time.Unix(createdAt, 0)
	project.UpdatedAt = time.Unix(updatedAt, 0)

	if lastScanned.Valid {
		ts := time.Unix(lastScanned.Int64, 0)
		project.LastScannedAt = &ts
	}

	return &project, nil
}

func (r *ProjectRepository) GetByPath(path string) (*models.Project, error) {
	query := `
		SELECT id, name, path, description, status, progress, created_at, updated_at, last_scanned_at, is_git_repo, git_remote, git_branch, notes
		FROM projects WHERE path = ?
	`

	var project models.Project
	var lastScanned sql.NullInt64
	var createdAt, updatedAt int64

	err := r.db.QueryRow(query, path).Scan(
		&project.ID,
		&project.Name,
		&project.Path,
		&project.Description,
		&project.Status,
		&project.Progress,
		&createdAt,
		&updatedAt,
		&lastScanned,
		&project.IsGitRepo,
		&project.GitRemote,
		&project.GitBranch,
		&project.Notes,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	project.CreatedAt = time.Unix(createdAt, 0)
	project.UpdatedAt = time.Unix(updatedAt, 0)

	if lastScanned.Valid {
		ts := time.Unix(lastScanned.Int64, 0)
		project.LastScannedAt = &ts
	}

	return &project, nil
}

func (r *ProjectRepository) List(status string, limit, offset int) ([]models.Project, error) {
	var query string
	var args []interface{}

	if status != "" {
		query = `
			SELECT id, name, path, description, status, progress, created_at, updated_at, last_scanned_at, is_git_repo, git_remote, git_branch, notes
			FROM projects WHERE status = ? ORDER BY updated_at DESC LIMIT ? OFFSET ?
		`
		args = []interface{}{status, limit, offset}
	} else {
		query = `
			SELECT id, name, path, description, status, progress, created_at, updated_at, last_scanned_at, is_git_repo, git_remote, git_branch, notes
			FROM projects ORDER BY updated_at DESC LIMIT ? OFFSET ?
		`
		args = []interface{}{limit, offset}
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []models.Project
	for rows.Next() {
		var project models.Project
		var lastScanned sql.NullInt64
		var createdAt, updatedAt int64

		err := rows.Scan(
			&project.ID,
			&project.Name,
			&project.Path,
			&project.Description,
			&project.Status,
			&project.Progress,
			&createdAt,
			&updatedAt,
			&lastScanned,
			&project.IsGitRepo,
			&project.GitRemote,
			&project.GitBranch,
			&project.Notes,
		)
		if err != nil {
			return nil, err
		}

		project.CreatedAt = time.Unix(createdAt, 0)
		project.UpdatedAt = time.Unix(updatedAt, 0)

		if lastScanned.Valid {
			ts := time.Unix(lastScanned.Int64, 0)
			project.LastScannedAt = &ts
		}

		projects = append(projects, project)
	}

	return projects, nil
}

func (r *ProjectRepository) Delete(id string) error {
	query := `DELETE FROM projects WHERE id = ?`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *ProjectRepository) Count(status string) (int, error) {
	var query string
	var args []interface{}

	if status != "" {
		query = `SELECT COUNT(*) FROM projects WHERE status = ?`
		args = []interface{}{status}
	} else {
		query = `SELECT COUNT(*) FROM projects`
	}

	var count int
	err := r.db.QueryRow(query, args...).Scan(&count)
	return count, err
}

func (r *ProjectRepository) Search(query string) ([]models.Project, error) {
	sqlQuery := `
		SELECT id, name, path, description, status, progress, created_at, updated_at, last_scanned_at, is_git_repo, git_remote, git_branch, notes
		FROM projects 
		WHERE name LIKE ? OR description LIKE ? OR path LIKE ?
		ORDER BY updated_at DESC
	`

	searchTerm := "%" + query + "%"
	rows, err := r.db.Query(sqlQuery, searchTerm, searchTerm, searchTerm)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []models.Project
	for rows.Next() {
		var project models.Project
		var lastScanned sql.NullInt64
		var createdAt, updatedAt int64

		err := rows.Scan(
			&project.ID,
			&project.Name,
			&project.Path,
			&project.Description,
			&project.Status,
			&project.Progress,
			&createdAt,
			&updatedAt,
			&lastScanned,
			&project.IsGitRepo,
			&project.GitRemote,
			&project.GitBranch,
			&project.Notes,
		)
		if err != nil {
			return nil, err
		}

		project.CreatedAt = time.Unix(createdAt, 0)
		project.UpdatedAt = time.Unix(updatedAt, 0)

		if lastScanned.Valid {
			ts := time.Unix(lastScanned.Int64, 0)
			project.LastScannedAt = &ts
		}

		projects = append(projects, project)
	}

	return projects, nil
}

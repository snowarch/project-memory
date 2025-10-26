package repository

import (
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/snowarch/project-memory/internal/models"
)

func setupTestDB(t *testing.T) *sql.DB {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Crear schema básico
	schema := `
	CREATE TABLE IF NOT EXISTS projects (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		path TEXT UNIQUE NOT NULL,
		description TEXT,
		status TEXT DEFAULT 'active',
		progress INTEGER DEFAULT 0,
		created_at INTEGER NOT NULL,
		updated_at INTEGER NOT NULL,
		last_scanned_at INTEGER,
		is_git_repo BOOLEAN DEFAULT 0,
		git_remote TEXT,
		git_branch TEXT,
		notes TEXT
	);
	`
	
	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("Failed to create schema: %v", err)
	}

	return db
}

func TestProjectRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewProjectRepository(db)
	now := time.Now()

	project := &models.Project{
		ID:          "test123",
		Name:        "Test Project",
		Path:        "/home/user/test",
		Description: "A test project",
		Status:      models.StatusActive,
		Progress:    0,
		CreatedAt:   now,
		UpdatedAt:   now,
		IsGitRepo:   false,
	}

	err := repo.Create(project)
	if err != nil {
		t.Fatalf("Create() failed: %v", err)
	}

	// Verificar que se creó
	retrieved, err := repo.GetByID("test123")
	if err != nil {
		t.Fatalf("GetByID() failed: %v", err)
	}

	if retrieved.Name != project.Name {
		t.Errorf("Retrieved name = %s, want %s", retrieved.Name, project.Name)
	}
	
	if retrieved.Path != project.Path {
		t.Errorf("Retrieved path = %s, want %s", retrieved.Path, project.Path)
	}
}

func TestProjectRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewProjectRepository(db)
	now := time.Now()

	project := &models.Project{
		ID:          "test456",
		Name:        "Original Name",
		Path:        "/home/user/original",
		Description: "Original description",
		Status:      models.StatusActive,
		Progress:    0,
		CreatedAt:   now,
		UpdatedAt:   now,
		IsGitRepo:   false,
	}

	if err := repo.Create(project); err != nil {
		t.Fatalf("Create() failed: %v", err)
	}

	// Actualizar
	project.Name = "Updated Name"
	project.Description = "Updated description"
	project.Progress = 50
	project.Status = models.StatusPaused

	if err := repo.Update(project); err != nil {
		t.Fatalf("Update() failed: %v", err)
	}

	// Verificar actualización
	retrieved, err := repo.GetByID("test456")
	if err != nil {
		t.Fatalf("GetByID() failed: %v", err)
	}

	if retrieved.Name != "Updated Name" {
		t.Errorf("Retrieved name = %s, want 'Updated Name'", retrieved.Name)
	}

	if retrieved.Progress != 50 {
		t.Errorf("Retrieved progress = %d, want 50", retrieved.Progress)
	}

	if retrieved.Status != models.StatusPaused {
		t.Errorf("Retrieved status = %s, want %s", retrieved.Status, models.StatusPaused)
	}
}

func TestProjectRepository_GetByPath(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewProjectRepository(db)
	now := time.Now()

	project := &models.Project{
		ID:        "test789",
		Name:      "Test",
		Path:      "/unique/path",
		Status:    models.StatusActive,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := repo.Create(project); err != nil {
		t.Fatalf("Create() failed: %v", err)
	}

	// Buscar por path
	retrieved, err := repo.GetByPath("/unique/path")
	if err != nil {
		t.Fatalf("GetByPath() failed: %v", err)
	}

	if retrieved == nil {
		t.Fatal("GetByPath() returned nil")
	}

	if retrieved.ID != "test789" {
		t.Errorf("Retrieved ID = %s, want 'test789'", retrieved.ID)
	}

	// Buscar path inexistente
	notFound, err := repo.GetByPath("/nonexistent")
	if err != nil {
		t.Fatalf("GetByPath() with non-existent path failed: %v", err)
	}

	if notFound != nil {
		t.Error("GetByPath() should return nil for non-existent path")
	}
}

func TestProjectRepository_List(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewProjectRepository(db)
	now := time.Now()

	// Crear varios proyectos
	projects := []*models.Project{
		{
			ID:        "active1",
			Name:      "Active Project 1",
			Path:      "/path1",
			Status:    models.StatusActive,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:        "active2",
			Name:      "Active Project 2",
			Path:      "/path2",
			Status:    models.StatusActive,
			CreatedAt: now,
			UpdatedAt: now.Add(time.Hour),
		},
		{
			ID:        "paused1",
			Name:      "Paused Project",
			Path:      "/path3",
			Status:    models.StatusPaused,
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	for _, p := range projects {
		if err := repo.Create(p); err != nil {
			t.Fatalf("Create() failed: %v", err)
		}
	}

	// Listar todos
	all, err := repo.List("", 10, 0)
	if err != nil {
		t.Fatalf("List() failed: %v", err)
	}

	if len(all) != 3 {
		t.Errorf("List() returned %d projects, want 3", len(all))
	}

	// Listar solo activos
	active, err := repo.List("active", 10, 0)
	if err != nil {
		t.Fatalf("List() with status filter failed: %v", err)
	}

	if len(active) != 2 {
		t.Errorf("List(active) returned %d projects, want 2", len(active))
	}
}

func TestProjectRepository_Search(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewProjectRepository(db)
	now := time.Now()

	projects := []*models.Project{
		{
			ID:          "search1",
			Name:        "React App",
			Path:        "/projects/react-app",
			Description: "A React application",
			Status:      models.StatusActive,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          "search2",
			Name:        "Vue Project",
			Path:        "/projects/vue-app",
			Description: "A Vue.js project",
			Status:      models.StatusActive,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
	}

	for _, p := range projects {
		if err := repo.Create(p); err != nil {
			t.Fatalf("Create() failed: %v", err)
		}
	}

	// Buscar por nombre
	results, err := repo.Search("React")
	if err != nil {
		t.Fatalf("Search() failed: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Search('React') returned %d results, want 1", len(results))
	}

	if results[0].Name != "React App" {
		t.Errorf("Search result name = %s, want 'React App'", results[0].Name)
	}

	// Buscar por descripción
	vueResults, err := repo.Search("Vue.js")
	if err != nil {
		t.Fatalf("Search() failed: %v", err)
	}

	if len(vueResults) != 1 {
		t.Errorf("Search('Vue.js') returned %d results, want 1", len(vueResults))
	}
}

func TestProjectRepository_Count(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewProjectRepository(db)
	now := time.Now()

	projects := []*models.Project{
		{
			ID:        "count1",
			Name:      "Project 1",
			Path:      "/p1",
			Status:    models.StatusActive,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:        "count2",
			Name:      "Project 2",
			Path:      "/p2",
			Status:    models.StatusActive,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:        "count3",
			Name:      "Project 3",
			Path:      "/p3",
			Status:    models.StatusCompleted,
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	for _, p := range projects {
		if err := repo.Create(p); err != nil {
			t.Fatalf("Create() failed: %v", err)
		}
	}

	// Contar todos
	totalCount, err := repo.Count("")
	if err != nil {
		t.Fatalf("Count() failed: %v", err)
	}

	if totalCount != 3 {
		t.Errorf("Count('') = %d, want 3", totalCount)
	}

	// Contar activos
	activeCount, err := repo.Count("active")
	if err != nil {
		t.Fatalf("Count('active') failed: %v", err)
	}

	if activeCount != 2 {
		t.Errorf("Count('active') = %d, want 2", activeCount)
	}
}

func TestProjectRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewProjectRepository(db)
	now := time.Now()

	project := &models.Project{
		ID:        "delete1",
		Name:      "To Delete",
		Path:      "/delete/me",
		Status:    models.StatusActive,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := repo.Create(project); err != nil {
		t.Fatalf("Create() failed: %v", err)
	}

	// Eliminar
	if err := repo.Delete("delete1"); err != nil {
		t.Fatalf("Delete() failed: %v", err)
	}

	// Verificar que se eliminó
	retrieved, err := repo.GetByID("delete1")
	if err == nil {
		t.Error("GetByID() should return error for deleted project")
	}

	if retrieved != nil {
		t.Error("GetByID() should return nil for deleted project")
	}
}

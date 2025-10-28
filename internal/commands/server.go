package commands

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/spf13/cobra"
	"github.com/snowarch/project-memory/internal/models"
	"github.com/snowarch/project-memory/internal/repository"
	"github.com/snowarch/project-memory/internal/scanner"
	"github.com/snowarch/project-memory/internal/utils"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start REST API server for agent integration",
	RunE: func(cmd *cobra.Command, args []string) error {
		port, _ := cmd.Flags().GetInt("port")
		host, _ := cmd.Flags().GetString("host")
		
		return startAPIServer(host, port)
	},
}

type APIServer struct {
	projectRepo *repository.ProjectRepository
	techRepo    *repository.TechnologyRepository
	router      *mux.Router
}

type ProjectResponse struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Path        string                 `json:"path"`
	Status      string                 `json:"status"`
	Progress    int                    `json:"progress"`
	Description string                 `json:"description"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	IsGitRepo   bool                   `json:"is_git_repo"`
	GitRemote   string                 `json:"git_remote,omitempty"`
	GitBranch   string                 `json:"git_branch,omitempty"`
	Notes       string                 `json:"notes,omitempty"`
	Technologies []models.Technology   `json:"technologies,omitempty"`
	Context     *utils.AgentContext    `json:"context,omitempty"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Code    int    `json:"code"`
}

func startAPIServer(host string, port int) error {
	server := &APIServer{
		projectRepo: repository.NewProjectRepository(db.Conn()),
		techRepo:    repository.NewTechnologyRepository(db.Conn()),
		router:      mux.NewRouter(),
	}
	
	server.setupRoutes()
	
	addr := fmt.Sprintf("%s:%d", host, port)
	fmt.Printf("Starting Project Memory API Server on %s\n", addr)
	fmt.Printf("Available endpoints:\n")
	fmt.Printf("  GET    /api/v1/projects              - List all projects\n")
	fmt.Printf("  GET    /api/v1/projects/{id}         - Get project details\n")
	fmt.Printf("  GET    /api/v1/projects/{id}/context - Get project context\n")
	fmt.Printf("  POST   /api/v1/projects/{id}/open    - Open project in IDE\n")
	fmt.Printf("  GET    /api/v1/health                - Health check\n")
	fmt.Printf("  GET    /api/v1/agents/info           - Agent integration info\n")
	
	return http.ListenAndServe(addr, server.router)
}

func (s *APIServer) setupRoutes() {
	s.router.HandleFunc("/api/v1/health", s.healthHandler).Methods("GET")
	s.router.HandleFunc("/api/v1/agents/info", s.agentInfoHandler).Methods("GET")
	
	// Project endpoints
	s.router.HandleFunc("/api/v1/projects", s.listProjectsHandler).Methods("GET")
	s.router.HandleFunc("/api/v1/projects/{id}", s.getProjectHandler).Methods("GET")
	s.router.HandleFunc("/api/v1/projects/{id}/context", s.getProjectContextHandler).Methods("GET")
	s.router.HandleFunc("/api/v1/projects/{id}/open", s.openProjectHandler).Methods("POST")
	s.router.HandleFunc("/api/v1/projects/{id}/handoff", s.generateHandoffHandler).Methods("POST")
	
	// Search endpoints
	s.router.HandleFunc("/api/v1/search", s.searchProjectsHandler).Methods("GET")
	s.router.HandleFunc("/api/v1/search/technologies", s.searchByTechnologyHandler).Methods("GET")
	
	// Agent-specific endpoints
	s.router.HandleFunc("/api/v1/agents/context", s.agentContextHandler).Methods("POST")
	s.router.HandleFunc("/api/v1/agents/discover", s.discoverProjectsHandler).Methods("GET")
}

func (s *APIServer) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
		"version":   "1.0.0",
		"service":   "project-memory-api",
	})
}

func (s *APIServer) agentInfoHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	projects, err := s.projectRepo.List("", 100, 0)
	if err != nil {
		s.sendError(w, "Failed to get projects", http.StatusInternalServerError)
		return
	}
	
	var techCount int
	for _, p := range projects {
		techs, _ := s.techRepo.GetByProject(p.ID)
		techCount += len(techs)
	}
	
	info := map[string]interface{}{
		"agent_capabilities": []string{
			"project_discovery",
			"context_generation",
			"ide_integration", 
			"handoff_generation",
			"technology_analysis",
			"git_integration",
		},
		"supported_formats": []string{"json", "markdown"},
		"total_projects":    len(projects),
		"total_technologies": techCount,
		"api_version":       "v1",
		"documentation":     "https://github.com/snowarch/project-memory",
	}
	
	json.NewEncoder(w).Encode(info)
}

func (s *APIServer) listProjectsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	status := r.URL.Query().Get("status")
	limitStr := r.URL.Query().Get("limit")
	
	limit := 50
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}
	
	projects, err := s.projectRepo.List(status, limit, 0)
	if err != nil {
		s.sendError(w, "Failed to list projects", http.StatusInternalServerError)
		return
	}
	
	var responses []ProjectResponse
	for _, p := range projects {
		techs, _ := s.techRepo.GetByProject(p.ID)
		responses = append(responses, ProjectResponse{
			ID:           p.ID,
			Name:         p.Name,
			Path:         p.Path,
			Status:       string(p.Status),
			Progress:     p.Progress,
			Description:  p.Description,
			CreatedAt:    p.CreatedAt,
			UpdatedAt:    p.UpdatedAt,
			IsGitRepo:    p.IsGitRepo,
			GitRemote:    p.GitRemote,
			GitBranch:    p.GitBranch,
			Notes:        p.Notes,
			Technologies: techs,
		})
	}
	
	json.NewEncoder(w).Encode(responses)
}

func (s *APIServer) getProjectHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	vars := mux.Vars(r)
	projectID := vars["id"]
	
	project, err := s.projectRepo.GetByID(projectID)
	if err != nil {
		s.sendError(w, "Project not found", http.StatusNotFound)
		return
	}
	
	techs, _ := s.techRepo.GetByProject(project.ID)
	
	response := ProjectResponse{
		ID:           project.ID,
		Name:         project.Name,
		Path:         project.Path,
		Status:       string(project.Status),
		Progress:     project.Progress,
		Description:  project.Description,
		CreatedAt:    project.CreatedAt,
		UpdatedAt:    project.UpdatedAt,
		IsGitRepo:    project.IsGitRepo,
		GitRemote:    project.GitRemote,
		GitBranch:    project.GitBranch,
		Notes:        project.Notes,
		Technologies: techs,
	}
	
	json.NewEncoder(w).Encode(response)
}

func (s *APIServer) getProjectContextHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	vars := mux.Vars(r)
	projectID := vars["id"]
	
	project, err := s.projectRepo.GetByID(projectID)
	if err != nil {
		s.sendError(w, "Project not found", http.StatusNotFound)
		return
	}
	
	generator := utils.NewContextGenerator(project.Path)
	context, err := generator.GenerateAgentContext(
		project.Name,
		string(project.Status),
		project.Progress,
		project.Notes,
	)
	if err != nil {
		s.sendError(w, "Failed to generate context", http.StatusInternalServerError)
		return
	}
	
	response := ProjectResponse{
		ID:       project.ID,
		Name:     project.Name,
		Path:     project.Path,
		Status:   string(project.Status),
		Progress: project.Progress,
		Context:  context,
	}
	
	json.NewEncoder(w).Encode(response)
}

func (s *APIServer) openProjectHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	vars := mux.Vars(r)
	projectID := vars["id"]
	
	project, err := s.projectRepo.GetByID(projectID)
	if err != nil {
		s.sendError(w, "Project not found", http.StatusNotFound)
		return
	}
	
	// Get IDE from request body
	var request struct {
		IDE string `json:"ide"`
	}
	
	json.NewDecoder(r.Body).Decode(&request)
	
	detector := utils.NewIDEDetector()
	ideName := request.IDE
	
	if ideName == "" {
		ide := detector.GetPreferredIDE()
		if ide == nil {
			s.sendError(w, "No IDEs detected", http.StatusBadRequest)
			return
		}
		ideName = ide.Name
	}
	
	err = detector.LaunchProjectInBackground(project.Path, ideName)
	if err != nil {
		s.sendError(w, fmt.Sprintf("Failed to open project: %v", err), http.StatusInternalServerError)
		return
	}
	
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":     "opened",
		"project":    project.Name,
		"ide":        ideName,
		"path":       project.Path,
		"timestamp":  time.Now().UTC(),
	})
}

func (s *APIServer) generateHandoffHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	vars := mux.Vars(r)
	projectID := vars["id"]
	
	project, err := s.projectRepo.GetByID(projectID)
	if err != nil {
		s.sendError(w, "Project not found", http.StatusNotFound)
		return
	}
	
	handoffDoc, err := generateHandoffDocContent(project)
	if err != nil {
		s.sendError(w, "Failed to generate handoff", http.StatusInternalServerError)
		return
	}
	
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":     "generated",
		"project":    project.Name,
		"handoff":    handoffDoc,
		"timestamp":  time.Now().UTC(),
	})
}

func (s *APIServer) searchProjectsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	query := r.URL.Query().Get("q")
	if query == "" {
		s.sendError(w, "Query parameter 'q' is required", http.StatusBadRequest)
		return
	}
	
	projects, err := s.projectRepo.Search(query)
	if err != nil {
		s.sendError(w, "Search failed", http.StatusInternalServerError)
		return
	}
	
	var responses []ProjectResponse
	for _, p := range projects {
		techs, _ := s.techRepo.GetByProject(p.ID)
		responses = append(responses, ProjectResponse{
			ID:           p.ID,
			Name:         p.Name,
			Path:         p.Path,
			Status:       string(p.Status),
			Progress:     p.Progress,
			Description:  p.Description,
			Technologies: techs,
		})
	}
	
	json.NewEncoder(w).Encode(responses)
}

func (s *APIServer) searchByTechnologyHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	tech := r.URL.Query().Get("technology")
	if tech == "" {
		s.sendError(w, "Technology parameter is required", http.StatusBadRequest)
		return
	}
	
	projects, err := s.projectRepo.List("", 100, 0)
	if err != nil {
		s.sendError(w, "Failed to get projects", http.StatusInternalServerError)
		return
	}
	
	var responses []ProjectResponse
	for _, p := range projects {
		techs, _ := s.techRepo.GetByProject(p.ID)
		for _, t := range techs {
			if t.Name == tech {
				responses = append(responses, ProjectResponse{
					ID:           p.ID,
					Name:         p.Name,
					Path:         p.Path,
					Status:       string(p.Status),
					Progress:     p.Progress,
					Description:  p.Description,
					Technologies: techs,
				})
				break
			}
		}
	}
	
	json.NewEncoder(w).Encode(responses)
}

func (s *APIServer) agentContextHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	var request struct {
		Projects []string `json:"projects"`
		Format   string   `json:"format"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		s.sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	if request.Format == "" {
		request.Format = "json"
	}
	
	var contexts []map[string]interface{}
	
	for _, projectName := range request.Projects {
		projects, err := s.projectRepo.Search(projectName)
		if err != nil || len(projects) == 0 {
			continue
		}
		
		project := projects[0]
		generator := utils.NewContextGenerator(project.Path)
		context, err := generator.GenerateAgentContext(
			project.Name,
			string(project.Status),
			project.Progress,
			project.Notes,
		)
		if err != nil {
			continue
		}
		
		var contextData interface{}
		switch request.Format {
		case "json":
			contextData, _ = json.Marshal(context)
		case "markdown":
			contextData = context.ExportToMarkdown()
		default:
			contextData = context
		}
		
		contexts = append(contexts, map[string]interface{}{
			"project": project.Name,
			"context": contextData,
		})
	}
	
	json.NewEncoder(w).Encode(map[string]interface{}{
		"contexts":  contexts,
		"format":    request.Format,
		"timestamp": time.Now().UTC(),
	})
}

func (s *APIServer) discoverProjectsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	path := r.URL.Query().Get("path")
	if path == "" {
		// Use current working directory
		if cwd, err := os.Getwd(); err == nil {
			path = cwd
		}
	}
	
	// Scan for projects in the specified path
	scanner := scanner.New(path)
	projects, err := scanner.ScanProjects()
	if err != nil {
		s.sendError(w, "Failed to scan directory", http.StatusInternalServerError)
		return
	}
	
	var discoveries []map[string]interface{}
	for _, p := range projects {
		discoveries = append(discoveries, map[string]interface{}{
			"name":        p.Name,
			"path":        p.Path,
			"status":      "discovered",
			"is_git_repo": p.IsGitRepo,
		})
	}
	
	json.NewEncoder(w).Encode(map[string]interface{}{
		"discovered": discoveries,
		"path":       path,
		"count":      len(discoveries),
		"timestamp":  time.Now().UTC(),
	})
}

func (s *APIServer) sendError(w http.ResponseWriter, message string, code int) {
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(ErrorResponse{
		Error:   "error",
		Message: message,
		Code:    code,
	})
}

func init() {
	serverCmd.Flags().IntP("port", "p", 8080, "Port for the API server")
	serverCmd.Flags().StringP("host", "H", "localhost", "Host for the API server")
	rootCmd.AddCommand(serverCmd)
}

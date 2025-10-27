package scanner

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// ProjectStateAnalyzer provides intelligent project state detection
type ProjectStateAnalyzer struct {
	projectPath string
}

func NewProjectStateAnalyzer(projectPath string) *ProjectStateAnalyzer {
	return &ProjectStateAnalyzer{projectPath: projectPath}
}

// AnalyzeProjectActivity determines project activity level and suggested progress
func (psa *ProjectStateAnalyzer) AnalyzeProjectActivity() (activityLevel string, suggestedProgress int, confidence float64, insights []string) {
	insights = []string{}
	
	// Check recent file modifications
	recentFiles, avgAge := psa.getRecentFileActivity()
	if len(recentFiles) > 5 {
		activityLevel = "active"
		suggestedProgress = minInt(50, len(recentFiles)*10)
		insights = append(insights, "High file modification activity detected")
	} else if len(recentFiles) > 0 {
		activityLevel = "moderate"
		suggestedProgress = minInt(25, len(recentFiles)*5)
		insights = append(insights, "Moderate development activity")
	} else {
		activityLevel = "stale"
		suggestedProgress = 0
		insights = append(insights, "No recent activity detected")
	}
	
	// Check git commit activity
	if gitInsights, gitProgress := psa.analyzeGitActivity(); len(gitInsights) > 0 {
		insights = append(insights, gitInsights...)
		suggestedProgress = minInt(100, suggestedProgress+gitProgress)
	}
	
	// Check for build artifacts and deployment files
	if psa.hasBuildArtifacts() {
		suggestedProgress = minInt(100, suggestedProgress+20)
		insights = append(insights, "Build artifacts detected - project may be deployable")
	}
	
	// Check for test coverage
	if testCoverage := psa.estimateTestCoverage(); testCoverage > 0 {
		suggestedProgress = minInt(100, suggestedProgress+testCoverage/5)
		insights = append(insights, "Test coverage detected")
	}
	
	// Calculate confidence based on data availability
	confidence = psa.calculateConfidence(recentFiles, avgAge)
	
	return activityLevel, suggestedProgress, confidence, insights
}

func (psa *ProjectStateAnalyzer) getRecentFileActivity() ([]string, time.Duration) {
	var recentFiles []string
	var totalAge time.Duration
	count := 0
	
	cutoff := time.Now().AddDate(0, 0, -7) // Last 7 days
	
	err := filepath.Walk(psa.projectPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		
		// Skip hidden files and common directories
		if strings.HasPrefix(info.Name(), ".") || 
		   strings.Contains(path, "node_modules") ||
		   strings.Contains(path, ".git") ||
		   strings.Contains(path, "target") ||
		   strings.Contains(path, "build") {
			return nil
		}
		
		if info.ModTime().After(cutoff) {
			recentFiles = append(recentFiles, path)
			totalAge += time.Since(info.ModTime())
			count++
		}
		
		return nil
	})
	
	if err == nil && count > 0 {
		return recentFiles, totalAge / time.Duration(count)
	}
	
	return recentFiles, 0
}

func (psa *ProjectStateAnalyzer) analyzeGitActivity() ([]string, int) {
	insights := []string{}
	progressBoost := 0
	
	// Check if it's a git repo
	gitDir := filepath.Join(psa.projectPath, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		return insights, 0
	}
	
	// Get recent commits
	cmd := exec.Command("git", "-C", psa.projectPath, "log", "--since=2.weeks", "--oneline")
	output, err := cmd.Output()
	if err == nil {
		commits := strings.Split(strings.TrimSpace(string(output)), "\n")
		if len(commits) > 0 && commits[0] != "" {
			progressBoost = minInt(30, len(commits)*5)
			insights = append(insights, "Recent git activity: "+string(rune(len(commits)))+" commits in last 2 weeks")
		}
	}
	
	// Check for uncommitted changes
	cmd = exec.Command("git", "-C", psa.projectPath, "status", "--porcelain")
	output, err = cmd.Output()
	if err == nil && len(strings.TrimSpace(string(output))) > 0 {
		insights = append(insights, "Uncommitted changes detected")
		progressBoost = minInt(100, progressBoost+10)
	}
	
	return insights, progressBoost
}

func (psa *ProjectStateAnalyzer) hasBuildArtifacts() bool {
	indicators := []string{
		"dist", "build", "target", "bin", "out",
		"Dockerfile", "docker-compose.yml",
		"package-lock.json", "yarn.lock",
		"go.sum", "Cargo.lock",
	}
	
	for _, indicator := range indicators {
		if _, err := os.Stat(filepath.Join(psa.projectPath, indicator)); err == nil {
			return true
		}
	}
	
	return false
}

func (psa *ProjectStateAnalyzer) estimateTestCoverage() int {
	testFiles := 0
	sourceFiles := 0
	
	err := filepath.Walk(psa.projectPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		
		// Skip vendor and dependencies
		if strings.Contains(path, "node_modules") || 
		   strings.Contains(path, "vendor") ||
		   strings.Contains(path, ".git") {
			return nil
		}
		
		ext := strings.ToLower(filepath.Ext(path))
		switch ext {
		case ".go", ".js", ".ts", ".py", ".rs", ".java":
			sourceFiles++
			if strings.Contains(strings.ToLower(info.Name()), "test") {
				testFiles++
			}
		}
		
		return nil
	})
	
	if err == nil && sourceFiles > 0 {
		return (testFiles * 100) / sourceFiles
	}
	
	return 0
}

func (psa *ProjectStateAnalyzer) calculateConfidence(recentFiles []string, avgAge time.Duration) float64 {
	confidence := 0.5 // Base confidence
	
	if len(recentFiles) > 0 {
		confidence += 0.2
	}
	
	if avgAge < 24*time.Hour {
		confidence += 0.2
	} else if avgAge < 7*24*time.Hour {
		confidence += 0.1
	}
	
	// Check for README
	if _, err := os.Stat(filepath.Join(psa.projectPath, "README.md")); err == nil {
		confidence += 0.1
	}
	
	return minFloat64(1.0, confidence)
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func minFloat64(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

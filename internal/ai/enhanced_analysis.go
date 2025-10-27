package ai

import (
	"fmt"
	"strings"
	"time"
)

// Enhanced AI analysis for better project context and state assessment
func (c *GroqClient) AnalyzeProjectEnhanced(projectName, description, technologies, readme, activityInsights string) (string, int, error) {
	systemPrompt := `You are a senior software engineer and project manager analyzing development projects. 
Provide comprehensive insights about project state, progress, blockers, and next steps. 
Focus on actionable recommendations that help developers understand exactly where they left off and what to do next.
Consider recent activity, technical debt, and completion probability.`

	userPrompt := fmt.Sprintf(`Analyze this project comprehensively:

Project: %s
Description: %s
Technologies: %s

README excerpt:
%s

Recent Activity Insights:
%s

Provide detailed analysis covering:
1. Current State Assessment (2-3 sentences)
2. Completion Percentage (0-100) with reasoning
3. Where Developer Left Off (specific last actions/stopping points)
4. Immediate Next Steps (3-5 prioritized actions)
5. Technical Blockers or Risks
6. Estimated Time to Complete (hours/days)
7. Code Quality Indicators (if detectable)
8. Recommendations for Efficiency

Format as clear, structured text with bullet points for action items.`, 
		projectName, description, technologies, readme, activityInsights)

	return c.Analyze(systemPrompt, userPrompt)
}

func (c *GroqClient) GenerateProjectSummary(projectName, status string, progress int, technologies []string, lastActivity time.Time, notes string) (string, int, error) {
	systemPrompt := `You are creating a concise project summary for developer handoff. 
Provide essential context that helps a new developer understand the project quickly and continue work effectively.`

	techList := strings.Join(technologies, ", ")
	activityAgo := time.Since(lastActivity).Round(time.Hour)
	
	userPrompt := fmt.Sprintf(`Create a developer handoff summary:

Project: %s
Status: %s
Progress: %d%%
Technologies: %s
Last Activity: %s ago
Notes: %s

Generate a concise summary including:
1. Project Overview (1 sentence)
2. Current Development Stage
3. Key Technologies in Use
4. Known Issues or Blockers
5. Quick Start Instructions (3-4 steps)

Keep it practical and developer-focused.`, 
		projectName, status, progress, techList, activityAgo, notes)

	return c.Analyze(systemPrompt, userPrompt)
}

func (c *GroqClient) SuggestNextActions(projectName, currentStatus string, progress int, blockers []string, availableTime string) (string, int, error) {
	systemPrompt := `You are a senior developer suggesting next actions for a project. 
Consider the current state, available time, and blockers to provide realistic, actionable recommendations.`

	blockerList := strings.Join(blockers, ", ")
	
	userPrompt := fmt.Sprintf(`Suggest next development actions:

Project: %s
Current Status: %s
Progress: %d%%
Known Blockers: %s
Available Time: %s

Provide:
1. Quick Wins (under 30 minutes)
2. Main Development Tasks (for available time)
3. Blocker Resolution Steps
4. Progress Milestones to Target
5. Efficiency Tips

Be specific and realistic about time estimates.`, 
		projectName, currentStatus, progress, blockerList, availableTime)

	return c.Analyze(systemPrompt, userPrompt)
}

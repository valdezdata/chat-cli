package assessment

import (
	"strings"

	"github.com/fatih/color"
)

// AssessmentResult holds the evaluation data for one criterion
type AssessmentResult struct {
	Name           string
	Score          int
	Rating         string
	Description    string
	Recommendation string
}

// PromptAssessment contains all evaluation results
type PromptAssessment struct {
	Criteria      map[string]AssessmentResult
	TotalScore    int
	OverallRating string
}

// ScoreLabels maps numeric scores to text ratings
var ScoreLabels = map[int]string{
	1: "Poor",
	2: "Needs Improvement",
	3: "Average",
	4: "Good",
	5: "Excellent",
}

// ScoreIcons maps numeric scores to visual indicators
var ScoreIcons = map[int]string{
	1: "❌",  // Red X - Poor
	2: "⚠️", // Warning - Needs Improvement
	3: "⚙️", // Gear - Average
	4: "✓",  // Checkmark - Good
	5: "✅",  // Green Checkmark - Excellent
}

// Criteria definitions and evaluation functions

// evaluateClarity assesses how clear and understandable the prompt is
func evaluateClarity(text string, _ int) AssessmentResult {
	result := AssessmentResult{
		Name:        "Clarity",
		Score:       5,
		Rating:      ScoreLabels[5],
		Description: "The prompt is clear, specific, and easy to understand.",
	}

	if len(text) < 5 {
		result.Score = 1
		result.Rating = ScoreLabels[1]
		result.Description = "The prompt is too short to be clear."
	} else if len(text) < 15 {
		result.Score = 2
		result.Rating = ScoreLabels[2]
		result.Description = "The prompt is vague or incomplete."
	} else if !strings.ContainsAny(text, ".?!") {
		result.Score = 2
		result.Rating = ScoreLabels[2]
		result.Description = "The prompt lacks proper structure or punctuation."
	} else if len(text) < 40 {
		result.Score = 3
		result.Rating = ScoreLabels[3]
		result.Description = "The prompt is moderately clear but could be more detailed."
	} else if !strings.Contains(strings.ToLower(text), "please") && !strings.Contains(strings.ToLower(text), "could you") {
		result.Score = 4
		result.Rating = ScoreLabels[4]
		result.Description = "The prompt is clear but could be more polite."
	}

	return result
}

// evaluateRelevance assesses how relevant the prompt is to a specific task
func evaluateRelevance(text string, _ int) AssessmentResult {
	result := AssessmentResult{
		Name:        "Relevance",
		Score:       5,
		Rating:      ScoreLabels[5],
		Description: "The prompt is highly relevant to a specific task.",
	}

	if len(text) < 3 {
		result.Score = 1
		result.Rating = ScoreLabels[1]
		result.Description = "The prompt is too short to determine relevance."
	} else if len(text) < 10 {
		result.Score = 2
		result.Rating = ScoreLabels[2]
		result.Description = "The prompt is too brief to be sufficiently relevant."
	} else if len(text) < 30 && !strings.ContainsAny(text, ",-+") {
		result.Score = 3
		result.Rating = ScoreLabels[3]
		result.Description = "The prompt has moderate relevance but lacks specific details."
	} else if len(text) < 60 {
		result.Score = 4
		result.Rating = ScoreLabels[4]
		result.Description = "The prompt is relevant but could be more focused."
	}

	return result
}

// evaluateSpecificity assesses how specific and detailed the prompt is
func evaluateSpecificity(text string, _ int) AssessmentResult {
	result := AssessmentResult{
		Name:        "Specificity",
		Score:       5,
		Rating:      ScoreLabels[5],
		Description: "The prompt is very specific and well-defined.",
	}

	taskWords := []string{
		"write", "explain", "generate", "describe", "list", "summarize",
		"create", "analyze", "compare", "evaluate",
	}

	hasTask := containsAny(text, taskWords)

	if !hasTask {
		if len(text) < 15 {
			result.Score = 1
			result.Rating = ScoreLabels[1]
			result.Description = "The prompt lacks any specific task or direction."
			result.Recommendation = "Include a clear action verb (e.g., explain, describe, list)."
		} else {
			result.Score = 2
			result.Rating = ScoreLabels[2]
			result.Description = "The prompt lacks a clear directive."
			result.Recommendation = "Be explicit about what you want (e.g., 'analyze this code')."
		}
	} else if !strings.ContainsAny(text, ".,;:") {
		result.Score = 3
		result.Rating = ScoreLabels[3]
		result.Description = "The prompt has a task but lacks structure."
		result.Recommendation = "Add more details and proper punctuation."
	} else if len(text) < 50 && !strings.Contains(text, "for example") && !strings.Contains(text, "such as") {
		result.Score = 4
		result.Rating = ScoreLabels[4]
		result.Description = "The prompt is specific but could include examples."
		result.Recommendation = "Consider adding examples to clarify intent."
	}

	return result
}

// evaluateContext assesses the background information provided in the prompt
func evaluateContext(text string, wordCount int) AssessmentResult {
	result := AssessmentResult{
		Name:        "Context",
		Score:       5,
		Rating:      ScoreLabels[5],
		Description: "Rich context is provided with background information.",
	}

	if wordCount < 5 {
		result.Score = 1
		result.Rating = ScoreLabels[1]
		result.Description = "No context provided at all."
		result.Recommendation = "Add background information about your request."
	} else if wordCount < 10 && !strings.Contains(text, "because") && !strings.Contains(text, "about") {
		result.Score = 2
		result.Rating = ScoreLabels[2]
		result.Description = "Minimal context provided."
		result.Recommendation = "Add relevant background details about the subject."
	} else if wordCount < 20 && !strings.Contains(text, "since") && !strings.Contains(text, "given") {
		result.Score = 3
		result.Rating = ScoreLabels[3]
		result.Description = "Some context provided but could be more detailed."
		result.Recommendation = "Expand on the background or situation."
	} else if wordCount < 40 {
		result.Score = 4
		result.Rating = ScoreLabels[4]
		result.Description = "Good context provided but could be more comprehensive."
		result.Recommendation = "Consider adding more specific details."
	}

	return result
}

// evaluateRichness assesses the level of detail and descriptiveness in the prompt
func evaluateRichness(text string, wordCount int) AssessmentResult {
	result := AssessmentResult{
		Name:        "Richness",
		Score:       5,
		Rating:      ScoreLabels[5],
		Description: "The prompt is rich with details and encourages creativity.",
	}

	if wordCount < 5 {
		result.Score = 1
		result.Rating = ScoreLabels[1]
		result.Description = "The prompt is too minimal to provide richness."
		result.Recommendation = "Expand your prompt substantially with details."
	} else if wordCount < 15 && !strings.Contains(text, "like") && !strings.Contains(text, "example") {
		result.Score = 2
		result.Rating = ScoreLabels[2]
		result.Description = "The prompt lacks depth and examples."
		result.Recommendation = "Add examples or descriptive details."
	} else if wordCount < 30 && !strings.Contains(text, "detailed") && !strings.Contains(text, "specific") {
		result.Score = 3
		result.Rating = ScoreLabels[3]
		result.Description = "The prompt has moderate richness but could be enhanced."
		result.Recommendation = "Add more descriptive elements or constraints."
	} else if wordCount < 50 {
		result.Score = 4
		result.Rating = ScoreLabels[4]
		result.Description = "The prompt has good richness but could be more elaborate."
		result.Recommendation = "Consider adding more nuanced details."
	}

	return result
}

// evaluatePersona assesses whether the prompt defines a role or persona
func evaluatePersona(text string, _ int) AssessmentResult {
	result := AssessmentResult{
		Name:        "Persona",
		Score:       5,
		Rating:      ScoreLabels[5],
		Description: "A clear, specific role or persona is well-defined.",
	}

	roleWords := []string{
		"as a", "like a", "act as", "you are", "pretend", "role", "expert",
	}

	hasPersona := containsAny(text, roleWords)

	if !hasPersona {
		result.Score = 2
		result.Rating = ScoreLabels[2]
		result.Description = "No specific role guidance provided."
		result.Recommendation = "Specify the role you want the AI to take (e.g., 'You are an expert coder')."
	} else if !strings.Contains(strings.ToLower(text), "expert") && !strings.Contains(strings.ToLower(text), "professional") {
		result.Score = 3
		result.Rating = ScoreLabels[3]
		result.Description = "A basic role is defined but lacks expertise level."
		result.Recommendation = "Specify the expertise level (e.g., 'as an expert scientist')."
	} else if !strings.Contains(strings.ToLower(text), "perspective") {
		result.Score = 4
		result.Rating = ScoreLabels[4]
		result.Description = "A good role is defined but lacks perspective guidance."
		result.Recommendation = "Consider specifying the perspective to take."
	}

	return result
}

// evaluateInstruction assesses the clarity of the task or instruction
func evaluateInstruction(text string, wordCount int) AssessmentResult {
	result := AssessmentResult{
		Name:        "Instruction",
		Score:       5,
		Rating:      ScoreLabels[5],
		Description: "The task is extremely well-defined and specific.",
	}

	taskWords := []string{
		"write", "explain", "generate", "describe", "list", "summarize",
		"create", "analyze", "compare", "evaluate",
	}

	hasTask := containsAny(text, taskWords)

	if !hasTask && wordCount < 10 {
		result.Score = 1
		result.Rating = ScoreLabels[1]
		result.Description = "No clear task or instruction provided."
	} else if !hasTask {
		result.Score = 2
		result.Rating = ScoreLabels[2]
		result.Description = "The task is implied but not explicitly stated."
	} else if !strings.ContainsAny(text, ".,;:") {
		result.Score = 3
		result.Rating = ScoreLabels[3]
		result.Description = "A basic task is provided but lacks structure."
	} else if !strings.Contains(text, "step") && !strings.Contains(text, "first") && !strings.Contains(text, "then") {
		result.Score = 4
		result.Rating = ScoreLabels[4]
		result.Description = "The task is clear but could benefit from sequencing."
	}

	return result
}

// evaluateFormat assesses whether the prompt specifies the desired output format
func evaluateFormat(text string, _ int) AssessmentResult {
	result := AssessmentResult{
		Name:        "Format",
		Score:       5,
		Rating:      ScoreLabels[5],
		Description: "The output format is precisely specified with clear structure.",
	}

	formatWords := []string{
		"list", "table", "paragraph", "json", "bullet", "code block",
		"format", "style", "markdown",
	}

	hasFormat := containsAny(text, formatWords)

	if !hasFormat {
		result.Score = 2
		result.Rating = ScoreLabels[2]
		result.Description = "No output format specified."
		result.Recommendation = "Specify the desired format (e.g., bullet points, table)."
	} else if !strings.Contains(strings.ToLower(text), "detailed") && !strings.Contains(strings.ToLower(text), "brief") {
		result.Score = 3
		result.Rating = ScoreLabels[3]
		result.Description = "A format is mentioned but lacks detail about length or depth."
		result.Recommendation = "Specify whether you want a detailed or brief response."
	} else if !strings.Contains(strings.ToLower(text), "example") {
		result.Score = 4
		result.Rating = ScoreLabels[4]
		result.Description = "The format is well-specified but lacks example structure."
		result.Recommendation = "Consider providing an example of the structure you want."
	}

	return result
}

// evaluateAudience assesses whether the prompt defines a target audience
func evaluateAudience(text string, _ int) AssessmentResult {
	result := AssessmentResult{
		Name:        "Audience",
		Score:       5,
		Rating:      ScoreLabels[5],
		Description: "The target audience is precisely defined with clear adaptation guidance.",
	}

	audienceWords := []string{
		"for ", "to ", "beginners", "experts", "students", "tutorial",
		"audience", "reader", "user",
	}

	hasAudience := containsAny(text, audienceWords)

	if !hasAudience {
		result.Score = 2
		result.Rating = ScoreLabels[2]
		result.Description = "No target audience specified."
		result.Recommendation = "Define who this is for (e.g., 'for a 12-year-old')."
	} else if !strings.Contains(strings.ToLower(text), "level") && !strings.Contains(strings.ToLower(text), "background") {
		result.Score = 3
		result.Rating = ScoreLabels[3]
		result.Description = "An audience is mentioned but their knowledge level is unclear."
		result.Recommendation = "Specify the audience's knowledge level."
	} else if !strings.Contains(strings.ToLower(text), "familiar") && !strings.Contains(strings.ToLower(text), "understand") {
		result.Score = 4
		result.Rating = ScoreLabels[4]
		result.Description = "The audience is well-defined but their familiarity with the topic is unclear."
		result.Recommendation = "Specify how familiar the audience is with the topic."
	}

	return result
}

// evaluateTone assesses whether the prompt specifies the desired tone
func evaluateTone(text string, _ int) AssessmentResult {
	result := AssessmentResult{
		Name:        "Tone",
		Score:       5,
		Rating:      ScoreLabels[5],
		Description: "The desired tone is precisely specified with clear guidance.",
	}

	toneWords := []string{
		"formal", "casual", "friendly", "professional", "encouraging",
		"tone", "style", "voice", "simple", "technical",
	}

	hasTone := containsAny(text, toneWords)

	if !hasTone {
		result.Score = 2
		result.Rating = ScoreLabels[2]
		result.Description = "No tone specification."
		result.Recommendation = "Define the tone (e.g., 'in a friendly tone')."
	} else if !strings.Contains(strings.ToLower(text), "level") && !strings.Contains(strings.ToLower(text), "very") {
		result.Score = 3
		result.Rating = ScoreLabels[3]
		result.Description = "A tone is mentioned but its intensity is unclear."
		result.Recommendation = "Specify how formal/casual the tone should be."
	} else if !strings.Contains(strings.ToLower(text), "example") {
		result.Score = 4
		result.Rating = ScoreLabels[4]
		result.Description = "The tone is well-specified but lacks an example."
		result.Recommendation = "Consider providing an example of the desired tone."
	}

	return result
}

// evaluateData assesses the richness of information or data provided in the prompt
func evaluateData(text string, wordCount int) AssessmentResult {
	result := AssessmentResult{
		Name:        "Data",
		Score:       5,
		Rating:      ScoreLabels[5],
		Description: "The prompt includes comprehensive relevant data and examples.",
	}

	if wordCount < 5 {
		result.Score = 1
		result.Rating = ScoreLabels[1]
		result.Description = "No data or information provided at all."
		result.Recommendation = "Include specific information or examples."
	} else if wordCount < 10 && !strings.Contains(text, "example") && !strings.Contains(text, "data") {
		result.Score = 2
		result.Rating = ScoreLabels[2]
		result.Description = "Very little information provided."
		result.Recommendation = "Add key information or examples related to the task."
	} else if wordCount < 30 && !strings.Contains(text, "detail") {
		result.Score = 3
		result.Rating = ScoreLabels[3]
		result.Description = "Some data provided but could be more comprehensive."
		result.Recommendation = "Include more specific details or examples."
	} else if !strings.Contains(text, "context") && wordCount < 70 {
		result.Score = 4
		result.Rating = ScoreLabels[4]
		result.Description = "Good data provided but could be more contextualized."
		result.Recommendation = "Add more context to your data."
	}

	return result
}

// Helper functions

// containsAny checks if the text contains any of the words in the list
func containsAny(text string, words []string) bool {
	lowercaseText := strings.ToLower(text)
	for _, word := range words {
		if strings.Contains(lowercaseText, word) {
			return true
		}
	}
	return false
}

// calculateOverallScore calculates the percentage score based on all criteria
func calculateOverallScore(criteria map[string]AssessmentResult) (int, string) {
	totalPoints := 0
	maxPoints := len(criteria) * 5

	for _, result := range criteria {
		totalPoints += result.Score
	}

	percentageScore := totalPoints * 100 / maxPoints

	var rating string
	if percentageScore >= 90 {
		rating = "Excellent"
	} else if percentageScore >= 75 {
		rating = "Very Good"
	} else if percentageScore >= 60 {
		rating = "Good"
	} else if percentageScore >= 40 {
		rating = "Average"
	} else if percentageScore >= 25 {
		rating = "Needs Improvement"
	} else {
		rating = "Poor"
	}

	return percentageScore, rating
}

// evaluatePrompt performs all evaluations and returns a complete assessment
func evaluatePrompt(text string) PromptAssessment {
	wordCount := len(strings.Fields(text))

	criteria := make(map[string]AssessmentResult)

	// Perform all evaluations
	criteria["Clarity"] = evaluateClarity(text, wordCount)
	criteria["Relevance"] = evaluateRelevance(text, wordCount)
	criteria["Specificity"] = evaluateSpecificity(text, wordCount)
	criteria["Context"] = evaluateContext(text, wordCount)
	criteria["Richness"] = evaluateRichness(text, wordCount)
	criteria["Persona"] = evaluatePersona(text, wordCount)
	criteria["Instruction"] = evaluateInstruction(text, wordCount)
	criteria["Format"] = evaluateFormat(text, wordCount)
	criteria["Audience"] = evaluateAudience(text, wordCount)
	criteria["Tone"] = evaluateTone(text, wordCount)
	criteria["Data"] = evaluateData(text, wordCount)

	scorePercentage, overallRating := calculateOverallScore(criteria)

	return PromptAssessment{
		Criteria:      criteria,
		TotalScore:    scorePercentage,
		OverallRating: overallRating,
	}
}

// renderAssessment displays the assessment results to the user
func renderAssessment(assessment PromptAssessment) {
	assessmentColor := color.New(color.FgHiCyan)

	assessmentColor.Println("\nPrompt Assessment:")
	assessmentColor.Println("=== Prompt Quality Assessment ===")

	// Display individual criteria scores
	for _, name := range []string{
		"Clarity", "Relevance", "Specificity", "Context", "Richness",
	} {
		result := assessment.Criteria[name]
		assessmentColor.Printf("- %s [%d/5]: %s. %s %s\n",
			result.Name,
			result.Score,
			result.Rating,
			result.Description,
			ScoreIcons[result.Score],
		)
	}

	// Display advanced prompt structure assessment
	assessmentColor.Println("\n=== Advanced Prompt Structure Assessment ===")

	for _, name := range []string{
		"Persona", "Instruction", "Format", "Audience", "Tone", "Data",
	} {
		result := assessment.Criteria[name]
		assessmentColor.Printf("- %s [%d/5]: %s. %s %s\n",
			result.Name,
			result.Score,
			result.Rating,
			result.Description,
			ScoreIcons[result.Score],
		)
	}

	// Display overall score
	assessmentColor.Printf("\n📊 Overall Score: %d%% - %s\n",
		assessment.TotalScore,
		assessment.OverallRating,
	)

	// Display checklist summary
	assessmentColor.Println("\n🔍 Assessment Summary:")

	for _, category := range []string{
		"Clarity", "Relevance", "Specificity", "Context", "Richness",
		"Persona", "Instruction", "Format", "Audience", "Tone", "Data",
	} {
		result := assessment.Criteria[category]
		assessmentColor.Printf("%s %s (%s)\n",
			ScoreIcons[result.Score],
			result.Name,
			result.Rating,
		)
	}

	// Display recommendations
	assessmentColor.Println("\n💡 Recommendations:")
	hasRecommendations := false

	for _, result := range assessment.Criteria {
		if result.Recommendation != "" {
			assessmentColor.Printf("- %s\n", result.Recommendation)
			hasRecommendations = true
		}
	}

	if !hasRecommendations {
		assessmentColor.Println("- None needed, great prompt!")
	}
}

// AssessPrompt evaluates the quality and structure of a prompt
func AssessPrompt(text string) {
	assessment := evaluatePrompt(text)
	renderAssessment(assessment)
	// Improved prompt section has been removed
}

// EvaluatePromptForHistory evaluates a prompt and returns the assessment without displaying it
// This is used for logging the assessment to history
func EvaluatePromptForHistory(text string) PromptAssessment {
	return evaluatePrompt(text)
}

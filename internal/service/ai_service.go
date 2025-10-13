package service

import (
	"context"
	"fmt"

	"schools-be/internal/config"
	"schools-be/internal/models"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type AIService struct {
	config *config.Config
	client *genai.Client
}

func NewAIService(ctx context.Context, config *config.Config) (*AIService, error) {
	if config.GeminiAPIKey == "" {
		return nil, fmt.Errorf("Gemini API key is not configured")
	}

	client, err := genai.NewClient(ctx, option.WithAPIKey(config.GeminiAPIKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini API client: %w", err)
	}

	return &AIService{
		config: config,
		client: client,
	}, nil
}

func (s *AIService) Close() error {
	if s.client != nil {
		return s.client.Close()
	}
	return nil
}

// GenerateSchoolSummary generates a comprehensive summary for a school using Gemini AI
func (s *AIService) GenerateSchoolSummary(ctx context.Context, school *models.EnrichedSchool) (string, error) {
	if s.client == nil {
		return "", fmt.Errorf("AI client is not initialized")
	}

	// Build the prompt with complete school information
	prompt := s.createEnrichedSchoolPrompt(school)

	// Get the generative model
	model := s.client.GenerativeModel("gemini-2.5-flash")

	// Generate content
	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", fmt.Errorf("failed to generate content: %w", err)
	}

	// Extract the text from the response
	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no content generated")
	}

	summary := ""
	for _, part := range resp.Candidates[0].Content.Parts {
		if text, ok := part.(genai.Text); ok {
			summary += string(text)
		}
	}

	return summary, nil
}

func (s *AIService) createEnrichedSchoolPrompt(data *models.EnrichedSchool) string {
	school := data.School
	details := data.Details
	languageStat := data.LanguageStat
	citizenshipStats := data.CitizenshipStats
	residenceStats := data.ResidenceStats
	absenceStat := data.AbsenceStat
	statistics := data.Statistics

	var stats *models.SchoolStatistic
	if len(statistics) > 0 {
		stats = &statistics[0]
	}

	prompt := fmt.Sprintf(`You are an expert educational consultant creating brief, informative profiles of Berlin schools for international parents. Your tone should be professional yet accessible, clear, and direct.

**School Basic Information:**
- Name: %s
- Type: %s
- Operator: %s
- Address: %s %s, %s Berlin
- District: %s, %s
- Website: %s
- Phone: %s
- Email: %s
`, school.Name, school.SchoolCategory, school.Operator, school.Street, school.HouseNumber, school.PostalCode, school.District, school.Neighborhood, stringOrNA2(school.Website), stringOrNA2(school.Phone), stringOrNA2(school.Email))

	// Add statistics if available
	if stats != nil {
		prompt += fmt.Sprintf(`
**Student & Teacher Statistics (%s):**
- Total Students: %s (%s female, %s male)
- Total Teachers: %s (%s female, %s male)
- Total Classes: %s
`, stats.SchoolYear, stats.Students, stats.StudentsFemale, stats.StudentsMale, stats.Teachers, stats.TeachersFemale, stats.TeachersMale, stats.Classes)
	}

	// Add language/heritage statistics
	if languageStat != nil {
		prompt += fmt.Sprintf(`
**Language & Heritage Statistics:**
- Total Students: %d
- Students with Non-German Heritage: %d (%.1f%%)
  - Female: %d, Male: %d
`, languageStat.TotalStudents, languageStat.NDHTotal, languageStat.NDHPercentage, languageStat.NDHFemaleStudents, languageStat.NDHMaleStudents)
	}

	// Add citizenship statistics
	if len(citizenshipStats) > 0 {
		var totalRow *models.SchoolCitizenshipStat
		for i := range citizenshipStats {
			if citizenshipStats[i].Citizenship == "Insgesamt" || citizenshipStats[i].Citizenship == "insgesamt" {
				totalRow = &citizenshipStats[i]
				break
			}
		}

		if totalRow != nil {
			prompt += fmt.Sprintf(`
**Citizenship Statistics:**
- Students with Non-German Citizenship: %d (%d female, %d male)
`, totalRow.Total, totalRow.FemaleStudents, totalRow.MaleStudents)

			// Add regional distribution
			regionalDist := "- Regional Distribution: "
			count := 0
			for _, stat := range citizenshipStats {
				if stat.Citizenship != "Insgesamt" && stat.Citizenship != "insgesamt" && count < 5 {
					if count > 0 {
						regionalDist += ", "
					}
					regionalDist += fmt.Sprintf("%s: %d", stat.Citizenship, stat.Total)
					count++
				}
			}
			if count > 0 {
				prompt += regionalDist + "\n"
			}
		}
	}

	// Add residence statistics
	if len(residenceStats) > 0 {
		prompt += "\n**Student Residence Distribution:**\n- Top districts where students live: "

		// Get top 5 districts by student count
		topCount := 5
		if len(residenceStats) < topCount {
			topCount = len(residenceStats)
		}

		for i := 0; i < topCount; i++ {
			if i > 0 {
				prompt += ", "
			}
			prompt += fmt.Sprintf("%s (%d)", residenceStats[i].District, residenceStats[i].StudentCount)
		}
		prompt += "\n"
	}

	// Add absence statistics
	if absenceStat != nil {
		prompt += fmt.Sprintf(`
**Absence Statistics:**
- School Absence Rate: %.1f%% (Unexcused: %.1f%%)
- School Type Average: %.1f%% (Unexcused: %.1f%%)
- Berlin Average: %.1f%% (Unexcused: %.1f%%)
`, absenceStat.SchoolAbsenceRate, absenceStat.SchoolUnexcusedRate, absenceStat.SchoolTypeAbsenceRate, absenceStat.SchoolTypeUnexcusedRate, absenceStat.BerlinAbsenceRate, absenceStat.BerlinUnexcusedRate)
	}

	// Add detailed school information
	if details != nil {
		if details.Languages != "" {
			prompt += fmt.Sprintf("\n**Languages Offered:**\n%s\n", details.Languages)
		}

		if details.Courses != "" {
			prompt += fmt.Sprintf("\n**Advanced Courses (Leistungskurse):**\n%s\n", details.Courses)
		}

		if details.Offerings != "" {
			prompt += fmt.Sprintf("\n**Programs & Special Offerings:**\n%s\n", details.Offerings)
		}

		if details.Equipment != "" {
			prompt += fmt.Sprintf("\n**Equipment & Facilities:**\n%s\n", details.Equipment)
		}

		if details.WorkingGroups != "" {
			prompt += fmt.Sprintf("\n**Working Groups & Extracurricular Activities:**\n%s\n", details.WorkingGroups)
		}

		if details.Partners != "" {
			prompt += fmt.Sprintf("\n**External Partners:**\n%s\n", details.Partners)
		}

		if details.Differentiation != "" {
			prompt += fmt.Sprintf("\n**Differentiation & Teaching Methods:**\n%s\n", details.Differentiation)
		}

		if details.LunchInfo != "" {
			prompt += fmt.Sprintf("\n**Lunch & Meal Services:**\n%s\n", details.LunchInfo)
		}

		if details.DualLearning != "" {
			prompt += fmt.Sprintf("\n**Dual Learning Programs:**\n%s\n", details.DualLearning)
		}

		if details.AdditionalInfo != "" {
			prompt += fmt.Sprintf("\n**Additional Information:**\n%s\n", details.AdditionalInfo)
		}

		available := "No"
		if details.AvailableAfter4thGrade {
			available = "Yes"
		}
		prompt += fmt.Sprintf("\n**Enrollment:** Accepts students after 4th grade: %s\n", available)
	}

	prompt += fmt.Sprintf(`
**Task:**
Synthesize all the data above into a concise, informative school profile.

1. **Prioritize the official website** (%s) for additional qualitative information if needed.
2. **Use all the provided data** to create an accurate, comprehensive summary.
3. **Focus on unique characteristics** that distinguish this school.

**Output Requirements:**
- **Total Length:** Must be under 300 words (given the rich data available, be comprehensive but concise).
- **Structure:** Use the following **bold** headers:
    - **Profile:**
    - **Academics & Languages:**
    - **Diversity & Student Body:**
    - **Extracurriculars & Facilities:**
- **Formatting:** Use short, concise bullet points (•).
- **Style:** Be factual and specific. Use actual numbers and statistics from the data. Avoid generic statements and conversational filler. Focus on concrete details that help parents make informed decisions.
`, stringOrNA2(school.Website))

	return prompt
}

// Helper functions
func stringOrNA(s *string) string {
	if s == nil || *s == "" {
		return "N/A"
	}
	return *s
}

func stringOrNA2(s string) string {
	if s == "" {
		return "N/A"
	}
	return s
}

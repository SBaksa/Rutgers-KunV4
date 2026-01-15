package commands

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

// Rutgers API structures
type RutgersCourse struct {
	CourseNumber string     `json:"courseNumber"`
	Subject      string     `json:"subject"`
	Title        string     `json:"title"`
	PreReqNotes  string     `json:"preReqNotes"`
	SynopsisURL  string     `json:"synopsisUrl"`
	CoreCodes    []CoreCode `json:"coreCodes"`
	Sections     []Section  `json:"sections"`
}

type CoreCode struct {
	Code        string `json:"code"`
	Description string `json:"description"`
}

type Section struct {
	Number       string        `json:"number"`
	Index        string        `json:"index"`
	Instructors  []Instructor  `json:"instructors"`
	MeetingTimes []MeetingTime `json:"meetingTimes"`
	Notes        string        `json:"notes"`
}

type Instructor struct {
	Name string `json:"name"`
}

type MeetingTime struct {
	MeetingDay      string `json:"meetingDay"`
	StartTime       string `json:"startTime"`
	EndTime         string `json:"endTime"`
	PmCode          string `json:"pmCode"`
	CampusName      string `json:"campusName"`
	BuildingCode    string `json:"buildingCode"`
	RoomNumber      string `json:"roomNumber"`
	MeetingModeDesc string `json:"meetingModeDesc"`
}

var (
	// Regex for parsing course codes: subject:course or subject:course:section
	courseRegex = regexp.MustCompile(`^(?:(\d{2}):)?(\d{3}):(\d{3})(?::(\w{2}))?$`)
	httpClient  = &http.Client{Timeout: 15 * time.Second}
)

// Course command handler
func Course(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	if len(args) == 0 {
		s.ChannelMessageSend(m.ChannelID, "Please provide a course code like `198:111` or `01:198:111:01`")
		return
	}

	// Parse the course code
	matches := courseRegex.FindStringSubmatch(args[0])
	if matches == nil {
		s.ChannelMessageSend(m.ChannelID, "❌ Invalid course format. Use `subject:course` (e.g., `198:111`) or `school:subject:course:section` (e.g., `01:198:111:01`)")
		return
	}

	subject := matches[2] // e.g., "198"
	course := matches[3]  // e.g., "111"
	section := matches[4] // e.g., "01" (optional)

	// Determine semester and level
	semester := getCurrentSemester()
	level := "UG"
	if courseNum, err := strconv.Atoi(course); err == nil && courseNum >= 500 {
		level = "G" // Graduate level
	}

	// Make API request
	courseData, err := fetchCourseFromAPI(subject, semester, "NB", level)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("❌ Error fetching course data: %v", err))
		return
	}

	// Find the specific course
	var targetCourse *RutgersCourse
	for i := range courseData {
		if courseData[i].CourseNumber == course {
			targetCourse = &courseData[i]
			break
		}
	}

	if targetCourse == nil {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Course %s:%s not found for current semester", subject, course))
		return
	}

	// Create embed
	embed := createCourseEmbed(targetCourse, section)

	// Send the message
	_, err = s.ChannelMessageSendEmbed(m.ChannelID, embed)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("❌ Error sending course info: %v", err))
	}
}

// fetchCourseFromAPI queries the Rutgers SIS API
func fetchCourseFromAPI(subject, semester, campus, level string) ([]RutgersCourse, error) {
	url := fmt.Sprintf("http://sis.rutgers.edu/oldsoc/courses.json?subject=%s&semester=%s&campus=%s&level=%s",
		subject, semester, campus, level)

	resp, err := httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var courses []RutgersCourse
	if err := json.NewDecoder(resp.Body).Decode(&courses); err != nil {
		return nil, fmt.Errorf("failed to decode JSON: %w", err)
	}

	return courses, nil
}

// getCurrentSemester determines current semester based on date
func getCurrentSemester() string {
	now := time.Now()
	year := now.Year()
	month := now.Month()
	day := now.Day()

	var season string

	// Determine season based on date ranges from JS implementation
	if month >= 8 || (month == 7 && day >= 12) {
		// Fall semester
		if month >= 8 {
			season = "9"
		} else {
			season = "9"
		}
	} else if month <= 4 || (month == 5 && day < 23) {
		// Spring semester
		if month >= 1 || (month == 1 && day >= 12) {
			season = "1"
		} else {
			// Winter session (previous year)
			season = "0"
			year--
		}
	} else {
		// Summer session
		season = "7"
	}

	return fmt.Sprintf("%s%d", season, year)
}

// createCourseEmbed builds a Discord embed for the course
func createCourseEmbed(course *RutgersCourse, requestedSection string) *discordgo.MessageEmbed {
	embed := &discordgo.MessageEmbed{
		Title: course.Title,
		Color: 0xCC0033, // Rutgers red
	}

	// Add URL if available
	if course.SynopsisURL != "" {
		embed.URL = course.SynopsisURL
	}

	// Add prerequisites
	if course.PreReqNotes != "" {
		prereqs := strings.ReplaceAll(course.PreReqNotes, "<em>", "")
		prereqs = strings.ReplaceAll(prereqs, "</em>", "")
		prereqs = strings.ReplaceAll(prereqs, " )", ")")

		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:  "Prerequisites",
			Value: prereqs,
		})
	}

	// Add core requirements
	if len(course.CoreCodes) > 0 {
		var coreList []string
		for _, core := range course.CoreCodes {
			coreList = append(coreList, fmt.Sprintf("%s (%s)", core.Code, core.Description))
		}
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:  "Requirements Satisfied",
			Value: strings.Join(coreList, "\n"),
		})
	}

	// Handle sections
	if requestedSection != "" {
		// Show specific section details
		addSectionDetails(embed, course, requestedSection)
	} else {
		// Show professors and their sections
		addProfessorSections(embed, course)
	}

	return embed
}

// addSectionDetails adds detailed info for a specific section
func addSectionDetails(embed *discordgo.MessageEmbed, course *RutgersCourse, sectionNum string) {
	var targetSection *Section
	for i := range course.Sections {
		if course.Sections[i].Number == sectionNum {
			targetSection = &course.Sections[i]
			break
		}
	}

	if targetSection == nil {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:  "Error",
			Value: fmt.Sprintf("Section %s not found", sectionNum),
		})
		return
	}

	embed.Description = fmt.Sprintf("**Section %s**\nIndex: %s", targetSection.Number, targetSection.Index)

	// Add instructors
	if len(targetSection.Instructors) > 0 {
		var instructorNames []string
		for _, instructor := range targetSection.Instructors {
			instructorNames = append(instructorNames, instructor.Name)
		}
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:  "Instructors",
			Value: strings.Join(instructorNames, "\n"),
		})
	}

	// Add meeting times
	if len(targetSection.MeetingTimes) > 0 {
		for _, meeting := range targetSection.MeetingTimes {
			fieldName := fmt.Sprintf("%s %s",
				dayCodeToFull(meeting.MeetingDay),
				modeCodeToFull(meeting.MeetingModeDesc))

			if meeting.CampusName != "" {
				fieldName += " on " + meeting.CampusName
			} else {
				fieldName += " online"
			}

			location := ""
			if meeting.BuildingCode != "" {
				location = fmt.Sprintf("%s-%s", meeting.BuildingCode, meeting.RoomNumber)
			}

			timeStr := formatTime(meeting.StartTime, meeting.EndTime, meeting.PmCode)

			value := ""
			if location != "" {
				value = fmt.Sprintf("%s from %s", location, timeStr)
			} else {
				value = timeStr
			}

			embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
				Name:  fieldName,
				Value: value,
			})
		}
	}

	// Add notes
	if targetSection.Notes != "" {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:  "Notes",
			Value: targetSection.Notes,
		})
	}
}

// addProfessorSections adds professor info for all sections
func addProfessorSections(embed *discordgo.MessageEmbed, course *RutgersCourse) {
	professorMap := make(map[string][]string)

	// Group sections by professor
	for _, section := range course.Sections {
		if len(section.Instructors) > 0 {
			for _, instructor := range section.Instructors {
				professorMap[instructor.Name] = append(professorMap[instructor.Name], section.Number)
			}
		}
	}

	// Add professor fields (limit to prevent embed size issues)
	count := 0
	for professor, sections := range professorMap {
		if count >= 20 { // Discord embed field limit
			embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
				Name:  "Note",
				Value: "Results truncated due to length",
			})
			break
		}

		sectionList := "None"
		if len(sections) > 0 {
			sectionList = strings.Join(sections, ", ")
		}

		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:  professor,
			Value: fmt.Sprintf("Sections: %s", sectionList),
		})
		count++
	}
}

// Helper functions for formatting

func dayCodeToFull(code string) string {
	switch code {
	case "M":
		return "Monday"
	case "T":
		return "Tuesday"
	case "W":
		return "Wednesday"
	case "TH":
		return "Thursday"
	case "F":
		return "Friday"
	default:
		return "Unknown"
	}
}

func modeCodeToFull(code string) string {
	switch code {
	case "LEC":
		return "Lecture"
	case "RECIT":
		return "Recitation"
	case "WORKSHOP":
		return "Workshop"
	case "REMOTE-SYNCH":
		return "Remote Synchronous"
	case "REMOTE-ASYNCH":
		return "Remote Asynchronous"
	default:
		return "Unknown"
	}
}

func formatTime(start, end, pmCode string) string {
	if start == "" || end == "" {
		return "No time provided"
	}

	ampm := "?M"
	if strings.ToLower(pmCode) == "p" {
		ampm = "PM"
	} else if strings.ToLower(pmCode) == "a" {
		ampm = "AM"
	}

	startFormatted := fmt.Sprintf("%s:%s", start[:2], start[2:])
	endFormatted := fmt.Sprintf("%s:%s", end[:2], end[2:])

	return fmt.Sprintf("%s %s to %s %s", startFormatted, ampm, endFormatted, ampm)
}

package commands

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/SBaksa/Rutgers-KunV4/logger"
	"github.com/SBaksa/Rutgers-KunV4/verification"
	"github.com/bwmarrin/discordgo"
)

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

type courseCacheEntry struct {
	data      []RutgersCourse
	timestamp time.Time
}

var (
	courseRegex = regexp.MustCompile(`^(?:(\d{2}):)?(\d{3}):(\d{3})(?::(\w{2}))?$`)
	httpClient  = &http.Client{Timeout: 15 * time.Second}

	apiEndpoints = []string{
		"https://sis.rutgers.edu/soc/api/courses.gzip?year=%s&term=%s&campus=%s",
		"https://classes.rutgers.edu/soc/api/courses.gzip?year=%s&term=%s&campus=%s",
	}

	courseCache = make(map[string]courseCacheEntry)
	cacheMutex  sync.RWMutex
	cacheTTL    = 24 * time.Hour
)

func Course(s *discordgo.Session, m *discordgo.MessageCreate, args []string, log *logger.Logger, vm *verification.VerificationManager) error {
	if len(args) == 0 {
		s.ChannelMessageSend(m.ChannelID, "Please provide a course code like `198:111` or `01:198:111:01`")
		return nil
	}

	matches := courseRegex.FindStringSubmatch(args[0])
	if matches == nil {
		s.ChannelMessageSend(m.ChannelID, "❌ Invalid course format. Use `subject:course` (e.g., `198:111`) or `school:subject:course:section` (e.g., `01:198:111:01`)")
		return nil
	}

	subject := matches[2]
	course := matches[3]
	section := matches[4]

	semester := getCurrentSemester()
	level := "UG"
	if courseNum, err := strconv.Atoi(course); err == nil && courseNum >= 500 {
		level = "G"
	}

	courseData, err := fetchCourseFromAPI(subject, semester, "NB", level)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("❌ Error fetching course data: %v", err))
		return nil
	}

	var targetCourse *RutgersCourse
	for i := range courseData {
		if courseData[i].CourseNumber == course {
			targetCourse = &courseData[i]
			break
		}
	}

	if targetCourse == nil {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Course %s:%s not found for current semester", subject, course))
		return nil
	}

	embed := createCourseEmbed(targetCourse, section)

	_, err = s.ChannelMessageSendEmbed(m.ChannelID, embed)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("❌ Error sending course info: %v", err))
		return err
	}

	return nil
}

func fetchCourseFromAPI(subject, semester, campus, level string) ([]RutgersCourse, error) {
	var year, term string
	if len(semester) >= 5 {
		term = semester[:1]
		year = semester[1:]
	} else {
		return nil, fmt.Errorf("invalid semester format: %s", semester)
	}

	cacheKey := fmt.Sprintf("%s:%s:%s:%s", subject, semester, campus, level)

	cacheMutex.RLock()
	entry, found := courseCache[cacheKey]
	cacheMutex.RUnlock()

	if found && time.Since(entry.timestamp) < cacheTTL {
		return entry.data, nil
	}

	var resp *http.Response
	var lastErr error
	var url string

	for _, endpoint := range apiEndpoints {
		url = fmt.Sprintf(endpoint, year, term, campus)
		resp, lastErr = httpClient.Get(url)

		if lastErr == nil && resp.StatusCode == 200 {
			break
		}

		if resp != nil {
			resp.Body.Close()
		}
	}

	if resp == nil || resp.StatusCode != 200 {
		if lastErr != nil {
			return nil, fmt.Errorf("all API endpoints failed: %w", lastErr)
		}
		return nil, fmt.Errorf("all API endpoints returned non-200 status")
	}
	defer resp.Body.Close()

	var reader io.Reader
	if resp.Header.Get("Content-Encoding") == "gzip" || strings.HasSuffix(url, ".gzip") {
		gzipReader, err := gzip.NewReader(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to create gzip reader: %w", err)
		}
		defer gzipReader.Close()
		reader = gzipReader
	} else {
		reader = resp.Body
	}

	var courses []RutgersCourse
	if err := json.NewDecoder(reader).Decode(&courses); err != nil {
		return nil, fmt.Errorf("failed to decode JSON: %w", err)
	}

	var filtered []RutgersCourse
	for _, course := range courses {
		if course.Subject == subject {
			filtered = append(filtered, course)
		}
	}

	cacheMutex.Lock()
	courseCache[cacheKey] = courseCacheEntry{
		data:      filtered,
		timestamp: time.Now(),
	}
	cacheMutex.Unlock()

	return filtered, nil
}

func getCurrentSemester() string {
	now := time.Now()
	year := now.Year()
	month := now.Month()
	day := now.Day()

	var season string

	if month >= 8 || (month == 7 && day >= 12) {
		season = "9"
	} else if month <= 4 || (month == 5 && day < 23) {
		if month > 1 || (month == 1 && day >= 12) {
			season = "1"
		} else {
			season = "0"
			year--
		}
	} else {
		season = "7"
	}

	return fmt.Sprintf("%s%d", season, year)
}

func createCourseEmbed(course *RutgersCourse, requestedSection string) *discordgo.MessageEmbed {
	embed := &discordgo.MessageEmbed{
		Title: course.Title,
		Color: 0xCC0033,
	}

	if course.SynopsisURL != "" {
		embed.URL = course.SynopsisURL

		if description := scrapeCourseDescription(course.SynopsisURL); description != "" {
			if len(description) > 1024 {
				description = description[:1021] + "..."
			}
			embed.Description = description
		}
	}

	if course.PreReqNotes != "" {
		prereqs := strings.ReplaceAll(course.PreReqNotes, "<em>", "")
		prereqs = strings.ReplaceAll(prereqs, "</em>", "")
		prereqs = strings.ReplaceAll(prereqs, " )", ")")

		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:  "Prerequisites",
			Value: prereqs,
		})
	}

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

	if requestedSection != "" {
		addSectionDetails(embed, course, requestedSection)
	} else {
		addProfessorSections(embed, course)
	}

	return embed
}

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

	if targetSection.Notes != "" {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:  "Notes",
			Value: targetSection.Notes,
		})
	}
}

func addProfessorSections(embed *discordgo.MessageEmbed, course *RutgersCourse) {
	professorMap := make(map[string][]string)

	for _, section := range course.Sections {
		if len(section.Instructors) > 0 {
			for _, instructor := range section.Instructors {
				professorMap[instructor.Name] = append(professorMap[instructor.Name], section.Number)
			}
		}
	}

	count := 0
	for professor, sections := range professorMap {
		if count >= 20 {
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

	if len(start) < 4 || len(end) < 4 {
		return "No time provided"
	}
	startFormatted := fmt.Sprintf("%s:%s", start[:2], start[2:])
	endFormatted := fmt.Sprintf("%s:%s", end[:2], end[2:])

	return fmt.Sprintf("%s %s to %s %s", startFormatted, ampm, endFormatted, ampm)
}

func scrapeCourseDescription(synopsisURL string) string {
	if synopsisURL == "" {
		return ""
	}

	resp, err := httpClient.Get(synopsisURL)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return ""
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return ""
	}

	var description string

	doc.Find(".course-description, .description, p").Each(func(i int, s *goquery.Selection) {
		text := strings.TrimSpace(s.Text())
		if len(text) > 50 && description == "" {
			description = text
		}
	})

	if description == "" {
		if metaDesc, exists := doc.Find("meta[name='description']").Attr("content"); exists {
			description = strings.TrimSpace(metaDesc)
		}
	}

	description = strings.ReplaceAll(description, "\n", " ")
	description = strings.ReplaceAll(description, "\t", " ")
	description = regexp.MustCompile(`\s+`).ReplaceAllString(description, " ")
	description = strings.TrimSpace(description)

	return description
}

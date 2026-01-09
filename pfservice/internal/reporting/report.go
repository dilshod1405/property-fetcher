package reporting

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type ReportStats struct {
	Date              time.Time
	PropertiesCreated int
	PropertiesUpdated int
	ImagesDownloaded  int
	UsersCreated      int
	UsersUpdated      int
	Errors            int
}

var ReportFile = getReportFile()

func getReportFile() string {
	if v := os.Getenv("REPORT_FILE"); v != "" {
		return v
	}
	return "/mhp/report.txt"
}

func getTashkentTime() time.Time {
	loc, err := time.LoadLocation("Asia/Tashkent")
	if err != nil {
		// Fallback to UTC if timezone loading fails
		return time.Now().UTC()
	}
	return time.Now().In(loc)
}

// GetTashkentTime returns current time in Tashkent timezone
func GetTashkentTime() time.Time {
	return getTashkentTime()
}

// WriteReport writes statistics to reports.txt file
func WriteReport(stats ReportStats) error {
	// Ensure directory exists
	dir := filepath.Dir(ReportFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create report directory: %w", err)
	}

	// Open file in append mode
	file, err := os.OpenFile(ReportFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open report file: %w", err)
	}
	defer file.Close()

	// Get Tashkent time
	tashkentTime := getTashkentTime()
	if !stats.Date.IsZero() {
		tashkentTime = stats.Date
	}

	// Format date and time
	dateStr := tashkentTime.Format("2006-01-02")
	timeStr := tashkentTime.Format("15:04:05")

	// Write header if file is empty or new day
	fileInfo, _ := file.Stat()
	if fileInfo.Size() == 0 {
		writeHeader(file)
	}

	// Write report entry as table row
	line := fmt.Sprintf("| %s | %s | %6d | %6d | %6d | %6d | %6d | %6d |\n",
		dateStr,
		timeStr,
		stats.PropertiesCreated,
		stats.PropertiesUpdated,
		stats.ImagesDownloaded,
		stats.UsersCreated,
		stats.UsersUpdated,
		stats.Errors,
	)

	_, err = file.WriteString(line)
	if err != nil {
		return fmt.Errorf("failed to write report: %w", err)
	}

	return nil
}

func writeHeader(file *os.File) {
	header := `+------------+----------+--------------+--------------+------------------+--------------+--------------+--------+
|    Date    |   Time   | Prop Created | Prop Updated | Images Downloaded | User Created | User Updated | Errors |
+------------+----------+--------------+--------------+------------------+--------------+--------------+--------+
`
	file.WriteString(header)
}

// WriteSummary writes a summary line at the end
func WriteSummary(stats ReportStats) error {
	file, err := os.OpenFile(ReportFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open report file: %w", err)
	}
	defer file.Close()

	tashkentTime := getTashkentTime()
	if !stats.Date.IsZero() {
		tashkentTime = stats.Date
	}

	summary := fmt.Sprintf(`
+------------+----------+------------+------------+----------------+------------+------------+--------+
| SUMMARY    | %s | %6d | %6d | %6d | %6d | %6d | %6d |
+------------+----------+------------+------------+----------------+------------+------------+--------+
`,
		tashkentTime.Format("2006-01-02 15:04:05"),
		stats.PropertiesCreated,
		stats.PropertiesUpdated,
		stats.ImagesDownloaded,
		stats.UsersCreated,
		stats.UsersUpdated,
		stats.Errors,
	)

	_, err = file.WriteString(summary)
	return err
}

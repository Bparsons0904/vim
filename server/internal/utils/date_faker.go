package utils

import (
	"math/rand"
	"strconv"
	"time"
)

type DateFaker struct {
	random *rand.Rand
}

type FakeDataOptions struct {
	Count        int
	StartYear    int
	EndYear      int
	IncludeTime  bool
	FormatMix    bool // If true, generates variety of formats; if false, uses specific format
	TargetFormat DateFormat
}

func NewDateFaker() *DateFaker {
	return &DateFaker{
		random: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (df *DateFaker) SetSeed(seed int64) {
	df.random = rand.New(rand.NewSource(seed))
}

// GenerateFakeDates generates a slice of fake date strings in various formats
func (df *DateFaker) GenerateFakeDates(options FakeDataOptions) []string {
	if options.Count <= 0 {
		options.Count = 10
	}
	if options.StartYear <= 0 {
		options.StartYear = 1990
	}
	if options.EndYear <= 0 {
		options.EndYear = 2030
	}

	dates := make([]string, options.Count)
	formats := df.getFormatsToUse(options)

	for i := 0; i < options.Count; i++ {
		// Generate random time within the year range
		randomTime := df.generateRandomTime(options.StartYear, options.EndYear, options.IncludeTime)
		
		// Select format
		var selectedFormat DateFormat
		if options.FormatMix {
			selectedFormat = formats[df.random.Intn(len(formats))]
		} else if options.TargetFormat != "" {
			selectedFormat = options.TargetFormat
		} else {
			selectedFormat = formats[df.random.Intn(len(formats))]
		}

		dates[i] = df.formatTimeWithFormat(randomTime, selectedFormat)
	}

	return dates
}

func (df *DateFaker) getFormatsToUse(options FakeDataOptions) []DateFormat {
	if !options.FormatMix && options.TargetFormat != "" {
		return []DateFormat{options.TargetFormat}
	}

	baseFormats := []DateFormat{
		FormatISO8601Date,
		FormatUSDate,
		FormatEuropeanDate,
		FormatDashDate,
		FormatDotDate,
		FormatMonthDay,
		FormatShortMonth,
		FormatYearMonth,
		FormatUnixTime,
	}

	if options.IncludeTime {
		timeFormats := []DateFormat{
			FormatISO8601,
			FormatUSDateTime,
			FormatRFC3339,
			FormatRFC822,
			FormatTime24,
			FormatTime12,
		}
		baseFormats = append(baseFormats, timeFormats...)
	}

	return baseFormats
}

func (df *DateFaker) generateRandomTime(startYear, endYear int, includeTime bool) time.Time {
	// Generate random year, month, day
	year := startYear + df.random.Intn(endYear-startYear+1)
	month := time.Month(1 + df.random.Intn(12))
	
	// Get the number of days in the selected month/year
	daysInMonth := df.getDaysInMonth(year, month)
	day := 1 + df.random.Intn(daysInMonth)

	var hour, minute, second int
	if includeTime {
		hour = df.random.Intn(24)
		minute = df.random.Intn(60)
		second = df.random.Intn(60)
	}

	return time.Date(year, month, day, hour, minute, second, 0, time.UTC)
}

func (df *DateFaker) getDaysInMonth(year int, month time.Month) int {
	// Get the first day of the next month, then subtract one day
	firstOfNextMonth := time.Date(year, month+1, 1, 0, 0, 0, 0, time.UTC)
	lastOfThisMonth := firstOfNextMonth.AddDate(0, 0, -1)
	return lastOfThisMonth.Day()
}

func (df *DateFaker) formatTimeWithFormat(t time.Time, format DateFormat) string {
	switch format {
	case FormatUnixTime:
		return strconv.FormatInt(t.Unix(), 10)
	default:
		return t.Format(string(format))
	}
}

// GenerateSpecificFormat generates dates in a specific format
func (df *DateFaker) GenerateSpecificFormat(format DateFormat, count int) []string {
	options := FakeDataOptions{
		Count:        count,
		StartYear:    1990,
		EndYear:      2030,
		IncludeTime:  true,
		FormatMix:    false,
		TargetFormat: format,
	}
	return df.GenerateFakeDates(options)
}

// GenerateMixedFormats generates dates in various formats for comprehensive testing
func (df *DateFaker) GenerateMixedFormats(count int) []string {
	options := FakeDataOptions{
		Count:       count,
		StartYear:   1990,
		EndYear:     2030,
		IncludeTime: true,
		FormatMix:   true,
	}
	return df.GenerateFakeDates(options)
}

// GenerateTestDataSet generates a comprehensive test dataset with known formats
func (df *DateFaker) GenerateTestDataSet() map[DateFormat][]string {
	testData := make(map[DateFormat][]string)
	
	formats := []DateFormat{
		FormatISO8601,
		FormatISO8601Date,
		FormatUSDate,
		FormatUSDateTime,
		FormatEuropeanDate,
		FormatDashDate,
		FormatDotDate,
		FormatUnixTime,
		FormatRFC3339,
		FormatMonthDay,
		FormatShortMonth,
		FormatYearMonth,
		FormatTime24,
		FormatTime12,
	}

	for _, format := range formats {
		testData[format] = df.GenerateSpecificFormat(format, 5)
	}

	return testData
}

// GenerateEdgeCases generates edge case dates for thorough testing
func (df *DateFaker) GenerateEdgeCases() []string {
	edgeCases := []string{
		"2000-02-29", // Leap year
		"1900-02-28", // Not a leap year (century rule)
		"2000-01-01", // Y2K
		"1970-01-01", // Unix epoch
		"2038-01-19", // Unix timestamp limit (32-bit)
		"0001-01-01", // Minimum Go time
		"9999-12-31", // Maximum reasonable year
	}

	// Add some generated edge cases
	now := time.Now()
	edgeCases = append(edgeCases, 
		now.Format("2006-01-02"),                    // Today
		now.AddDate(0, 0, 1).Format("2006-01-02"),  // Tomorrow
		now.AddDate(0, 0, -1).Format("2006-01-02"), // Yesterday
		now.AddDate(1, 0, 0).Format("2006-01-02"),  // Next year
		now.AddDate(-1, 0, 0).Format("2006-01-02"), // Last year
	)

	return edgeCases
}

// GenerateInvalidDates generates invalid date strings for negative testing
func (df *DateFaker) GenerateInvalidDates() []string {
	return []string{
		"2023-13-01", // Invalid month
		"2023-02-30", // Invalid day for February
		"2023-04-31", // Invalid day for April
		"invalid",    // Not a date
		"",          // Empty string
		"2023",      // Incomplete date
		"32/01/2023", // Invalid day
		"01/13/2023", // Invalid month (US format)
		"2023/02/30", // Invalid day
		"-1",        // Negative Unix timestamp
		"abc123",    // Mixed invalid
		"2023-1-1",  // Non-zero padded (might be valid in some contexts)
	}
}

// GetRandomFormat returns a random date format
func (df *DateFaker) GetRandomFormat() DateFormat {
	formats := []DateFormat{
		FormatISO8601,
		FormatISO8601Date,
		FormatUSDate,
		FormatUSDateTime,
		FormatEuropeanDate,
		FormatDashDate,
		FormatDotDate,
		FormatUnixTime,
		FormatRFC3339,
		FormatMonthDay,
		FormatShortMonth,
		FormatYearMonth,
	}
	
	return formats[df.random.Intn(len(formats))]
}

// GenerateWithCallback generates dates and applies a callback function to each
func (df *DateFaker) GenerateWithCallback(count int, callback func(date string, format DateFormat) bool) []string {
	var results []string
	formats := df.getFormatsToUse(FakeDataOptions{FormatMix: true, IncludeTime: true})
	
	for len(results) < count {
		randomTime := df.generateRandomTime(1990, 2030, true)
		selectedFormat := formats[df.random.Intn(len(formats))]
		dateStr := df.formatTimeWithFormat(randomTime, selectedFormat)
		
		if callback(dateStr, selectedFormat) {
			results = append(results, dateStr)
		}
	}
	
	return results
}
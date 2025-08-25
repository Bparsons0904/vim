package utils

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type DateFormat string

const (
	FormatISO8601      DateFormat = "2006-01-02T15:04:05Z07:00"
	FormatISO8601Date  DateFormat = "2006-01-02"
	FormatUSDate       DateFormat = "01/02/2006"
	FormatUSDateTime   DateFormat = "01/02/2006 15:04:05"
	FormatEuropeanDate DateFormat = "02/01/2006"
	FormatDashDate     DateFormat = "02-01-2006"
	FormatDotDate      DateFormat = "02.01.2006"
	FormatUnixTime     DateFormat = "unix"
	FormatRFC3339      DateFormat = "2006-01-02T15:04:05Z"
	FormatRFC822       DateFormat = "02 Jan 06 15:04 MST"
	FormatRFC850       DateFormat = "Monday, 02-Jan-06 15:04:05 MST"
	FormatMonthDay     DateFormat = "January 2, 2006"
	FormatShortMonth   DateFormat = "Jan 2, 2006"
	FormatYearMonth    DateFormat = "2006-01"
	FormatTime24       DateFormat = "15:04:05"
	FormatTime12       DateFormat = "3:04:05 PM"
)

type DateValidator struct {
	supportedFormats []DateFormat
	standardFormat   DateFormat
}

type ValidationResult struct {
	IsValid        bool
	DetectedFormat DateFormat
	ParsedTime     time.Time
	StandardFormat string
	OriginalValue  string
}

func NewDateValidator() *DateValidator {
	return &DateValidator{
		supportedFormats: []DateFormat{
			FormatISO8601,
			FormatISO8601Date,
			FormatUSDate,
			FormatUSDateTime,
			FormatEuropeanDate,
			FormatDashDate,
			FormatDotDate,
			FormatUnixTime,
			FormatRFC3339,
			FormatRFC822,
			FormatRFC850,
			FormatMonthDay,
			FormatShortMonth,
			FormatYearMonth,
			FormatTime24,
			FormatTime12,
		},
		standardFormat: FormatISO8601,
	}
}

func (dv *DateValidator) SetStandardFormat(format DateFormat) {
	dv.standardFormat = format
}

func (dv *DateValidator) ValidateAndConvert(input string) ValidationResult {
	result := ValidationResult{
		IsValid:       false,
		OriginalValue: input,
	}

	input = strings.TrimSpace(input)
	if input == "" {
		return result
	}

	// Try Unix timestamp first (integer)
	if unixTime, err := strconv.ParseInt(input, 10, 64); err == nil {
		if unixTime > 0 && unixTime < 4102444800 { // Valid range: 1970-2100
			parsedTime := time.Unix(unixTime, 0).UTC()
			result.IsValid = true
			result.DetectedFormat = FormatUnixTime
			result.ParsedTime = parsedTime
			result.StandardFormat = parsedTime.Format(string(dv.standardFormat))
			return result
		}
	}

	// Try each supported format
	for _, format := range dv.supportedFormats {
		if format == FormatUnixTime {
			continue // Already handled above
		}

		if parsedTime, err := time.Parse(string(format), input); err == nil {
			// Additional validation for ambiguous formats
			if dv.isValidForFormat(input, format) {
				result.IsValid = true
				result.DetectedFormat = format
				result.ParsedTime = parsedTime
				result.StandardFormat = parsedTime.Format(string(dv.standardFormat))
				return result
			}
		}
	}

	// Try some flexible parsing patterns
	if parsedTime, format := dv.tryFlexibleParsing(input); !parsedTime.IsZero() {
		result.IsValid = true
		result.DetectedFormat = format
		result.ParsedTime = parsedTime
		result.StandardFormat = parsedTime.Format(string(dv.standardFormat))
		return result
	}

	return result
}

func (dv *DateValidator) isValidForFormat(input string, format DateFormat) bool {
	switch format {
	case FormatUSDate, FormatUSDateTime:
		// MM/DD/YYYY - month should be 1-12, day should be 1-31
		return dv.validateUSDateFormat(input)
	case FormatEuropeanDate:
		// DD/MM/YYYY - day should be 1-31, month should be 1-12
		return dv.validateEuropeanDateFormat(input)
	case FormatYearMonth:
		// YYYY-MM - basic format validation
		pattern := regexp.MustCompile(`^\d{4}-\d{2}$`)
		return pattern.MatchString(input)
	default:
		return true
	}
}

func (dv *DateValidator) validateUSDateFormat(input string) bool {
	// MM/DD/YYYY format validation
	pattern := regexp.MustCompile(`^(\d{1,2})/(\d{1,2})/(\d{4})`)
	matches := pattern.FindStringSubmatch(input)
	if len(matches) < 4 {
		return false
	}

	month, _ := strconv.Atoi(matches[1])
	day, _ := strconv.Atoi(matches[2])
	
	return month >= 1 && month <= 12 && day >= 1 && day <= 31
}

func (dv *DateValidator) validateEuropeanDateFormat(input string) bool {
	// DD/MM/YYYY format validation
	pattern := regexp.MustCompile(`^(\d{1,2})/(\d{1,2})/(\d{4})`)
	matches := pattern.FindStringSubmatch(input)
	if len(matches) < 4 {
		return false
	}

	day, _ := strconv.Atoi(matches[1])
	month, _ := strconv.Atoi(matches[2])
	
	return month >= 1 && month <= 12 && day >= 1 && day <= 31
}

func (dv *DateValidator) tryFlexibleParsing(input string) (time.Time, DateFormat) {
	// Try common variations and flexible patterns
	flexibleFormats := []string{
		"2006-01-02 15:04:05",
		"2006/01/02 15:04:05",
		"02-01-2006 15:04:05",
		"Jan 02, 2006 15:04:05",
		"January 02, 2006 15:04:05",
		"2006-01-02T15:04:05",
		"01-02-2006",
		"2006/01/02",
		"02/01/2006 15:04",
		"01/02/2006 15:04",
		// 2-digit year formats
		"1/2/06",      // M/D/YY
		"01/02/06",    // MM/DD/YY
		"1/2/2006",    // M/D/YYYY
		"1-2-06",      // M-D-YY
		"01-02-06",    // MM-DD-YY
		"1-2-2006",    // M-D-YYYY
		"2-1-06",      // D-M-YY (European style)
		"02-01-06",    // DD-MM-YY (European style)
		"2-1-2006",    // D-M-YYYY (European style)
		"2/1/06",      // D/M/YY (European style)
		"02/01/06",    // DD/MM/YY (European style)  
		"2/1/2006",    // D/M/YYYY (European style)
		"06-1-2",      // YY-M-D
		"06-01-02",    // YY-MM-DD
		"2006-1-2",    // YYYY-M-D
	}

	for _, format := range flexibleFormats {
		if parsedTime, err := time.Parse(format, input); err == nil {
			// For 2-digit years, assume years 00-29 are 2000-2029, 30-99 are 1930-1999
			if strings.Contains(format, "/06") || strings.Contains(format, "-06") {
				year := parsedTime.Year()
				if year < 30 {
					// Convert to 20XX
					parsedTime = parsedTime.AddDate(2000-year, 0, 0)
				} else if year < 100 {
					// Convert to 19XX
					parsedTime = parsedTime.AddDate(1900-year+100, 0, 0)
				}
			}
			return parsedTime, DateFormat(format)
		}
	}

	// Try manual parsing for very flexible formats
	return dv.tryManualParsing(input)
}

func (dv *DateValidator) tryManualParsing(input string) (time.Time, DateFormat) {
	// Handle common separators and short years
	separators := []string{"/", "-", "."}
	
	for _, sep := range separators {
		if strings.Contains(input, sep) {
			parts := strings.Split(input, sep)
			if len(parts) == 3 {
				// Try to parse as numbers
				nums := make([]int, 3)
				valid := true
				for i, part := range parts {
					if num, err := strconv.Atoi(strings.TrimSpace(part)); err == nil {
						nums[i] = num
					} else {
						valid = false
						break
					}
				}
				
				if valid {
					// Try different combinations
					combinations := [][]int{
						{0, 1, 2}, // M/D/Y or MM/DD/YY
						{1, 0, 2}, // D/M/Y or DD/MM/YY  
						{2, 0, 1}, // Y/M/D or YY/MM/DD
						{2, 1, 0}, // Y/D/M or YY/DD/MM
					}
					
					for _, combo := range combinations {
						month := nums[combo[0]]
						day := nums[combo[1]]
						year := nums[combo[2]]
						
						// Handle 2-digit years
						if year < 100 {
							if year <= 29 {
								year += 2000
							} else {
								year += 1900
							}
						}
						
						// Basic validation
						if month >= 1 && month <= 12 && day >= 1 && day <= 31 && year >= 1900 && year <= 2100 {
							if parsedTime := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC); !parsedTime.IsZero() {
								return parsedTime, DateFormat("flexible_" + sep + "_format")
							}
						}
					}
				}
			}
		}
	}
	
	return time.Time{}, ""
}

func (dv *DateValidator) GetSupportedFormats() []DateFormat {
	return dv.supportedFormats
}

func (dv *DateValidator) AddCustomFormat(format DateFormat) {
	dv.supportedFormats = append(dv.supportedFormats, format)
}

// ValidateBatch validates multiple date strings and returns results
func (dv *DateValidator) ValidateBatch(inputs []string) []ValidationResult {
	results := make([]ValidationResult, len(inputs))
	for i, input := range inputs {
		results[i] = dv.ValidateAndConvert(input)
	}
	return results
}

// GetFormatExample returns an example of the given format
func GetFormatExample(format DateFormat) string {
	now := time.Now()
	
	switch format {
	case FormatUnixTime:
		return fmt.Sprintf("%d", now.Unix())
	default:
		return now.Format(string(format))
	}
}
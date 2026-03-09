package dpmaconnect

import (
	"fmt"
	"time"

	"github.com/patent-dev/dpma-connect-plus/query"
)

// Period constants for register extract queries
const (
	PeriodDaily   = "daily"
	PeriodWeekly  = "weekly"
	PeriodMonthly = "monthly"
	PeriodYearly  = "yearly"
)

// Service name constants for GetVersion
const (
	ServicePatent    = "DPMAregisterPatService"
	ServiceDesign    = "DPMAregisterGsmService"
	ServiceTrademark = "DPMAregisterMarkeService"
)

// FormatPublicationWeek formats year and week into YYYYWW format.
// Returns an error if year < 1 or week is outside [1, 53].
// Example: FormatPublicationWeek(2024, 45) returns "202445"
func FormatPublicationWeek(year, week int) (string, error) {
	if year < 1 {
		return "", fmt.Errorf("invalid year %d: must be positive", year)
	}
	if week < 1 || week > 53 {
		return "", fmt.Errorf("invalid week %d: must be between 1 and 53", week)
	}
	return fmt.Sprintf("%04d%02d", year, week), nil
}

// FormatDate formats a time.Time into YYYY-MM-DD format for register extract queries.
// The date is formatted in the input's location (no timezone conversion).
// Example: FormatDate(time.Date(2024, 10, 23, 0, 0, 0, 0, time.UTC)) returns "2024-10-23"
func FormatDate(date time.Time) string {
	return date.Format("2006-01-02")
}

// ValidatePatentQuery parses and validates a query against patent field codes.
// Returns nil if valid, or an error describing the validation failure.
func ValidatePatentQuery(q string) error {
	parsed, err := query.ParseQuery(q, query.ServicePatent)
	if err != nil {
		return err
	}
	return parsed.Validate()
}

// ValidateDesignQuery parses and validates a query against design field codes.
// Returns nil if valid, or an error describing the validation failure.
func ValidateDesignQuery(q string) error {
	parsed, err := query.ParseQuery(q, query.ServiceDesign)
	if err != nil {
		return err
	}
	return parsed.Validate()
}

// ValidateTrademarkQuery parses and validates a query against trademark field codes.
// Returns nil if valid, or an error describing the validation failure.
func ValidateTrademarkQuery(q string) error {
	parsed, err := query.ParseQuery(q, query.ServiceTrademark)
	if err != nil {
		return err
	}
	return parsed.Validate()
}

// ParsePublicationWeek parses a publication week string (YYYYWW) into year and week integers
// Returns an error if the format is invalid
func ParsePublicationWeek(pubWeek string) (year, week int, err error) {
	if len(pubWeek) != 6 {
		return 0, 0, fmt.Errorf("invalid publication week format: expected YYYYWW, got %s", pubWeek)
	}
	_, err = fmt.Sscanf(pubWeek, "%04d%02d", &year, &week)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to parse publication week %s: %w", pubWeek, err)
	}
	if year < 1 {
		return 0, 0, fmt.Errorf("invalid year %d: must be positive", year)
	}
	if week < 1 || week > 53 {
		return 0, 0, fmt.Errorf("invalid week number %d: must be between 1 and 53", week)
	}
	return year, week, nil
}

// ValidatePeriod checks that a period string is one of the valid values.
func ValidatePeriod(period string) error {
	switch period {
	case PeriodDaily, PeriodWeekly, PeriodMonthly, PeriodYearly:
		return nil
	default:
		return fmt.Errorf("invalid period %q: must be one of %q, %q, %q, %q",
			period, PeriodDaily, PeriodWeekly, PeriodMonthly, PeriodYearly)
	}
}

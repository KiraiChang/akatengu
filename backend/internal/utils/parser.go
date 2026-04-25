package utils

import "time"

const DefaultDateFormat = "2006-01-02"
const DefaultDateTimeFormat = "2006-01-02 15:04:05"

func ConvertSQLiteDate(time2 time.Time) string {
	return time2.Format(DefaultDateFormat)
}

func ConvertSQLiteTime(time2 time.Time) string {
	return time2.Format(DefaultDateTimeFormat)
}

func ParseSQLiteTime(s string) (time.Time, error) {
	if t, err := time.Parse(time.RFC3339Nano, s); err == nil {
		return t, nil
	}
	return time.ParseInLocation(
		DefaultDateTimeFormat,
		s,
		time.UTC,
	)
}

func ParseSQLiteDate(s string) (time.Time, error) {
	if t, err := time.Parse(DefaultDateFormat, s); err == nil {
		return t, nil
	}
	return time.ParseInLocation(
		DefaultDateFormat,
		s,
		time.UTC,
	)
}

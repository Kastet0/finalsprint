package nextdate

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

var (
	ErrInvalidDateFormat   = errors.New("invalid dstart date format")
	ErrEmptyRepeatRule     = errors.New("empty repeat rule")
	ErrInvalidRepeatFormat = errors.New("invalid repeat rule format")
	ErrUnsupportedRepeat   = errors.New("unsupported repeat rule")
	ErrIntervalTooLarge    = errors.New("day interval exceeds maximum of 400")
	ErrInvalidWeekday      = errors.New("invalid weekday number, must be 1-7")
	ErrInvalidDayOfMonth   = errors.New("invalid day of month number, must be 1-31, -1, or -2")
	ErrInvalidMonth        = errors.New("invalid month number, must be 1-12")
)

const (
	dateFormat     = "20060102"
	maxDayInterval = 400
)

func AfterNow(date, now time.Time) bool {
	dateDay := date.Format(dateFormat)
	nowDay := now.Format(dateFormat)

	return dateDay > nowDay
}

func parseRepeatIntervals(s string, min, max int) (map[int]bool, error) {
	parts := strings.Split(s, ",")
	intervals := make(map[int]bool)
	for _, p := range parts {
		num, err := strconv.Atoi(p)
		if err != nil {
			return nil, fmt.Errorf("invalid number in list: %s", p)
		}

		isDayOfMonth := (min == 1 && max == 31)
		if isDayOfMonth && (num == -1 || num == -2) {
			intervals[num] = true
			continue
		}

		if num < min || num > max {
			return nil, fmt.Errorf("number %d out of range %d-%d", num, min, max)
		}

		intervals[num] = true
	}
	return intervals, nil
}

func NextDate(now time.Time, dstart string, repeat string) (string, error) {
	repeat = strings.TrimSpace(repeat)
	if repeat == "" {
		return "", ErrEmptyRepeatRule
	}

	date, err := time.Parse(dateFormat, dstart)
	if err != nil {
		return "", ErrInvalidDateFormat
	}

	parts := strings.Fields(repeat)
	if len(parts) == 0 {
		return "", ErrInvalidRepeatFormat
	}

	ruleType := parts[0]

	switch ruleType {
	case "y":

		if AfterNow(date, now) {
			date = date.AddDate(1, 0, 0)
			return date.Format(dateFormat), nil
		}
		for !AfterNow(date, now) {
			date = date.AddDate(1, 0, 0)
		}
		return date.Format(dateFormat), nil

	case "d":
		if len(parts) != 2 {
			return "", ErrInvalidRepeatFormat
		}
		interval, err := strconv.Atoi(parts[1])
		if err != nil || interval <= 0 {
			return "", ErrInvalidRepeatFormat
		}
		if interval > maxDayInterval {
			return "", ErrIntervalTooLarge
		}

		if AfterNow(date, now) {
			date = date.AddDate(0, 0, interval)
			return date.Format(dateFormat), nil
		}

		for !AfterNow(date, now) {
			date = date.AddDate(0, 0, interval)
		}

		return date.Format(dateFormat), nil

	case "w":
		if len(parts) != 2 {
			return "", ErrInvalidRepeatFormat
		}

		weekdaysMap, err := parseRepeatIntervals(parts[1], 1, 7)
		if err != nil {
			if strings.Contains(err.Error(), "out of range 1-7") {
				return "", ErrInvalidWeekday
			}
			return "", ErrInvalidRepeatFormat
		}

		for i := 0; i < 366*2; i++ {
			currentWeekday := (int(date.Weekday())+6)%7 + 1

			if weekdaysMap[currentWeekday] && AfterNow(date, now) {
				return date.Format(dateFormat), nil
			}

			date = date.AddDate(0, 0, 1)
		}

		return "", errors.New("failed to find next weekday")

	case "m":
		if len(parts) < 2 || len(parts) > 3 {
			return "", ErrInvalidRepeatFormat
		}

		daysMap, err := parseRepeatIntervals(parts[1], 1, 31)
		if err != nil {
			if strings.Contains(err.Error(), "out of range 1-31") {
				return "", ErrInvalidDayOfMonth
			}
			return "", ErrInvalidRepeatFormat
		}

		monthsMap := make(map[int]bool)
		if len(parts) == 3 {
			monthsMap, err = parseRepeatIntervals(parts[2], 1, 12)
			if err != nil {
				if strings.Contains(err.Error(), "out of range 1-12") {
					return "", ErrInvalidMonth
				}
				return "", ErrInvalidRepeatFormat
			}
		}

		for i := 0; i < 366*2; i++ {
			monthNum := int(date.Month())
			if len(monthsMap) > 0 && !monthsMap[monthNum] {
				date = date.AddDate(0, 0, 1)
				continue
			}

			firstOfNextMonth := time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, date.Location()).AddDate(0, 1, 0)
			lastDayOfMonth := firstOfNextMonth.AddDate(0, 0, -1)

			dayNum := date.Day()
			isMatch := false

			if daysMap[-1] {
				if dayNum == lastDayOfMonth.Day() {
					isMatch = true
				}
			}
			if daysMap[-2] {
				if dayNum == lastDayOfMonth.Day()-1 {
					isMatch = true
				}
			}

			if daysMap[dayNum] {
				if dayNum <= lastDayOfMonth.Day() {
					isMatch = true
				}
			}

			if isMatch && AfterNow(date, now) {
				return date.Format(dateFormat), nil
			}

			date = date.AddDate(0, 0, 1)
		}

		return "", errors.New("failed to find next monthly date")

	default:
		return "", ErrUnsupportedRepeat
	}
}

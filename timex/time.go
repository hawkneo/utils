package timex

import "time"

func GetWeekStartTime(t time.Time, firstWeekDay time.Weekday) time.Time {
	weekday := t.Weekday()
	daysAgo := int(weekday - firstWeekDay)
	if daysAgo < 0 {
		daysAgo += 7
	}
	year, month, day := t.AddDate(0, 0, -daysAgo).Date()
	return time.Date(year, month, day, 0, 0, 0, 0, t.Location())
}

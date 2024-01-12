package now

import (
	"time"
)

var (
	now       = time.Now
	date      = time.Date
	StartTest time.Time
)

func Now() time.Time {
	return now().UTC()
}

func Date(year int, month time.Month, day int, hour int, min int, sec int, nsec int, loc *time.Location) time.Time {
	return date(year, month, day, hour, min, sec, nsec, loc).UTC()
}

func SetNow(f func() time.Time) {
	now = f
}
package timeutil

import (
	"time"

	"github.com/dromara/carbon/v2"
)

func InitCarbon() {
	carbon.SetDefault(carbon.Default{
		Layout:       carbon.DateTimeLayout,
		Timezone:     carbon.UTC,
		Locale:       "en",
		WeekStartsAt: carbon.Monday,
		WeekendDays:  []carbon.Weekday{carbon.Saturday, carbon.Sunday},
	})
}

func DateID(t time.Time) string {
	return t.Format("20060102")
}

package timeutil

import "time"

var MaxTime = time.Date(9999, 12, 31, 23, 59, 59, 0, time.Local)
var MinTime = time.Date(0, 1, 1, 0, 0, 0, 0, time.Local)

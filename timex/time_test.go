package timex

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestGetWeekStartTime(t *testing.T) {
	time.Local = time.UTC

	tt, err := time.Parse(time.RFC3339, "2024-04-07T15:04:05Z")
	require.NoError(t, err)

	weekStartTime := GetWeekStartTime(tt, time.Monday)
	println(weekStartTime.Format(time.RFC3339))
	require.Equal(t, 2024, weekStartTime.Year())
	require.Equal(t, 4, int(weekStartTime.Month()))
	require.Equal(t, 1, weekStartTime.Day())
	require.Equal(t, 0, weekStartTime.Hour())
	require.Equal(t, 0, weekStartTime.Minute())
	require.Equal(t, 0, weekStartTime.Second())
	require.Equal(t, 0, weekStartTime.Nanosecond())

	weekStartTime = GetWeekStartTime(tt, time.Sunday)
	println(weekStartTime.Format(time.RFC3339))
	require.Equal(t, 2024, weekStartTime.Year())
	require.Equal(t, 4, int(weekStartTime.Month()))
	require.Equal(t, 7, weekStartTime.Day())
	require.Equal(t, 0, weekStartTime.Hour())
	require.Equal(t, 0, weekStartTime.Minute())
	require.Equal(t, 0, weekStartTime.Second())
	require.Equal(t, 0, weekStartTime.Nanosecond())
}

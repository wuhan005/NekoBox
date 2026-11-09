package response

import (
	"encoding/json"
	"testing"
	"time"
)

func TestTime_MarshalJSON(t *testing.T) {
	sh, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		t.Fatalf("load Asia/Shanghai: %v", err)
	}
	ny, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatalf("load America/New_York: %v", err)
	}
	// fixedUTC emulates pgx's behaviour for `timestamp with time zone`
	// columns under TZ=UTC: the location prints as "UTC" but is *not*
	// the package-level time.UTC sentinel.
	fixedUTC := time.FixedZone("UTC", 0)

	tests := []struct {
		name string
		in   time.Time
		want string
	}{
		{
			name: "naive UTC from `timestamp` column (wall-clock is actually Asia/Shanghai)",
			in:   time.Date(2026, 5, 2, 2, 9, 23, 243_000_000, time.UTC),
			want: `"2026-05-02T02:09:23.243+08:00"`,
		},
		{
			name: "tz-aware UTC instant from `timestamptz` column under TZ=UTC container",
			in:   time.Date(2026, 5, 1, 18, 9, 23, 243_000_000, fixedUTC),
			want: `"2026-05-02T02:09:23.243+08:00"`,
		},
		{
			name: "tz-aware Asia/Shanghai value from `timestamptz` column under TZ=Asia/Shanghai container",
			in:   time.Date(2026, 5, 2, 2, 9, 23, 243_000_000, sh),
			want: `"2026-05-02T02:09:23.243+08:00"`,
		},
		{
			name: "any other tz is converted to Asia/Shanghai (same instant)",
			in:   time.Date(2026, 5, 1, 14, 9, 23, 0, ny),
			want: `"2026-05-02T02:09:23+08:00"`,
		},
		{
			name: "zero value",
			in:   time.Time{},
			want: `"0001-01-01T00:00:00+08:00"`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := json.Marshal(Time(tc.in))
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			if string(got) != tc.want {
				t.Errorf("got %s, want %s", got, tc.want)
			}
		})
	}
}

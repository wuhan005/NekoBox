package response

import (
	"time"
)

// shanghaiLoc holds the Asia/Shanghai location, falling back to a fixed
// +08:00 zone if the system tzdata cannot be loaded (e.g. in a minimal
// container without tzdata).
var shanghaiLoc = func() *time.Location {
	if loc, err := time.LoadLocation("Asia/Shanghai"); err == nil {
		return loc
	}
	return time.FixedZone("CST", 8*60*60)
}()

// Time wraps time.Time so JSON-marshalled values always use the
// Asia/Shanghai (+08:00) offset, regardless of the runtime TZ or how the
// underlying database column is typed.
//
// Why this exists:
//
// Production stores created_at / updated_at as `timestamp without time
// zone`, with the wall-clock written in Asia/Shanghai. The pgx driver,
// having no tz info for those columns, hands us a time.Time whose
// Location is the package-level time.UTC sentinel. Marshalling that
// straight to JSON yields a "Z" suffix; dayjs in the browser then
// re-shifts by +8 hours and renders 10:09 instead of 02:09 — the bug
// this type fixes.
//
// Locally (and in any future migration) the column is
// `timestamp with time zone`. In that case pgx returns a time.Time
// whose Location is a non-sentinel zone (Asia/Shanghai when the
// container runs under TZ=Asia/Shanghai, or a freshly-built UTC zone
// when the container runs under TZ=UTC). Either way it is *not* the
// time.UTC sentinel.
//
// We exploit that subtle but reliable distinction:
//   - Location == time.UTC sentinel  →  treat as "naive Shanghai
//     wall-clock" and re-tag it with Asia/Shanghai (no offset shift).
//   - Anything else                  →  it already encodes a real
//     instant, so just convert it into Asia/Shanghai before formatting.
//
// Both branches produce "+08:00" output, so the API contract is stable
// no matter which database schema or container TZ is used.
type Time time.Time

func (t Time) MarshalJSON() ([]byte, error) {
	tt := time.Time(t)
	if tt.IsZero() {
		return []byte(`"0001-01-01T00:00:00+08:00"`), nil
	}

	var out time.Time
	if tt.Location() == time.UTC {
		out = time.Date(
			tt.Year(), tt.Month(), tt.Day(),
			tt.Hour(), tt.Minute(), tt.Second(), tt.Nanosecond(),
			shanghaiLoc,
		)
	} else {
		out = tt.In(shanghaiLoc)
	}

	return []byte(`"` + out.Format(time.RFC3339Nano) + `"`), nil
}

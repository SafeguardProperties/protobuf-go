package timestamppb

import (
	"database/sql/driver"
	"fmt"
	"log"
	"time"
	"github.com/pkg/errors"
)

var americaNewYork *time.Location

func (m *Timestamp) IsZero() bool {
	if m == nil {
		return true
	}
	if m.Nanos == 0 && m.Seconds == 0 {
		return true
	}
	return false
}

func (m *Timestamp) NilIfZero() *Timestamp {
	if m.IsZero() {
		return nil
	}
	return m
}

// Conform to the Scanner interface for database/sql
func (m *Timestamp) Scan(value interface{}) error {

	// We want a time.Time.   We assume that if the driver gives us a time.Time that it is in the appropriate timezone.
	t, ok := value.(time.Time)
	if ok {
		return m.StampFromTime(t)
	}

	// FIXME -  Ok, tis is a horrible hack.
	// Safeguard does not store date/time values with timezone.   This means that all dates/times have been stored in America/New_York.
	// This means we must interpret them as America/New_York

	// Lets try the strings.
	tString, ok := value.(string)
	if ok {
		// Try RFC3339?
		t, err := time.ParseInLocation(time.RFC3339, tString, getAmericaNewYork())
		if err == nil {
			// Success!
			return m.StampFromTime(t)
		}

		// How about RFC3339Nano?
		t, err = time.ParseInLocation(time.RFC3339Nano, tString, getAmericaNewYork())
		if err == nil {
			return m.StampFromTime(t)
		}

		// How about an eastern standard doohickey.
		t, err = time.ParseInLocation("2006-01-02 15:04:05", tString, getAmericaNewYork())
		if err == nil {
			return m.StampFromTime(t)
		}

		// Last try, something simple.
		t, err = time.ParseInLocation("2006-01-02", tString, getAmericaNewYork())
		if err == nil {
			return m.StampFromTime(t)
		}

		return errors.Errorf("Unable to parse time, value not understood: %s", value)

	}

	return errors.Errorf("Unexpected type: %T", value)
	//return errors.New("incompatible type passed, expected time.Time, or string.")
}

func (m *Timestamp) StampFromTime(t time.Time) error {
	seconds := t.UTC().Unix()
	nanos := int32(t.Sub(time.Unix(seconds, 0)))
	m.Seconds = seconds
	m.Nanos = nanos
	return m.validateTimestamp()
}

// Value satisfies the valuer interface for database/sql.  Copied from ptypes.
func (m *Timestamp) Value() (driver.Value, error) {
	var t time.Time
	if m == nil {
		t = time.Unix(0, 0).UTC() // treat nil like the empty Timestamp
	} else {
		t = time.Unix(m.Seconds, int64(m.Nanos)).UTC()
	}
	return t, m.validateTimestamp()
}

const (
	// Seconds field of the earliest valid Timestamp.
	// This is time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC).Unix().
	minValidSeconds = -62135596800
	// Seconds field just after the latest valid Timestamp.
	// This is time.Date(10000, 1, 1, 0, 0, 0, 0, time.UTC).Unix().
	maxValidSeconds = 253402300800
)

// validateTimestamp determines whether a Timestamp is valid.
// A valid timestamp represents a time in the range
// [0001-01-01, 10000-01-01) and has a Nanos field
// in the range [0, 1e9).
//
// If the Timestamp is valid, validateTimestamp returns nil.
// Otherwise, it returns an error that describes
// the problem.
//
// Every valid Timestamp can be represented by a time.Time, but the converse is not true.
func (m *Timestamp) validateTimestamp() error {
	if m.Seconds < minValidSeconds {
		return fmt.Errorf("timestamp: %v before 0001-01-01", m)
	}
	if m.Seconds >= maxValidSeconds {
		return fmt.Errorf("timestamp: %v after 10000-01-01", m)
	}
	if m.Nanos < 0 || m.Nanos >= 1e9 {
		return fmt.Errorf("timestamp: %v: nanos not in range [0, 1e9)", m)
	}
	return nil
}

func getAmericaNewYork() *time.Location {
	if americaNewYork == nil {
		var err error
		americaNewYork, err = time.LoadLocation("America/New_York")
		if err != nil {
			log.Printf("Error loading timezone: %v", err)
		}
	}
	return americaNewYork
}

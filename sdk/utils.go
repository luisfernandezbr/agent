package sdk

import (
	"encoding/json"
	"reflect"
	"time"

	"github.com/pinpt/go-common/v10/datetime"
	"github.com/pinpt/go-common/v10/hash"
	"github.com/pinpt/go-common/v10/log"
	ps "github.com/pinpt/go-common/v10/strings"
)

// StringPointer return a string pointer from a value
func StringPointer(val interface{}) *string {
	return ps.Pointer(val)
}

// Hash will convert all objects to a string and return a SHA256 of the concatenated values.
// Uses xxhash to calculate a faster hash value that is not cryptographically secure but is OK since
// we use hashing mainfully for generating consistent key values or equality checks.
func Hash(val ...interface{}) string {
	return hash.Values(val...)
}

// Logger is a logger interface
type Logger = log.Logger

// LogDebug will log an debug level log to logger
func LogDebug(logger Logger, msg string, kv ...interface{}) error {
	return log.Debug(logger, msg, kv...)
}

// LogInfo will log an info level log to logger
func LogInfo(logger Logger, msg string, kv ...interface{}) error {
	return log.Info(logger, msg, kv...)
}

// LogWarn will log an warning level log to logger
func LogWarn(logger Logger, msg string, kv ...interface{}) error {
	return log.Warn(logger, msg, kv...)
}

// LogError will log an error level log to logger
func LogError(logger Logger, msg string, kv ...interface{}) error {
	return log.Error(logger, msg, kv...)
}

// LogFatal will log an fatal level log to logger
func LogFatal(logger Logger, msg string, kv ...interface{}) {
	log.Fatal(logger, msg, kv...)
}

// LogWith will return a new logger adding keyvalues to all logs
func LogWith(logger Logger, keyvals ...interface{}) Logger {
	return log.With(logger, keyvals...)
}

// Date is a date structure with epoch, offset and RFC3339 timestamp format
type Date = datetime.Date

// NewDateWithTime will return a Date from a time
func NewDateWithTime(tv time.Time) (*Date, error) {
	return datetime.NewDateWithTime(tv)
}

// TimeToEpoch returns an epoch time from a time.Time
func TimeToEpoch(tv time.Time) int64 {
	return datetime.TimeToEpoch(tv)
}

// EpochNow will return the current time as an epoch value
func EpochNow() int64 {
	return datetime.EpochNow()
}

// ConvertTimeToDateModel will fill the dataModel struct with the current time
func ConvertTimeToDateModel(ts time.Time, dateModel interface{}) {
	if ts.IsZero() {
		return
	}

	date, err := datetime.NewDateWithTime(ts)
	if err != nil {
		// this will never happen NewDateWithTime, always returns nil
		panic(err)
	}

	t := reflect.ValueOf(dateModel).Elem()
	t.FieldByName("Rfc3339").Set(reflect.ValueOf(date.Rfc3339))
	t.FieldByName("Epoch").Set(reflect.ValueOf(date.Epoch))
	t.FieldByName("Offset").Set(reflect.ValueOf(date.Offset))
}

// MapToStruct will unmarshal a map into the target
func MapToStruct(m map[string]interface{}, target interface{}) error {
	b, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, target)
}

// JoinURL creates a URL string from parts
func JoinURL(elem ...string) string {
	return ps.JoinURL(elem...)
}

// Keys will return the keys that are true from a map[string]bool
func Keys(m map[string]bool) []string {
	a := make([]string, 0)
	for k, v := range m {
		if v {
			a = append(a, k)
		}
	}
	return a
}

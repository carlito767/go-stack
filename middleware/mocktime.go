package middleware

import "time"

type TimeProvider interface {
	Now() time.Time
	Since(time.Time) time.Duration
}

//
// Mock Time Provider
//

type MockTimeProvider struct{}

func (_ MockTimeProvider) Now() time.Time {
	return time.Now()
}

func (_ MockTimeProvider) Since(t time.Time) time.Duration {
	return time.Since(t)
}

//
// Fake Time Provider
//

type FakeTimeProvider struct{}

func (_ FakeTimeProvider) Now() time.Time {
	return time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
}

func (_ FakeTimeProvider) Since(t time.Time) time.Duration {
	return 42 * time.Second
}

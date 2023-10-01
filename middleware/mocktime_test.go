package middleware_test

import (
	"testing"
	"time"

	"github.com/carlito767/go-stack/middleware"
)

func TestMockTime(t *testing.T) {
	mtp := middleware.MockTimeProvider{}
	now := mtp.Now()
	mtp.Since(now)

	ftp := middleware.FakeTimeProvider{}
	now = ftp.Now()
	expectedNow := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
	if now != expectedNow {
		t.Errorf("now expected:%v, got:%v", expectedNow, now)
	}
	since := ftp.Since(now)
	expectedSince := 42 * time.Second
	if since != expectedSince {
		t.Errorf("since expected:%v, got:%v", expectedSince, since)
	}
}

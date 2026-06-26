package wails_service

import "errors"

// ErrNotFound はリソース未検出。
var ErrNotFound = errors.New("not found")

// ErrUpdaterUnavailable は updater が初期化されていない。
var ErrUpdaterUnavailable = errors.New("updater unavailable")

package wails_service

import "time"

// debugLogSlowThreshold は gowrap debug ログの end 出力閾値。
// これ未満の呼び出しは end ログを出さない。
const debugLogSlowThreshold = 200 * time.Millisecond

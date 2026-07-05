package sqlitedsn

// DSN は GORM 用 SQLite DSN を返す。
// foreign_keys・WAL・synchronous(NORMAL)・busy_timeout を接続ごとに適用する。
func DSN(dbPath string) string {
	return dbPath + "?_pragma=foreign_keys(1)" +
		"&_pragma=journal_mode(WAL)" +
		"&_pragma=synchronous(NORMAL)" +
		"&_pragma=busy_timeout(5000)"
}

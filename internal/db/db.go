// Package db 负责 SQLite 连接装配与建表。
//
// 说明：首轮基线使用单文件 schema.sql（通过 go:embed 打包），
// 这是 B3 坏味道 2「Schema 单文件混写」的刻意留点。
package db

import (
	"database/sql"
	"embed"
	"fmt"

	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var schemaFS embed.FS

// NewDB 打开 SQLite 连接并执行 schema.sql 建表。
// path 传 ":memory:" 可得到隔离的内存库（测试用），传文件路径则为磁盘库。
func NewDB(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite %s: %w", path, err)
	}
	// 强制外键生效（否则 ON DELETE CASCADE 不工作）
	if _, err := db.Exec("PRAGMA foreign_keys = ON;"); err != nil {
		db.Close()
		return nil, fmt.Errorf("enable foreign_keys pragma: %w", err)
	}
	// 执行 schema.sql
	schemaBytes, _ := schemaFS.ReadFile("schema.sql")
	if _, err := db.Exec(string(schemaBytes)); err != nil {
		db.Close()
		return nil, fmt.Errorf("exec schema.sql: %w", err)
	}
	return db, nil
}
package main

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

// ZhiPath 数据库迁移 runner。
// 直接复用 DATABASE_URL（go-sql-driver DSN 格式）连库；
// 维护 schema_migrations 表，按版本号顺序执行 migrations/*.up.sql。
//
// 用法：
//   go run ./cmd/migrate up [dir]
//   go run ./cmd/migrate down [dir]
//
// dir 默认取运行目录下的 migrations/。

type migration struct {
	version int
	name    string
	up      string
	down    string
}

var fileRe = regexp.MustCompile(`^(\d+)_(.+)\.(up|down)\.sql$`)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: migrate <up|down> [dir]")
		os.Exit(1)
	}
	cmd := os.Args[1]
	dir := "migrations"
	if len(os.Args) > 2 {
		dir = os.Args[2]
	}

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = buildDSN()
	}
	if !strings.Contains(dsn, "multiStatements") {
		sep := "?"
		if strings.Contains(dsn, "?") {
			sep = "&"
		}
		dsn += sep + "multiStatements=true"
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		fatal("open db: %v", err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		fatal("ping db: %v", err)
	}

	ensureSchemaTable(db)
	migs := loadMigrations(dir)
	sort.Slice(migs, func(i, j int) bool { return migs[i].version < migs[j].version })

	switch cmd {
	case "up":
		applied := appliedVersions(db)
		n := 0
		for _, m := range migs {
			if applied[m.version] {
				continue
			}
			applyUp(db, m)
			n++
		}
		fmt.Printf("✓ migrate up: %d applied (total %d migrations)\n", n, len(migs))
	case "down":
		applied := appliedVersions(db)
		maxV := -1
		for v := range applied {
			if v > maxV {
				maxV = v
			}
		}
		if maxV < 0 {
			fmt.Println("nothing to roll back")
			return
		}
		var m *migration
		for i := range migs {
			if migs[i].version == maxV {
				m = &migs[i]
			}
		}
		if m == nil {
			fmt.Printf("version %d not found locally\n", maxV)
			return
		}
		applyDown(db, *m)
		fmt.Printf("✓ migrate down: version %d (%s) rolled back\n", maxV, m.name)
	default:
		fatal("unknown command: %s", cmd)
	}
}

func buildDSN() string {
	host := env("MYSQL_HOST", "localhost")
	port := env("MYSQL_PORT", "3306")
	user := env("MYSQL_USER", "consult")
	pass := env("MYSQL_PASSWORD", "consult")
	db := env("MYSQL_DATABASE", "zhipath")
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&loc=Local", user, pass, host, port, db)
}

func env(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func ensureSchemaTable(db *sql.DB) {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		version BIGINT PRIMARY KEY,
		dirty BOOLEAN NOT NULL DEFAULT FALSE,
		updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`)
	if err != nil {
		fatal("create schema_migrations: %v", err)
	}
}

func appliedVersions(db *sql.DB) map[int]bool {
	rows, err := db.Query("SELECT version FROM schema_migrations")
	if err != nil {
		fatal("query applied: %v", err)
	}
	defer rows.Close()
	m := map[int]bool{}
	for rows.Next() {
		var v int
		if err := rows.Scan(&v); err != nil {
			fatal("scan version: %v", err)
		}
		m[v] = true
	}
	return m
}

func loadMigrations(dir string) []migration {
	entries, err := os.ReadDir(dir)
	if err != nil {
		fatal("read migrations dir %s: %v", dir, err)
	}
	byVer := map[int]*migration{}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		mt := fileRe.FindStringSubmatch(e.Name())
		if mt == nil {
			continue
		}
		ver, _ := strconv.Atoi(mt[1])
		content, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			fatal("read %s: %v", e.Name(), err)
		}
		mm := byVer[ver]
		if mm == nil {
			mm = &migration{version: ver, name: mt[2]}
			byVer[ver] = mm
		}
		if mt[3] == "up" {
			mm.up = string(content)
		} else {
			mm.down = string(content)
		}
	}
	out := make([]migration, 0, len(byVer))
	for _, v := range byVer {
		out = append(out, *v)
	}
	return out
}

func applyUp(db *sql.DB, m migration) {
	tx, err := db.Begin()
	if err != nil {
		fatal("begin tx (up v%d): %v", m.version, err)
	}
	if _, err := tx.Exec(m.up); err != nil {
		tx.Rollback()
		fatal("apply up v%d (%s): %v", m.version, m.name, err)
	}
	if _, err := tx.Exec("INSERT INTO schema_migrations (version, dirty) VALUES (?, FALSE)", m.version); err != nil {
		tx.Rollback()
		fatal("record up v%d: %v", m.version, err)
	}
	if err := tx.Commit(); err != nil {
		fatal("commit up v%d: %v", m.version, err)
	}
	fmt.Printf("  ✓ v%d %s\n", m.version, m.name)
}

func applyDown(db *sql.DB, m migration) {
	tx, err := db.Begin()
	if err != nil {
		fatal("begin tx (down v%d): %v", m.version, err)
	}
	if m.down != "" {
		if _, err := tx.Exec(m.down); err != nil {
			tx.Rollback()
			fatal("apply down v%d (%s): %v", m.version, m.name, err)
		}
	}
	if _, err := tx.Exec("DELETE FROM schema_migrations WHERE version = ?", m.version); err != nil {
		tx.Rollback()
		fatal("record down v%d: %v", m.version, err)
	}
	if err := tx.Commit(); err != nil {
		fatal("commit down v%d: %v", m.version, err)
	}
}

func fatal(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "✗ "+format+"\n", args...)
	os.Exit(1)
}

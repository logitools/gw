package pgsql

import (
	"fmt"
	"io/fs"
	"log"
	"path/filepath"
	"strings"

	"github.com/logitools/gw/db/sqldb"
)

var rawStmtStore = sqldb.NewRawStore()

func LoadRawStmtsToStore(sqlFS fs.FS) error {
	stmtCnt := 0
	err := fs.WalkDir(sqlFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			// error reading a directory
			return err
		}
		if d.IsDir() {
			// Skip directory itself. still walking into it.
			return nil
		}
		ext := filepath.Ext(path) // with the leading dot
		if ext == "" {
			return nil
		}
		plainExt := strings.TrimPrefix(ext, ".")
		if plainExt != DBType && plainExt != "sql" {
			return nil
		}
		data, err := fs.ReadFile(sqlFS, path)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", path, err)
		}
		// Build key ("foo/bar/find" → "foo.bar.find")
		// fs.FS always stores paths using forward slashes '/', regardless of the OS.
		key := strings.TrimSuffix(path, ext)
		key = strings.TrimPrefix(key, "./") // just in case that fs is not an embed.fs
		key = strings.ReplaceAll(key, "/", ".")

		if plainExt == DBType {
			// exact matching file extension -> use it as-is for dialects
			rawStmtStore.Set(key, string(data))
			stmtCnt++
			return nil
		}
		// *.sql (Standard SQL) fallback
		if _, exists := rawStmtStore.Get(key); !exists {
			// Placeholder Conversion
			sql := sqldb.ReplaceStaticPlaceholders(string(data), DefaultPlaceholderPrefix)
			rawStmtStore.Set(key, sql)
			stmtCnt++
		}
		return nil
	})
	if err != nil {
		return err
	}
	log.Printf("[INFO][%s] %d sql raw stmts loaded", DBType, stmtCnt)
	return nil
}

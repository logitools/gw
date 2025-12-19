package tpl

import (
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

const FileSuffix = ".gohtml"

type HTMLTemplateStore struct {
	FileTemplates map[string]*template.Template // each file → html/template.New(key).Parse(string(fileBytes))
	Derived       map[string]*template.Template // composed templates
}

func NewHTMLTemplateStore() *HTMLTemplateStore {
	return &HTMLTemplateStore{
		FileTemplates: make(map[string]*template.Template),
		Derived:       make(map[string]*template.Template),
	}
}

func (s *HTMLTemplateStore) LoadFileTemplates(tplRoot string) error {
	tplRoot = filepath.Clean(tplRoot)

	fileInfo, err := os.Stat(tplRoot)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("template root does not exist: %s", tplRoot)
		}
		return fmt.Errorf("cannot access template root %s: %w", tplRoot, err)
	}
	if !fileInfo.IsDir() {
		return fmt.Errorf("template root is not a directory: %s", tplRoot)
	}

	err = filepath.WalkDir( // Pre-order Depth-first Traversal
		tplRoot,
		func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			name := d.Name()
			// Skip hidden dirs/files
			if strings.HasPrefix(name, ".") {
				if d.IsDir() {
					return fs.SkipDir // don't even walk into it
				}
				return nil // skip the file
			}
			if d.IsDir() {
				return nil // just walk into it
			}
			if !strings.HasSuffix(path, FileSuffix) {
				return nil
			}
			// Read file
			fileBytes, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			// UTF-8 only
			if !utf8.Valid(fileBytes) {
				return fmt.Errorf("file %s is not valid UTF-8", path)
			}
			// Template key = relative path to the tplRoot without extension
			rel, _ := filepath.Rel(tplRoot, path) // walking dirs. no chance of errors
			key := strings.TrimSuffix(filepath.ToSlash(rel), FileSuffix)
			// Duplicate
			if _, exists := s.FileTemplates[key]; exists {
				return fmt.Errorf("duplicate template key detected: %s (file=%s)", key, path)
			}
			// Parse
			t, err := template.New(key).Parse(string(fileBytes))
			if err != nil {
				return fmt.Errorf("parse error in %s: %w", path, err)
			}
			s.FileTemplates[key] = t
			return nil
		},
	)
	if err != nil {
		return err
	}
	log.Printf("[INFO][TEMPLATE] Loaded %d templates from %s", len(s.FileTemplates), tplRoot)
	return nil
}

package main

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var videoExtensions = map[string]bool{
	".mkv": true, ".mp4": true, ".avi": true, ".m4v": true,
	".mov": true, ".wmv": true, ".flv": true, ".webm": true, ".ts": true,
}

// Sorter watches a source directory and moves video files into organised destinations.
type Sorter struct {
	cfg       Config
	prevSizes map[string]int64 // path → size from last scan; stable when unchanged
	mu        sync.Mutex
}

func NewSorter(cfg Config) *Sorter {
	return &Sorter{
		cfg:       cfg,
		prevSizes: make(map[string]int64),
	}
}

// Scan walks the source directory, waits for files to stabilise across two scans,
// then moves them to the appropriate destination.
func (s *Sorter) Scan() {
	if _, err := os.Stat(s.cfg.SourcePath); os.IsNotExist(err) {
		log.Printf("[Scan] Source path not available yet: %s", s.cfg.SourcePath)
		return
	}

	log.Printf("[Scan] Scanning %s", s.cfg.SourcePath)
	found := make(map[string]int64)

	_ = filepath.Walk(s.cfg.SourcePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Printf("[Scan] Cannot access %s: %v", path, err)
			return nil
		}
		if info.IsDir() {
			return nil
		}
		if !videoExtensions[strings.ToLower(filepath.Ext(path))] {
			return nil
		}
		found[path] = info.Size()
		return nil
	})

	s.mu.Lock()
	defer s.mu.Unlock()

	moved, pending := 0, 0
	for path, size := range found {
		prev, seen := s.prevSizes[path]
		if !seen || prev != size {
			// New or still growing — record and wait for next scan
			s.prevSizes[path] = size
			pending++
			continue
		}
		// Size identical to last scan — file is stable
		if s.processFile(path) {
			delete(s.prevSizes, path)
			moved++
		} else {
			pending++
		}
	}

	// Drop tracking for files no longer present
	for path := range s.prevSizes {
		if _, exists := found[path]; !exists {
			delete(s.prevSizes, path)
		}
	}

	log.Printf("[Scan] Complete — moved: %d, pending next scan: %d", moved, pending)
}

func (s *Sorter) processFile(src string) bool {
	filename := filepath.Base(src)

	tvInfo := parseTVShowInfo(filename)
	if tvInfo.HasSeasonInfo {
		destDir, err := buildTVShowPath(s.cfg.TVShowPath, tvInfo)
		if err != nil {
			log.Printf("[Sort] Failed to create directory for %s: %v", filename, err)
			return false
		}
		return s.moveFile(src, filepath.Join(destDir, tvInfo.StandardName))
	}

	// Treat as movie
	movieInfo := parseMovieInfo(filename)
	destName := filename
	if movieInfo.HasMovieInfo {
		destName = buildMovieFilename(movieInfo)
	}
	if err := os.MkdirAll(s.cfg.MoviesPath, 0755); err != nil {
		log.Printf("[Sort] Failed to create movies directory: %v", err)
		return false
	}
	return s.moveFile(src, filepath.Join(s.cfg.MoviesPath, destName))
}

func (s *Sorter) moveFile(src, dst string) bool {
	if s.cfg.DryRun {
		log.Printf("[DryRun] %s\n         → %s", src, dst)
		return true
	}

	if _, err := os.Stat(dst); err == nil {
		log.Printf("[Move] Skipping — already exists: %s", dst)
		return true
	}

	log.Printf("[Move] %s\n       → %s", src, dst)

	// Attempt atomic rename (works when source and dest are on the same filesystem)
	if err := os.Rename(src, dst); err == nil {
		s.pruneEmptyDirs(filepath.Dir(src))
		return true
	}

	// Cross-filesystem fallback: copy then delete
	if err := copyFile(src, dst); err != nil {
		log.Printf("[Move] Copy failed for %s: %v", filepath.Base(src), err)
		os.Remove(dst)
		return false
	}
	if err := os.Remove(src); err != nil {
		log.Printf("[Move] Warning: could not remove source %s: %v", src, err)
	}
	s.pruneEmptyDirs(filepath.Dir(src))
	return true
}

// pruneEmptyDirs removes empty directories up the tree toward SourcePath.
func (s *Sorter) pruneEmptyDirs(dir string) {
	for dir != s.cfg.SourcePath && dir != "." && dir != "/" {
		entries, err := os.ReadDir(dir)
		if err != nil || len(entries) > 0 {
			break
		}
		if err := os.Remove(dir); err != nil {
			break
		}
		log.Printf("[Cleanup] Removed empty dir: %s", dir)
		dir = filepath.Dir(dir)
	}
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Sync()
}

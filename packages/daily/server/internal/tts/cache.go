package tts

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	cacheDirRel  = ".xkit/cache/daily/tts"
	defaultTTL   = 7 * 24 * time.Hour
	maxCacheSize = 200 * 1024 * 1024 // 200MB
)

// CacheDir returns the global TTS cache directory (~/.xkit/cache/daily/tts).
func CacheDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("get home dir: %w", err)
	}
	dir := filepath.Join(home, cacheDirRel)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("create cache dir: %w", err)
	}
	return dir, nil
}

// CacheKey builds a filename key from memo UUID prefix + section + expires_at unix.
// Example: a1b2c3d4_T_1717603200.wav  (with TTL)
//
//	e5f6g7h8_T_0.wav                    (no TTL)
func CacheKey(uuid, section string, expiresAt *time.Time) string {
	prefix := uuid
	if len(prefix) > 8 {
		prefix = prefix[:8]
	}
	ts := int64(0)
	if expiresAt != nil {
		ts = expiresAt.Unix()
	}
	return fmt.Sprintf("%s_%s_%d.wav", prefix, section, ts)
}

// parseCacheFile parses a cache filename into its components.
// Format: {uuid8}_{section}_{ts}.wav
func parseCacheFile(name string) (key string, uuidPrefix string, section string, expiresAt int64, ok bool) {
	base := strings.TrimSuffix(name, ".wav")
	parts := strings.Split(base, "_")
	if len(parts) < 3 {
		return "", "", "", 0, false
	}
	ts, err := strconv.ParseInt(parts[len(parts)-1], 10, 64)
	if err != nil {
		return "", "", "", 0, false
	}
	section = parts[len(parts)-2]
	uuidPrefix = parts[0]
	key = strings.Join(parts[:len(parts)-1], "_")
	return key, uuidPrefix, section, ts, true
}

// FindCache looks up a cached WAV file for the given memo UUID + section.
// Returns the file path if found and still valid, empty string otherwise.
func FindCache(dir, uuid, section string) string {
	prefix := uuid
	if len(prefix) > 8 {
		prefix = prefix[:8]
	}
	// Search for files matching {prefix}_{section}_*.wav
	pattern := filepath.Join(dir, fmt.Sprintf("%s_%s_*.wav", prefix, section))
	matches, err := filepath.Glob(pattern)
	if err != nil || len(matches) == 0 {
		return ""
	}

	now := time.Now()
	for _, f := range matches {
		_, _, _, ts, ok := parseCacheFile(filepath.Base(f))
		if !ok {
			continue
		}
		// ts == 0 means no TTL, valid until cleanup (7 days)
		if ts == 0 {
			// Check file age for default 7-day TTL
			info, err := os.Stat(f)
			if err != nil {
				continue
			}
			if now.Sub(info.ModTime()) > defaultTTL {
				continue
			}
			return f
		}
		// Has explicit TTL
		if now.Unix() < ts {
			return f
		}
	}
	return ""
}

// CacheStatus holds info about a single cache entry for display.
type CacheStatus struct {
	Key       string
	FilePath  string
	Size      int64
	ExpiresAt int64 // unix timestamp, 0 = no TTL
	IsValid   bool
}

// GetCacheStatus returns cache info for a specific memo UUID + section.
func GetCacheStatus(dir, uuid, section string) CacheStatus {
	prefix := uuid
	if len(prefix) > 8 {
		prefix = prefix[:8]
	}
	pattern := filepath.Join(dir, fmt.Sprintf("%s_%s_*.wav", prefix, section))
	matches, err := filepath.Glob(pattern)
	if err != nil || len(matches) == 0 {
		return CacheStatus{IsValid: false}
	}

	now := time.Now()
	best := ""
	var bestTS int64
	for _, f := range matches {
		_, _, _, ts, ok := parseCacheFile(filepath.Base(f))
		if !ok {
			continue
		}
		valid := false
		if ts == 0 {
			info, err := os.Stat(f)
			if err == nil && now.Sub(info.ModTime()) <= defaultTTL {
				valid = true
			}
		} else if now.Unix() < ts {
			valid = true
		}
		if valid {
			best = f
			bestTS = ts
			break
		}
	}

	if best == "" {
		return CacheStatus{IsValid: false}
	}

	info, err := os.Stat(best)
	if err != nil {
		return CacheStatus{IsValid: false}
	}

	return CacheStatus{
		Key:       fmt.Sprintf("%s_%s", prefix, section),
		FilePath:  best,
		Size:      info.Size(),
		ExpiresAt: bestTS,
		IsValid:   true,
	}
}

// CleanupCache removes expired and oversized cache files.
func CleanupCache(dir string) (int, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0, nil // not an error if dir doesn't exist
	}

	now := time.Now()
	removed := 0
	type fileEntry struct {
		path    string
		modTime time.Time
		size    int64
	}
	var valid []fileEntry

	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".wav") {
			continue
		}
		_, _, _, ts, ok := parseCacheFile(e.Name())
		if !ok {
			continue
		}

		path := filepath.Join(dir, e.Name())
		info, err := os.Stat(path)
		if err != nil {
			continue
		}

		expired := false
		if ts == 0 && now.Sub(info.ModTime()) > defaultTTL {
			expired = true
		} else if ts > 0 && now.Unix() > ts {
			expired = true
		}

		if expired {
			os.Remove(path)
			removed++
		} else {
			valid = append(valid, fileEntry{path: path, modTime: info.ModTime(), size: info.Size()})
		}
	}

	// LRU eviction if over size limit
	var totalSize int64
	for _, v := range valid {
		totalSize += v.size
	}
	if totalSize > maxCacheSize && len(valid) > 1 {
		// Sort by modTime ascending (oldest first) — simple bubble sort for small lists
		for i := 0; i < len(valid); i++ {
			for j := i + 1; j < len(valid); j++ {
				if valid[i].modTime.After(valid[j].modTime) {
					valid[i], valid[j] = valid[j], valid[i]
				}
			}
		}
		for _, v := range valid {
			if totalSize <= maxCacheSize {
				break
			}
			os.Remove(v.path)
			totalSize -= v.size
			removed++
		}
	}

	return removed, nil
}

// CacheStats holds aggregate cache statistics.
type CacheStats struct {
	Count     int
	TotalSize int64
}

// GetCacheStats returns aggregate cache info.
func GetCacheStats(dir string) CacheStats {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return CacheStats{}
	}

	var stats CacheStats
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".wav") {
			continue
		}
		info, err := os.Stat(filepath.Join(dir, e.Name()))
		if err != nil {
			continue
		}
		stats.Count++
		stats.TotalSize += info.Size()
	}
	return stats
}

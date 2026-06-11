package strata

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

// Options configures how Load and Save operate.
type Options struct {
	Namespace  string // Key in ~/.xkit/config.json, e.g. "daily"
	ConfigFile string // Tool-specific config file path, e.g. ~/.daily/config.json (optional)
}

// Load reads configuration from multiple sources with priority:
//
//	struct envDefault tags → .env file → ~/.xkit/config.json[ns] → ConfigFile → explicit env vars
//
// target must be a pointer to a struct with `json` and `env` tags.
// Returns a Sources map tracking where each value came from.
func Load(target interface{}, opts Options) (Sources, error) {
	sources := make(Sources)

	// 1. Fill struct defaults via envDefault tags
	if err := env.Parse(target); err != nil {
		return nil, fmt.Errorf("strata: parse defaults: %w", err)
	}

	// 2. Mark all fields as default
	markFields(target, "", sources, SourceDefault)

	// 3. Detect which env vars are explicitly set (before we merge anything)
	envSet := detectExplicitEnv(target)

	// 4. Merge .env file (without polluting os.Environ)
	if envFile, err := godotenv.Read(); err == nil {
		mergeFromMap(target, envFile, sources, "env:")
	}

	// 5. Merge ~/.xkit/config.json[namespace]
	if opts.Namespace != "" {
		if home, err := os.UserHomeDir(); err == nil {
			xkitPath := filepath.Join(home, ".xkit", "config.json")
			mergeFromXkitFile(target, xkitPath, opts.Namespace, sources)
		}
	}

	// 6. Merge tool-specific config file (higher priority)
	if opts.ConfigFile != "" {
		mergeFromConfigFile(target, opts.ConfigFile, opts.Namespace, sources)
	}

	// 7. Merge explicit env vars (highest priority)
	mergeFromEnvVars(target, envSet, sources)

	return sources, nil
}

// Save writes the target config to the ConfigFile path.
// It reads any existing file first and updates only the relevant section,
// preserving other settings.
func Save(target interface{}, opts Options) error {
	if opts.ConfigFile == "" {
		return fmt.Errorf("strata: ConfigFile is required for Save")
	}

	// Ensure directory exists
	dir := filepath.Dir(opts.ConfigFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("strata: create config dir: %w", err)
	}

	// Read existing file to preserve other sections
	existing := map[string]interface{}{}
	if data, err := os.ReadFile(opts.ConfigFile); err == nil {
		_ = json.Unmarshal(data, &existing)
	}

	// Marshal target and merge into existing
	targetData, err := json.Marshal(target)
	if err != nil {
		return fmt.Errorf("strata: marshal target: %w", err)
	}
	var targetMap map[string]interface{}
	if err := json.Unmarshal(targetData, &targetMap); err != nil {
		return fmt.Errorf("strata: unmarshal target: %w", err)
	}

	// If namespace is set, put under that key; otherwise write top-level
	if opts.Namespace != "" {
		existing[opts.Namespace] = targetMap
	} else {
		for k, v := range targetMap {
			existing[k] = v
		}
	}

	data, err := json.MarshalIndent(existing, "", "  ")
	if err != nil {
		return fmt.Errorf("strata: marshal config: %w", err)
	}

	return os.WriteFile(opts.ConfigFile, data, 0644)
}

// ── Internal helpers ──

// markFields walks the struct and sets the source for every leaf field.
func markFields(v interface{}, prefix string, sources Sources, src Source) {
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return
	}
	rt := rv.Type()
	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)
		fv := rv.Field(i)

		// Build dot-path key
		jsonTag := field.Tag.Get("json")
		name := strings.Split(jsonTag, ",")[0]
		if name == "" || name == "-" {
			continue
		}
		key := name
		if prefix != "" {
			key = prefix + "." + name
		}

		switch fv.Kind() {
		case reflect.Struct:
			// Recurse into nested structs
			markFields(fv.Addr().Interface(), key, sources, src)
		default:
			sources[key] = src
		}
	}
}

// detectExplicitEnv returns a set of env var names that are explicitly set in the environment.
func detectExplicitEnv(v interface{}) map[string]bool {
	set := make(map[string]bool)
	detectEnvFromStruct(v, set)
	return set
}

func detectEnvFromStruct(v interface{}, set map[string]bool) {
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return
	}
	rt := rv.Type()
	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)
		fv := rv.Field(i)

		if fv.Kind() == reflect.Struct {
			detectEnvFromStruct(fv.Addr().Interface(), set)
			continue
		}

		envTag := field.Tag.Get("env")
		if envTag == "" {
			continue
		}
		for _, tag := range strings.Split(envTag, ",") {
			tag = strings.TrimSpace(tag)
			if tag != "" {
				if _, ok := os.LookupEnv(tag); ok {
					set[tag] = true
				}
			}
		}
	}
}

// mergeFromMap overlays values from a map onto the struct.
// mapKeys should be prefixed with "env:" to match env tag names.
func mergeFromMap(v interface{}, m map[string]string, sources Sources, keyPrefix string) {
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return
	}
	rt := rv.Type()
	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)
		fv := rv.Field(i)

		jsonTag := field.Tag.Get("json")
		name := strings.Split(jsonTag, ",")[0]
		if name == "" || name == "-" {
			continue
		}

		if fv.Kind() == reflect.Struct {
			mergeFromMap(fv.Addr().Interface(), m, sources, keyPrefix)
			continue
		}

		envTag := field.Tag.Get("env")
		if envTag == "" {
			continue
		}
		envName := strings.Split(envTag, ",")[0]
		envName = strings.TrimSpace(envName)
		if envName == "" {
			continue
		}

		if val, ok := m[envName]; ok && val != "" {
			setFieldValue(fv, val)
			sources[name] = SourceFile
		}
	}
}

// mergeFromXkitFile reads ~/.xkit/config.json and overlays the namespace section.
func mergeFromXkitFile(v interface{}, path string, namespace string, sources Sources) {
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return
	}
	sectionRaw, ok := raw[namespace]
	if !ok {
		return
	}

	// Unmarshal section into a temp copy of the target type
	tmp := reflect.New(reflect.TypeOf(v).Elem()).Interface()
	if err := json.Unmarshal(sectionRaw, tmp); err != nil {
		return
	}

	// Overlay non-zero values
	overlayStruct(v, tmp, "", sources, SourceFile)
}

// mergeFromConfigFile reads a tool-specific config file and overlays fields.
// Handles both namespaced ({"daily": {...}}) and flat ({...}) formats.
func mergeFromConfigFile(v interface{}, path string, namespace string, sources Sources) {
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}

	// Try to detect namespace wrapper
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err == nil {
		if namespace != "" {
			if sectionRaw, ok := raw[namespace]; ok {
				tmp := reflect.New(reflect.TypeOf(v).Elem()).Interface()
				if err := json.Unmarshal(sectionRaw, tmp); err == nil {
					overlayStruct(v, tmp, "", sources, SourceFile)
					return
				}
			}
		}
		if len(raw) == 1 {
			// Single key — might be a namespace wrapper. Try to use it.
			for _, sectionRaw := range raw {
				tmp := reflect.New(reflect.TypeOf(v).Elem()).Interface()
				if err := json.Unmarshal(sectionRaw, tmp); err == nil {
					overlayStruct(v, tmp, "", sources, SourceFile)
					return
				}
			}
		}
	}

	// Fall back to flat config
	tmp := reflect.New(reflect.TypeOf(v).Elem()).Interface()
	if err := json.Unmarshal(data, tmp); err != nil {
		return
	}
	overlayStruct(v, tmp, "", sources, SourceFile)
}

// mergeFromEnvVars overlays explicitly-set env vars onto the struct.
func mergeFromEnvVars(v interface{}, envSet map[string]bool, sources Sources) {
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return
	}
	rt := rv.Type()
	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)
		fv := rv.Field(i)

		jsonTag := field.Tag.Get("json")
		name := strings.Split(jsonTag, ",")[0]
		if name == "" || name == "-" {
			continue
		}

		if fv.Kind() == reflect.Struct {
			mergeFromEnvVars(fv.Addr().Interface(), envSet, sources)
			continue
		}

		envTag := field.Tag.Get("env")
		if envTag == "" {
			continue
		}
		envName := strings.Split(envTag, ",")[0]
		envName = strings.TrimSpace(envName)
		if envName == "" {
			continue
		}

		if envSet[envName] {
			if val, ok := os.LookupEnv(envName); ok {
				setFieldValue(fv, val)
				sources[name] = SourceEnv
			}
		}
	}
}

// overlayStruct copies non-zero values from src into dst, tracking sources.
func overlayStruct(dst, src interface{}, prefix string, sources Sources, srcType Source) {
	dstRv := reflect.ValueOf(dst)
	srcRv := reflect.ValueOf(src)
	if dstRv.Kind() == reflect.Ptr {
		dstRv = dstRv.Elem()
	}
	if srcRv.Kind() == reflect.Ptr {
		srcRv = srcRv.Elem()
	}
	if dstRv.Kind() != reflect.Struct || srcRv.Kind() != reflect.Struct {
		return
	}

	dstRt := dstRv.Type()
	for i := 0; i < dstRt.NumField(); i++ {
		field := dstRt.Field(i)
		dstFv := dstRv.Field(i)
		srcFv := srcRv.Field(i)

		jsonTag := field.Tag.Get("json")
		name := strings.Split(jsonTag, ",")[0]
		if name == "" || name == "-" {
			continue
		}
		key := name
		if prefix != "" {
			key = prefix + "." + name
		}

		if dstFv.Kind() == reflect.Struct {
			overlayStruct(dstFv.Addr().Interface(), srcFv.Addr().Interface(), key, sources, srcType)
			continue
		}

		// Check if src value is non-zero
		if srcFv.IsZero() {
			continue
		}

		dstFv.Set(srcFv)
		sources[key] = srcType
	}
}

// setFieldValue sets a struct field from a string value.
func setFieldValue(fv reflect.Value, val string) {
	switch fv.Kind() {
	case reflect.String:
		fv.SetString(val)
	case reflect.Bool:
		fv.SetBool(val == "true" || val == "1")
	case reflect.Int, reflect.Int64:
		var n int64
		fmt.Sscanf(val, "%d", &n)
		fv.SetInt(n)
	case reflect.Float64:
		var f float64
		fmt.Sscanf(val, "%f", &f)
		fv.SetFloat(f)
	}
}

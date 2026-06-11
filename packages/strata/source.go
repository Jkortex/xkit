package strata

// Source tracks where a config value came from.
type Source string

const (
	SourceDefault Source = "default"
	SourceFile    Source = "file"
	SourceEnv     Source = "env"
)

// Sources maps config field dot-paths to their source.
type Sources map[string]Source

package config

import "encoding/json"

// migrationStep rewrites a raw config file from one schema version to the next.
// Steps chain, so each one only has to know about the version it upgrades from.
type migrationStep func(raw []byte) ([]byte, error)

// migrations[v] upgrades a file written at schema version v to v+1. Empty
// today: v1 is the first released schema, so no file predates it.
var migrations = map[int]migrationStep{}

// schemaVersionOf reports the schema_version recorded in raw. A missing or
// unparseable field reads as 0.
func schemaVersionOf(raw []byte) int {
	var probe struct {
		SchemaVersion int `json:"schema_version"`
	}
	_ = json.Unmarshal(raw, &probe)
	return probe.SchemaVersion
}

// migrateToCurrent walks raw up to CurrentSchemaVersion a step at a time. A
// version with no registered step ends the walk instead of failing.
func migrateToCurrent(raw []byte) ([]byte, bool, error) {
	changed := false
	for version := schemaVersionOf(raw); version < CurrentSchemaVersion; version++ {
		step, ok := migrations[version]
		if !ok {
			break
		}
		next, err := step(raw)
		if err != nil {
			return nil, false, err
		}
		raw, changed = next, true
	}
	return raw, changed, nil
}

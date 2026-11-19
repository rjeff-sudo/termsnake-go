// Package migrations holds the SQL schema for the local score database and
// embeds it into the binary so the game ships as a single file.
package migrations

import _ "embed"

// Init creates the scores and settings tables.
//
//go:embed 0001_init.sql
var Init string

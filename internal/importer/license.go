package importer

import (
	"regexp"
	"strings"
)

// leadingCC matches a Creative Commons license named at the start of
// useStmt/conditions, e.g. "CC BY-NC-ND 4.0 (https://…): non-commercial …".
var leadingCC = regexp.MustCompile(`^CC (BY(?:-NC)?(?:-SA|-ND)?) 4\.0\b`)

// LicenseID returns the SPDX identifier (e.g. "CC-BY-NC-ND-4.0") of the
// license a study's conditions start with, so clients can show its button
// without reading the prose. "" when they start with anything else: other
// terms of use (e.g. Statistikamt Nord's) or no license stated.
func LicenseID(conditions string) string {
	m := leadingCC.FindStringSubmatch(strings.TrimSpace(conditions))
	if m == nil {
		return ""
	}
	return "CC-" + m[1] + "-4.0"
}

package updater

import (
	"strconv"
	"strings"
)

// Version represents a semantic version
type Version struct {
	Major int
	Minor int
	Patch int
	Raw   string
}

// ParseVersion parses a version string like "v1.2.3" or "1.2.3"
func ParseVersion(v string) Version {
	raw := v
	v = strings.TrimPrefix(v, "v")
	parts := strings.Split(v, ".")

	var major, minor, patch int
	if len(parts) >= 1 {
		major, _ = strconv.Atoi(parts[0])
	}
	if len(parts) >= 2 {
		minor, _ = strconv.Atoi(parts[1])
	}
	if len(parts) >= 3 {
		patch, _ = strconv.Atoi(parts[2])
	}

	return Version{
		Major: major,
		Minor: minor,
		Patch: patch,
		Raw:   raw,
	}
}

// String returns the version as a string
func (v Version) String() string {
	return v.Raw
}

// IsNewerThan returns true if v is newer than other
func (v Version) IsNewerThan(other Version) bool {
	if v.Major != other.Major {
		return v.Major > other.Major
	}
	if v.Minor != other.Minor {
		return v.Minor > other.Minor
	}
	return v.Patch > other.Patch
}

// Equal returns true if versions are equal
func (v Version) Equal(other Version) bool {
	return v.Major == other.Major && v.Minor == other.Minor && v.Patch == other.Patch
}

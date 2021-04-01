package version

import (
	"bytes"
	"fmt"
)

var (
	// Whether cgo is enabled or not; set at build time
	CgoEnabled bool

	Version     = "unknown"
	VPrerelease = "unknown"
	VMetadata   = ""
)

// VInfo
type VInfo struct {
	Revision    string
	Version     string
	VPrerelease string
	VMetadata   string
}

func GetVersion() *VInfo {
	ver := Version
	rel := VPrerelease
	md := VMetadata

	return &VInfo{
		Version:     ver,
		VPrerelease: rel,
		VMetadata:   md,
	}
}

func (c *VInfo) VersionNumber() string {
	if Version == "unknown" {
		return "(version unknown)"
	}

	version := fmt.Sprintf("%s", c.Version)

	if c.VPrerelease != "" {
		version = fmt.Sprintf("%s-%s", version, c.VPrerelease)
	}

	if c.VMetadata != "" {
		version = fmt.Sprintf("%s+%s", version, c.VMetadata)
	}

	return version
}

func (c *VInfo) FullVersionNumber(rev bool) string {
	var versionString bytes.Buffer

	if Version == "unknown" {
		return "Vault (version unknown)"
	}

	fmt.Fprintf(&versionString, "v%s", c.Version)
	if c.VPrerelease != "" {
		fmt.Fprintf(&versionString, "-%s", c.VPrerelease)
	}

	if c.VMetadata != "" {
		fmt.Fprintf(&versionString, "+%s", c.VMetadata)
	}

	if rev && c.Revision != "" {
		fmt.Fprintf(&versionString, " (%s)", c.Revision)
	}

	return versionString.String()
}

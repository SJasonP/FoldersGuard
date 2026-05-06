package fsmeta

import (
	"fmt"
	"io/fs"
	"os"
	"sort"
	"strings"
	"time"
)

const (
	CapabilityMode              = "mode"
	CapabilityModTime           = "mod_time"
	CapabilityAccessTime        = "access_time"
	CapabilityBirthTime         = "birth_time"
	CapabilityWindowsAttributes = "windows_attributes"
)

type Metadata struct {
	Mode              uint32
	ModTime           time.Time
	AccessTime        *time.Time
	BirthTime         *time.Time
	WindowsAttributes *uint32
	Capabilities      []string
}

func Capture(path string, info fs.FileInfo) (Metadata, error) {
	if info == nil {
		return Metadata{}, fmt.Errorf("file info is required")
	}
	metadata := Metadata{
		Mode:         uint32(info.Mode()),
		ModTime:      info.ModTime().UTC(),
		Capabilities: []string{CapabilityMode, CapabilityModTime},
	}
	if err := capturePlatform(path, &metadata); err != nil {
		return Metadata{}, err
	}
	metadata.Capabilities = normalizeCapabilities(metadata.Capabilities)
	return metadata, nil
}

func Apply(path string, metadata Metadata) error {
	capabilities := capabilitySet(metadata.Capabilities)
	if capabilities[CapabilityMode] {
		if err := os.Chmod(path, fs.FileMode(metadata.Mode)); err != nil {
			return fmt.Errorf("restore mode: %w", err)
		}
	}
	if err := applyPlatformAttributes(path, metadata, capabilities); err != nil {
		return err
	}
	if capabilities[CapabilityModTime] || capabilities[CapabilityAccessTime] {
		atime := time.Time{}
		if capabilities[CapabilityAccessTime] && metadata.AccessTime != nil {
			atime = metadata.AccessTime.UTC()
		}
		mtime := time.Time{}
		if capabilities[CapabilityModTime] {
			mtime = metadata.ModTime.UTC()
		}
		if err := os.Chtimes(path, atime, mtime); err != nil {
			return fmt.Errorf("restore timestamps: %w", err)
		}
	}
	if err := applyPlatformTimes(path, metadata, capabilities); err != nil {
		return err
	}
	return nil
}

func CapabilitiesString(capabilities []string) string {
	return strings.Join(normalizeCapabilities(capabilities), ",")
}

func ParseCapabilities(value string) []string {
	if value == "" {
		return nil
	}
	return normalizeCapabilities(strings.Split(value, ","))
}

func capabilitySet(capabilities []string) map[string]bool {
	output := make(map[string]bool, len(capabilities))
	for _, capability := range capabilities {
		if capability != "" {
			output[capability] = true
		}
	}
	return output
}

func addCapability(metadata *Metadata, capability string) {
	metadata.Capabilities = append(metadata.Capabilities, capability)
}

func normalizeCapabilities(capabilities []string) []string {
	seen := make(map[string]struct{}, len(capabilities))
	var output []string
	for _, capability := range capabilities {
		capability = strings.TrimSpace(capability)
		if capability == "" {
			continue
		}
		if _, ok := seen[capability]; ok {
			continue
		}
		seen[capability] = struct{}{}
		output = append(output, capability)
	}
	sort.Strings(output)
	return output
}

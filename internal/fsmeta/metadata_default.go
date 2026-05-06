//go:build !darwin && !freebsd && !linux && !netbsd && !openbsd && !windows

package fsmeta

func capturePlatform(path string, metadata *Metadata) error {
	return nil
}

func applyPlatformAttributes(path string, metadata Metadata, capabilities map[string]bool) error {
	return nil
}

func applyPlatformTimes(path string, metadata Metadata, capabilities map[string]bool) error {
	return nil
}

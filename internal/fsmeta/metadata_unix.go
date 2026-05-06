//go:build darwin || freebsd || linux || netbsd || openbsd

package fsmeta

func applyPlatformAttributes(path string, metadata Metadata, capabilities map[string]bool) error {
	return nil
}

func applyPlatformTimes(path string, metadata Metadata, capabilities map[string]bool) error {
	return nil
}

//go:build linux

package fsmeta

import (
	"fmt"
	"time"

	"golang.org/x/sys/unix"
)

func capturePlatform(path string, metadata *Metadata) error {
	var stat unix.Stat_t
	if err := unix.Lstat(path, &stat); err != nil {
		return fmt.Errorf("stat filesystem metadata: %w", err)
	}
	accessTime := time.Unix(0, unix.TimespecToNsec(stat.Atim)).UTC()
	metadata.AccessTime = &accessTime
	addCapability(metadata, CapabilityAccessTime)
	return nil
}

package app

import (
	"fmt"
	"os"
	"path/filepath"
)

const diskSpaceReserve int64 = 16 * 1024 * 1024

type diskSpaceInfo struct {
	Available int64
	VolumeID  string
}

var inspectDiskSpace = platformDiskSpace

func ensureOperationSpace(sourcePath, outputPath string, outputs, freed []int64, incremental bool) error {
	if len(outputs) != len(freed) {
		return fmt.Errorf("internal disk-space estimate mismatch")
	}
	outputSpace, err := inspectDiskSpace(existingPath(outputPath))
	if err != nil {
		return fmt.Errorf("inspect output disk space: %w", err)
	}
	required := totalOutputSize(outputs)
	if incremental {
		sourceSpace, err := inspectDiskSpace(existingPath(sourcePath))
		if err != nil {
			return fmt.Errorf("inspect source disk space: %w", err)
		}
		if sourceSpace.VolumeID == outputSpace.VolumeID {
			required = incrementalPeakSize(outputs, freed)
		}
	}
	return requireAvailableSpace(outputSpace.Available, required)
}

func ensureOperationSpaceForSources(sourcePaths []string, outputPath string, outputs, freed []int64, incremental bool) error {
	if len(sourcePaths) != len(outputs) || len(outputs) != len(freed) {
		return fmt.Errorf("internal disk-space estimate mismatch")
	}
	outputSpace, err := inspectDiskSpace(existingPath(outputPath))
	if err != nil {
		return fmt.Errorf("inspect output disk space: %w", err)
	}
	required := totalOutputSize(outputs)
	if incremental {
		reclaimable := append([]int64(nil), freed...)
		volumes := make(map[string]string)
		for i, sourcePath := range sourcePaths {
			root := existingPath(sourcePath)
			volumeID, ok := volumes[root]
			if !ok {
				sourceSpace, err := inspectDiskSpace(root)
				if err != nil {
					return fmt.Errorf("inspect source disk space: %w", err)
				}
				volumeID = sourceSpace.VolumeID
				volumes[root] = volumeID
			}
			if volumeID != outputSpace.VolumeID {
				reclaimable[i] = 0
			}
		}
		required = incrementalPeakSize(outputs, reclaimable)
	}
	return requireAvailableSpace(outputSpace.Available, required)
}

func requireAvailableSpace(available, required int64) error {
	required += diskSpaceReserve
	if available < required {
		return fmt.Errorf("%w: output volume has %d bytes available, but this operation requires at least %d bytes", ErrInsufficientDiskSpace, available, required)
	}
	return nil
}

func existingPath(path string) string {
	current, err := filepath.Abs(path)
	if err != nil {
		current = filepath.Clean(path)
	}
	for {
		if _, err := os.Stat(current); err == nil {
			return current
		}
		parent := filepath.Dir(current)
		if parent == current {
			return current
		}
		current = parent
	}
}

func totalOutputSize(outputs []int64) int64 {
	var total int64
	for _, size := range outputs {
		total += size
	}
	return total
}

func incrementalPeakSize(outputs, freed []int64) int64 {
	var consumed, peak int64
	for i, output := range outputs {
		if candidate := consumed + output; candidate > peak {
			peak = candidate
		}
		consumed += output - freed[i]
		if consumed < 0 {
			consumed = 0
		}
	}
	return peak
}

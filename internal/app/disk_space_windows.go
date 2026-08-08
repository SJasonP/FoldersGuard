//go:build windows

package app

import (
	"path/filepath"
	"strings"

	"golang.org/x/sys/windows"
)

func platformDiskSpace(path string) (diskSpaceInfo, error) {
	pathPtr, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return diskSpaceInfo{}, err
	}
	var available uint64
	if err := windows.GetDiskFreeSpaceEx(pathPtr, &available, nil, nil); err != nil {
		return diskSpaceInfo{}, err
	}
	return diskSpaceInfo{
		Available: int64(available),
		VolumeID:  strings.ToLower(filepath.VolumeName(path)),
	}, nil
}

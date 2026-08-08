//go:build !windows

package app

import (
	"fmt"
	"syscall"
)

func platformDiskSpace(path string) (diskSpaceInfo, error) {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		return diskSpaceInfo{}, err
	}
	var fileStat syscall.Stat_t
	if err := syscall.Stat(path, &fileStat); err != nil {
		return diskSpaceInfo{}, err
	}
	return diskSpaceInfo{
		Available: int64(stat.Bavail) * int64(stat.Bsize),
		VolumeID:  fmt.Sprint(fileStat.Dev),
	}, nil
}

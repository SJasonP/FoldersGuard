//go:build windows

package fsmeta

import (
	"fmt"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

const supportedWindowsAttributes = windows.FILE_ATTRIBUTE_READONLY |
	windows.FILE_ATTRIBUTE_HIDDEN |
	windows.FILE_ATTRIBUTE_SYSTEM |
	windows.FILE_ATTRIBUTE_ARCHIVE

func capturePlatform(path string, metadata *Metadata) error {
	name, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return fmt.Errorf("encode metadata path: %w", err)
	}
	var data windows.Win32FileAttributeData
	if err := windows.GetFileAttributesEx(name, windows.GetFileExInfoStandard, (*byte)(unsafe.Pointer(&data))); err != nil {
		return fmt.Errorf("stat filesystem metadata: %w", err)
	}
	accessTime := filetimeToTime(data.LastAccessTime)
	birthTime := filetimeToTime(data.CreationTime)
	attributes := data.FileAttributes & supportedWindowsAttributes
	metadata.AccessTime = &accessTime
	metadata.BirthTime = &birthTime
	metadata.WindowsAttributes = &attributes
	addCapability(metadata, CapabilityAccessTime)
	addCapability(metadata, CapabilityBirthTime)
	addCapability(metadata, CapabilityWindowsAttributes)
	return nil
}

func applyPlatformAttributes(path string, metadata Metadata, capabilities map[string]bool) error {
	if !capabilities[CapabilityWindowsAttributes] || metadata.WindowsAttributes == nil {
		return nil
	}
	name, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return fmt.Errorf("encode metadata path: %w", err)
	}
	current, err := windows.GetFileAttributes(name)
	if err != nil {
		return fmt.Errorf("read windows attributes: %w", err)
	}
	attributes := (current &^ supportedWindowsAttributes) | (*metadata.WindowsAttributes & supportedWindowsAttributes)
	if err := windows.SetFileAttributes(name, attributes); err != nil {
		return fmt.Errorf("restore windows attributes: %w", err)
	}
	return nil
}

func applyPlatformTimes(path string, metadata Metadata, capabilities map[string]bool) error {
	if !capabilities[CapabilityBirthTime] || metadata.BirthTime == nil {
		return nil
	}
	name, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return fmt.Errorf("encode metadata path: %w", err)
	}
	handle, err := windows.CreateFile(
		name,
		windows.FILE_WRITE_ATTRIBUTES,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE,
		nil,
		windows.OPEN_EXISTING,
		windows.FILE_FLAG_BACKUP_SEMANTICS,
		0,
	)
	if err != nil {
		return fmt.Errorf("open path for creation time restore: %w", err)
	}
	defer windows.CloseHandle(handle)

	creation := windows.NsecToFiletime(metadata.BirthTime.UTC().UnixNano())
	if err := windows.SetFileTime(handle, &creation, nil, nil); err != nil {
		return fmt.Errorf("restore creation time: %w", err)
	}
	return nil
}

func filetimeToTime(value windows.Filetime) time.Time {
	return time.Unix(0, (&value).Nanoseconds()).UTC()
}

package noise

import "strings"

const (
	ModeIgnoreEverywhere              = "ignore_everywhere"
	ModeIgnoreDuringVerifyAndMatching = "ignore_during_verify_and_matching"
	ModeDoNotIgnore                   = "do_not_ignore"
)

func IsName(name string) bool {
	switch name {
	case ".DS_Store", "._.DS_Store", "Thumbs.db", "ehthumbs.db", "desktop.ini":
		return true
	case ".Spotlight-V100", ".Trashes", ".fseventsd":
		return true
	default:
		return strings.HasPrefix(name, "._")
	}
}

func IgnoreDuringSourceScan(mode string) bool {
	return mode == "" || mode == ModeIgnoreEverywhere
}

func IgnoreDuringMatching(mode string) bool {
	return mode == "" || mode == ModeIgnoreEverywhere || mode == ModeIgnoreDuringVerifyAndMatching
}

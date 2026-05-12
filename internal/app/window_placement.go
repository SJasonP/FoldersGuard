package app

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type WindowPlacement struct {
	X         int  `json:"x"`
	Y         int  `json:"y"`
	Width     int  `json:"width"`
	Height    int  `json:"height"`
	Maximised bool `json:"maximised"`
}

func (s Service) WindowPlacementPath() string {
	return filepath.Join(s.DataDir, "window-placement.json")
}

func (s Service) ReadWindowPlacement(minWidth, minHeight int) (WindowPlacement, bool, error) {
	if err := s.EnsureDataDir(); err != nil {
		return WindowPlacement{}, false, err
	}

	data, err := os.ReadFile(s.WindowPlacementPath())
	if err != nil {
		if os.IsNotExist(err) {
			return WindowPlacement{}, false, nil
		}
		return WindowPlacement{}, false, fmt.Errorf("read window placement: %w", err)
	}

	var placement WindowPlacement
	if err := json.Unmarshal(data, &placement); err != nil {
		return WindowPlacement{}, false, fmt.Errorf("decode window placement: %w", err)
	}
	placement, ok := normalizeWindowPlacement(placement, minWidth, minHeight)
	return placement, ok, nil
}

func (s Service) SaveWindowPlacement(placement WindowPlacement, minWidth, minHeight int) error {
	if err := s.EnsureDataDir(); err != nil {
		return err
	}

	normalized, ok := normalizeWindowPlacement(placement, minWidth, minHeight)
	if !ok {
		return nil
	}
	data, err := json.MarshalIndent(normalized, "", "  ")
	if err != nil {
		return fmt.Errorf("encode window placement: %w", err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(s.WindowPlacementPath(), data, 0o600); err != nil {
		return fmt.Errorf("write window placement: %w", err)
	}
	return nil
}

func normalizeWindowPlacement(placement WindowPlacement, minWidth, minHeight int) (WindowPlacement, bool) {
	if placement.Width < minWidth || placement.Height < minHeight {
		return WindowPlacement{}, false
	}
	return placement, true
}

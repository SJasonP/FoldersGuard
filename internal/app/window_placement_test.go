package app

import (
	"path/filepath"
	"reflect"
	"testing"
)

func TestWindowPlacementRoundTrip(t *testing.T) {
	service, err := NewService(filepath.Join(t.TempDir(), "data"))
	if err != nil {
		t.Fatal(err)
	}
	want := WindowPlacement{
		X:         120,
		Y:         80,
		Width:     1440,
		Height:    900,
		Maximised: true,
	}

	if err := service.SaveWindowPlacement(want, 700, 500); err != nil {
		t.Fatal(err)
	}
	got, ok, err := service.ReadWindowPlacement(700, 500)
	if err != nil {
		t.Fatal(err)
	}
	if !ok || !reflect.DeepEqual(got, want) {
		t.Fatalf("window placement = %+v ok=%v, want %+v ok=true", got, ok, want)
	}
}

func TestWindowPlacementRejectsTooSmallSize(t *testing.T) {
	service, err := NewService(filepath.Join(t.TempDir(), "data"))
	if err != nil {
		t.Fatal(err)
	}

	if err := service.SaveWindowPlacement(WindowPlacement{Width: 500, Height: 400}, 700, 500); err != nil {
		t.Fatal(err)
	}
	_, ok, err := service.ReadWindowPlacement(700, 500)
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatal("expected invalid placement to be ignored")
	}
}

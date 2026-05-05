package project

import (
	"testing"
	"time"

	"github.com/google/uuid"

	"foldersguard/internal/fswalk"
	"foldersguard/internal/model"
)

func TestPlannerPlan(t *testing.T) {
	ids := deterministicUUIDs()
	planner := Planner{
		MaxPartSize: 5,
		Now: func() time.Time {
			return time.Date(2026, 5, 5, 10, 0, 0, 0, time.UTC)
		},
		NewUUID: ids.next,
		NewKey: func() ([]byte, error) {
			return make([]byte, 32), nil
		},
	}

	plan, err := planner.Plan(fswalk.ScanResult{
		Root: fswalk.Entry{AbsolutePath: "/tmp/root", Type: fswalk.EntryTypeFolder},
		Entries: []fswalk.Entry{
			{RootRelativePath: "dir/file.txt", Type: fswalk.EntryTypeFile, Size: 12},
			{RootRelativePath: "dir", Type: fswalk.EntryTypeFolder},
			{RootRelativePath: "small.txt", Type: fswalk.EntryTypeFile, Size: 3},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	if plan.RootItem.RealName != "root" {
		t.Fatalf("root name = %q, want root", plan.RootItem.RealName)
	}
	if len(plan.Items) != 3 {
		t.Fatalf("items = %d, want 3", len(plan.Items))
	}
	if len(plan.Folders) != 1 {
		t.Fatalf("folders = %d, want 1 child folder", len(plan.Folders))
	}
	if len(plan.Files) != 2 {
		t.Fatalf("files = %d, want 2", len(plan.Files))
	}
	if len(plan.Parts) != 3 {
		t.Fatalf("parts = %d, want 3", len(plan.Parts))
	}

	var splitFile *model.File
	for i := range plan.Files {
		if plan.Files[i].OriginalSize == 12 {
			splitFile = &plan.Files[i]
			break
		}
	}
	if splitFile == nil {
		t.Fatal("missing split file")
	}
	if splitFile.StorageKind != model.StorageKindSplit {
		t.Fatalf("storage kind = %q, want split", splitFile.StorageKind)
	}

	partSizes := []int64{}
	for _, part := range plan.Parts {
		partSizes = append(partSizes, part.Size)
	}
	if got, want := partSizes, []int64{4, 4, 4}; !equalInt64s(got, want) {
		t.Fatalf("part sizes = %v, want %v", got, want)
	}
}

func deterministicUUIDs() *uuidSeq {
	return &uuidSeq{nextID: 1}
}

type uuidSeq struct {
	nextID byte
}

func (s *uuidSeq) next() uuid.UUID {
	var id uuid.UUID
	id[15] = s.nextID
	s.nextID++
	return id
}

func equalInt64s(left, right []int64) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}

package model

import "testing"

func TestPlanBalancedSplitSingle(t *testing.T) {
	plan, err := PlanBalancedSplit(100, 100)
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Parts) != 1 {
		t.Fatalf("parts = %d, want 1", len(plan.Parts))
	}
	if plan.Parts[0].Offset != 0 || plan.Parts[0].Size != 100 {
		t.Fatalf("part = %+v, want offset 0 size 100", plan.Parts[0])
	}
}

func TestPlanBalancedSplitEvenDistribution(t *testing.T) {
	plan, err := PlanBalancedSplit(10, 4)
	if err != nil {
		t.Fatal(err)
	}

	wantSizes := []int64{4, 3, 3}
	if len(plan.Parts) != len(wantSizes) {
		t.Fatalf("parts = %d, want %d", len(plan.Parts), len(wantSizes))
	}
	for i, want := range wantSizes {
		if plan.Parts[i].Size != want {
			t.Fatalf("part %d size = %d, want %d", i, plan.Parts[i].Size, want)
		}
	}
	if plan.Parts[2].Offset != 7 {
		t.Fatalf("last offset = %d, want 7", plan.Parts[2].Offset)
	}
}

func TestPlanBalancedSplitRejectsBadInputs(t *testing.T) {
	if _, err := PlanBalancedSplit(-1, 1); err == nil {
		t.Fatal("expected error for negative original size")
	}
	if _, err := PlanBalancedSplit(1, 0); err == nil {
		t.Fatal("expected error for zero max part size")
	}
}

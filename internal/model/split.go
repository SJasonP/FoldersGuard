package model

import "fmt"

type SplitPlan struct {
	OriginalSize int64
	MaxPartSize  int64
	Parts        []PartSpan
}

type PartSpan struct {
	Index  int
	Offset int64
	Size   int64
}

func PlanBalancedSplit(originalSize, maxPartSize int64) (SplitPlan, error) {
	if originalSize < 0 {
		return SplitPlan{}, fmt.Errorf("original size must be non-negative")
	}
	if maxPartSize <= 0 {
		return SplitPlan{}, fmt.Errorf("max part size must be positive")
	}
	if originalSize == 0 || originalSize <= maxPartSize {
		return SplitPlan{
			OriginalSize: originalSize,
			MaxPartSize:  maxPartSize,
			Parts: []PartSpan{{
				Index:  0,
				Offset: 0,
				Size:   originalSize,
			}},
		}, nil
	}

	partCount := int((originalSize + maxPartSize - 1) / maxPartSize)
	baseSize := originalSize / int64(partCount)
	remainder := originalSize % int64(partCount)

	parts := make([]PartSpan, 0, partCount)
	var offset int64
	for i := range partCount {
		size := baseSize
		if int64(i) < remainder {
			size++
		}
		parts = append(parts, PartSpan{
			Index:  i,
			Offset: offset,
			Size:   size,
		})
		offset += size
	}

	return SplitPlan{
		OriginalSize: originalSize,
		MaxPartSize:  maxPartSize,
		Parts:        parts,
	}, nil
}

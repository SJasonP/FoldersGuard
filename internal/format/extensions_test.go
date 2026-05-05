package format

import "testing"

func TestExtensionForShape(t *testing.T) {
	tests := []struct {
		name  string
		shape DatabaseShape
		want  string
	}{
		{
			name:  "single top directory project",
			shape: DatabaseShape{TopLevelObjects: 1, SingleTopIsDir: true},
			want:  ProjectExtension,
		},
		{
			name:  "single top file",
			shape: DatabaseShape{TopLevelObjects: 1, SingleTopIsDir: false},
			want:  SetExtension,
		},
		{
			name:  "multiple tops",
			shape: DatabaseShape{TopLevelObjects: 2, SingleTopIsDir: true},
			want:  SetExtension,
		},
		{
			name:  "shared single top directory",
			shape: DatabaseShape{TopLevelObjects: 1, SingleTopIsDir: true, FromShare: true},
			want:  SetExtension,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ExtensionForShape(tt.shape); got != tt.want {
				t.Fatalf("ExtensionForShape() = %q, want %q", got, tt.want)
			}
		})
	}
}

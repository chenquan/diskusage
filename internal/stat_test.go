package internal

import (
	"testing"

	"github.com/jedib0t/go-pretty/v6/list"
)

func TestBuildInfoFileReverse(t *testing.T) {
	files := []*file{
		{name: "large", size: 3, print: true},
		{name: "medium", size: 2, print: true},
		{name: "small", size: 1, print: true},
	}

	tests := []struct {
		name    string
		reverse bool
		want    []float64
	}{
		{name: "descending by default", want: []float64{3, 2, 1}},
		{name: "ascending when reversed", reverse: true, want: []float64{1, 2, 3}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			infoFiles := buildInfoFile(list.NewWriter(), files, 0, 1, "B", 6, false, tt.reverse)
			if len(infoFiles) != len(tt.want) {
				t.Fatalf("got %d files, want %d", len(infoFiles), len(tt.want))
			}

			for i, want := range tt.want {
				if infoFiles[i].size != want {
					t.Errorf("infoFiles[%d].size = %v, want %v", i, infoFiles[i].size, want)
				}
			}
		})
	}
}

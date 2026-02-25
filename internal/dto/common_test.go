package dto

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPaginationParams_Normalize(t *testing.T) {
	tests := []struct {
		name        string
		input       PaginationParams
		wantPage    int
		wantPerPage int
	}{
		{
			name:        "zero values get defaults",
			input:       PaginationParams{Page: 0, PerPage: 0},
			wantPage:    1,
			wantPerPage: 20,
		},
		{
			name:        "negative page gets default",
			input:       PaginationParams{Page: -5, PerPage: 10},
			wantPage:    1,
			wantPerPage: 10,
		},
		{
			name:        "negative per_page gets default",
			input:       PaginationParams{Page: 1, PerPage: -10},
			wantPage:    1,
			wantPerPage: 20,
		},
		{
			name:        "both negative get defaults",
			input:       PaginationParams{Page: -1, PerPage: -1},
			wantPage:    1,
			wantPerPage: 20,
		},
		{
			name:        "per_page capped at 100",
			input:       PaginationParams{Page: 1, PerPage: 200},
			wantPage:    1,
			wantPerPage: 100,
		},
		{
			name:        "per_page exactly 100 stays",
			input:       PaginationParams{Page: 1, PerPage: 100},
			wantPage:    1,
			wantPerPage: 100,
		},
		{
			name:        "per_page 101 capped to 100",
			input:       PaginationParams{Page: 1, PerPage: 101},
			wantPage:    1,
			wantPerPage: 100,
		},
		{
			name:        "valid values stay unchanged",
			input:       PaginationParams{Page: 3, PerPage: 25},
			wantPage:    3,
			wantPerPage: 25,
		},
		{
			name:        "page 1 per_page 1",
			input:       PaginationParams{Page: 1, PerPage: 1},
			wantPage:    1,
			wantPerPage: 1,
		},
		{
			name:        "large page number stays",
			input:       PaginationParams{Page: 9999, PerPage: 50},
			wantPage:    9999,
			wantPerPage: 50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := tt.input
			p.Normalize()
			assert.Equal(t, tt.wantPage, p.Page, "page")
			assert.Equal(t, tt.wantPerPage, p.PerPage, "per_page")
		})
	}
}

func TestPaginationParams_Offset(t *testing.T) {
	tests := []struct {
		name       string
		page       int
		perPage    int
		wantOffset int
	}{
		{
			name:       "page 1 has offset 0",
			page:       1,
			perPage:    20,
			wantOffset: 0,
		},
		{
			name:       "page 2 with per_page 20",
			page:       2,
			perPage:    20,
			wantOffset: 20,
		},
		{
			name:       "page 3 with per_page 10",
			page:       3,
			perPage:    10,
			wantOffset: 20,
		},
		{
			name:       "page 5 with per_page 50",
			page:       5,
			perPage:    50,
			wantOffset: 200,
		},
		{
			name:       "page 1 with per_page 1",
			page:       1,
			perPage:    1,
			wantOffset: 0,
		},
		{
			name:       "page 100 with per_page 100",
			page:       100,
			perPage:    100,
			wantOffset: 9900,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &PaginationParams{Page: tt.page, PerPage: tt.perPage}
			assert.Equal(t, tt.wantOffset, p.Offset())
		})
	}
}

func TestPaginationParams_NormalizeThenOffset(t *testing.T) {
	t.Run("normalize then offset with defaults", func(t *testing.T) {
		p := &PaginationParams{Page: 0, PerPage: 0}
		p.Normalize()
		assert.Equal(t, 0, p.Offset(), "page 1 offset is 0")
	})

	t.Run("normalize then offset page 3", func(t *testing.T) {
		p := &PaginationParams{Page: 3, PerPage: 0}
		p.Normalize()
		// Page 3, PerPage defaults to 20
		assert.Equal(t, 40, p.Offset())
	})
}

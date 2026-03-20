package tui

import (
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestLayoutModeBreakpoint(t *testing.T) {
	tests := []struct {
		name  string
		width int
		want  layoutMode
	}{
		{"wide terminal selects horizontal", 120, layoutHorizontal},
		{"exactly at breakpoint selects horizontal", 80, layoutHorizontal},
		{"narrow terminal selects stacked", 79, layoutStacked},
		{"very narrow terminal selects stacked", 40, layoutStacked},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewApp(nil)
			updated, _ := m.Update(tea.WindowSizeMsg{Width: tt.width, Height: 40})
			got := updated.(appModel).layout
			if got != tt.want {
				t.Errorf("width=%d: layout = %d, want %d", tt.width, got, tt.want)
			}
		})
	}
}

func TestStackedSidebarHeight(t *testing.T) {
	tests := []struct {
		name             string
		totalHeight      int
		fileCount        int
		contentItemCount int
		want             int
	}{
		{"clamps to minimum 4", 50, 0, 0, 4},
		{"uses file count plus header", 50, 6, 0, 7},
		{"includes content items", 50, 3, 3, 7},
		{"clamps to maximum 10", 50, 15, 5, 10},
		{"caps at 40% of total height", 20, 15, 0, 8},
		{"40% cap doesn't go below min 4", 8, 0, 0, 4},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stackedSidebarHeight(tt.totalHeight, tt.fileCount, tt.contentItemCount)
			if got != tt.want {
				t.Errorf("stackedSidebarHeight(%d, %d, %d) = %d, want %d",
					tt.totalHeight, tt.fileCount, tt.contentItemCount, got, tt.want)
			}
		})
	}
}

func TestWidthAllocationHorizontal(t *testing.T) {
	m := NewApp(nil)
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	app := updated.(appModel)

	if app.layout != layoutHorizontal {
		t.Fatalf("expected horizontal layout at width 120")
	}

	// Sidebar should be clamped to [30, 50]
	if app.sidebar.width < 30 || app.sidebar.width > 50 {
		t.Errorf("sidebar.width = %d, want [30, 50]", app.sidebar.width)
	}

	// Diff view should get the remaining space
	sidebarOuter := app.sidebar.width + 2 // border
	expectedDiffW := 120 - sidebarOuter - 2
	if app.diffView.width != expectedDiffW {
		t.Errorf("diffView.width = %d, want %d", app.diffView.width, expectedDiffW)
	}
}

func TestWidthAllocationStacked(t *testing.T) {
	m := NewApp(nil)
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 60, Height: 40})
	app := updated.(appModel)

	if app.layout != layoutStacked {
		t.Fatalf("expected stacked layout at width 60")
	}

	expectedW := 60 - 2 // full width minus border
	if app.sidebar.width != expectedW {
		t.Errorf("sidebar.width = %d, want %d", app.sidebar.width, expectedW)
	}
	if app.diffView.width != expectedW {
		t.Errorf("diffView.width = %d, want %d", app.diffView.width, expectedW)
	}
}

func TestLayoutTransitionOnResize(t *testing.T) {
	m := NewApp(nil)

	// Start wide → horizontal
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	app := updated.(appModel)
	if app.layout != layoutHorizontal {
		t.Fatal("expected horizontal at width 120")
	}

	// Resize narrow → stacked
	updated, _ = app.Update(tea.WindowSizeMsg{Width: 60, Height: 40})
	app = updated.(appModel)
	if app.layout != layoutStacked {
		t.Fatal("expected stacked at width 60")
	}

	// Resize wide again → horizontal
	updated, _ = app.Update(tea.WindowSizeMsg{Width: 100, Height: 40})
	app = updated.(appModel)
	if app.layout != layoutHorizontal {
		t.Fatal("expected horizontal at width 100")
	}
}

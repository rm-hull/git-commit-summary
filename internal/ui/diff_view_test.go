package ui

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
)

func TestDiffViewModel_DiffHotkeyTogglesViews(t *testing.T) {
	view := initialDiffViewModel(72, 20)
	ctrlD := tea.KeyPressMsg(tea.Key{Code: 'd', Mod: tea.ModCtrl})

	view.Update(diffColorMsg("diff --git a/file.go b/file.go"))
	_, cmd := view.Update(ctrlD)
	assert.NotNil(t, cmd)
	assert.IsType(t, showDiffSummaryViewMsg{}, cmd())

	view.Update(diffSummaryMsg("file.go | 2 +-"))
	_, cmd = view.Update(ctrlD)
	assert.NotNil(t, cmd)
	assert.IsType(t, showDiffViewMsg{}, cmd())

	view.Update(diffColorMsg("diff --git a/file.go b/file.go"))
	assert.Contains(t, view.View().Content, "Raw diff")
	view.Update(diffSummaryMsg("file.go | 2 +-"))
	assert.Contains(t, view.View().Content, "Compact summary")
}

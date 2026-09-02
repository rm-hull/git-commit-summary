package help

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRenderMarkdownHelp(t *testing.T) {
	rendered := RenderMarkdownHelp("# Git Commit Summary\n\n- one\n- two", 72)

	assert.NotEmpty(t, rendered)
	assert.Contains(t, rendered, "Git")
	assert.Contains(t, rendered, "Summary")
	assert.Contains(t, rendered, "one")
	assert.Contains(t, rendered, "two")
	assert.Contains(t, rendered, "\x1b[")
}

package ui

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHistory(t *testing.T) {
	t.Run("NewHistory", func(t *testing.T) {
		h := NewHistory("initial")
		assert.Equal(t, "initial", h.Value(), "NewHistory should initialize with the given value")
		assert.Equal(t, 0, h.index, "NewHistory should initialize index to 0")
		assert.Len(t, h.stack, 1, "NewHistory should have one item in stack")
	})

	t.Run("Add and Value", func(t *testing.T) {
		h := NewHistory("initial")
		h.Add("second")
		assert.Equal(t, "second", h.Value(), "Value should return the last added item")
		assert.Equal(t, 1, h.index, "Add should increment index")
		assert.Len(t, h.stack, 2, "Add should add item to stack")
	})

	t.Run("Undo", func(t *testing.T) {
		h := NewHistory("initial")
		h.Add("second")
		val, ok := h.Undo()
		assert.True(t, ok, "Undo should be successful")
		assert.Equal(t, "initial", val, "Undo should return the previous value")
		assert.Equal(t, "initial", h.Value(), "Value should reflect the undone state")
		assert.Equal(t, 0, h.index, "Undo should decrement index")
	})

	t.Run("Redo", func(t *testing.T) {
		h := NewHistory("initial")
		h.Add("second")
		h.Undo()
		val, ok := h.Redo()
		assert.True(t, ok, "Redo should be successful")
		assert.Equal(t, "second", val, "Redo should return the next value")
		assert.Equal(t, "second", h.Value(), "Value should reflect the redone state")
		assert.Equal(t, 1, h.index, "Redo should increment index")
	})

	t.Run("Undo at beginning", func(t *testing.T) {
		h := NewHistory("initial")
		val, ok := h.Undo()
		assert.False(t, ok, "Undo should fail at the beginning")
		assert.Equal(t, "initial", val, "Undo at beginning should return current value")
		assert.Equal(t, "initial", h.Value(), "Value should remain unchanged")
		assert.Equal(t, 0, h.index, "Index should remain 0")
	})

	t.Run("Redo at end", func(t *testing.T) {
		h := NewHistory("initial")
		h.Add("second")
		val, ok := h.Redo()
		assert.False(t, ok, "Redo should fail at the end")
		assert.Equal(t, "second", val, "Redo at end should return current value")
		assert.Equal(t, "second", h.Value(), "Value should remain unchanged")
		assert.Equal(t, 1, h.index, "Index should remain at max")
	})

	t.Run("Add truncates future history", func(t *testing.T) {
		h := NewHistory("initial")
		h.Add("second")
		h.Add("third")
		h.Undo() // back to "second"
		h.Add("new third")

		assert.Equal(t, "new third", h.Value(), "Value should be the newly added item")
		assert.Equal(t, 2, h.index, "Index should be updated after Add")
		assert.Len(t, h.stack, 3, "Stack length should be correct after truncation")

		// Redo should not be possible
		val, ok := h.Redo()
		assert.False(t, ok, "Redo should fail after adding a new value")
		assert.Equal(t, "new third", val, "Redo should return current value")

		// Undo should go to "second"
		val, ok = h.Undo()
		assert.True(t, ok, "Undo should be successful")
		assert.Equal(t, "second", val, "Undo should return 'second'")
		assert.Equal(t, "second", h.Value(), "Value should be 'second'")
	})

	t.Run("Multiple Undos and Redos", func(t *testing.T) {
		h := NewHistory("1")
		h.Add("2")
		h.Add("3")
		h.Add("4")

		val, ok := h.Undo() // 3
		assert.True(t, ok)
		assert.Equal(t, "3", val)
		assert.Equal(t, "3", h.Value())

		val, ok = h.Undo() // 2
		assert.True(t, ok)
		assert.Equal(t, "2", val)
		assert.Equal(t, "2", h.Value())

		val, ok = h.Redo() // 3
		assert.True(t, ok)
		assert.Equal(t, "3", val)
		assert.Equal(t, "3", h.Value())

		val, ok = h.Redo() // 4
		assert.True(t, ok)
		assert.Equal(t, "4", val)
		assert.Equal(t, "4", h.Value())

		_, ok = h.Redo() // end
		assert.False(t, ok)

		val, ok = h.Undo() // 3
		assert.True(t, ok)
		assert.Equal(t, "3", val)

		h.Add("5") // truncates 4, adds 5
		assert.Equal(t, "5", h.Value())
		assert.Len(t, h.stack, 4) // 1, 2, 3, 5
		assert.Equal(t, 3, h.index)

		_, ok = h.Redo() // end
		assert.False(t, ok)

		val, ok = h.Undo() // 3
		assert.True(t, ok)
		assert.Equal(t, "3", val)
	})
}

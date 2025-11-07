package ui

type History struct {
	stack []string
	index int
}

func NewHistory(initialValue string) *History {
	return &History{
		stack: []string{initialValue},
		index: 0,
	}
}

func (h *History) Undo() (string, bool) {
	if h.index > 0 {
		h.index--
		return h.stack[h.index], true
	}
	return h.stack[h.index], false
}

func (h *History) Redo() (string, bool) {
	if h.index < len(h.stack)-1 {
		h.index++
		return h.stack[h.index], true
	}
	return h.stack[h.index], false
}

func (h *History) Add(value string) {
	if h.index < len(h.stack)-1 {
		h.stack = h.stack[:h.index+1] // truncate the future history
	}
	h.stack = append(h.stack, value)
	h.index++
}

func (h *History) Value() string {
	return h.stack[h.index]
}

package ui

import (
	"time"

	"github.com/briandowns/spinner"
	"github.com/gookit/color"
)

type Client struct {
	spinner *spinner.Spinner
}

func (c *Client) TextArea(value string) (string, bool, error) {
	return textArea(value)
}

func (c *Client) StartSpinner(message string) {
	c.spinner = spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	c.spinner.Suffix = color.Render(message)
	c.spinner.Start()
}

func (c *Client) UpdateSpinner(message string) {
	if c.spinner != nil {
		c.spinner.Suffix = color.Render(message)
	}
}

func (c *Client) StopSpinner() {
	if c.spinner != nil {
		c.spinner.Stop()
	}
}

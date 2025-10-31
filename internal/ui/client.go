package ui

import (
	"time"

	"github.com/briandowns/spinner"
	"github.com/gookit/color"
)

type Client struct {
	spinner *spinner.Spinner
}

func NewClient() *Client {
	return &Client{
		spinner: spinner.New(spinner.CharSets[14], 100*time.Millisecond),
	}
}

func (c *Client) TextArea(value string) (string, bool, error) {
	return textArea(value)
}

func (c *Client) StartSpinner(message string) {
	c.spinner.Suffix = color.Render(message)
	c.spinner.Start()
}

func (c *Client) UpdateSpinner(message string) {
	c.spinner.Suffix = color.Render(message)
}

func (c *Client) StopSpinner() {
	c.spinner.Stop()
}

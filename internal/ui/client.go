package ui

type Client struct{}

func (c *Client) TextArea(value string) (string, bool, error) {
	return textArea(value)
}

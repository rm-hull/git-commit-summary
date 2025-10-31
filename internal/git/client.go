package git

type Client struct{}

func (c *Client) Diff() (string, error) {
	out, err := diff()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func (c *Client) Commit(message string) error {
	return commit(message)
}

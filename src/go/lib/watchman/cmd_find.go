package watchman

// Find lists all files that match the optional patterns under the specified dir.
// When no patterns are provided, all files are returned.
// https://facebook.github.io/watchman/docs/cmd/find.html
func (client *Client) Find(opts FindOptions) (resp FindResponse, err error) {
	rawCmd := RawPDUCommand{"find", opts.Directory}

	for _, pattern := range opts.Patterns {
		rawCmd = append(rawCmd, pattern)
	}

	err = client.Exec(rawCmd, &resp)
	return
}

// FindResponse
// https://facebook.github.io/watchman/docs/cmd/find.html
type FindResponse struct {
	Version string `json:"version" mapstructure:"version"`
	Clock   string `json:"clock" mapstructure:"clock"`
	Files   []File `json:"files" mapstructure:"files"`
}

// FindOptions
// https://facebook.github.io/watchman/docs/cmd/find.html
type FindOptions struct {
	Directory string
	Patterns  []string
}

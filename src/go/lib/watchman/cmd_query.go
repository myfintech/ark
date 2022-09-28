package watchman

// Query executes a query against the specified root.
// https://facebook.github.io/watchman/docs/cmd/query.html
func (client *Client) Query(opts QueryOptions) (resp QueryResponse, err error) {
	rawCmd := RawPDUCommand{"query", opts.Directory}

	if opts.Filter != nil {
		rawCmd = append(rawCmd, opts.Filter)
	}
	err = client.Exec(rawCmd, &resp)
	return
}

// QueryResponse
// https://facebook.github.io/watchman/docs/cmd/find.html
type QueryResponse struct {
	Version         string `json:"version" mapstructure:"version"`
	Clock           string `json:"clock" mapstructure:"clock"`
	IsFreshInstance bool   `json:"is_fresh_instance" mapstructure:"is_fresh_instance"`
	Files           []File `json:"files" mapstructure:"files"`
}

// QueryOptions
// https://facebook.github.io/watchman/docs/cmd/find.html
type QueryOptions struct {
	Directory string
	Filter    *QueryFilter
}

// QueryFilter the optional query parameters for a query
type QueryFilter struct {
	Since               string        `json:"since,omitempty" mapstructure:"since"`                               // The clock to use for querying changes
	Fields              []string      `json:"fields,omitempty" mapstructure:"fields"`                             // a list of fields to project file objects
	Expression          []interface{} `json:"expression,omitempty" mapstructure:"expression"`                     // any valid expression to query the file system
	Glob                []string      `json:"glob,omitempty" mapstructure:"glob"`                                 // https://facebook.github.io/watchman/docs/file-query.html#glob-generator
	DeferVcs            bool          `json:"defer_vcs,omitempty" mapstructure:"defer_vcs"`                       // https://facebook.github.io/watchman/docs/cmd/subscribe.html#filesystem-settling
	Defer               []string      `json:"defer,omitempty" mapstructure:"defer"`                               // https://facebook.github.io/watchman/docs/cmd/subscribe.html#defer
	Drop                []string      `json:"drop,omitempty" mapstructure:"drop"`                                 // https://facebook.github.io/watchman/docs/cmd/subscribe.html#drop
	DedupResults        bool          `json:"dedup_results,omitempty" mapstructure:"dedup_results"`               // https://facebook.github.io/watchman/docs/file-query.html#de-duplicating-results
	GlobIncludeDotFiles bool          `json:"glob_includedotfiles,omitempty" mapstructure:"glob_includedotfiles"` // https://github.com/facebook/watchman/issues/647#issuecomment-422252118
}

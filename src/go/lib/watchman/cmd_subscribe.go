package watchman

import (
	"golang.org/x/net/context"
)

/*
Subscribe
The updates will continue to be sent while the connection is open. If the connection is closed, the subscription is implicitly removed.
https://facebook.github.io/watchman/docs/cmd/subscribe.html

In some circumstances it is desirable for a client to observe the creation of the control files at the start of a version control operation. You may specify that you want this behavior by passing the defer_vcs flag to your subscription command invocation
https://facebook.github.io/watchman/docs/cmd/subscribe.html#filesystem-settling

Some applications will want to know that the update is in progress and continue to process notifications. Others may want to defer processing the notifications until the update completes, and some may wish to drop any notifications produced while the update was in progress.
https://facebook.github.io/watchman/docs/cmd/subscribe.html#advanced-settling
*/
func Subscribe(opts SubscribeOptions, cancelOnFirstError bool, ignoreInitialResponse bool) (sub *Subscription, err error) {
	client, err := Connect(context.Background(), 10)

	if err != nil {
		return
	}

	rawCmd := RawPDUCommand{"subscribe", opts.Root, opts.Name}

	if opts.Filter != nil {
		rawCmd = append(rawCmd, opts.Filter)
	}

	err = client.Send(rawCmd)
	if err != nil {
		return
	}

	_, err = client.Receive()
	if err != nil {
		return
	}

	// log.Println(utils.MarshalJSONSafe(resp, true))

	sub = &Subscription{
		client:                client,
		Cancel:                make(chan bool, 1),
		Errors:                make(chan error),
		ChangeFeed:            make(chan SubscriptionChangeFeedResponse),
		CancelOnFirstError:    cancelOnFirstError,
		IgnoreInitialResponse: ignoreInitialResponse,
		RawCommand:            rawCmd,
	}

	go subscriptionChangeFeedWorker(sub)

	return
}

// SubscribeOptions
// https://facebook.github.io/watchman/docs/cmd/subscribe.html
type SubscribeOptions struct {
	Name   string `json:"name" mapstructure:"name"` // the name of the subscription used to provide unilateral responses
	Root   string `json:"root" mapstructure:"root"` // the relative root to subscribe for changes
	Filter *QueryFilter
}

// SubscriptionChangeFeedResponse
// https://facebook.github.io/watchman/docs/cmd/subscribe.html
type SubscriptionChangeFeedResponse struct {
	Version         string `json:"version" mapstructure:"version"`
	Unilateral      bool   `json:"unilateral" mapstructure:"unilateral"`
	Clock           string `json:"clock" mapstructure:"clock"`
	Files           []File `json:"files" mapstructure:"files"`
	Root            string `json:"root" mapstructure:"root"`
	Subscription    string `json:"subscription" mapstructure:"subscription"`
	IsFreshInstance bool   `json:"is_fresh_instance" mapstructure:"is_fresh_instance"`
}

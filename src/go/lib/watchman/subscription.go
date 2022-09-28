package watchman

func subscriptionChangeFeedWorker(sub *Subscription) {
	defer func() { _ = sub.client.Close() }()
	var ingoredInitResponse bool
	for {
		resp, err := sub.client.Receive()
		if err != nil {
			sub.PushError(err)
			continue
		}

		if sub.IgnoreInitialResponse && !ingoredInitResponse {
			ingoredInitResponse = true
			continue
		}

		change := SubscriptionChangeFeedResponse{}
		err = resp.Decode(&change)
		if err != nil {
			sub.PushError(err)
			continue
		}

		sub.PushChange(change)

		if sub.Canceled() {
			break
		}
	}
}

// Subscription a sub
type Subscription struct {
	client                *Client
	Cancel                chan bool
	Errors                chan error
	ChangeFeed            chan SubscriptionChangeFeedResponse
	CancelOnFirstError    bool
	IgnoreInitialResponse bool
	RawCommand            RawPDUCommand
}

// OnErrorNotification a callback that can handle watchman socket errors
type OnErrorNotification func(err error)

// OnErrorNotification executes the callback function in a go routine for every error returned from the socket
func (sub *Subscription) OnErrorNotification(onError OnErrorNotification) {
	go func() {
		for err := range sub.Errors {
			onError(err)
		}
	}()
}

// OnChangeNotification a callback that can handle change feed notifications
type OnChangeNotification func(change SubscriptionChangeFeedResponse)

// OnChangeNotification executes the callback function in a go routine for every change detected on the file system
func (sub *Subscription) OnChangeNotification(onChange OnChangeNotification) {
	go func() {
		for change := range sub.ChangeFeed {
			onChange(change)
		}
	}()
}

// Wait synchronously blocks awaiting the cancellation signal
func (sub *Subscription) Wait() bool {
	return <-sub.Cancel
}

// PushError sends an error over the errors channel
func (sub *Subscription) PushError(err error) bool {
	select {
	case sub.Errors <- err:
		if sub.CancelOnFirstError {
			sub.Unsubscribe()
		}
		return true
	default:
		return false
	}
}

// PushChange sends a change over the change feed channel
func (sub *Subscription) PushChange(change SubscriptionChangeFeedResponse) bool {
	select {
	case sub.ChangeFeed <- change:
		return true
	default:
		return false
	}
}

// Canceled checks if the subscription has been canceled
func (sub *Subscription) Canceled() bool {
	canceled := false
	select {
	case canceled = <-sub.Cancel:
		return canceled
	default:
		return canceled
	}
}

// Unsubscribe shuts down the subscription
func (sub *Subscription) Unsubscribe() bool {
	select {
	case sub.Cancel <- true:
		close(sub.Errors)
		close(sub.ChangeFeed)
		return true
	default:
		return false
	}
}

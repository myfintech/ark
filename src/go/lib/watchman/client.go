package watchman

import (
	"bufio"
	"encoding/json"
	"net"
	"sync"
	"time"

	"golang.org/x/net/context"

	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
)

// Client a client interface for the watchman conn protocol
type Client struct {
	conn   net.Conn
	mutex  sync.Mutex
	socket *bufio.ReadWriter

	version      string
	capabilities map[string]string
}

// Connect will attempt to connect to the watchman conn service for IPC
func Connect(ctx context.Context, timeoutSeconds int) (*Client, error) {
	client := &Client{}
	socketName, err := GetSocketName()
	if err != nil {
		return client, errors.Wrap(err, "failed to get socket name")
	}

	dialer := net.Dialer{
		Timeout: time.Duration(timeoutSeconds) * time.Second,
	}
	conn, err := dialer.DialContext(ctx, "unix", socketName)
	if err != nil {
		return client, errors.Wrap(err, "failed to connect to the socket service")
	}

	client.conn = conn
	client.socket = bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))
	return client, nil
}

// Close closes the connection to the watchman socket service
func (client *Client) Close() error {
	return client.conn.Close()
}

// RawPDUCommand is the raw structure required to serialize a command for the socket server
type RawPDUCommand []interface{}

// RawPDUResponse is the raw PDU response structure from the socket
type RawPDUResponse map[string]interface{}

// RawPDUCommandOptions is a raw PDU options object
type RawPDUCommandOptions map[string]interface{}

// Decode decodes the raw PDU response into a struct pointer
// Note: this will panic if you do not supply a pointer
func (resp RawPDUResponse) Decode(v interface{}) error {
	if err := mapstructure.Decode(resp, &v); err != nil {
		return errors.Wrap(err, "failed to decode PDU response into struct")
	}
	return nil
}

// Send issues a command to the watchman socket service
func (client *Client) Send(rawCmd RawPDUCommand) error {
	client.mutex.Lock()
	defer client.mutex.Unlock()

	// FIXME: increase performance with a streaming JSON implementation
	jsonBytes, err := json.Marshal(rawCmd)
	if err != nil {
		return errors.Wrap(err, "failed to marshal rawCmd into a JSON PDU")
	}

	if _, err = client.conn.Write(jsonBytes); err != nil {
		return errors.Wrap(err, "failed to send JSON PDU to socket service")
	}

	if _, err = client.conn.Write([]byte("\n")); err != nil {
		return errors.Wrap(err, "failed to write command termination after JSON PDU to the socket service")
	}

	return nil
}

// Receive attempts to read a single JSON encoded PDU response line from the watchman socket service
// if the response contains an error key it will raise an error with its value
// if the response contains an error key, and it's value is a string this process will panic
func (client *Client) Receive() (resp RawPDUResponse, err error) {
	client.mutex.Lock()
	defer client.mutex.Unlock()

	// FIXME: increase performance with a streaming JSON implementation
	// decoder := json.NewDecoder(client.socket)
	line, err := client.socket.ReadBytes('\n')
	if err != nil {
		return nil, err
	}
	if err = json.Unmarshal(line, &resp); err != nil {
		return resp, errors.Wrapf(err, "failed to decode JSON PDU response from socket %v", resp)
	}

	if errValue, ok := resp["error"]; ok {
		return resp, errors.New(errValue.(string))
	}
	return
}

// Exec executes a command on the server synchronously and marshals the response into the given struct
// This function will raise an error if you do not supply for the destResp
func (client *Client) Exec(cmd RawPDUCommand, destResp interface{}) error {
	err := client.Send(cmd)

	if err != nil {
		return err
	}

	resp, err := client.Receive()
	if err != nil {
		return err
	}

	if err = resp.Decode(destResp); err != nil {
		return err
	}

	return nil
}

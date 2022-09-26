package watchman

import (
	"encoding/json"
	"os"
	"os/exec"
	"reflect"
	"sync"

	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
)

var watchmanBinaryPath string
var watchmanContext sync.Once

// Path returns a string containing the path to the watchman binary
func Path() (string, error) {
	return exec.LookPath("watchman")
}

// PathOnce cached the path to the watchman binary
func PathOnce() (path string, err error) {
	watchmanContext.Do(func() {
		path, err = exec.LookPath("watchman")
		watchmanBinaryPath = path
	})
	return watchmanBinaryPath, err
}

// IsBinaryInstalled checks if watchman is visible in the os path
func IsBinaryInstalled() bool {
	watchmanPath, _ := Path()
	return watchmanPath != ""
}

// GetSocketName returns a strong containing the location of the watchman conn
func GetSocketName() (string, error) {
	socketOverride := os.Getenv("WATCHMAN_SOCK")
	if socketOverride != "" {
		return socketOverride, nil
	}

	resp := struct {
		SocketName string `mapstructure:"sockname"`
	}{}

	wmPath, err := PathOnce()
	if err != nil {
		return resp.SocketName, err
	}

	cmd := exec.Command(wmPath, "get-sockname")

	if err := JSONUnmarshalCommand(cmd, &resp); err != nil {
		return resp.SocketName, err
	}

	return resp.SocketName, nil
}

/*
JSONUnmarshalCommand
accepts a cmd pointer, and an arbitrary interface executes that command and attempts to unmarshal the output as JSON
if the command returns a JSON value that contains an error key it will attempt to raise a new error with the string value of that error key
	if the error key contains a value other than a string this function will panic
*/
func JSONUnmarshalCommand(cmd *exec.Cmd, v interface{}) error {
	output, err := cmd.Output()

	if reflect.ValueOf(v).Kind() != reflect.Ptr {
		return errors.Errorf("Value of %v must be a pointer", v)
	}

	if err != nil {
		return errors.Wrapf(err, "failed to execute %s", cmd.String())
	}

	var result map[string]interface{}
	if err = json.Unmarshal(output, &result); err != nil {
		return errors.Wrapf(err, "failed to parse JSON from %s", cmd.String())
	}

	if errValue, ok := result["error"]; ok {
		errString := errValue.(string)
		return errors.New(errString)
	}

	if err = mapstructure.Decode(result, &v); err != nil {
		return errors.Wrapf(err, "failed to parse JSON from %s", cmd.String())
	}

	return nil
}

package utils

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"

	"github.com/myfintech/ark/src/go/lib/log"
)

// EnvLookup looks up env vars, returns default string if not found
func EnvLookup(key, defaultValue string) string {
	envVar, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}
	return envVar
}

// EnvLookupInt looks up env vars, returns default string if not found

// EnvLookupAll returns a map[string]string of all environment variables
func EnvLookupAll() map[string]string {
	env := make(map[string]string)
	vars := os.Environ()
	for _, v := range vars {
		pair := strings.Split(v, "=")
		env[pair[0]] = pair[1]
	}
	return env
}

func MarshalJSON(obj interface{}, explode bool) (string, error) {

	if explode {
		bytes, err := json.MarshalIndent(obj, "", "  ")

		if err != nil {
			return "", err
		}

		return string(bytes), nil
	}

	bytes, err := json.Marshal(obj)

	if err != nil {
		return "", err
	}

	return string(bytes), nil
}

func MarshalJSONSafe(obj interface{}, explode bool) string {
	data, err := MarshalJSON(obj, explode)
	if err != nil {
		log.Error(errors.Wrap(err, "Failed to marshal JSON"))
		return ""
	}
	return data
}

// SingleJoiningSlash joins two strings and returns what its name implies

// MergeMaps merges n number of maps[string]interface
func MergeMaps(maps ...map[string]interface{}) map[string]interface{} {
	merged := map[string]interface{}{}
	for _, m := range maps {
		for k, v := range m {
			merged[k] = v
		}
	}
	return merged
}

// MergeMapStringString merges n number of map[string]string

func EncodeBytesToString(data []byte, format string) (string, error) {
	switch format {
	case "raw":
		return string(data), nil
	case "hex":
		return hex.EncodeToString(data), nil
	case "base64":
		return base64.StdEncoding.EncodeToString(data), nil
	default:
		return "", errors.Errorf("invalid format (%s) must be one of (raw|hex|base64)", format)
	}
}

func DecodeStringToBytes(data, format string) ([]byte, error) {
	switch format {
	case "raw":
		return []byte(data), nil
	case "hex":
		return hex.DecodeString(data)
	case "base64":
		return base64.StdEncoding.DecodeString(data)
	default:
		return []byte{}, errors.Errorf("invalid format (%s) must be one of (raw|hex|base64)", format)
	}
}

func BasicAuth(username, password string) string {
	return base64.StdEncoding.EncodeToString([]byte(username + ":" + password))
}

func BytesToString(data []byte) string {
	return string(data)
}

func StringToBytes(data string) []byte {
	return []byte(data)
}

func UUIDV4() string {
	return uuid.NewV4().String()
}

func ISODateNow() string {
	return time.Now().Format(time.RFC3339)
}

// ExtractDebugInformationFromHTTPRequest be careful this might read the entire request body. It cannot not be read again

// CheckErrFatal
func CheckErrFatal(err error, message string) {
	if err != nil {
		log.Fatal(errors.Wrap(err, message))
	}
}

// GetHostname

func JoinStrings(separator string, stringValues ...string) string {
	return strings.Join(stringValues, separator)
}

// ExpandEnvOnArgs expands environmental variables from all strings in the arg list
func ExpandEnvOnArgs(args []string) []string {
	expanded := make([]string, len(args))
	// Allows the CLI tool to be the entry point and still expand $ENV vars line in bash
	for i, arg := range args {
		expanded[i] = os.ExpandEnv(arg)
	}
	return expanded
}

// BuildDeepMapString uses a slice of keys to build an n dimensional map[string]interface{}
// Be aware, providing an empty keys slice will result in a noop function
func BuildDeepMapString(val interface{}, keys []string, level map[string]interface{}) {
	if len(keys) == 0 {
		return
	}

	if len(keys) == 1 {
		level[keys[0]] = val
		return
	}

	if _, levelExists := level[keys[0]]; !levelExists {
		level[keys[0]] = make(map[string]interface{})
	}

	BuildDeepMapString(val, keys[1:], level[keys[0]].(map[string]interface{}))
}

// GetRuntime returns runtime info as a map[string]string
func GetRuntime() map[string]string {
	return map[string]string{
		"os":   runtime.GOOS,
		"arch": runtime.GOARCH,
	}
}

// MapToEnvStringSlice converts a map of environmental variables to a string slice of kv pairs
// Useful for interfaces that require this data structure like docker
func MapToEnvStringSlice(env map[string]string) []string {
	var envSlice []string
	for k, v := range env {
		envSlice = append(envSlice, fmt.Sprintf("%s=%s", k, v))
	}
	return envSlice
}

// IsK8sContextSafe return a slice of k8s safe contexts
func IsK8sContextSafe(defaultContexts []string, envKey, currentContext string) bool {
	safeContexts := append(defaultContexts, strings.Split(os.Getenv(envKey), ",")...)
	for _, c := range safeContexts {
		if currentContext == c {
			return true
		}
	}
	return false
}

// GetFreePort attempts to resolve localhost:0
// *nix operating systems will return an unused port instead of resolving port 0
func GetFreePort() (string, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return "0", err
	}
	listen, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return "0", nil
	}
	defer listen.Close()
	return strconv.Itoa(listen.Addr().(*net.TCPAddr).Port), nil
}

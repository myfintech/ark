package templateutils

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"text/template"

	"github.com/myfintech/ark/src/go/lib/utils"
	"github.com/myfintech/ark/src/go/lib/utils/cryptoutils"
)

var ConvenienceFunctions = template.FuncMap{
	// stdlib
	"base64Encode": base64.StdEncoding.EncodeToString,
	"base64Decode": base64.StdEncoding.DecodeString,
	"hexEncode":    hex.EncodeToString,
	"hexDecode":    hex.DecodeString,

	// stdutils
	"basicAuth":     utils.BasicAuth,
	"bytesToString": utils.BytesToString,
	"stringToBytes": utils.StringToBytes,
	"encodeBytes":   utils.EncodeBytesToString,
	"decodeBytes":   utils.DecodeStringToBytes,
	"uuidv4":        utils.UUIDV4,
	"uuidV4":        utils.UUIDV4,
	"isoDateNow":    utils.ISODateNow,
	"join":          utils.JoinStrings,

	// cryptoutils
	"md5":              cryptoutils.MD5Sum,
	"sha1":             cryptoutils.SHA1Sum,
	"sha256":           cryptoutils.SHA256Sum,
	"aes256CBCEncrypt": cryptoutils.AES256CBCEncrypt,
	"aes256CBCDecrypt": cryptoutils.AES256CBCDecrypt,
	"randIV":           cryptoutils.RandomIV,
	"hmac256":          cryptoutils.HMAC256Sum,
}

// RenderTemplateAsString executes a template and returns a string
func RenderTemplateAsString(tmp *template.Template, data interface{}) (string, error) {
	buff := bytes.NewBufferString("")
	if err := tmp.Execute(buff, data); err != nil {
		return buff.String(), err
	}
	return buff.String(), nil
}

// RenderFunc is a function that accepts a template and its data, executes it and returns the string representation
// This is a convenience method to the standard template library which accepts an io.Writer instead of returning the rendered template
type RenderFunc func(data interface{}) (string, error)

// ParseFunc is a function that accepts the body of a template and returns a core.RenderFunc
type ParseFunc func(body string) RenderFunc

// NewTemplate is a curried function that allows for boot time template parsing

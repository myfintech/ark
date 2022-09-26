package tcputils

import (
	"bufio"
	"net"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/myfintech/ark/src/go/lib/log"
	"github.com/myfintech/ark/src/go/lib/utils"
)

func handler(conn net.Conn) {
	defer conn.Close()

	contextLogger := log.WithFields(log.Fields{
		"remote_address": conn.RemoteAddr().String(),
		"trace_id":       utils.UUIDV4(),
	})

	if err := conn.SetDeadline(time.Now().Add(time.Second * 5)); err != nil {
		contextLogger.Error(errors.Wrap(err, "failed to set deadline on connection"))
		return
	}

	// maxBytes := 1024
	// bytesRead := make([]byte, maxBytes)
	bytesCopied := 0
	reader := bufio.NewReader(conn)

	for {
		line, _, err := reader.ReadLine()
		length := len(line)
		contextLogger.WithField("length", length).Debugf("Read Line %s", line)

		if err != nil {
			contextLogger.Error(errors.Wrap(err, "failed to read bytes from connection"))
			return
		}

		if length == 0 {
			break
		}

		byteCount, err := conn.Write(line)

		if err != nil {
			contextLogger.Error(errors.Wrap(err, "failed to write bytes back to connection"))
			return
		}

		bytesCopied += byteCount

		contextLogger.Debugf("Wrote %d bytes", byteCount)

		if strings.HasSuffix(string(line), "\r") {
			contextLogger.Debugf("Found special line ending")
			break
		}

	}

	log.Infof("successfully echoed %d bytes", bytesCopied)

}

// ListenAndServe is used for unit testing TCP connections where no public server is available

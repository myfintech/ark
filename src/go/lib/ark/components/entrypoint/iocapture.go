package entrypoint

import (
	"bufio"
	"os"
	"time"

	"github.com/pkg/errors"

	"github.com/myfintech/ark/src/go/lib/utils"

	"github.com/myfintech/ark/src/go/lib/ark/components/log_sink_server"
	"github.com/myfintech/ark/src/go/lib/iomux"
	"github.com/myfintech/ark/src/go/lib/log"
)

// CaptureAllProcessLogs intercepts StdIO so it can be multiplexed to multiple writers
// one of which is sending log lines to the log sink server via the log sink client
func CaptureAllProcessLogs(readyChannel chan string, loggingClient log_sink_server.LogSink_RecordClient) error {
	defer close(readyChannel)
	r, w, err := os.Pipe()
	if err != nil {
		return err
	}

	pipeWriter, err := iomux.StdIOCapture(w)
	if err != nil {
		return err
	}

	log.SetOutput(pipeWriter)

	close(readyChannel)

	reader := bufio.NewReader(r)
	for {
		line, _, readErr := reader.ReadLine()
		if readErr != nil {
			return readErr
		}

		if err = loggingClient.Send(&log_sink_server.LogLine{
			TargetAddress: "test.thing.whatever",
			TargetHash:    "8675309",
			Data:          line,
			CreatedAt:     time.Now().Format(time.RFC3339),
			OrgId:         utils.EnvLookup("ARK_ORG_ID", "mantl"),
			ProjectId:     utils.EnvLookup("ARK_PROJECT_ID", "mantl"),
		}); err != nil {
			return errors.Wrap(err, "error sending log line to sink server")
		}

		if _, err = loggingClient.Recv(); err != nil {
			return errors.Wrap(err, "error received from gRPC server")
		}
	}
}

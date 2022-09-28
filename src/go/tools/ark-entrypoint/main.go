package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"time"

	"github.com/grpc-ecosystem/go-grpc-middleware/util/metautils"

	"google.golang.org/grpc/metadata"

	"github.com/myfintech/ark/src/go/lib/ark/components/log_sink_server"

	"github.com/myfintech/ark/src/go/lib/utils"

	"github.com/moby/buildkit/util/appcontext"

	"google.golang.org/grpc"

	"github.com/myfintech/ark/src/go/lib/log"

	"golang.org/x/sync/errgroup"

	"github.com/myfintech/ark/src/go/lib/ark/components/entrypoint"
)

func createNewGrpcServer(ctx context.Context, grpcServer *grpc.Server) error {
	grpcPort := utils.EnvLookup("ARK_EP_GRPC_PORT", "9000")
	lis, err := net.Listen("tcp", net.JoinHostPort("", grpcPort))
	if err != nil {
		return err
	}

	go func() {
		select {
		case <-ctx.Done():
			log.Infof("shutting down ark entrypoint gRPC server")
			timer := time.AfterFunc(time.Second*10, grpcServer.Stop)
			defer timer.Stop()
			defer log.Infof("grpc server shutdown")
			grpcServer.GracefulStop()
		}
	}()

	if err = grpcServer.Serve(lis); err != nil {
		return err
	}
	return nil
}

func main() {
	flag.Parse()
	eg, ctx := errgroup.WithContext(appcontext.Context())
	syncServer := entrypoint.New(flag.Args(), ctx)
	grpcServer := grpc.NewServer()
	entrypoint.RegisterSyncServer(grpcServer, syncServer)

	userToken := utils.EnvLookup("ARK_USER_TOKEN", "")
	sinkServerURL := utils.EnvLookup("ARK_LOG_SINK_URL", "")

	if userToken != "" && sinkServerURL != "" {
		log.Info("user token and log sink url are present; setting up log streaming")
		grpcMetadata := metadata.Pairs("authorization", fmt.Sprintf("bearer %s", userToken))
		recordCtx := metautils.NiceMD(grpcMetadata).ToOutgoing(ctx)

		logCtx, _ := context.WithTimeout(ctx, time.Second*10) // ignoring cancel func; will not cause context leak because the timeout will cancel the context
		logClientConnection, err := grpc.DialContext(logCtx, sinkServerURL, grpc.WithInsecure())
		if err != nil {
			log.Fatalf("there was an error creating a client connection to the log sink server: %v", err)
		}
		logClient := log_sink_server.NewLogSinkClient(logClientConnection)
		logRecordClient, err := logClient.Record(recordCtx)
		if err != nil {
			log.Fatalf("there was an error creating a client to record log events to the log sink server: %v", err)
		}

		logSetupChan := make(chan string)

		eg.Go(func() error {
			return entrypoint.CaptureAllProcessLogs(logSetupChan, logRecordClient)
		})

		<-logSetupChan
	}

	eg.Go(func() error {
		return createNewGrpcServer(ctx, grpcServer)
	})

	eg.Go(func() error {
		syncServer.Watch(ctx)
		return nil
	})

	if err := eg.Wait(); err != nil {
		log.Fatalf("there was an error from a Go routine: %s", err)
	}

	log.Info("ark entrypoint server shutdown was successful")
	return
}

func init() {
	log.SetFormat("json")
}

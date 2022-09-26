package grpc_jwt

import (
	"context"
	"fmt"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	grpc_testing "github.com/grpc-ecosystem/go-grpc-middleware/testing"
	pb_testproto "github.com/grpc-ecosystem/go-grpc-middleware/testing/testproto"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/grpc-ecosystem/go-grpc-middleware/util/metautils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

var (
	authToken = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxZjBkMzk5MC1lMmFmLTRmYzktOWI1ZC0yMGM4ZDliZjdkYmEiLCJpYXQiOjE2MDcwMTgwNTgsImV4cCI6MTY3MDA5MDEwOX0.m0hVC6ejYjxJk-akTWXPGV_B8wByoF9OLMuZDUsZ6Jk"
	subject   = "1f0d3990-e2af-4fc9-9b5d-20c8d9bf7dba"
	goodPing  = &pb_testproto.PingRequest{Value: "something", SleepTimeMs: 9999}
)

// All of these tests were pulled from the github.com/grpc-ecosystem/go-grpc-middleware/auth library
// I needed a way to test the AuthFunc we've writen to ensure it properly attaches the user_id to the context

func requireUserIDToExist(t *testing.T, ctx context.Context) {
	require.Equal(t, subject, ctx.Value(ContextKey("user_id")).(string), "user_id from CtxUpdateFunc must be passed around")
}

type assertingPingService struct {
	pb_testproto.TestServiceServer
	T *testing.T
}

func (s *assertingPingService) PingError(ctx context.Context, ping *pb_testproto.PingRequest) (*pb_testproto.Empty, error) {
	requireUserIDToExist(s.T, ctx)
	return s.TestServiceServer.PingError(ctx, ping)
}

func (s *assertingPingService) PingList(ping *pb_testproto.PingRequest, stream pb_testproto.TestService_PingListServer) error {
	requireUserIDToExist(s.T, stream.Context())
	return s.TestServiceServer.PingList(ping, stream)
}

func ctxWithToken(ctx context.Context, scheme string, token string) context.Context {
	md := metadata.Pairs("authorization", fmt.Sprintf("%s %v", scheme, token))
	nCtx := metautils.NiceMD(md).ToOutgoing(ctx)
	return nCtx
}

func TestAuthTestSuite(t *testing.T) {
	authFunc := UpdateCtx()
	s := &AuthTestSuite{
		InterceptorTestSuite: &grpc_testing.InterceptorTestSuite{
			TestService: &assertingPingService{&grpc_testing.TestPingService{T: t}, t},
			ServerOpts: []grpc.ServerOption{
				grpc.StreamInterceptor(grpc_auth.StreamServerInterceptor(authFunc)),
				grpc.UnaryInterceptor(grpc_auth.UnaryServerInterceptor(authFunc)),
			},
		},
	}
	suite.Run(t, s)
}

type AuthTestSuite struct {
	*grpc_testing.InterceptorTestSuite
}

func (s *AuthTestSuite) TestUnary_NoAuth() {
	_, err := s.Client.Ping(s.SimpleCtx(), goodPing)
	require.Error(s.T(), err, "there must be an error")
	require.Equal(s.T(), codes.Unauthenticated, status.Code(err), "must error with unauthenticated")
}

func (s *AuthTestSuite) TestUnary_BadAuth() {
	_, err := s.Client.Ping(ctxWithToken(s.SimpleCtx(), "bearer", "bad_token"), goodPing)
	require.Error(s.T(), err, "there must be an error")
	require.Equal(s.T(), codes.PermissionDenied, status.Code(err), "must error with permission denied")
}

func (s *AuthTestSuite) TestUnary_PassesAuth() {
	_, err := s.Client.Ping(ctxWithToken(s.SimpleCtx(), "bearer", authToken), goodPing)
	require.NoError(s.T(), err, "no error must occur")
}

func (s *AuthTestSuite) TestStream_NoAuth() {
	stream, err := s.Client.PingList(s.SimpleCtx(), goodPing)
	require.NoError(s.T(), err, "should not fail on establishing the stream")
	_, err = stream.Recv()
	require.Error(s.T(), err, "there must be an error")
	require.Equal(s.T(), codes.Unauthenticated, status.Code(err), "must error with unauthenticated")
}

func (s *AuthTestSuite) TestStream_BadAuth() {
	stream, err := s.Client.PingList(ctxWithToken(s.SimpleCtx(), "bearer", "bad_token"), goodPing)
	require.NoError(s.T(), err, "should not fail on establishing the stream")
	_, err = stream.Recv()
	require.Error(s.T(), err, "there must be an error")
	require.Equal(s.T(), codes.PermissionDenied, status.Code(err), "must error with permission denied")
}

func (s *AuthTestSuite) TestStream_PassesAuth() {
	stream, err := s.Client.PingList(ctxWithToken(s.SimpleCtx(), "bearer", authToken), goodPing)
	require.NoError(s.T(), err, "should not fail on establishing the stream")
	pong, err := stream.Recv()
	require.NoError(s.T(), err, "no error must occur")
	require.NotNil(s.T(), pong, "pong must not be nil")
}

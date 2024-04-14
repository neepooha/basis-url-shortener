package grpc

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	grpclog "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	grpcretry "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/retry"
	ssov2 "github.com/neepooha/protos/gen/go/sso"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	auth ssov2.AuthClient
	perm ssov2.PermissionsClient
	log  *slog.Logger
}

func New(ctx context.Context, log *slog.Logger, addr string, timeout time.Duration, retriesCount int) (*Client, error) {
	const op = "grpc.New"

	retryOpts := []grpcretry.CallOption{
		grpcretry.WithCodes(codes.NotFound, codes.Aborted, codes.DeadlineExceeded),
		grpcretry.WithMax(uint(retriesCount)),
		grpcretry.WithPerRetryTimeout(timeout),
	}
	logOpts := []grpclog.Option{
		grpclog.WithLogOnEvents(grpclog.PayloadSent, grpclog.PayloadReceived),
	}

	cc, err := grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(
			grpclog.UnaryClientInterceptor(InterceptorLogger(log), logOpts...),
			grpcretry.UnaryClientInterceptor(retryOpts...),
		),
	)

	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Client{
		auth: ssov2.NewAuthClient(cc),
		perm: ssov2.NewPermissionsClient(cc),
		log:  log,
	}, nil
}

func (c *Client) IsAdmin(ctx context.Context, userID uint64, appID int) (bool, error) {
	const op = "grpc.IsAdmin"
	resp, err := c.perm.IsAdmin(ctx, &ssov2.IsAdminRequest{
		UserId: userID,
		AppId: int32(appID),
	})
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}
	return resp.GetIsAdmin(), nil
}

func (c *Client) SetAdmin(ctx context.Context, email string, appID int) (bool, error) {
	const op = "grpc.SetAdmin"
	
	resp, err := c.perm.SetAdmin(ctx, &ssov2.SetAdminRequest{
		Email: email,
		AppId: int32(appID),
	})
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}
	return resp.GetSetAdmin(), nil
}

func (c *Client) DelAdmin(ctx context.Context, email string, appID int) (bool, error) {
	const op = "grpc.DelAdmin"
	resp, err := c.perm.DelAdmin(ctx, &ssov2.DelAdminRequest{
		Email: email,
		AppId: int32(appID),
	})
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}
	return resp.GetDelAdmin(), nil
}

func InterceptorLogger(log *slog.Logger) grpclog.Logger {
	return grpclog.LoggerFunc(func(ctx context.Context, level grpclog.Level, msg string, fields ...any) {
		log.Log(ctx, slog.Level(level), msg, fields...)
	})
}

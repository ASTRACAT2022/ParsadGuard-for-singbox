package singbox

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	statsService "github.com/xtls/xray-core/app/stats/command"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

	"github.com/pasarguard/node/common"
)

// statsClient talks to the sing-box experimental v2ray_api endpoint,
// which exposes the same StatsService that Xray uses.
type statsClient struct {
	conn   *grpc.ClientConn
	client statsService.StatsServiceClient
}

func newStatsClient(apiPort int) (*statsClient, error) {
	target := fmt.Sprintf("127.0.0.1:%d", apiPort)
	dialer := &net.Dialer{Timeout: 5 * time.Second}
	conn, err := grpc.NewClient(
		target,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(func(ctx context.Context, addr string) (net.Conn, error) {
			return dialer.DialContext(ctx, "tcp", addr)
		}),
	)
	if err != nil {
		return nil, err
	}
	return &statsClient{
		conn:   conn,
		client: statsService.NewStatsServiceClient(conn),
	}, nil
}

func (s *statsClient) Close() error {
	if s.conn != nil {
		return s.conn.Close()
	}
	return nil
}

func (s *statsClient) GetSysStats(ctx context.Context) (*common.BackendStatsResponse, error) {
	resp, err := s.client.GetSysStats(ctx, &statsService.SysStatsRequest{})
	if err != nil {
		return nil, status.Errorf(codes.Unknown, "failed to get sys stats: %v", err)
	}
	return &common.BackendStatsResponse{
		NumGoroutine: resp.NumGoroutine,
		NumGc:        resp.NumGC,
		Alloc:        resp.Alloc,
		TotalAlloc:   resp.TotalAlloc,
		Sys:          resp.Sys,
		Mallocs:      resp.Mallocs,
		Frees:        resp.Frees,
		LiveObjects:  resp.LiveObjects,
		PauseTotalNs: resp.PauseTotalNs,
		Uptime:       resp.Uptime,
	}, nil
}

func (s *statsClient) queryStats(ctx context.Context, pattern string, reset bool) (*statsService.QueryStatsResponse, error) {
	return s.client.QueryStats(ctx, &statsService.QueryStatsRequest{Pattern: pattern, Reset_: reset})
}

func parseUserStat(name string) (user, link, statType string, ok bool) {
	parts := strings.Split(name, ">>>")
	if len(parts) < 4 {
		return
	}
	return parts[1], parts[2], parts[3], true
}

func parseTagStat(name string) (tag, link, statType string, ok bool) {
	parts := strings.Split(name, ">>>")
	if len(parts) < 4 {
		return
	}
	return parts[1], parts[2], parts[3], true
}

func (s *statsClient) statsByPattern(ctx context.Context, pattern string, reset bool, parser func(string) (string, string, string, bool)) (*common.StatResponse, error) {
	resp, err := s.queryStats(ctx, pattern, reset)
	if err != nil {
		return nil, err
	}
	out := &common.StatResponse{}
	for _, st := range resp.GetStat() {
		name, link, statType, ok := parser(st.GetName())
		if !ok {
			continue
		}
		out.Stats = append(out.Stats, &common.Stat{
			Name:  name,
			Type:  statType,
			Link:  link,
			Value: st.GetValue(),
		})
	}
	return out, nil
}

func (s *statsClient) GetUsersStats(ctx context.Context, reset bool) (*common.StatResponse, error) {
	return s.statsByPattern(ctx, "user>>>", reset, parseTagStat)
}

func (s *statsClient) GetInboundsStats(ctx context.Context, reset bool) (*common.StatResponse, error) {
	return s.statsByPattern(ctx, "inbound>>>", reset, parseTagStat)
}

func (s *statsClient) GetOutboundsStats(ctx context.Context, reset bool) (*common.StatResponse, error) {
	return s.statsByPattern(ctx, "outbound>>>", reset, parseTagStat)
}

func (s *statsClient) GetUserStats(ctx context.Context, email string, reset bool) (*common.StatResponse, error) {
	if email == "" {
		return nil, errors.New("email required")
	}
	resp, err := s.queryStats(ctx, fmt.Sprintf("user>>>%s>>>", email), reset)
	if err != nil {
		return nil, err
	}
	out := &common.StatResponse{}
	for _, st := range resp.GetStat() {
		parts := strings.Split(st.GetName(), ">>>")
		if len(parts) < 4 {
			continue
		}
		out.Stats = append(out.Stats, &common.Stat{
			Name:  parts[1],
			Type:  parts[2],
			Link:  parts[3],
			Value: st.GetValue(),
		})
	}
	return out, nil
}

func (s *statsClient) GetInboundStats(ctx context.Context, tag string, reset bool) (*common.StatResponse, error) {
	if tag == "" {
		return nil, errors.New("tag required")
	}
	return s.statsByPattern(ctx, fmt.Sprintf("inbound>>>%s>>>", tag), reset, parseTagStat)
}

func (s *statsClient) GetOutboundStats(ctx context.Context, tag string, reset bool) (*common.StatResponse, error) {
	if tag == "" {
		return nil, errors.New("tag required")
	}
	return s.statsByPattern(ctx, fmt.Sprintf("outbound>>>%s>>>", tag), reset, parseTagStat)
}

func (s *statsClient) GetUserOnlineStats(ctx context.Context, email string) (*common.OnlineStatResponse, error) {
	if email == "" {
		return nil, errors.New("email required")
	}
	resp, err := s.client.GetStatsOnline(ctx, &statsService.GetStatsRequest{Name: fmt.Sprintf("user>>>%s>>>online", email)})
	if err != nil {
		return nil, err
	}
	return &common.OnlineStatResponse{Name: email, Value: resp.GetStat().GetValue()}, nil
}

func (s *statsClient) GetUserOnlineIpListStats(ctx context.Context, email string) (*common.StatsOnlineIpListResponse, error) {
	if email == "" {
		return nil, errors.New("email required")
	}
	resp, err := s.client.GetStatsOnlineIpList(ctx, &statsService.GetStatsRequest{Name: fmt.Sprintf("user>>>%s>>>online", email)})
	if err != nil {
		return nil, err
	}
	return &common.StatsOnlineIpListResponse{Name: email, Ips: resp.GetIps()}, nil
}

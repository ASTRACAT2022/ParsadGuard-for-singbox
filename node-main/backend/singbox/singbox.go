package singbox

import (
	"context"
	"errors"
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/pasarguard/node/common"
	"github.com/pasarguard/node/config"
)

// SingBox implements backend.Backend over a sing-box subprocess.
type SingBox struct {
	cfg        *config.Config
	config     *Config
	core       *Core
	stats      *statsClient
	apiPort    int
	cancelFunc context.CancelFunc
	mu         sync.RWMutex
}

// New starts sing-box with the given config + initial user list.
// apiPort must be a free local TCP port for the v2ray_api gRPC stats endpoint.
func New(ctx context.Context, sbConfig *Config, users []*common.User, apiPort int, cfg *config.Config) (*SingBox, error) {
	executablePath, err := filepath.Abs(singboxExecutablePath(cfg))
	if err != nil {
		return nil, err
	}
	generatedPath, err := filepath.Abs(cfg.GeneratedConfigPath)
	if err != nil {
		return nil, err
	}

	ctxCore, cancel := context.WithCancel(context.Background())
	sb := &SingBox{
		cancelFunc: cancel,
		cfg:        cfg,
		apiPort:    apiPort,
	}

	start := time.Now()

	if err := sbConfig.ApplyAPI(apiPort); err != nil {
		return nil, err
	}
	if len(users) > 0 {
		log.Printf("syncing %d users on startup", len(users))
		sbConfig.syncUsers(users)
	} else {
		log.Println("no users provided on startup")
	}
	sb.config = sbConfig

	log.Printf("sing-box config generated in %.2f seconds", time.Since(start).Seconds())

	core, err := NewCore(executablePath, generatedPath, cfg.LogBufferSize, cfg.StartupLogTailSize)
	if err != nil {
		return nil, err
	}
	if err := core.Start(sbConfig, cfg.Debug); err != nil {
		return nil, err
	}
	sb.core = core

	statsCli, err := newStatsClient(apiPort)
	if err != nil {
		sb.Shutdown()
		return nil, err
	}
	sb.stats = statsCli

	if err := sb.checkStatus(ctx); err != nil {
		sb.Shutdown()
		return nil, err
	}

	go sb.checkHealth(ctxCore)

	log.Println("sing-box started, Version:", sb.Version())
	return sb, nil
}

func singboxExecutablePath(cfg *config.Config) string {
	// Reuse XRAY_EXECUTABLE_PATH if SINGBOX_EXECUTABLE_PATH is unset; this lets
	// existing deployments keep using their pre-baked env files.
	if p := cfg.SingBoxExecutablePath; p != "" {
		return p
	}
	return "/usr/local/bin/sing-box"
}

// --- backend.Backend interface ---

func (s *SingBox) Started() bool   { s.mu.RLock(); defer s.mu.RUnlock(); return s.core.Started() }
func (s *SingBox) Version() string { s.mu.RLock(); defer s.mu.RUnlock(); return s.core.Version() }
func (s *SingBox) Logs() <-chan string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.core.Logs()
}

func (s *SingBox) Restart() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.core.Restart(s.config, s.cfg.Debug); err != nil {
		return err
	}
	return nil
}

func (s *SingBox) Shutdown() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.cancelFunc != nil {
		s.cancelFunc()
	}
	if s.core != nil {
		s.core.Stop()
	}
	if s.stats != nil {
		_ = s.stats.Close()
	}
}

// --- user sync ---
//
// sing-box has no public user-management RPC; every user change requires a
// config reload. We therefore implement all SyncUser/SyncUsers/UpdateUsers
// variants the same way: mutate the in-memory config and restart the core.

func (s *SingBox) SyncUser(ctx context.Context, user *common.User) error {
	return s.SyncUsersWithRestart(ctx, []*common.User{user}, false)
}

func (s *SingBox) SyncUsers(ctx context.Context, users []*common.User) error {
	s.mu.Lock()
	s.config.syncUsers(users)
	s.mu.Unlock()
	if err := s.Restart(); err != nil {
		return err
	}
	return s.checkStatus(ctx)
}

func (s *SingBox) UpdateUsers(ctx context.Context, users []*common.User) error {
	return s.SyncUsersWithRestart(ctx, users, true)
}

func (s *SingBox) UpdateUsersAndRestart(ctx context.Context, users []*common.User) error {
	return s.SyncUsersWithRestart(ctx, users, true)
}

// SyncUsersWithRestart applies an incremental user update and restarts sing-box.
// When `incremental` is true the existing user set is preserved.
func (s *SingBox) SyncUsersWithRestart(ctx context.Context, users []*common.User, incremental bool) error {
	s.mu.Lock()
	if incremental {
		s.config.updateUsers(users)
	} else {
		s.config.syncUsers(users)
	}
	s.mu.Unlock()
	if err := s.Restart(); err != nil {
		return err
	}
	return s.checkStatus(ctx)
}

// --- stats passthrough ---

func (s *SingBox) GetSysStats(ctx context.Context) (*common.BackendStatsResponse, error) {
	return s.stats.GetSysStats(ctx)
}

func (s *SingBox) GetUserOnlineStats(ctx context.Context, email string) (*common.OnlineStatResponse, error) {
	return s.stats.GetUserOnlineStats(ctx, email)
}

func (s *SingBox) GetUserOnlineIpListStats(ctx context.Context, email string) (*common.StatsOnlineIpListResponse, error) {
	return s.stats.GetUserOnlineIpListStats(ctx, email)
}

func (s *SingBox) GetStats(ctx context.Context, req *common.StatRequest) (*common.StatResponse, error) {
	switch req.GetType() {
	case common.StatType_Outbounds:
		return s.stats.GetOutboundsStats(ctx, req.GetReset_())
	case common.StatType_Outbound:
		return s.stats.GetOutboundStats(ctx, req.GetName(), req.GetReset_())
	case common.StatType_Inbounds:
		return s.stats.GetInboundsStats(ctx, req.GetReset_())
	case common.StatType_Inbound:
		return s.stats.GetInboundStats(ctx, req.GetName(), req.GetReset_())
	case common.StatType_UsersStat:
		return s.stats.GetUsersStats(ctx, req.GetReset_())
	case common.StatType_UserStat:
		return s.stats.GetUserStats(ctx, req.GetName(), req.GetReset_())
	default:
		return nil, errors.New("not implemented stat type")
	}
}

// --- health checks ---

func (s *SingBox) checkStatus(baseCtx context.Context) error {
	apiTicker := time.NewTicker(time.Second)
	defer apiTicker.Stop()
	deadline := time.After(15 * time.Second)
	for {
		select {
		case <-baseCtx.Done():
			return errors.New("context cancelled")
		case <-deadline:
			if failure := s.core.LatestStartupFailure(); failure != "" {
				return fmt.Errorf("failed to start sing-box: %s", failure)
			}
			tail := s.core.StartupLogTail(50)
			return fmt.Errorf("sing-box API did not become ready in time; recent logs:\n%s", strings.Join(tail, "\n"))
		case <-apiTicker.C:
			ctx, cancel := context.WithTimeout(baseCtx, time.Second)
			_, err := s.stats.GetSysStats(ctx)
			cancel()
			if err == nil {
				s.core.SwitchToRuntimeLogPhase()
				return nil
			}
			if failure := s.core.LatestStartupFailure(); failure != "" {
				return fmt.Errorf("failed to start sing-box: %s", failure)
			}
			if !s.core.Started() {
				tail := s.core.StartupLogTail(50)
				return fmt.Errorf("sing-box process stopped before API became ready; recent logs:\n%s", strings.Join(tail, "\n"))
			}
		}
	}
}

func (s *SingBox) checkHealth(baseCtx context.Context) {
	consecutiveFailures := 0
	const maxFailures = 3
	for {
		select {
		case <-baseCtx.Done():
			return
		default:
		}
		ctx, cancel := context.WithTimeout(baseCtx, 3*time.Second)
		_, err := s.stats.GetSysStats(ctx)
		cancel()
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return
			}
			consecutiveFailures++
			if consecutiveFailures >= maxFailures {
				log.Printf("sing-box health check failed %d times, restarting...", consecutiveFailures)
				if err := s.Restart(); err != nil {
					log.Println(err.Error())
				} else {
					log.Println("sing-box restarted")
					consecutiveFailures = 0
				}
			}
		} else {
			consecutiveFailures = 0
		}
		time.Sleep(5 * time.Second)
	}
}

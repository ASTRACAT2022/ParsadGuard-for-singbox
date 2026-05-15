package singbox

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	nodeLogger "github.com/pasarguard/node/logger"
)

const (
	logPhaseRuntime uint32 = iota
	logPhaseStartup
)

// Core wraps the sing-box subprocess.
type Core struct {
	executablePath string
	configPath     string
	version        string
	process        *exec.Cmd
	processPID     int
	restarting     bool
	logsChan       chan string
	logger         *nodeLogger.Logger
	cancelFunc     context.CancelFunc

	logPhase       uint32
	startupLogs    []string
	startupLogSize int
	startupFailure string
	startupEnabled bool

	mu        sync.Mutex
	startupMu sync.RWMutex
}

// NewCore prepares a Core. configPath is the directory where rendered configs are written.
func NewCore(executablePath, configPath string, logBufferSize, startupLogTailSize int) (*Core, error) {
	if startupLogTailSize <= 0 {
		startupLogTailSize = 200
	}

	c := &Core{
		executablePath: executablePath,
		configPath:     configPath,
		logsChan:       make(chan string, logBufferSize),
		logPhase:       logPhaseRuntime,
		startupLogSize: startupLogTailSize,
	}

	v, err := c.refreshVersion()
	if err != nil {
		return nil, err
	}
	c.version = v
	return c, nil
}

func (c *Core) refreshVersion() (string, error) {
	cmd := exec.Command(c.executablePath, "version")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return "", err
	}
	// "sing-box version 1.10.3" or multi-line with version on first line
	re := regexp.MustCompile(`(?:sing-box[^\d]*)?(\d+\.\d+\.\d+\S*)`)
	matches := re.FindStringSubmatch(out.String())
	if len(matches) > 1 {
		return matches[1], nil
	}
	return "", errors.New("could not parse sing-box version")
}

// Version returns the resolved sing-box version string.
func (c *Core) Version() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.version
}

// Started returns true while the subprocess is alive.
func (c *Core) Started() bool {
	if c.process == nil || c.process.Process == nil {
		return false
	}
	if c.process.ProcessState == nil {
		return true
	}
	return false
}

// GenerateConfigFile writes a pretty-printed config to disk (debug only).
func (c *Core) GenerateConfigFile(config []byte) error {
	var pretty bytes.Buffer
	if err := json.Indent(&pretty, config, "", "    "); err != nil {
		return err
	}
	if err := os.MkdirAll(c.configPath, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %v", err)
	}
	f, err := os.Create(filepath.Join(c.configPath, "singbox.json"))
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(pretty.Bytes())
	return err
}

// Logs returns the runtime log channel.
func (c *Core) Logs() <-chan string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.logsChan
}

// Start launches sing-box with the rendered config supplied on stdin.
func (c *Core) Start(cfg *Config, debug bool) error {
	bytesConfig, err := cfg.ToBytes()
	if err != nil {
		return err
	}
	if debug {
		if err := c.GenerateConfigFile(bytesConfig); err != nil {
			return err
		}
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.Started() {
		return errors.New("sing-box is already started")
	}

	c.enableStartupDiagnostics(c.startupLogSize)
	c.setStartupLogPhase()

	cmd := exec.Command(c.executablePath, "run", "-c", "/dev/stdin")
	setProcAttributes(cmd)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	c.logger = nodeLogger.New(debug)
	cmd.Stdin = bytes.NewBuffer(bytesConfig)

	if err := cmd.Start(); err != nil {
		return err
	}
	c.process = cmd
	c.processPID = cmd.Process.Pid

	go func() { _ = cmd.Wait() }()

	ctxCore, cancel := context.WithCancel(context.Background())
	c.cancelFunc = cancel

	go c.captureProcessLogs(ctxCore, stdout)
	go c.captureProcessLogs(ctxCore, stderr)

	return nil
}

// Stop terminates the sing-box subprocess.
func (c *Core) Stop() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.Started() && c.process == nil {
		return
	}

	if c.Started() {
		pid := c.process.Process.Pid
		_ = c.process.Process.Kill()
		done := make(chan error, 1)
		go func() { done <- c.process.Wait() }()
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			log.Printf("sing-box process %d did not terminate within timeout, force killing", pid)
			_ = killProcessTree(pid)
		}
	}

	c.process = nil
	c.processPID = 0

	if c.cancelFunc != nil {
		c.cancelFunc()
		c.cancelFunc = nil
	}
	if c.logger != nil {
		c.logger.Close()
		c.logger = nil
	}
	c.switchToRuntimeLogPhase()

	log.Println("sing-box core stopped")
}

// Restart performs a Stop + Start with the same config.
func (c *Core) Restart(cfg *Config, debug bool) error {
	c.mu.Lock()
	if c.restarting {
		c.mu.Unlock()
		return errors.New("sing-box is already restarting")
	}
	c.restarting = true
	c.mu.Unlock()

	defer func() {
		c.mu.Lock()
		c.restarting = false
		c.mu.Unlock()
	}()

	log.Println("restarting sing-box core...")
	c.Stop()
	return c.Start(cfg, debug)
}

// --- log handling ---

var accessLogPattern = regexp.MustCompile(`(?i)inbound connection|user .* outbound`)

func (c *Core) captureProcessLogs(ctx context.Context, pipe io.Reader) {
	scanner := bufio.NewScanner(pipe)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return
		default:
			line := scanner.Text()
			if c.isStartupLogPhase() {
				c.recordStartupLog(line)
			}
			select {
			case c.logsChan <- line:
			default:
			}
			c.detectLogType(line)
		}
	}
}

func (c *Core) detectLogType(line string) {
	if c.logger == nil {
		return
	}
	if accessLogPattern.MatchString(line) {
		c.logger.Log(nodeLogger.LogInfo, line)
		return
	}
	c.logger.Log(nodeLogger.LogError, line)
}

// --- startup diagnostics ---

func (c *Core) enableStartupDiagnostics(tailSize int) {
	c.startupMu.Lock()
	defer c.startupMu.Unlock()
	if tailSize <= 0 {
		tailSize = 200
	}
	c.startupLogs = make([]string, 0, tailSize)
	c.startupFailure = ""
	c.startupEnabled = true
}

func (c *Core) recordStartupLog(line string) {
	c.startupMu.Lock()
	defer c.startupMu.Unlock()
	if !c.startupEnabled {
		return
	}
	if len(c.startupLogs) >= cap(c.startupLogs) && cap(c.startupLogs) > 0 {
		c.startupLogs = append(c.startupLogs[1:], line)
	} else {
		c.startupLogs = append(c.startupLogs, line)
	}
	lower := strings.ToLower(line)
	if strings.Contains(lower, "fatal") ||
		strings.Contains(lower, "panic") ||
		strings.Contains(lower, "failed to start") ||
		strings.Contains(lower, "failed to listen") ||
		strings.Contains(lower, "failed to parse") ||
		strings.Contains(lower, "permission denied") {
		c.startupFailure = line
	}
}

func (c *Core) LatestStartupFailure() string {
	c.startupMu.RLock()
	defer c.startupMu.RUnlock()
	return c.startupFailure
}

func (c *Core) StartupLogTail(n int) []string {
	c.startupMu.RLock()
	defer c.startupMu.RUnlock()
	if n <= 0 || n > len(c.startupLogs) {
		n = len(c.startupLogs)
	}
	out := make([]string, n)
	copy(out, c.startupLogs[len(c.startupLogs)-n:])
	return out
}

func (c *Core) setStartupLogPhase()       { atomic.StoreUint32(&c.logPhase, logPhaseStartup) }
func (c *Core) switchToRuntimeLogPhase()  { atomic.StoreUint32(&c.logPhase, logPhaseRuntime); c.startupMu.Lock(); c.startupEnabled = false; c.startupMu.Unlock() }
func (c *Core) isStartupLogPhase() bool   { return atomic.LoadUint32(&c.logPhase) == logPhaseStartup }
func (c *Core) SwitchToRuntimeLogPhase()  { c.switchToRuntimeLogPhase() }

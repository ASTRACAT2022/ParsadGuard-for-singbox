package singbox

import (
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"
	"sync"

	"github.com/google/uuid"
	"github.com/pasarguard/node/common"
)

// Supported sing-box inbound protocols
const (
	Vless        = "vless"
	Vmess        = "vmess"
	Trojan       = "trojan"
	Shadowsocks  = "shadowsocks"
	Hysteria2    = "hysteria2"
	Hysteria     = "hysteria" // legacy
	Tuic         = "tuic"
	ShadowTLS    = "shadowtls"
	Naive        = "naive"
	AnyTLS       = "anytls"
	SocksInbound = "socks"
	HTTPInbound  = "http"
)

// Inbound is a thin wrapper around sing-box inbound JSON.
type Inbound struct {
	raw     map[string]any
	tag     string
	itype   string
	mu      sync.RWMutex
	exclude bool
	// runtime: email -> user object (already shaped for the protocol)
	users map[string]map[string]any
}

func newInbound(raw map[string]any) (*Inbound, error) {
	tag, _ := raw["tag"].(string)
	itype, _ := raw["type"].(string)
	if tag == "" || itype == "" {
		return nil, errors.New("inbound must have non-empty type and tag")
	}
	return &Inbound{
		raw:   raw,
		tag:   tag,
		itype: itype,
		users: make(map[string]map[string]any),
	}, nil
}

// Tag returns the inbound tag.
func (i *Inbound) Tag() string { return i.tag }

// Type returns the inbound protocol type.
func (i *Inbound) Type() string { return i.itype }

// Excluded returns whether this inbound is excluded from user sync.
func (i *Inbound) Excluded() bool {
	i.mu.RLock()
	defer i.mu.RUnlock()
	return i.exclude
}

// Config holds the parsed sing-box configuration with runtime-tracked users.
type Config struct {
	raw      map[string]any
	inbounds []*Inbound
	mu       sync.RWMutex
}

// NewConfig parses a sing-box JSON config and prepares per-inbound state.
func NewConfig(config string, exclude []string) (*Config, error) {
	var raw map[string]any
	if err := json.Unmarshal([]byte(config), &raw); err != nil {
		return nil, fmt.Errorf("invalid sing-box config json: %w", err)
	}

	inboundsRaw, _ := raw["inbounds"].([]any)
	cfg := &Config{raw: raw}
	for _, item := range inboundsRaw {
		inboundMap, ok := item.(map[string]any)
		if !ok {
			continue
		}
		in, err := newInbound(inboundMap)
		if err != nil {
			return nil, err
		}
		if slices.Contains(exclude, in.tag) {
			in.exclude = true
		}
		cfg.inbounds = append(cfg.inbounds, in)
	}
	return cfg, nil
}

// Inbounds returns inbounds slice (read-only).
func (c *Config) Inbounds() []*Inbound { return c.inbounds }

// findInbound returns inbound by tag or nil.
func (c *Config) findInbound(tag string) *Inbound {
	for _, in := range c.inbounds {
		if in.tag == tag {
			return in
		}
	}
	return nil
}

// ApplyAPI injects/replaces the experimental.v2ray_api block so that
// the stats service is exposed on the chosen local port.
func (c *Config) ApplyAPI(apiPort int) error {
	if apiPort <= 0 {
		return errors.New("invalid api port")
	}
	exp, _ := c.raw["experimental"].(map[string]any)
	if exp == nil {
		exp = map[string]any{}
	}
	exp["v2ray_api"] = map[string]any{
		"listen": fmt.Sprintf("127.0.0.1:%d", apiPort),
		"stats": map[string]any{
			"enabled":   true,
			"users":     c.collectUserEmails(),
			"inbounds":  c.collectInboundTags(),
			"outbounds": c.collectOutboundTags(),
		},
	}
	c.raw["experimental"] = exp
	return nil
}

func (c *Config) collectInboundTags() []string {
	tags := make([]string, 0, len(c.inbounds))
	for _, in := range c.inbounds {
		if in.exclude {
			continue
		}
		tags = append(tags, in.tag)
	}
	return tags
}

func (c *Config) collectOutboundTags() []string {
	out := []string{}
	outs, _ := c.raw["outbounds"].([]any)
	for _, o := range outs {
		m, ok := o.(map[string]any)
		if !ok {
			continue
		}
		if t, ok := m["tag"].(string); ok && t != "" {
			out = append(out, t)
		}
	}
	return out
}

func (c *Config) collectUserEmails() []string {
	seen := map[string]struct{}{}
	for _, in := range c.inbounds {
		if in.exclude {
			continue
		}
		in.mu.RLock()
		for email := range in.users {
			seen[email] = struct{}{}
		}
		in.mu.RUnlock()
	}
	out := make([]string, 0, len(seen))
	for e := range seen {
		out = append(out, e)
	}
	return out
}

// syncUsers replaces the user set for all non-excluded inbounds.
func (c *Config) syncUsers(users []*common.User) {
	for _, in := range c.inbounds {
		if in.exclude {
			continue
		}
		newUsers := map[string]map[string]any{}
		for _, u := range users {
			if !slices.Contains(u.GetInbounds(), in.tag) {
				continue
			}
			entry, ok := buildUserEntry(in, u)
			if !ok {
				continue
			}
			newUsers[u.GetEmail()] = entry
		}
		in.mu.Lock()
		in.users = newUsers
		in.mu.Unlock()
	}
}

// updateUsers applies an incremental update.
func (c *Config) updateUsers(users []*common.User) {
	for _, in := range c.inbounds {
		if in.exclude {
			continue
		}
		in.mu.Lock()
		for _, u := range users {
			email := u.GetEmail()
			if slices.Contains(u.GetInbounds(), in.tag) {
				if entry, ok := buildUserEntry(in, u); ok {
					in.users[email] = entry
				} else {
					delete(in.users, email)
				}
			} else {
				delete(in.users, email)
			}
		}
		in.mu.Unlock()
	}
}

// buildUserEntry builds the protocol-specific user object for an inbound.
// Returns false if the user has no matching credentials for this protocol.
func buildUserEntry(in *Inbound, u *common.User) (map[string]any, bool) {
	email := u.GetEmail()
	p := u.GetProxies()
	switch in.itype {
	case Vless:
		if p.GetVless() == nil {
			return nil, false
		}
		if _, err := uuid.Parse(p.GetVless().GetId()); err != nil {
			return nil, false
		}
		entry := map[string]any{
			"name": email,
			"uuid": p.GetVless().GetId(),
		}
		// Flow is only legal with TLS/reality + tcp transport in sing-box; the
		// panel-side validators should have already filtered it, but we keep it
		// optional and let sing-box reject invalid combinations on start.
		if flow := p.GetVless().GetFlow(); flow != "" {
			entry["flow"] = flow
		}
		return entry, true
	case Vmess:
		if p.GetVmess() == nil {
			return nil, false
		}
		return map[string]any{
			"name":     email,
			"uuid":     p.GetVmess().GetId(),
			"alterId":  0,
		}, true
	case Trojan:
		if p.GetTrojan() == nil {
			return nil, false
		}
		return map[string]any{
			"name":     email,
			"password": p.GetTrojan().GetPassword(),
		}, true
	case Hysteria2, Hysteria:
		if p.GetHysteria() == nil {
			return nil, false
		}
		return map[string]any{
			"name":     email,
			"password": p.GetHysteria().GetAuth(),
		}, true
	case Tuic:
		// tuic requires uuid+password; we reuse VLESS uuid and trojan password
		// if both are provided (panel may ship them together via different
		// proxy fields). Falling back to trojan-only is also acceptable.
		var id, pwd string
		if p.GetVless() != nil {
			id = p.GetVless().GetId()
		}
		if p.GetTrojan() != nil {
			pwd = p.GetTrojan().GetPassword()
		}
		if id == "" || pwd == "" {
			return nil, false
		}
		return map[string]any{
			"name":     email,
			"uuid":     id,
			"password": pwd,
		}, true
	case ShadowTLS:
		if p.GetTrojan() == nil {
			return nil, false
		}
		return map[string]any{
			"name":     email,
			"password": p.GetTrojan().GetPassword(),
		}, true
	case Shadowsocks:
		if p.GetShadowsocks() == nil {
			return nil, false
		}
		method, _ := in.raw["method"].(string)
		pwd := p.GetShadowsocks().GetPassword()
		if strings.HasPrefix(method, "2022-blake3") {
			pwd = common.EnsureBase64Password(pwd, method)
		}
		return map[string]any{
			"name":     email,
			"password": pwd,
		}, true
	case Naive, AnyTLS, SocksInbound, HTTPInbound:
		if p.GetTrojan() != nil {
			return map[string]any{
				"username": email,
				"password": p.GetTrojan().GetPassword(),
			}, true
		}
		return nil, false
	}
	return nil, false
}

// applyUsersToRaw flushes the runtime users map into each inbound's raw["users"].
func (c *Config) applyUsersToRaw() {
	for _, in := range c.inbounds {
		if in.exclude {
			continue
		}
		in.mu.RLock()
		usersList := make([]map[string]any, 0, len(in.users))
		for _, u := range in.users {
			usersList = append(usersList, u)
		}
		in.mu.RUnlock()

		switch in.itype {
		case Shadowsocks:
			method, _ := in.raw["method"].(string)
			if strings.HasPrefix(method, "2022-blake3") {
				// 2022 inbound supports per-user list
				in.raw["users"] = usersList
			} else if len(usersList) > 0 {
				// classic SS only supports a single password; fall back to first
				if pwd, ok := usersList[0]["password"].(string); ok {
					in.raw["password"] = pwd
				}
			}
		default:
			in.raw["users"] = usersList
		}
	}
}

// ToBytes returns the rendered JSON config ready for sing-box.
func (c *Config) ToBytes() ([]byte, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.applyUsersToRaw()

	// Re-attach inbound raw maps to the inbounds slice in raw config to make
	// sure any mutations land in the final document.
	updated := make([]any, 0, len(c.inbounds))
	for _, in := range c.inbounds {
		updated = append(updated, in.raw)
	}
	c.raw["inbounds"] = updated

	return json.MarshalIndent(c.raw, "", "  ")
}

// InboundTags returns active inbound tags (non-excluded).
func (c *Config) InboundTags() []string { return c.collectInboundTags() }

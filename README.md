# PasarGuard-for-singbox

**Community fork of [PasarGuard](https://github.com/PasarGuard/panel) with sing-box as the primary core instead of Xray.**

## Structure

| Directory | Description |
|-----------|-------------|
| `panel-main/` | Python panel (FastAPI + React dashboard) |
| `node-main/` | Go node agent (sing-box + WireGuard backend) |
| `subscription-template-main/` | Subscription page templates |

## What changed from upstream

- **sing-box** replaces Xray-core as the default backend (VLESS+Reality, Trojan, Hysteria2, TUIC, ShadowTLS)
- WireGuard support preserved
- All `xray_version` references renamed to `core_version`
- Docker images ship sing-box instead of xray

## Quick Start

See individual READMEs:

- [Panel README](panel-main/README.md)
- [Node README](node-main/README.md)
- [Subscription Template README](subscription-template-main/README.md)

## License

Same as upstream PasarGuard.

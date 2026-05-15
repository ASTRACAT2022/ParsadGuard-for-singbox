> **This is a community fork of PasarGuard that replaces Xray-core with sing-box.**

<h1 align="center">🛡️ PasarGuard</h1>

<p align="center">
    <strong>Unified & Censorship-Resistant Proxy Management Solution — sing-box Edition</strong>
</p>

---

<br/>
<p align="center">
    <a href="https://github.com/ASTRACAT2022/ParsadGuard-for-singbox" target="_blank">
        <img src="https://img.shields.io/github/license/ASTRACAT2022/ParsadGuard-for-singbox?style=flat-square" />
    </a>
    <a href="https://github.com/ASTRACAT2022/ParsadGuard-for-singbox" target="_blank">
        <img src="https://img.shields.io/github/stars/ASTRACAT2022/ParsadGuard-for-singbox?style=social" />
    </a>
</p>

<p align="center">
 <a href="./README-fa.md">
 🇮🇷 فارسی
 </a>
  /
  <a href="./README-zh-cn.md">
 🇨🇳 简体中文
 </a>
   /
  <a href="./README-ru.md">
 🇷🇺 Русский
 </a>
</p>

## 📋 Table of Contents

> **Quick Navigation** - Jump to any section below

-   [📖 Overview](#-overview)
    -   [🤔 Why using PasarGuard?](#-why-using-pasarguard)
        -   [✨ Features](#-features)
-   [🚀 Installation guide](#-installation-guide)
-   [📚 Documentation](#-documentation)

---

# 📖 Overview

> **What is PasarGuard?**

PasarGuard is a powerful proxy management tool that offers an intuitive and efficient interface for handling hundreds of proxy accounts. Built with Python and React.js it combines performance, scalability, and ease of use to simplify large-scale proxy management. This fork replaces Xray-core with [sing-box](https://github.com/SagerNet/sing-box) for maximum performance, alongside [WireGuard](https://www.wireguard.com/) support.

---

## 🤔 Why using PasarGuard?

> **Simple, Powerful, Reliable**

PasarGuard is a user-friendly, feature-rich, and reliable proxy management tool. It allows you to create and manage multiple proxies for your users without the need for complex configuration. With its built-in web interface, you can easily monitor activity, modify settings, and control user access limits — all from one convenient dashboard.

---

### ✨ Features

<div align="left">

**🌐 Web Interface & API**
- Built-in **Web UI** dashboard
- Fully **REST API** backend
- **Multi-Node** support for infrastructure distribution

**🔐 Protocols & Security**
- Supports **Vmess**, **VLESS**, **Trojan**, **Shadowsocks**, **WireGuard** and **Hysteria2**
- **TLS** and **REALITY** support
- **Multi-protocol** for a single user

**👥 User Management**
- **Multi-user** on a single inbound
- **Multi-inbound** on a **single port** (fallbacks support)
- **Traffic** and **expiry date** limitations
- **Periodic** traffic limit (daily, weekly, etc.)

**🔗 Subscriptions & Sharing**
- **Subscription link** compatible with **V2ray**, **Clash** and **ClashMeta**
- Automated **Share link** and **QRcode** generator
- System monitoring and **traffic statistics**

**🛠️ Tools & Customization**
- Customizable sing-box configuration
- Integrated **Telegram Bot**
- **Command Line Interface (CLI)**
- **Multi-language** support
- **Multi-admin** support (WIP)

</div>

---

# 🚀 Installation guide

> **Quick Start** - Get PasarGuard running in minutes

### For a quick setup, use the following commands based on your preferred database.

---

**TimescaleDB (Recommended):**
```bash
curl -fsSL https://github.com/ASTRACAT2022/ParsadGuard-for-singbox/raw/main/pasarguard.sh -o /tmp/pg.sh \
  && sudo bash /tmp/pg.sh install --database timescaledb
```

**SQLite:**
```bash
curl -fsSL https://github.com/ASTRACAT2022/ParsadGuard-for-singbox/raw/main/pasarguard.sh -o /tmp/pg.sh \
  && sudo bash /tmp/pg.sh install
```

**MySQL:**
```bash
curl -fsSL https://github.com/ASTRACAT2022/ParsadGuard-for-singbox/raw/main/pasarguard.sh -o /tmp/pg.sh \
  && sudo bash /tmp/pg.sh install --database mysql
```

**MariaDB:**
```bash
curl -fsSL https://github.com/ASTRACAT2022/ParsadGuard-for-singbox/raw/main/pasarguard.sh -o /tmp/pg.sh \
  && sudo bash /tmp/pg.sh install --database mariadb
```

**PostgreSQL:**
```bash
curl -fsSL https://github.com/ASTRACAT2022/ParsadGuard-for-singbox/raw/main/pasarguard.sh -o /tmp/pg.sh \
  && sudo bash /tmp/pg.sh install --database postgresql
```

### 📋 After installation:

<div align="left">

**📋 Watch the logs** (press `Ctrl+C` to stop)

**📁 Files are located at** `/opt/pasarguard`

**⚙️ Config file:** `/opt/pasarguard/.env` (see [Configuration](#-configuration) for details)

**💾 Data files:** `/var/lib/pasarguard`

**🔒 Important:** Dashboard requires SSL certificate for security
- Get SSL certificate (refer to documentation)
- Access: `https://YOUR_DOMAIN:8000/dashboard/`

**🔗 For testing without domain:** Use SSH port forwarding (see below)

</div>

---

```bash
ssh -L 8000:localhost:8000 user@serverip
```

Then access: `http://localhost:8000/dashboard/`

> ⚠️ **Testing only** - You'll lose access when you close the SSH terminal.

### 🔧 Next Steps:

```bash
# Create admin account
pasarguard cli admins --create <username>

# Get help
pasarguard --help
```



# 📚 Documentation

<div align="left">

**📖 Documentation** — Guides under development for the sing-box fork

🇺🇸 English | 🇮🇷 فارسی | 🇷🇺 Русский | 🇨🇳 简体中文

</div>

> **Original project:** [github.com/PasarGuard/panel](https://github.com/PasarGuard/panel) | Original docs: [docs.pasarguard.org](https://docs.pasarguard.org)

---

<p align="center">
  Made with ❤️ for Internet freedom
</p>

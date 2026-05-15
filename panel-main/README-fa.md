> **این یک فورک انجمنی از PasarGuard است که Xray-core را با sing-box جایگزین می‌کند.**

<h1 align="center">🛡️ پاسارگارد</h1>

<p align="center">
    <strong>راه‌حل یکپارچه و مقاوم در برابر سانسور برای مدیریت پروکسی — نسخه sing-box</strong>
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
 <a href="./README.md">
 🇺🇸 English
 </a>
 /
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

## 📋 فهرست مطالب

> **ناوبری سریع** - به هر بخش زیر پرش کنید

-   [📖 بررسی اجمالی](#-بررسی-اجمالی)
    -   [🤔 چرا از پاسارگارد استفاده کنیم؟](#-چرا-از-پاسارگارد-استفاده-کنیم)
        -   [✨ ویژگی‌ها](#-ویژگیها)
-   [🚀 راهنمای نصب](#-راهنمای-نصب)
-   [📚 مستندات](#-مستندات)

---

# 📖 بررسی اجمالی

> **پاسارگارد چیست؟**

پاسارگارد یک ابزار قدرتمند مدیریت پروکسی است که رابط کاربری بصری و کارآمدی برای مدیریت صدها حساب پروکسی ارائه می‌دهد. این ابزار با Python و React.js ساخته شده و عملکرد، مقیاس‌پذیری و سهولت استفاده را برای ساده‌سازی مدیریت پروکسی در مقیاس بزرگ ترکیب می‌کند. این فورک Xray-core را با [sing-box](https://github.com/SagerNet/sing-box) برای حداکثر عملکرد جایگزین می‌کند، همراه با پشتیبانی از [WireGuard](https://www.wireguard.com/).

---

## 🤔 چرا از پاسارگارد استفاده کنیم؟

> **ساده، قدرتمند، قابل اعتماد**

پاسارگارد یک ابزار مدیریت پروکسی کاربرپسند، غنی از ویژگی و قابل اعتماد است. این ابزار به شما امکان ایجاد و مدیریت چندین پروکسی برای کاربران بدون نیاز به پیکربندی پیچیده را می‌دهد. با رابط کاربری وب داخلی آن، می‌توانید به راحتی فعالیت‌ها را نظارت کنید، تنظیمات را تغییر دهید و محدودیت‌های دسترسی کاربران را کنترل کنید — همه از یک داشبورد مناسب.

---

### ✨ ویژگی‌ها

<div align="right">

**🌐 رابط کاربری وب و API**
- داشبورد **Web UI** داخلی
- بک‌اند کاملاً **REST API**
- پشتیبانی از **Multi-Node** برای توزیع زیرساخت

**🔐 پروتکل‌ها و امنیت**
- پشتیبانی از **Vmess**، **VLESS**، **Trojan**، **Shadowsocks**، **WireGuard** و **Hysteria2**
- پشتیبانی از **TLS** و **REALITY**
- **چند پروتکل** برای یک کاربر

**👥 مدیریت کاربران**
- **چند کاربر** روی یک inbound
- **چند inbound** روی **یک پورت** (پشتیبانی از fallbacks)
- محدودیت‌های **ترافیک** و **تاریخ انقضا**
- محدودیت ترافیک **دوره‌ای** (روزانه، هفتگی و غیره)

**🔗 اشتراک‌ها و اشتراک‌گذاری**
- **لینک اشتراک** سازگار با **V2ray**، **Clash** و **ClashMeta**
- تولیدکننده خودکار **لینک اشتراک** و **QRcode**
- نظارت بر سیستم و **آمار ترافیک**

**🛠️ ابزارها و سفارشی‌سازی**
- پیکربندی قابل تنظیم sing-box
- **ربات تلگرام** یکپارچه
- **رابط خط فرمان (CLI)**
- پشتیبانی از **چند زبان**
- پشتیبانی از **چند ادمین** (در حال توسعه)

</div>

---

# 🚀 راهنمای نصب

> **شروع سریع** - پاسارگارد را در چند دقیقه راه‌اندازی کنید

### برای راه‌اندازی سریع، از دستورات زیر بر اساس دیتابیس مورد نظرتان استفاده کنید.

---

**TimescaleDB (توصیه شده):**
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

### 📋 پس از نصب:

<div align="right">

**📋 لاگ‌ها را مشاهده کنید** (برای توقف `Ctrl+C` را فشار دهید)

**📁 فایل‌ها در مسیر** `/opt/pasarguard` قرار دارند

**⚙️ فایل پیکربندی:** `/opt/pasarguard/.env` (برای جزئیات [پیکربندی](#-پیکربندی) را ببینید)

**💾 فایل‌های داده:** `/var/lib/pasarguard`

**🔒 مهم:** داشبورد برای امنیت نیاز به گواهی SSL دارد
- دریافت گواهی SSL (به مستندات مراجعه کنید)
- دسترسی: `https://YOUR_DOMAIN:8000/dashboard/`

**🔗 برای تست بدون دامنه:** از SSH port forwarding استفاده کنید (پایین را ببینید)

</div>

---

```bash
ssh -L 8000:localhost:8000 user@serverip
```

سپس دسترسی: `http://localhost:8000/dashboard/`

> ⚠️ **فقط برای تست** - با بستن ترمینال SSH دسترسی خود را از دست خواهید داد.

### 🔧 مراحل بعدی:

```bash
# ایجاد حساب ادمین
pasarguard cli admins --create <username>

# دریافت راهنما
pasarguard --help
```

---

# 📚 مستندات

<div align="right">

**📖 مستندات** — راهنماها در حال توسعه برای فورک sing-box

🇺🇸 English | 🇮🇷 فارسی | 🇷🇺 Русский | 🇨🇳 简体中文

</div>

> **پروژه اصلی:** [github.com/PasarGuard/panel](https://github.com/PasarGuard/panel) | مستندات اصلی: [docs.pasarguard.org](https://docs.pasarguard.org)

---

<p align="center">
  Made with ❤️ for Internet freedom
</p>

> **Это форк сообщества PasarGuard, который заменяет Xray-core на sing-box.**

<h1 align="center">🛡️ PasarGuard</h1>

<p align="center">
    <strong>Унифицированное решение для управления прокси, устойчивое к цензуре — версия sing-box</strong>
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

## 📋 Содержание

> **Быстрая навигация** - Перейдите к любому разделу ниже

-   [📖 Обзор](#-обзор)
    -   [🤔 Зачем использовать PasarGuard?](#-зачем-использовать-pasarguard)
        -   [✨ Функции](#-функции)
-   [🚀 Руководство по установке](#-руководство-по-установке)
-   [📚 Документация](#-документация)

---

# 📖 Обзор

> **Что такое PasarGuard?**

PasarGuard — это мощный инструмент управления прокси-серверами, который предлагает интуитивно понятный и эффективный интерфейс для работы с сотнями прокси-аккаунтов. Построенный на Python и React.js, он сочетает производительность, масштабируемость и простоту использования для упрощения управления прокси в больших масштабах. Этот форк заменяет Xray-core на [sing-box](https://github.com/SagerNet/sing-box) для максимальной производительности, наряду с поддержкой [WireGuard](https://www.wireguard.com/).

---

## 🤔 Зачем использовать PasarGuard?

> **Простой, Мощный, Надежный**

PasarGuard — это удобный, многофункциональный и надежный инструмент управления прокси-серверами. Он позволяет создавать и управлять несколькими прокси для ваших пользователей без необходимости сложной настройки. С помощью встроенного веб-интерфейса вы можете легко отслеживать активность, изменять настройки и контролировать ограничения доступа пользователей — все из одного удобного панели управления.

---

### ✨ Функции

<div align="left">

**🌐 Веб-интерфейс и API**
- Встроенная панель управления **Web UI**
- Полностью функциональный бэкенд **REST API**
- Поддержка **Multi-Node** для распределения инфраструктуры

**🔐 Протоколы и безопасность**
- Поддержка **Vmess**, **VLESS**, **Trojan**, **Shadowsocks**, **WireGuard** и **Hysteria2**
- Поддержка **TLS** и **REALITY**
- **Мультипротокол** для одного пользователя

**👥 Управление пользователями**
- **Мультипользователь** на одном inbound
- **Мультиinbound** на **одном порту** (поддержка fallbacks)
- Ограничения по **трафику** и **сроку действия**
- **Периодические** ограничения трафика (ежедневно, еженедельно и т.д.)

**🔗 Подписки и обмен**
- **Ссылка подписки** совместимая с **V2ray**, **Clash** и **ClashMeta**
- Автоматический генератор **ссылок для обмена** и **QR-кодов**
- Мониторинг системы и **статистика трафика**

**🛠️ Инструменты и настройка**
- Настраиваемая конфигурация sing-box
- Встроенный **Telegram Bot**
- **Интерфейс командной строки (CLI)**
- Поддержка **многоязычности**
- Поддержка **множественных администраторов** (в разработке)

</div>

---

# 🚀 Руководство по установке

> **Быстрый старт** - Запустите PasarGuard за несколько минут

### Для быстрой настройки используйте следующие команды в зависимости от предпочитаемой базы данных.

---

**TimescaleDB (Рекомендуется):**
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

### 📋 После установки:

<div align="left">

**📋 Следите за логами** (нажмите `Ctrl+C` для остановки)

**📁 Файлы находятся в** `/opt/pasarguard`

**⚙️ Файл конфигурации:** `/opt/pasarguard/.env` (см. [Конфигурация](#-конфигурация) для деталей)

**💾 Файлы данных:** `/var/lib/pasarguard`

**🔒 Важно:** Панель управления требует SSL-сертификат для безопасности
- Получить SSL-сертификат (см. документацию)
- Доступ: `https://YOUR_DOMAIN:8000/dashboard/`

**🔗 Для тестирования без домена:** Используйте SSH port forwarding (см. ниже)

</div>

---

```bash
ssh -L 8000:localhost:8000 user@serverip
```

Затем доступ: `http://localhost:8000/dashboard/`

> ⚠️ **Только для тестирования** - Вы потеряете доступ при закрытии SSH-терминала.

### 🔧 Следующие шаги:

```bash
# Создать учетную запись администратора
pasarguard cli admins --create <username>

# Получить справку
pasarguard --help
```

---

# 📚 Документация

<div align="left">

**📖 Документация** — Руководства в разработке для sing-box форка

🇺🇸 English | 🇮🇷 فارسی | 🇷🇺 Русский | 🇨🇳 简体中文

</div>

> **Оригинальный проект:** [github.com/PasarGuard/panel](https://github.com/PasarGuard/panel) | Оригинальные доки: [docs.pasarguard.org](https://docs.pasarguard.org)

---

<p align="center">
  Made with ❤️ for Internet freedom
</p>

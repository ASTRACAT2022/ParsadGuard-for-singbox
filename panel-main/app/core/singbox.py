from __future__ import annotations

import json
from copy import deepcopy
from pathlib import PosixPath
from typing import Union

import commentjson

from app.core.abstract_core import AbstractCore
from app.models.core import CoreType
from app.models.protocol import ProxyProtocol
from app.utils.crypto import get_x25519_public_key


# sing-box inbound type -> panel-side protocol name in inbounds_by_tag
_SINGBOX_TO_PANEL_PROTOCOL = {
    "vless": "vless",
    "vmess": "vmess",
    "trojan": "trojan",
    "shadowsocks": "shadowsocks",
    "hysteria2": "hysteria",
    "hysteria": "hysteria",
    "tuic": "tuic",
    "shadowtls": "shadowtls",
    "naive": "naive",
    "anytls": "anytls",
}

# sing-box transport type -> panel "network" identifier
_SINGBOX_TRANSPORT_TO_NETWORK = {
    "http": "tcp",
    "ws": "ws",
    "grpc": "grpc",
    "httpupgrade": "httpupgrade",
    "quic": "quic",
}


def _protocols_from_inbounds_by_tag(inbounds_by_tag: dict[str, dict]) -> frozenset[ProxyProtocol]:
    return frozenset(
        protocol
        for inbound in inbounds_by_tag.values()
        if (protocol := ProxyProtocol.from_value(inbound.get("protocol"))) is not None
    )


class SingBoxConfig(dict, AbstractCore):
    """Normalized representation of a SagerNet/sing-box configuration.

    The instance behaves like a ``dict`` with the raw JSON document as content and
    additionally produces the same kind of ``inbounds_by_tag`` map that the
    existing subscription / hosts pipelines consume from :class:`XRayConfig`.
    """

    def __init__(
        self,
        config: Union[dict, str, PosixPath] | None = None,
        exclude_inbound_tags: set[str] | None = None,
        fallbacks_inbound_tags: set[str] | None = None,
        skip_validation: bool = False,
    ) -> None:
        if config is None:
            config = {}
        if isinstance(config, (str, PosixPath)):
            config = commentjson.loads(str(config) if isinstance(config, PosixPath) else config)
        if isinstance(config, dict):
            config = deepcopy(config)
        super().__init__(config)

        self._type = CoreType.singbox
        if exclude_inbound_tags is None:
            exclude_inbound_tags = set()
        if fallbacks_inbound_tags is None:
            fallbacks_inbound_tags = set()
        # sing-box has no Xray-style fallback inbound feature, but the panel
        # treats fallback tags as "exclude from user sync", so we still respect
        # the union.
        exclude_inbound_tags.update(fallbacks_inbound_tags)
        self.exclude_inbound_tags = exclude_inbound_tags
        self.fallbacks_inbound_tags = set(fallbacks_inbound_tags)

        self._inbounds: list[str] = []
        self._inbounds_by_tag: dict[str, dict] = {}
        self._protocols: frozenset[ProxyProtocol] = frozenset()

        if skip_validation:
            return

        self._validate()
        self._resolve_inbounds()

    # ---------------------------------------------------------------- helpers
    def _validate(self) -> None:
        if not self.get("inbounds"):
            raise ValueError("config doesn't have inbounds")
        for inbound in self["inbounds"]:
            tag = inbound.get("tag")
            if not tag:
                raise ValueError("all inbounds must have a unique tag")
            if "," in tag:
                raise ValueError("character «,» is not allowed in inbound tag")
            if "<=>" in tag:
                raise ValueError("character «<=>» is not allowed in inbound tag")
            if not inbound.get("type"):
                raise ValueError(f'inbound "{tag}" must have a type')

    def _resolve_inbounds(self) -> None:
        for inbound in self.get("inbounds", []):
            self._read_inbound(inbound)
        # WireGuard endpoints are top-level "endpoints" in sing-box 1.11+
        for endpoint in self.get("endpoints", []) or []:
            self._read_endpoint(endpoint)
        self._protocols = _protocols_from_inbounds_by_tag(self._inbounds_by_tag)

    # --- inbound parsing ------------------------------------------------
    def _read_inbound(self, inbound: dict) -> None:
        tag = inbound["tag"]
        itype = inbound["type"]
        protocol = _SINGBOX_TO_PANEL_PROTOCOL.get(itype)
        if not protocol:
            return
        if tag in self.exclude_inbound_tags:
            return

        settings = self._build_base_settings(inbound, protocol)

        # transport
        transport = inbound.get("transport") or {}
        ttype = transport.get("type")
        if ttype:
            settings["network"] = _SINGBOX_TRANSPORT_TO_NETWORK.get(ttype, ttype)
            self._handle_transport(transport, settings)

        # tls / reality
        tls = inbound.get("tls") or {}
        if tls.get("enabled"):
            reality = tls.get("reality") or {}
            if reality.get("enabled"):
                self._handle_reality(tls, reality, settings, tag)
            else:
                self._handle_tls(tls, settings)

        # protocol-specific fields
        if protocol == "vless":
            settings["flow"] = inbound.get("flow", "")
            settings["encryption"] = "none"
        elif protocol == "shadowsocks":
            self._handle_shadowsocks(inbound, settings)
        elif protocol == "hysteria":
            self._handle_hysteria(inbound, settings)
        elif protocol == "tuic":
            settings["congestion_control"] = inbound.get("congestion_control", "bbr")
        elif protocol == "shadowtls":
            settings["shadowtls_version"] = inbound.get("version", 3)

        if tag not in self._inbounds:
            self._inbounds.append(tag)
            self._inbounds_by_tag[tag] = settings

    def _build_base_settings(self, inbound: dict, protocol: str) -> dict:
        port = inbound.get("listen_port") or inbound.get("port")
        if port is None:
            raise ValueError(f'{inbound["tag"]} inbound is missing listen_port')
        return {
            "tag": inbound["tag"],
            "protocol": protocol,
            "port": port,
            "network": "tcp",
            "tls": "none",
            "sni": [],
            "host": [],
            "path": "",
            "header_type": "",
            "is_fallback": False,
            "fallbacks": [],
            "finalmask": None,
        }

    def _handle_transport(self, transport: dict, settings: dict) -> None:
        ttype = transport.get("type")
        if ttype == "ws":
            path = transport.get("path", "")
            if max_early := transport.get("max_early_data"):
                hdr = transport.get("early_data_header_name") or "Sec-WebSocket-Protocol"
                if hdr == "Sec-WebSocket-Protocol":
                    sep = "&" if "?" in path else "?"
                    path = f"{path}{sep}ed={max_early}"
            settings["path"] = path
            host = transport.get("headers", {}).get("Host") or transport.get("headers", {}).get("host")
            if isinstance(host, list):
                settings["host"] = [host[0]] if host else []
            elif isinstance(host, str):
                settings["host"] = [host]
        elif ttype == "grpc":
            settings["path"] = transport.get("service_name", "")
            settings["host"] = []
        elif ttype == "httpupgrade":
            settings["path"] = transport.get("path", "")
            host = transport.get("host", "")
            if host:
                settings["host"] = [host]
        elif ttype == "http":
            settings["path"] = transport.get("path", "")
            hosts = transport.get("host") or []
            if isinstance(hosts, list):
                settings["host"] = hosts
            elif isinstance(hosts, str):
                settings["host"] = [hosts]
            settings["header_type"] = "http"
        elif ttype == "quic":
            settings["network"] = "quic"

    def _handle_tls(self, tls: dict, settings: dict) -> None:
        settings["tls"] = "tls"
        sni = tls.get("server_name")
        if sni:
            settings["sni"] = [sni] if isinstance(sni, str) else list(sni)
        utls = tls.get("utls") or {}
        if utls.get("enabled") and (fp := utls.get("fingerprint")):
            settings["fp"] = fp
        alpn = tls.get("alpn")
        if alpn:
            settings["alpn"] = alpn if isinstance(alpn, list) else [alpn]

    def _handle_reality(self, tls: dict, reality: dict, settings: dict, tag: str) -> None:
        settings["tls"] = "reality"
        settings["fp"] = (tls.get("utls") or {}).get("fingerprint", "chrome")
        sni = tls.get("server_name")
        if sni:
            settings["sni"] = [sni] if isinstance(sni, str) else list(sni)

        pvk = reality.get("private_key")
        if not pvk:
            raise ValueError(f'{tag}: reality private_key is required')
        settings["pbk"] = get_x25519_public_key(pvk)
        if not settings.get("pbk"):
            raise ValueError(f"{tag}: failed to derive reality public key")

        short_ids = reality.get("short_id")
        if isinstance(short_ids, str):
            short_ids = [short_ids]
        if not short_ids:
            raise ValueError(f"{tag}: at least one short_id is required for reality")
        settings["sids"] = list(short_ids)

        # spider X path is not part of sing-box; leave empty
        settings["spx"] = ""

    def _handle_shadowsocks(self, inbound: dict, settings: dict) -> None:
        method = inbound.get("method", "")
        settings["method"] = method
        settings["is_2022"] = method.startswith("2022-blake3")
        settings["password"] = inbound.get("password", "")

    def _handle_hysteria(self, inbound: dict, settings: dict) -> None:
        # Pack hysteria2-specific parameters so that subscription/links code can
        # render salamander obfs and BBR/Brutal QUIC params.
        finalmask: dict = {}
        if obfs := inbound.get("obfs"):
            finalmask["obfsPassword"] = obfs.get("password", "")
        if up := inbound.get("up_mbps"):
            finalmask["brutalUp"] = f"{up}M"
        if down := inbound.get("down_mbps"):
            finalmask["brutalDown"] = f"{down}M"
        if hop := inbound.get("server_ports"):
            finalmask["udpHop"] = {"ports": hop}
        if finalmask:
            settings["finalmask"] = finalmask

    def _read_endpoint(self, endpoint: dict) -> None:
        if endpoint.get("type") != "wireguard":
            return
        tag = endpoint.get("tag")
        if not tag or tag in self.exclude_inbound_tags:
            return
        settings = {
            "tag": tag,
            "protocol": "wireguard",
            "port": (endpoint.get("peers") or [{}])[0].get("port"),
            "network": "wg",
            "tls": "none",
            "sni": [],
            "host": [],
            "path": "",
            "header_type": "",
            "is_fallback": False,
            "fallbacks": [],
            "finalmask": None,
            "listen_port": (endpoint.get("peers") or [{}])[0].get("port"),
            "public_key": (endpoint.get("peers") or [{}])[0].get("public_key", ""),
            "address": endpoint.get("address", []),
            "pre_shared_key": (endpoint.get("peers") or [{}])[0].get("pre_shared_key"),
        }
        self._inbounds.append(tag)
        self._inbounds_by_tag[tag] = settings

    # --- AbstractCore api -----------------------------------------------
    def to_str(self, **json_kwargs) -> str:
        return json.dumps(self, **json_kwargs)

    @property
    def inbounds_by_tag(self) -> dict:
        return self._inbounds_by_tag

    @property
    def inbounds(self) -> list[str]:
        return self._inbounds

    @property
    def protocols(self) -> frozenset[ProxyProtocol]:
        return self._protocols

    @property
    def type(self) -> str:
        return self._type

    def to_json(self) -> dict:
        return {
            "type": self.type,
            "config": dict(self),
            "exclude_inbound_tags": list(self.exclude_inbound_tags),
            "fallbacks_inbound_tags": list(self.fallbacks_inbound_tags),
            "inbounds": self.inbounds,
            "inbounds_by_tag": self.inbounds_by_tag,
        }

    @classmethod
    def from_json(cls, data: dict) -> "SingBoxConfig":
        instance = cls(
            config=data.get("config", {}),
            exclude_inbound_tags=set(data.get("exclude_inbound_tags", [])),
            fallbacks_inbound_tags=set(data.get("fallbacks_inbound_tags", []) or []),
            skip_validation=True,
        )
        if "inbounds" in data:
            instance._inbounds = data["inbounds"]
        if "inbounds_by_tag" in data:
            instance._inbounds_by_tag = data["inbounds_by_tag"]
        instance._protocols = _protocols_from_inbounds_by_tag(instance._inbounds_by_tag)
        return instance

    def copy(self) -> "SingBoxConfig":
        return deepcopy(self)

    def get_inbound(self, tag: str) -> dict | None:
        for inbound in self.get("inbounds", []):
            if inbound.get("tag") == tag:
                return inbound
        for endpoint in self.get("endpoints", []) or []:
            if endpoint.get("tag") == tag:
                return endpoint
        return None

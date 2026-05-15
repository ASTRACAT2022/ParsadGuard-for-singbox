NAME = pasarguard-node-$(GOOS)-$(GOARCH)

LDFLAGS = -s -w -buildid=
PARAMS = -trimpath -ldflags "$(LDFLAGS)" -v
MAIN = ./cmd/node
PREFIX ?= $(shell go env GOPATH)

# Default sing-box release; can be overridden, e.g. `make install_singbox SINGBOX_VERSION=v1.10.7`.
SINGBOX_VERSION ?= v1.10.7
SINGBOX_OS      ?= $(if $(GOOS),$(GOOS),linux)

# Map GOARCH to sing-box release tarball arch suffix.
SINGBOX_ARCH_MAP_amd64 = amd64
SINGBOX_ARCH_MAP_arm64 = arm64
SINGBOX_ARCH_MAP_armv7 = armv7
SINGBOX_ARCH_MAP_arm   = armv7
SINGBOX_ARCH_MAP_386   = 386
SINGBOX_ARCH         ?= $(SINGBOX_ARCH_MAP_$(GOARCH))

ifeq ($(GOOS),windows)
OUTPUT = $(NAME).exe
ADDITION = go build -o w$(NAME).exe -trimpath -ldflags "-H windowsgui $(LDFLAGS)" -v $(MAIN)
else
OUTPUT = $(NAME)
endif

ifeq ($(shell echo "$(GOARCH)" | grep -Eq "(mips|mipsle)" && echo true),true)
ADDITION = GOMIPS=softfloat go build -o $(NAME)_softfloat -trimpath -ldflags "$(LDFLAGS)" -v $(MAIN)
endif

.PHONY: clean build deps generate_grpc_code generate_server_cert generate_client_cert \
        update_os install_singbox install_wg test test-race-wireguard \
        test-integration test-integration-full test-integration-wireguard serve

build:
	CGO_ENABLED=0 go build -o $(OUTPUT) $(PARAMS) $(MAIN)
	$(ADDITION)

clean:
	go clean -v -i $(PWD)
	rm -f $(NAME)-* w$(NAME)-*.exe

deps:
	go mod download
	go mod tidy

generate_grpc_code:
	protoc \
	--go_out=. \
	--go_opt=paths=source_relative \
	--go-grpc_out=. \
	--go-grpc_opt=paths=source_relative \
	common/service.proto

CN ?= localhost
SAN ?= DNS:localhost,IP:127.0.0.1

generate_server_cert:
	mkdir -p ./certs
	openssl req -x509 -newkey rsa:4096 -keyout ./certs/ssl_key.pem \
	-out ./certs/ssl_cert.pem -days 36500 -nodes \
	-subj "/CN=$(CN)" \
	-addext "subjectAltName = $(SAN)"

generate_client_cert:
	mkdir -p ./certs
	openssl req -x509 -newkey rsa:4096 -keyout ./certs/ssl_client_key.pem \
 	-out ./certs/ssl_client_cert.pem -days 36500 -nodes \
	-subj "/CN=$(CN)" \
	-addext "subjectAltName = $(SAN)"

UNAME_S := $(shell uname -s)
UNAME_M := $(shell uname -m)
DISTRO  := $(shell . /etc/os-release 2>/dev/null && echo $$ID || echo "unknown")

update_os:
ifeq ($(UNAME_S),Linux)
	@echo "Detected OS: Linux"
	@echo "Distribution: $(DISTRO)"

	if [ "$(DISTRO)" = "debian" ] || [ "$(DISTRO)" = "ubuntu" ]; then \
		sudo apt-get update && \
		sudo apt-get install -y curl bash tar; \
	fi
	if [ "$(DISTRO)" = "alpine" ]; then \
		apk update && \
		apk add --no-cache curl bash tar; \
	fi
	if [ "$(DISTRO)" = "centos" ] || [ "$(DISTRO)" = "rhel" ] || [ "$(DISTRO)" = "fedora" ]; then \
		sudo yum update -y && \
		sudo yum install -y curl bash tar; \
	fi
	if [ "$(DISTRO)" = "arch" ]; then \
		sudo pacman -Sy --noconfirm curl bash tar; \
	fi
else
	@echo "Unsupported operating system: $(UNAME_S)"
	@exit 1
endif

install_singbox: update_os
ifeq ($(UNAME_S),Linux)
	@if [ -z "$(SINGBOX_ARCH)" ]; then \
		echo "Unsupported GOARCH '$(GOARCH)' for sing-box install"; exit 1; \
	fi
	@VER_NUM=$$(echo "$(SINGBOX_VERSION)" | sed 's/^v//'); \
	TARBALL="sing-box-$${VER_NUM}-$(SINGBOX_OS)-$(SINGBOX_ARCH).tar.gz"; \
	URL="https://github.com/SagerNet/sing-box/releases/download/$(SINGBOX_VERSION)/$${TARBALL}"; \
	echo "Downloading $${URL}"; \
	curl -fsSL -o /tmp/sing-box.tgz "$${URL}"; \
	tar -xzf /tmp/sing-box.tgz -C /tmp; \
	if [ "$(DISTRO)" = "debian" ] || [ "$(DISTRO)" = "ubuntu" ] || \
	   [ "$(DISTRO)" = "centos" ]   || [ "$(DISTRO)" = "rhel" ] || [ "$(DISTRO)" = "fedora" ] || \
	   [ "$(DISTRO)" = "arch" ]; then \
		sudo install -m 0755 /tmp/sing-box-$${VER_NUM}-$(SINGBOX_OS)-$(SINGBOX_ARCH)/sing-box /usr/local/bin/sing-box; \
	else \
		install -m 0755 /tmp/sing-box-$${VER_NUM}-$(SINGBOX_OS)-$(SINGBOX_ARCH)/sing-box /usr/local/bin/sing-box; \
	fi
	@sing-box version || /usr/local/bin/sing-box version
else
	@echo "Unsupported operating system: $(UNAME_S)"
	@exit 1
endif

install_wg: update_os
ifeq ($(UNAME_S),Linux)
	@echo "Installing WireGuard for $(DISTRO)..."
	if [ "$(DISTRO)" = "debian" ] || [ "$(DISTRO)" = "ubuntu" ]; then \
		sudo apt-get update && sudo apt-get install -y wireguard wireguard-tools; \
	elif [ "$(DISTRO)" = "centos" ] || [ "$(DISTRO)" = "rhel" ]; then \
		sudo yum install -y epel-release && sudo yum install -y wireguard-tools; \
	elif [ "$(DISTRO)" = "fedora" ]; then \
		sudo dnf install -y wireguard-tools; \
	elif [ "$(DISTRO)" = "arch" ] || [ "$(DISTRO)" = "manjaro" ]; then \
		sudo pacman -Sy --noconfirm wireguard-tools; \
	elif [ "$(DISTRO)" = "alpine" ]; then \
		apk add --no-cache wireguard-tools; \
	elif [ "$(DISTRO)" = "opensuse" ] || [ "$(DISTRO)" = "opensuse-leap" ] || [ "$(DISTRO)" = "opensuse-tumbleweed" ]; then \
		sudo zypper install -y wireguard-tools; \
	else \
		echo "Unsupported distribution: $(DISTRO)"; \
		exit 1; \
	fi
	@echo "WireGuard installed successfully"
	@wg --version
else
	@echo "Unsupported operating system: $(UNAME_S)"
	@exit 1
endif

test-integration:
	$(MAKE) test-integration-full

test-integration-full:
	GOTOOLCHAIN=auto TEST_INTEGRATION=true go test ./... -v -p 1
	GOTOOLCHAIN=auto go test -tags=integration -v -p 1 ./backend/wireguard

test-integration-wireguard:
	GOTOOLCHAIN=auto go test -tags=integration -v -p 1 ./backend/wireguard

test:
	GOTOOLCHAIN=auto TEST_INTEGRATION=false go test ./... -v -p 1

test-race-wireguard:
	GOTOOLCHAIN=auto CGO_ENABLED=1 go test -race -v -p 1 ./backend/wireguard

serve:
	go run ./cmd/node

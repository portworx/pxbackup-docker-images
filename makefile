# Name of the binary
BINARY_NAME := pxb-diags

# Base directory to store the binaries
BINARY_DIR := binary

# Directories for different platforms and architectures
LINUX_DIR := $(BINARY_DIR)/linux
MAC_DIR := $(BINARY_DIR)/mac
WINDOWS_DIR := $(BINARY_DIR)/windows

# Subdirectories for different architectures
LINUX64_DIR := $(LINUX_DIR)/x86_64
LINUX32_DIR := $(LINUX_DIR)/i386
LINUXARM64_DIR := $(LINUX_DIR)/arm64
LINUXARM32_DIR := $(LINUX_DIR)/arm
MAC64_DIR := $(MAC_DIR)/x86_64
MACARM64_DIR := $(MAC_DIR)/arm64
WINDOWS64_DIR := $(WINDOWS_DIR)/AMD64
WINDOWS32_DIR := $(WINDOWS_DIR)/x86
WINDOWSARM64_DIR := $(WINDOWS_DIR)/ARM64

# Go build flags
GO_FLAGS := -ldflags "-s -w"

# Targets for different platforms and architectures
all: linux64 linux32 linuxarm64 linuxarm32 mac64 macarm64 windows64 windows32 windowsarm64

# Build for Linux 64-bit
linux64: | $(LINUX64_DIR)
	GOOS=linux GOARCH=amd64 go build $(GO_FLAGS) -o $(LINUX64_DIR)/$(BINARY_NAME)

# Build for Linux 32-bit
linux32: | $(LINUX32_DIR)
	GOOS=linux GOARCH=386 go build $(GO_FLAGS) -o $(LINUX32_DIR)/$(BINARY_NAME)

# Build for Linux ARM 64-bit
linuxarm64: | $(LINUXARM64_DIR)
	GOOS=linux GOARCH=arm64 go build $(GO_FLAGS) -o $(LINUXARM64_DIR)/$(BINARY_NAME)

# Build for Linux ARM 32-bit
linuxarm32: | $(LINUXARM32_DIR)
	GOOS=linux GOARCH=arm go build $(GO_FLAGS) -o $(LINUXARM32_DIR)/$(BINARY_NAME)

# Build for macOS 64-bit
mac64: | $(MAC64_DIR)
	GOOS=darwin GOARCH=amd64 go build $(GO_FLAGS) -o $(MAC64_DIR)/$(BINARY_NAME)

# Build for macOS ARM 64-bit
macarm64: | $(MACARM64_DIR)
	GOOS=darwin GOARCH=arm64 go build $(GO_FLAGS) -o $(MACARM64_DIR)/$(BINARY_NAME)

# Build for Windows 64-bit
windows64: | $(WINDOWS64_DIR)
	GOOS=windows GOARCH=amd64 go build $(GO_FLAGS) -o $(WINDOWS64_DIR)/$(BINARY_NAME).exe

# Build for Windows 32-bit
windows32: | $(WINDOWS32_DIR)
	GOOS=windows GOARCH=386 go build $(GO_FLAGS) -o $(WINDOWS32_DIR)/$(BINARY_NAME).exe

# Build for Windows ARM 64-bit
windowsarm64: | $(WINDOWSARM64_DIR)
	GOOS=windows GOARCH=arm64 go build $(GO_FLAGS) -o $(WINDOWSARM64_DIR)/$(BINARY_NAME).exe

# Clean binaries
clean:
	rm -rf $(BINARY_DIR)

# Create the output directories if they don't exist
$(LINUX64_DIR) $(LINUX32_DIR) $(LINUXARM64_DIR) $(LINUXARM32_DIR) $(MAC64_DIR) $(MACARM64_DIR) $(WINDOWS64_DIR) $(WINDOWS32_DIR) $(WINDOWSARM64_DIR):
	mkdir -p $@

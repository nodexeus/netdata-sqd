#!/bin/bash

set -e

# Print usage information
function usage() {
    echo "Usage: $0 [OPTIONS]"
    echo "Build the SQD collector plugin for netdata"
    echo ""
    echo "Options:"
    echo "  -h, --help     Display this help message"
    echo "  -d, --debug    Build with debug symbols"
    echo "  -i, --install  Install to netdata plugin directory"
    echo ""
}

# Default values
DEBUG=0
INSTALL=0
NETDATA_PLUGIN_DIR="/usr/libexec/netdata/plugins.d"

# Parse arguments
while [[ $# -gt 0 ]]; do
    case "$1" in
        -h|--help)
            usage
            exit 0
            ;;
        -d|--debug)
            DEBUG=1
            shift
            ;;
        -i|--install)
            INSTALL=1
            shift
            ;;
        *)
            echo "Unknown option: $1"
            usage
            exit 1
            ;;
    esac
done

# Create build directory
mkdir -p build

# Build flags
BUILD_FLAGS="-trimpath"

if [ "$DEBUG" -eq 1 ]; then
    echo "Building with debug symbols..."
    BUILD_FLAGS="$BUILD_FLAGS -gcflags=all=-N -l"
else
    echo "Building optimized binary..."
    BUILD_FLAGS="$BUILD_FLAGS -ldflags=-w"
fi

# Build the plugin
echo "Building SQD collector plugin..."
go build $BUILD_FLAGS -o build/sqd.plugin ./cmd/go.d.plugin

if [ $? -eq 0 ]; then
    echo "Build successful! Binary located at: build/sqd.plugin"

    # Install if requested
    if [ "$INSTALL" -eq 1 ]; then
        echo "Installing to $NETDATA_PLUGIN_DIR..."
        
        if [ ! -d "$NETDATA_PLUGIN_DIR" ]; then
            echo "Error: Netdata plugin directory not found at $NETDATA_PLUGIN_DIR"
            echo "Please ensure netdata is installed or specify the correct directory"
            exit 1
        fi
        
        # Create config directories if they don't exist
        sudo mkdir -p /etc/netdata/go.d
        
        # Copy the binary
        sudo cp build/sqd.plugin "$NETDATA_PLUGIN_DIR/"
        
        # Make it executable
        sudo chmod 755 "$NETDATA_PLUGIN_DIR/sqd.plugin"
        
        # Copy default configuration
        echo "Creating default configuration..."
        cat > build/sqd.conf << EOF
# SQD collector configuration
update_every: 1

workers:
  - name: default
    prometheus_url: http://localhost:9090
    graphql_url: http://localhost:8080
    port: 9090
EOF
        sudo cp build/sqd.conf /etc/netdata/go.d/
        
        echo "Installation complete! Restart netdata to activate the plugin:"
        echo "sudo systemctl restart netdata"
    fi
else
    echo "Build failed!"
    exit 1
fi

#!/bin/bash

echo "=========================================="
echo "  Fixing npm Installation"
echo "=========================================="
echo ""

# First, let's completely remove old installations
echo "Cleaning up old installations..."
sudo apt-get purge -y nodejs npm
sudo apt-get autoremove -y

# Remove any leftover files
sudo rm -rf /usr/local/bin/npm /usr/local/share/man/man1/node* /usr/local/lib/dtrace/node.d
sudo rm -rf ~/.npm ~/.node-gyp /opt/local/bin/node /opt/local/include/node /opt/local/lib/node_modules
sudo rm -rf /usr/local/lib/node* /usr/local/include/node* /usr/local/bin/node*

echo ""
echo "Installing Node.js 20 LTS with npm..."
echo ""

# Install using NodeSource repository
curl -fsSL https://deb.nodesource.com/setup_20.x | sudo -E bash -
sudo apt-get install -y nodejs

echo ""
echo "Verifying installation..."
which node
which npm
node --version
npm --version

if ! command -v npm &> /dev/null; then
    echo ""
    echo "npm still not found. Installing manually..."
    sudo apt-get install -y npm
fi

echo ""
echo "=========================================="
echo "  Installation Summary"
echo "=========================================="
echo "Node.js: $(node --version)"
echo "npm: $(npm --version)"
echo "Node location: $(which node)"
echo "npm location: $(which npm)"
echo ""

if command -v npm &> /dev/null; then
    echo "✅ npm is now available!"
    echo ""
    echo "Now run:"
    echo "  cd /home/kaijage/model/inform/masuby-model"
    echo "  npm install"
    echo "  npm run dev"
else
    echo "❌ npm installation failed"
    echo ""
    echo "Try manual installation:"
    echo "  sudo apt-get update"
    echo "  sudo apt-get install -y nodejs npm"
fi

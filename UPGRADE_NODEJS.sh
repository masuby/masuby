#!/bin/bash

echo "=========================================="
echo "  Upgrading Node.js to Version 20 LTS"
echo "=========================================="
echo ""

echo "Current Node.js version: $(node --version)"
echo ""
echo "Vite 7 requires Node.js 20.19+ or 22.12+"
echo ""

read -p "Upgrade to Node.js 20 LTS? (y/n) " -n 1 -r
echo ""

if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Upgrade cancelled."
    exit 1
fi

echo ""
echo "📦 Removing old Node.js version..."
sudo apt-get remove -y nodejs npm

echo ""
echo "📥 Adding Node.js 20 LTS repository..."
curl -fsSL https://deb.nodesource.com/setup_20.x | sudo -E bash -

echo ""
echo "📦 Installing Node.js 20 and npm..."
sudo apt-get install -y nodejs

echo ""
echo "✅ Installation complete!"
echo ""
echo "New versions:"
echo "  Node.js: $(node --version)"
echo "  npm: $(npm --version)"
echo ""

echo "=========================================="
echo "  Now run the app with:"
echo "  bash /home/kaijage/model/inform/START_LOCAL.sh"
echo "=========================================="

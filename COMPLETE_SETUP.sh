#!/bin/bash

echo "=========================================="
echo "  INFORM Tanzania - Complete Setup"
echo "=========================================="
echo ""
echo "This will:"
echo "  1. Upgrade Node.js to version 20"
echo "  2. Install npm"
echo "  3. Install project dependencies"
echo "  4. Copy boundary files"
echo "  5. Start the development server"
echo ""
echo "You will be asked for your password (8b)"
echo ""

read -p "Continue? (y/n) " -n 1 -r
echo ""
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    exit 1
fi

echo ""
echo "=========================================="
echo "  Step 1: Upgrading Node.js"
echo "=========================================="
echo ""

# Remove old versions
echo "Removing old Node.js/npm..."
sudo apt-get remove -y nodejs npm 2>/dev/null

# Add Node.js 20 repository
echo ""
echo "Adding Node.js 20 LTS repository..."
curl -fsSL https://deb.nodesource.com/setup_20.x | sudo -E bash -

# Install Node.js and npm
echo ""
echo "Installing Node.js 20 and npm..."
sudo apt-get install -y nodejs

# Verify installation
echo ""
echo "✅ Installation complete!"
echo "   Node.js: $(node --version)"
echo "   npm: $(npm --version)"

echo ""
echo "=========================================="
echo "  Step 2: Installing Dependencies"
echo "=========================================="
echo ""

cd /home/kaijage/model/inform/masuby-model

echo "Installing project dependencies..."
echo "(This may take 2-3 minutes)"
npm install

if [ $? -ne 0 ]; then
    echo ""
    echo "❌ npm install failed!"
    exit 1
fi

echo ""
echo "=========================================="
echo "  Step 3: Setting Up Boundary Files"
echo "=========================================="
echo ""

mkdir -p public/geojson

if [ ! -f "public/geojson/ADM1.geojson" ]; then
    echo "Copying ADM1 boundaries..."
    cp /home/kaijage/model/inform/Boundaries/adm1/adm1.geojson public/geojson/ADM1.geojson
fi

if [ ! -f "public/geojson/ADM2.geojson" ]; then
    echo "Copying ADM2 boundaries..."
    cp /home/kaijage/model/inform/Boundaries/adm2/adm2.geojson public/geojson/ADM2.geojson
fi

echo "✅ Boundary files ready"

echo ""
echo "=========================================="
echo "  ✅ Setup Complete!"
echo "=========================================="
echo ""
echo "Starting development server..."
echo ""
echo "Open your browser to: http://localhost:5173"
echo ""
echo "To use INFORM visualizations:"
echo "  1. Load: Tanzania - Country Model Template.xlsx"
echo "  2. Click: Select Visualization"
echo "  3. Choose: 🎯 INFORM Risk Dashboard"
echo ""
echo "Press Ctrl+C to stop the server"
echo ""
echo "=========================================="
echo ""

# Start the dev server
npm run dev

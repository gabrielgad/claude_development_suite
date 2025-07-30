#!/bin/bash
# Claude Manager PATH Setup
# Run: source setup-path.sh

if [[ ":$PATH:" != *":$HOME/.local/bin:"* ]]; then
    export PATH="$HOME/.local/bin:$PATH"
    echo "✅ Added ~/.local/bin to PATH for this session"
    echo ""
    echo "To make permanent, add this line to your shell config:"
    echo "  echo 'export PATH=\"\$HOME/.local/bin:\$PATH\"' >> ~/.bashrc"
    echo "  # or for zsh: echo 'export PATH=\"\$HOME/.local/bin:\$PATH\"' >> ~/.zshrc"
    echo "  # or for fish: fish_add_path ~/.local/bin"
else
    echo "✅ ~/.local/bin already in PATH"
fi

echo ""
echo "Available commands:"
echo "  claude-manager  # Main binary"
echo "  cm              # Short alias"  
echo "  claude-web      # Descriptive alias"
echo "  cmgr            # Manager alias"
echo ""
echo "Test installation:"
echo "  claude-manager --version"
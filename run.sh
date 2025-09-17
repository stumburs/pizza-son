#!/bin/bash

# Check if venv exists
if [ ! -f ".venv/bin/activate" ]; then
    echo "Creating virtual environment..."
    python3 -m venv .venv
fi

# Activate venv
source .venv/bin/activate

# Install dependencies
echo "Installing dependencies..."
pip install --upgrade -r requirements.txt

# Run the bot
python3 -m bot.bot

# Deactivate venv
deactivate
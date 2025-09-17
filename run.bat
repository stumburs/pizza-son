@echo off

:: Check if venv exists
if not exist ".venv\Scripts\activate" (
    echo Creating virtual environment...
    python -m venv .venv
)

:: Activate venv
call .venv\Scripts\activate

:: Install dependencies
echo Installing dependencies...
pip install --upgrade -r requirements.txt

:: Run the bot
python -m bot.bot

:: Deactivate the .venv
deactivate
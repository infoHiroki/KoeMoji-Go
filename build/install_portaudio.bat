@echo off
echo Installing PortAudio using MSYS2...
C:\msys64\usr\bin\bash.exe -l -c "pacman -S --noconfirm mingw-w64-x86_64-portaudio"
echo.
echo Installation complete!
pause
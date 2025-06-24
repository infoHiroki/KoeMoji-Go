@echo off
echo Copying required DLL files...

set SOURCE_DIR=C:\msys64\mingw64\bin
set DEST_DIR=%~dp0

rem Copy essential DLLs
copy /Y "%SOURCE_DIR%\libportaudio.dll" "%DEST_DIR%" >nul
copy /Y "%SOURCE_DIR%\libgcc_s_seh-1.dll" "%DEST_DIR%" >nul
copy /Y "%SOURCE_DIR%\libwinpthread-1.dll" "%DEST_DIR%" >nul

echo DLL files copied successfully.

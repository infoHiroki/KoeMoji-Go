@echo off
setlocal enabledelayedexpansion

rem KoeMoji-Go Windows Build Script
rem Version: Dynamically extracted from version.go

rem スクリプトのディレクトリに移動
cd /d "%~dp0"

rem バージョン情報をversion.goから動的に取得
for /f "tokens=3 delims== " %%i in ('findstr /C:"const Version" ..\..\version.go') do (
    set VERSION=%%i
    set VERSION=!VERSION:"=!
)
set APP_NAME=koemoji-go
set DIST_DIR=dist
set SOURCE_DIR=..\..\cmd\koemoji-go
set COMMON_DIR=..\common
rem Get GOPATH from go env if not set
if "%GOPATH%"=="" (
    for /f "tokens=*" %%i in ('go env GOPATH') do set GOPATH=%%i
)
set GOPATH_BIN=%GOPATH%\bin

rem Check command line arguments
if "%1"=="" goto :build_all
if /i "%1"=="clean" goto :clean
if /i "%1"=="help" goto :show_help
if /i "%1"=="-h" goto :show_help
if /i "%1"=="/?" goto :show_help
echo Error: Unknown option: %1
goto :show_help

:show_help
echo Usage: build.bat [options]
echo.
echo Options:
echo   (no args)   Build Windows executable with icon
echo   clean       Clean build artifacts
echo   help        Show this help message
echo.
exit /b 0

:clean
echo Cleaning build artifacts...
if exist %DIST_DIR% rmdir /s /q %DIST_DIR%
if exist temp rmdir /s /q temp
echo Clean completed.
exit /b 0

:build_all
echo ========================================
echo   Building KoeMoji-Go for Windows
echo   Version: %VERSION%
echo ========================================
echo.

rem Add MSYS2 MinGW64 to PATH temporarily
set PATH=C:\msys64\mingw64\bin;%PATH%
echo Added MSYS2 MinGW64 to PATH temporarily

rem Set PKG_CONFIG_PATH for PortAudio
set PKG_CONFIG_PATH=C:\msys64\mingw64\lib\pkgconfig
echo Set PKG_CONFIG_PATH for libraries

rem Check Go installation
where go >nul 2>nul
if %errorlevel% neq 0 (
    echo Error: Go is not installed or not in PATH
    echo Please install Go 1.21 or later from https://golang.org/
    exit /b 1
)

rem Check Go version
echo Checking Go version...
go version

rem Clean and prepare directories
echo.
echo Preparing build directories...
if exist %DIST_DIR% rmdir /s /q %DIST_DIR%
mkdir %DIST_DIR%
if not exist temp mkdir temp

rem Check for goversioninfo
echo.
echo Checking for goversioninfo...
if not exist "%GOPATH_BIN%\goversioninfo.exe" (
    echo Installing goversioninfo...
    go install github.com/josephspurrier/goversioninfo/cmd/goversioninfo@latest
)

rem Generate Windows resource file
echo.
echo Generating Windows resource file...
set CURRENT_DIR=%cd%
set TEMPLATES_DIR=%~dp0..\common\templates\windows
set TEMP_DIR=%~dp0temp

rem Change to templates directory for goversioninfo to find the icon
cd /d "%TEMPLATES_DIR%"

if exist "%GOPATH_BIN%\goversioninfo.exe" (
    "%GOPATH_BIN%\goversioninfo.exe" -64 -o "%TEMP_DIR%\resource.syso" versioninfo.json
) else (
    goversioninfo -64 -o "%TEMP_DIR%\resource.syso" versioninfo.json
)
set ERROR_LEVEL=%errorlevel%

rem Return to original directory
cd /d "%CURRENT_DIR%"

if %ERROR_LEVEL% neq 0 (
    echo Error: Failed to generate Windows resource file
    echo Make sure goversioninfo is properly installed
    exit /b 1
)

rem Copy resource file to source directory
echo Copying resource file to source directory...
copy "%~dp0temp\resource.syso" "%~dp0%SOURCE_DIR%\" >nul

rem Build Windows executable
echo.
echo Building Windows executable with CGO enabled...
echo This may take a few minutes...

rem Set CGO flags for MinGW
set CGO_ENABLED=1

rem Build the executable
echo Current directory: %cd%
echo Building from: %~dp0%SOURCE_DIR%
echo Output to: %~dp0%DIST_DIR%\%APP_NAME%.exe
cd /d "%~dp0%SOURCE_DIR%"
echo Changed to: %cd%
go build -ldflags "-s -w -H=windowsgui -X main.version=%VERSION%" -o "%~dp0%DIST_DIR%\%APP_NAME%.exe" .
set BUILD_ERROR=%errorlevel%
cd /d "%~dp0"
if %BUILD_ERROR% neq 0 (
    echo Error: Build failed
    echo Make sure you have a C compiler (MinGW-w64 or MSYS2) installed
    exit /b 1
)

rem Clean up resource file
if exist "%~dp0temp\resource.syso" del "%~dp0temp\resource.syso"
if exist "%~dp0%SOURCE_DIR%\resource.syso" del "%~dp0%SOURCE_DIR%\resource.syso"

echo.
echo Build completed successfully!

echo.
echo ========================================
echo   Manual steps required for distribution:
echo ========================================
echo.
echo 1. Copy required DLL files manually:
echo    From: %~dp0*.dll
echo    To:   %~dp0%DIST_DIR%\
echo.
echo 2. Create distribution folder manually:
echo    Folder name: KoeMoji-Go-v%VERSION%-win
echo.
echo 3. Copy files to distribution folder:
echo    - %DIST_DIR%\%APP_NAME%.exe
echo    - Required DLL files (libportaudio.dll, etc.)
echo    - config.example.json (rename to config.json)
echo    - README_RELEASE.md (rename to README.md)
echo.
echo 4. Create ZIP file manually:
echo    ZIP name: KoeMoji-Go-v%VERSION%-win.zip
echo.
echo Executable location:
echo   %DIST_DIR%\%APP_NAME%.exe
echo.

exit /b 0

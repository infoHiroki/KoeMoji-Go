@echo off
setlocal enabledelayedexpansion

rem KoeMoji-Go Windows Build Script
rem Version: Dynamically extracted from version.go

rem バージョン情報をversion.goから動的に取得
for /f "tokens=3 delims==\" " %%i in ('findstr "const Version" ..\..\version.go') do set VERSION=%%i
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
cd %COMMON_DIR%\templates\windows
if exist "%GOPATH_BIN%\goversioninfo.exe" (
    "%GOPATH_BIN%\goversioninfo.exe" -64 -o ..\..\..\windows\temp\resource.syso versioninfo.json
) else (
    goversioninfo -64 -o ..\..\..\windows\temp\resource.syso versioninfo.json
)
if %errorlevel% neq 0 (
    echo Error: Failed to generate Windows resource file
    echo Make sure goversioninfo is properly installed
    cd ..\..\..\windows
    exit /b 1
)
cd ..\..\..\windows

rem Copy resource file to source directory
echo Copying resource file to source directory...
copy temp\resource.syso %SOURCE_DIR%\ >nul

rem Build Windows executable
echo.
echo Building Windows executable with CGO enabled...
echo This may take a few minutes...

rem Set CGO flags for MinGW
set CGO_ENABLED=1

rem Build the executable
cd %SOURCE_DIR%
go build -ldflags="-s -w -H=windowsgui -X main.version=%VERSION%" -o ..\..\build\windows\%DIST_DIR%\%APP_NAME%.exe .
if %errorlevel% neq 0 (
    echo Error: Build failed
    echo Make sure you have a C compiler (MinGW-w64 or MSYS2) installed
    cd ..\..\build\windows
    exit /b 1
)
cd ..\..\build\windows

rem Clean up resource file
if exist temp\resource.syso del temp\resource.syso

echo.
echo Build completed successfully!

rem Copy required DLLs (same directory)
echo Copying required DLL files...
copy /Y *.dll "%DIST_DIR%\" >nul
echo DLL files copied.

rem Create distribution package
echo.
echo Creating distribution package...

cd %DIST_DIR%
mkdir KoeMoji-Go-v%VERSION%
copy %APP_NAME%.exe KoeMoji-Go-v%VERSION%\
copy *.dll KoeMoji-Go-v%VERSION%\
copy %COMMON_DIR%\assets\config.example.json KoeMoji-Go-v%VERSION%\config.json
copy %COMMON_DIR%\assets\README_RELEASE.md KoeMoji-Go-v%VERSION%\README.md

rem Create ZIP package with new naming convention
echo Creating ZIP package...
set RELEASE_NAME=KoeMoji-Go-v%VERSION%-win
powershell -Command "Compress-Archive -Path 'KoeMoji-Go-v%VERSION%' -DestinationPath '%RELEASE_NAME%.zip'"

rem Move ZIP to releases directory
if not exist ..\releases mkdir ..\releases
move %RELEASE_NAME%.zip ..\releases\

rem Clean up temporary directory
rmdir /s /q KoeMoji-Go-v%VERSION%

cd ..

echo.
echo ========================================
echo   Build completed successfully!
echo ========================================
echo.
echo Distribution file created:
echo   ..\releases\%RELEASE_NAME%.zip
echo.
echo Executable location:
echo   %DIST_DIR%\%APP_NAME%.exe
echo.

exit /b 0

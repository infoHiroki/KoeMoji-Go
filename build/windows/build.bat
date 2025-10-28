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
if exist ..\releases\koemoji-go-*.zip del /q ..\releases\koemoji-go-*.zip
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

rem Check for goversioninfo (optional - icon embedding)
echo.
echo Checking for goversioninfo...
if not exist "%GOPATH_BIN%\goversioninfo.exe" (
    echo Installing goversioninfo...
    go install github.com/josephspurrier/goversioninfo/cmd/goversioninfo@latest
)

rem Generate Windows resource file (optional - will continue without icon if fails)
echo.
echo Generating Windows resource file...
set CURRENT_DIR=%cd%
set TEMPLATES_DIR=%~dp0..\common\templates\windows
set TEMP_DIR=%~dp0temp
set RESOURCE_GENERATED=0

rem Change to templates directory for goversioninfo to find the icon
cd /d "%TEMPLATES_DIR%"

if exist "%GOPATH_BIN%\goversioninfo.exe" (
    "%GOPATH_BIN%\goversioninfo.exe" -64 -o "%TEMP_DIR%\resource.syso" versioninfo.json
    if %errorlevel% equ 0 (
        set RESOURCE_GENERATED=1
        echo [OK] Resource file generated successfully
    ) else (
        echo [WARNING] goversioninfo failed - continuing without icon
    )
) else (
    goversioninfo -64 -o "%TEMP_DIR%\resource.syso" versioninfo.json
    if %errorlevel% equ 0 (
        set RESOURCE_GENERATED=1
        echo [OK] Resource file generated successfully
    ) else (
        echo [WARNING] goversioninfo failed - continuing without icon
    )
)

rem Return to original directory
cd /d "%CURRENT_DIR%"

rem Copy resource file to source directory (only if generated successfully)
if %RESOURCE_GENERATED% equ 1 (
    echo Copying resource file to source directory...
    copy "%~dp0temp\resource.syso" "%~dp0%SOURCE_DIR%\" >nul
    echo [OK] Icon will be embedded in executable
) else (
    echo [INFO] Building without embedded icon
)

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
    echo [ERROR] Build failed
    echo Make sure you have a C compiler ^(MinGW-w64 or MSYS2^) available
    exit /b 1
)

rem Clean up resource file
if exist "%~dp0temp\resource.syso" del "%~dp0temp\resource.syso"
if exist "%~dp0%SOURCE_DIR%\resource.syso" del "%~dp0%SOURCE_DIR%\resource.syso"

echo.
echo Build completed successfully!

rem Copy required DLLs (same directory)
echo.
echo Copying required DLL files...
copy /Y *.dll "%DIST_DIR%\" >nul
if %errorlevel% neq 0 (
    echo Warning: Failed to copy DLL files
    echo Make sure DLL files exist in build\windows directory
    exit /b 1
)
echo DLL files copied.

rem Create distribution package
echo.
echo Creating distribution package...

cd /d "%~dp0%DIST_DIR%"
if not exist "koemoji-go-%VERSION%" mkdir "koemoji-go-%VERSION%"
copy "%APP_NAME%.exe" "koemoji-go-%VERSION%\" >nul
copy "*.dll" "koemoji-go-%VERSION%\" >nul
copy "%~dp0..\common\assets\config.example.json" "koemoji-go-%VERSION%\config.json" >nul
copy "%~dp0..\common\assets\README_RELEASE.md" "koemoji-go-%VERSION%\README.md" >nul

rem Create ZIP package with new naming convention
echo Creating ZIP package...
set RELEASE_NAME=koemoji-go-%VERSION%
if exist "%RELEASE_NAME%.zip" del "%RELEASE_NAME%.zip"
powershell -Command "Compress-Archive -Path 'koemoji-go-%VERSION%' -DestinationPath '%RELEASE_NAME%.zip' -Force"
if %errorlevel% neq 0 (
    echo Error: Failed to create ZIP package
    cd /d "%~dp0"
    exit /b 1
)

rem Move ZIP to releases directory
echo Moving ZIP to releases directory...
if not exist "%~dp0..\releases" mkdir "%~dp0..\releases"
move /Y "%RELEASE_NAME%.zip" "%~dp0..\releases\" >nul

rem Clean up temporary distribution directory
echo Cleaning up temporary files...
rmdir /s /q "koemoji-go-%VERSION%"

cd /d "%~dp0"

echo.
echo ========================================
echo   Build completed successfully!
echo ========================================
echo.
echo Distribution file created:
echo   build\releases\%RELEASE_NAME%.zip
echo.
echo Executable location:
echo   build\windows\%DIST_DIR%\%APP_NAME%.exe
echo.
echo Press any key to close...
pause >nul
exit /b 0

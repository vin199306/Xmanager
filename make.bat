@echo off
echo.
echo ========================================
echo     Program Manager Build Tool
echo ========================================
echo.

if "%1"=="" goto :help
if "%1"=="help" goto :help
if "%1"=="go" goto :build_go
if "%1"=="linux" goto :build_linux
if "%1"=="web" goto :build_web
if "%1"=="dev" goto :dev_web
if "%1"=="deps" goto :install_deps
if "%1"=="all" goto :full_build
if "%1"=="clean" goto :clean
goto :help

:help
echo Usage: make.bat [command]
echo.
echo Commands:
echo   go     - Build Go backend for Windows
echo   linux  - Build Go backend for Linux
echo   all    - Full project build
echo   clean  - Clean build files
echo   help   - Show this help
echo.
echo Examples:
echo   make.bat go
echo   make.bat linux
echo   make.bat all
goto :eof

:build_go
echo Building Go backend for Windows...
cd /d "%~dp0"
set GOOS=windows
set GOARCH=amd64
set CGO_ENABLED=0
go build -o program-manager.exe
if %errorlevel% neq 0 (
    echo Go build failed!
    exit /b %errorlevel%
)
echo Go backend built successfully: program-manager.exe
goto :eof

:build_linux
echo Building Go backend for Linux...
cd /d "%~dp0"
set GOOS=linux
set GOARCH=amd64
set CGO_ENABLED=0
go build -ldflags "-s -w" -trimpath -o program-manager-linux .
if %errorlevel% neq 0 (
    echo Go build failed!
    exit /b %errorlevel%
)
echo Go backend built successfully: program-manager-linux
goto :eof




:full_build
echo Starting full project build...
echo Step 1: Building Go backend...
cd /d "%~dp0"
set GOOS=windows
set GOARCH=amd64
set CGO_ENABLED=0
go build -o program-manager.exe
if %errorlevel% neq 0 (
    echo Go build failed!
    exit /b %errorlevel%
)

echo.
echo ========================================
echo     Full build completed successfully!
echo ========================================
goto :eof

:clean
echo Cleaning build files...
if exist "%~dp0program-manager.exe" (
    del /q "%~dp0program-manager.exe"
    echo Removed program-manager.exe
)
echo Clean completed!
goto :eof
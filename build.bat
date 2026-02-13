@echo off
setlocal

set PROJECT=openclipboard
set FLAGS=-ldflags="-s -w" -trimpath -buildvcs=false
set OUTDIR=build

if exist %OUTDIR% rd /s /q %OUTDIR%
mkdir %OUTDIR%

echo ========================================
echo  OpenClipboard Multi-Platform Build
echo ========================================
echo.

:: Windows amd64
echo [1/8] Building windows/amd64...
set GOOS=windows
set GOARCH=amd64
go build %FLAGS% -o %OUTDIR%\%PROJECT%-windows-amd64.exe main.go
if %errorlevel% neq 0 echo FAILED! & goto :next1
echo       OK
:next1

:: Windows arm64
echo [2/8] Building windows/arm64...
set GOOS=windows
set GOARCH=arm64
go build %FLAGS% -o %OUTDIR%\%PROJECT%-windows-arm64.exe main.go
if %errorlevel% neq 0 echo FAILED! & goto :next2
echo       OK
:next2

:: Linux amd64
echo [3/8] Building linux/amd64...
set GOOS=linux
set GOARCH=amd64
go build %FLAGS% -o %OUTDIR%\%PROJECT%-linux-amd64 main.go
if %errorlevel% neq 0 echo FAILED! & goto :next3
echo       OK
:next3

:: Linux arm64
echo [4/8] Building linux/arm64...
set GOOS=linux
set GOARCH=arm64
go build %FLAGS% -o %OUTDIR%\%PROJECT%-linux-arm64 main.go
if %errorlevel% neq 0 echo FAILED! & goto :next4
echo       OK
:next4

:: Linux arm (ARMv7)
echo [5/8] Building linux/arm (v7)...
set GOOS=linux
set GOARCH=arm
set GOARM=7
go build %FLAGS% -o %OUTDIR%\%PROJECT%-linux-armv7 main.go
if %errorlevel% neq 0 echo FAILED! & goto :next5
echo       OK
:next5
set GOARM=

:: macOS amd64
echo [6/8] Building darwin/amd64...
set GOOS=darwin
set GOARCH=amd64
go build %FLAGS% -o %OUTDIR%\%PROJECT%-darwin-amd64 main.go
if %errorlevel% neq 0 echo FAILED! & goto :next6
echo       OK
:next6

:: macOS arm64 (Apple Silicon)
echo [7/8] Building darwin/arm64...
set GOOS=darwin
set GOARCH=arm64
go build %FLAGS% -o %OUTDIR%\%PROJECT%-darwin-arm64 main.go
if %errorlevel% neq 0 echo FAILED! & goto :next7
echo       OK
:next7

:: FreeBSD amd64
echo [8/8] Building freebsd/amd64...
set GOOS=freebsd
set GOARCH=amd64
go build %FLAGS% -o %OUTDIR%\%PROJECT%-freebsd-amd64 main.go
if %errorlevel% neq 0 echo FAILED! & goto :done
echo       OK

:done
echo.
echo ========================================
echo  Build complete! Output: %OUTDIR%\
echo ========================================
dir /b %OUTDIR%

endlocal

@echo off
setlocal

set SERVER=openclipboard
set CLIENT=opencb-client
set FLAGS=-ldflags="-s -w" -trimpath -buildvcs=false
set OUTDIR=build
set CLIENTOUT=public\client

if not exist %OUTDIR%    mkdir %OUTDIR%
if not exist %CLIENTOUT% mkdir %CLIENTOUT%

echo ========================================
echo  Step 1: Build CLI clients
echo ========================================
echo.

pushd client

set CGO_ENABLED=0

echo [1/8] Client windows/amd64...
set GOOS=windows
set GOARCH=amd64
go build %FLAGS% -o ..\%CLIENTOUT%\%CLIENT%-windows-amd64.exe .
if %errorlevel% neq 0 (echo FAILED!) else (echo       OK)

echo [2/8] Client windows/arm64...
set GOOS=windows
set GOARCH=arm64
go build %FLAGS% -o ..\%CLIENTOUT%\%CLIENT%-windows-arm64.exe .
if %errorlevel% neq 0 (echo FAILED!) else (echo       OK)

echo [3/8] Client linux/amd64...
set GOOS=linux
set GOARCH=amd64
go build %FLAGS% -o ..\%CLIENTOUT%\%CLIENT%-linux-amd64 .
if %errorlevel% neq 0 (echo FAILED!) else (echo       OK)

echo [4/8] Client linux/arm64...
set GOOS=linux
set GOARCH=arm64
go build %FLAGS% -o ..\%CLIENTOUT%\%CLIENT%-linux-arm64 .
if %errorlevel% neq 0 (echo FAILED!) else (echo       OK)

echo [5/8] Client linux/armv7...
set GOOS=linux
set GOARCH=arm
set GOARM=7
go build %FLAGS% -o ..\%CLIENTOUT%\%CLIENT%-linux-armv7 .
if %errorlevel% neq 0 (echo FAILED!) else (echo       OK)
set GOARM=

echo [6/8] Client darwin/amd64...
set GOOS=darwin
set GOARCH=amd64
go build %FLAGS% -o ..\%CLIENTOUT%\%CLIENT%-darwin-amd64 .
if %errorlevel% neq 0 (echo FAILED!) else (echo       OK)

echo [7/8] Client darwin/arm64...
set GOOS=darwin
set GOARCH=arm64
go build %FLAGS% -o ..\%CLIENTOUT%\%CLIENT%-darwin-arm64 .
if %errorlevel% neq 0 (echo FAILED!) else (echo       OK)

echo [8/8] Client freebsd/amd64...
set GOOS=freebsd
set GOARCH=amd64
go build %FLAGS% -o ..\%CLIENTOUT%\%CLIENT%-freebsd-amd64 .
if %errorlevel% neq 0 (echo FAILED!) else (echo       OK)

set GOOS=
set GOARCH=
set CGO_ENABLED=

popd

echo.
echo ========================================
echo  Step 2: Build server (embeds public/)
echo ========================================
echo.

echo [1/8] Server windows/amd64...
set GOOS=windows
set GOARCH=amd64
go build %FLAGS% -o %OUTDIR%\%SERVER%-windows-amd64.exe main.go
if %errorlevel% neq 0 (echo FAILED!) else (echo       OK)

echo [2/8] Server windows/arm64...
set GOOS=windows
set GOARCH=arm64
go build %FLAGS% -o %OUTDIR%\%SERVER%-windows-arm64.exe main.go
if %errorlevel% neq 0 (echo FAILED!) else (echo       OK)

echo [3/8] Server linux/amd64...
set GOOS=linux
set GOARCH=amd64
go build %FLAGS% -o %OUTDIR%\%SERVER%-linux-amd64 main.go
if %errorlevel% neq 0 (echo FAILED!) else (echo       OK)

echo [4/8] Server linux/arm64...
set GOOS=linux
set GOARCH=arm64
go build %FLAGS% -o %OUTDIR%\%SERVER%-linux-arm64 main.go
if %errorlevel% neq 0 (echo FAILED!) else (echo       OK)

echo [5/8] Server linux/armv7...
set GOOS=linux
set GOARCH=arm
set GOARM=7
go build %FLAGS% -o %OUTDIR%\%SERVER%-linux-armv7 main.go
if %errorlevel% neq 0 (echo FAILED!) else (echo       OK)
set GOARM=

echo [6/8] Server darwin/amd64...
set GOOS=darwin
set GOARCH=amd64
go build %FLAGS% -o %OUTDIR%\%SERVER%-darwin-amd64 main.go
if %errorlevel% neq 0 (echo FAILED!) else (echo       OK)

echo [7/8] Server darwin/arm64...
set GOOS=darwin
set GOARCH=arm64
go build %FLAGS% -o %OUTDIR%\%SERVER%-darwin-arm64 main.go
if %errorlevel% neq 0 (echo FAILED!) else (echo       OK)

echo [8/8] Server freebsd/amd64...
set GOOS=freebsd
set GOARCH=amd64
go build %FLAGS% -o %OUTDIR%\%SERVER%-freebsd-amd64 main.go
if %errorlevel% neq 0 (echo FAILED!) else (echo       OK)

set GOOS=
set GOARCH=

echo.
echo ========================================
echo  Build complete!
echo   Clients : %CLIENTOUT%\
echo   Servers : %OUTDIR%\
echo ========================================

endlocal

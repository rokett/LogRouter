@echo off
SETLOCAL

set APP=LogRouter
set VERSION=1.0.0
set BINARY-X86=%APP%_%VERSION%_Windows_32bit.exe
set BINARY-X64=%APP%_%VERSION%_Windows_64bit.exe

REM Set build number from git commit hash
for /f %%i in ('git rev-parse HEAD') do set BUILD=%%i

set LDFLAGS=-ldflags "-X main.version=%VERSION% -X main.build=%BUILD% -s -w -extldflags '-static'"

goto build

:build
    set GOOS=windows

    echo "=== Building x86 ==="
    set GOARCH=386

    go build -mod=vendor -o %BINARY-X86% %LDFLAGS%

    echo "=== Building x64 ==="
    set GOARCH=amd64

    go build -mod=vendor -o %BINARY-X64% %LDFLAGS%

    goto :finalise

:finalise
    set GOOS=
    set GOARCH=

    goto :EOF

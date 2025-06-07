@echo off
REM Ensure we are in the root of the Go module
cd /d "%~dp0\.."

REM Define paths
set "PROTO_PATH=pkg/api/proto"
set "OUTPUT_PATH=pkg/api/proto"

REM Define service
set "SERVICE=auth/v1"

echo Generating Go code for %SERVICE%...

protoc ^
    --proto_path=%PROTO_PATH% ^
    --go_out=%OUTPUT_PATH% ^
    --go_opt=paths=source_relative ^
    --go-grpc_out=%OUTPUT_PATH% ^
    --go-grpc_opt=paths=source_relative ^
    %PROTO_PATH%/%SERVICE%/auth_service.proto

echo Generation complete.
pause
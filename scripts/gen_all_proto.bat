@echo off

echo ========================================
echo Generating Auth Protos
echo ========================================

call "%~dp0gen_auth_proto.bat"

echo.
echo ========================================
echo Generating Catalog Protos
echo ========================================

call "%~dp0gen_catalog_proto.bat"

echo.
echo ========================================
echo All proto files generated successfully
echo ========================================

pause
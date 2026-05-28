@echo off
chcp 65001 > nul
echo ========================================
echo   纪念站 - 编译脚本
echo ========================================

set GOPROXY=https://goproxy.cn,direct

echo.
echo [1/2] 编译 Windows 版本 (本地测试)...
go build -o memorial-site.exe .
if %errorlevel% neq 0 goto :error
echo        memorial-site.exe ✓

echo [2/2] 编译 Linux 版本 (服务器部署)...
set CGO_ENABLED=0
set GOOS=linux
set GOARCH=amd64
go build -o memorial-site .
if %errorlevel% neq 0 goto :error
echo        memorial-site ✓

echo.
echo ========================================
echo   编译完成!
echo.
echo   本地测试: memorial-site.exe
echo   服务器:    memorial-site (Linux)
echo ========================================
goto :end

:error
echo.
echo   编译失败，请检查错误信息
pause

:end

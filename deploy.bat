@echo off
chcp 65001 > nul
echo ========================================
echo   纪念站 - 部署到云服务器
echo ========================================

REM ====== 修改这里 ======
set SERVER_IP=47.99.159.141
set SERVER_PATH=/opt/memorial
REM =====================

echo.
echo 服务器: %SERVER_IP%
echo 路径:   %SERVER_PATH%
echo.
echo 将上传以下内容:
echo   - memorial-site (Linux 二进制)
echo   - templates\ (模板文件)
echo   - static\    (CSS/JS)
echo.
echo 注意: data\ 和 uploads\ 不会上传，用户数据安全
echo.
set /p CONFIRM=确认上传? (Y/N):
if /i not "%CONFIRM%"=="Y" goto :cancel

echo.
echo [1/3] 停止服务器上的旧进程...
ssh root@%SERVER_IP% "supervisorctl stop memorial"
echo        ✓

echo [2/3] 上传文件...
scp memorial-site root@%SERVER_IP%:%SERVER_PATH%/
scp -r templates\* root@%SERVER_IP%:%SERVER_PATH%/templates/
scp -r static\* root@%SERVER_IP%:%SERVER_PATH%/static/
echo        ✓

echo [3/3] 启动新版本...
ssh root@%SERVER_IP% "supervisorctl start memorial"
echo        ✓

echo.
echo ========================================
echo   部署完成!
echo   访问: http://%SERVER_IP%
echo ========================================
goto :end

:cancel
echo 已取消

:end
pause

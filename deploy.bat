@echo off
REM ===================================
REM  Deploy to Cloud Server
REM ===================================

set SERVER_IP=47.99.159.141
set SERVER_PATH=/opt/memorial

echo.
echo Server: %SERVER_IP%
echo Path:   %SERVER_PATH%
echo.
echo Uploading:
echo   - memorial-site (Linux binary)
echo   - templates\
echo   - static\
echo.
echo data\ and uploads\ will NOT be touched.
echo.

set /p CONFIRM=Continue? (Y/N):
if /i not "%CONFIRM%"=="Y" goto :cancel

echo.
echo [1/3] Stopping old process...
ssh root@%SERVER_IP% "supervisorctl stop memorial"

echo [2/3] Uploading files...
scp memorial-site root@%SERVER_IP%:%SERVER_PATH%/
scp -r templates\* root@%SERVER_IP%:%SERVER_PATH%/templates/
scp -r static\* root@%SERVER_IP%:%SERVER_PATH%/static/

echo [3/3] Starting new version...
ssh root@%SERVER_IP% "supervisorctl start memorial"

echo.
echo Done! http://%SERVER_IP%
goto :end

:cancel
echo Cancelled.

:end
pause

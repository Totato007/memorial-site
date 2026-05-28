# Deploy to cloud server
$SERVER_IP = "47.99.159.141"
$SERVER_PATH = "/opt/memorial"

Write-Host "========================================"
Write-Host "   Deploy to $SERVER_IP"
Write-Host "========================================"
Write-Host ""
Write-Host "Uploading:"
Write-Host "  - memorial-site (Linux binary)"
Write-Host "  - templates\"
Write-Host "  - static\"
Write-Host ""
Write-Host "data\ and uploads\ will NOT be touched."
Write-Host ""

$confirm = Read-Host "Continue? (Y/N)"
if ($confirm -ne "Y" -and $confirm -ne "y") { Write-Host "Cancelled."; return }

Write-Host "`n[1/3] Stopping server..." -ForegroundColor Yellow
ssh root@$SERVER_IP "supervisorctl stop memorial"

Write-Host "[2/3] Uploading..." -ForegroundColor Yellow
scp memorial-site root@${SERVER_IP}:${SERVER_PATH}/
scp -r templates/* root@${SERVER_IP}:${SERVER_PATH}/templates/
scp -r static/* root@${SERVER_IP}:${SERVER_PATH}/static/

Write-Host "[3/3] Restarting..." -ForegroundColor Yellow
ssh root@$SERVER_IP "supervisorctl start memorial"

Write-Host "`nDone! http://$SERVER_IP" -ForegroundColor Green

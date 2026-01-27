# Docker Sequential Build Script
# Build services one by one to prevent memory exhaustion and Docker crashes
# Usage: .\docker-build-sequential.ps1

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "   SkyeIM Sequential Build Script" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# Enable BuildKit
$env:DOCKER_BUILDKIT = "1"
$env:COMPOSE_DOCKER_CLI_BUILD = "1"

Write-Host "[OK] Docker BuildKit enabled" -ForegroundColor Green
Write-Host ""

# Define build stages
# Stage 1: Infrastructure services (no build needed, just pull images)
$infraServices = @("etcd", "redis", "mysql", "minio")

# Stage 2: RPC services (bottom layer)
$rpcServices = @("user-rpc", "friend-rpc", "group-rpc", "message-rpc")

# Stage 3: API services
$apiServices = @("auth-api", "user-api", "friend-api", "message-api", "group-api", "upload-api")

# Stage 4: Application layer services
$appServices = @("ws-server", "gateway")

# Statistics
$totalServices = $rpcServices.Count + $apiServices.Count + $appServices.Count
$currentStep = 0
$failedServices = @()

Write-Host "========================================" -ForegroundColor Yellow
Write-Host "Build Plan:" -ForegroundColor Yellow
Write-Host "  Stage 1: Infrastructure (pull images only, no build)" -ForegroundColor Gray
Write-Host "  Stage 2: RPC services ($($rpcServices.Count) services)" -ForegroundColor Gray
Write-Host "  Stage 3: API services ($($apiServices.Count) services)" -ForegroundColor Gray
Write-Host "  Stage 4: App layer ($($appServices.Count) services)" -ForegroundColor Gray
Write-Host "  Total: $totalServices services to build" -ForegroundColor Yellow
Write-Host "========================================" -ForegroundColor Yellow
Write-Host ""

# Function to build a single service
function Build-Service {
    param (
        [string]$ServiceName,
        [int]$Current,
        [int]$Total
    )
    
    $script:currentStep++
    Write-Host "[$script:currentStep/$Total] Building: $ServiceName" -ForegroundColor Cyan
    Write-Host "  Start time: $(Get-Date -Format 'HH:mm:ss')" -ForegroundColor Gray
    
    $startTime = Get-Date
    
    # Execute build
    docker-compose build $ServiceName 2>&1 | ForEach-Object {
        Write-Host "  $_" -ForegroundColor DarkGray
    }
    
    $endTime = Get-Date
    $duration = ($endTime - $startTime).TotalSeconds
    
    if ($LASTEXITCODE -eq 0) {
        Write-Host "  [OK] $ServiceName built successfully (Duration: $([math]::Round($duration, 1))s)" -ForegroundColor Green
        return $true
    } else {
        Write-Host "  [ERROR] $ServiceName build failed!" -ForegroundColor Red
        $script:failedServices += $ServiceName
        return $false
    }
}

# Confirm before starting
$confirm = Read-Host "Start sequential build? (y/n, default: y)"
if ($confirm -eq "n" -or $confirm -eq "N") {
    Write-Host "Build cancelled" -ForegroundColor Yellow
    exit 0
}

Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Starting build..." -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# ============================================
# Stage 1: Start infrastructure services
# ============================================
Write-Host ""
Write-Host "=== Stage 1: Starting Infrastructure Services ===" -ForegroundColor Magenta
Write-Host ""

foreach ($service in $infraServices) {
    Write-Host "Starting service: $service" -ForegroundColor Cyan
    docker-compose up -d $service
    if ($LASTEXITCODE -eq 0) {
        Write-Host "  [OK] $service started" -ForegroundColor Green
    } else {
        Write-Host "  [ERROR] $service failed to start!" -ForegroundColor Red
    }
    Start-Sleep -Seconds 2
}

Write-Host ""
Write-Host "Waiting for infrastructure health checks..." -ForegroundColor Yellow
Start-Sleep -Seconds 10

# ============================================
# Stage 2: Build RPC services
# ============================================
Write-Host ""
Write-Host "=== Stage 2: Building RPC Services ===" -ForegroundColor Magenta
Write-Host ""

foreach ($service in $rpcServices) {
    Build-Service -ServiceName $service -Current $currentStep -Total $totalServices
    Write-Host ""
    
    # Wait 2 seconds after each build to release resources
    Start-Sleep -Seconds 2
}

# ============================================
# Stage 3: Build API services
# ============================================
Write-Host ""
Write-Host "=== Stage 3: Building API Services ===" -ForegroundColor Magenta
Write-Host ""

foreach ($service in $apiServices) {
    Build-Service -ServiceName $service -Current $currentStep -Total $totalServices
    Write-Host ""
    
    # Wait 2 seconds after each build to release resources
    Start-Sleep -Seconds 2
}

# ============================================
# Stage 4: Build application layer services
# ============================================
Write-Host ""
Write-Host "=== Stage 4: Building Application Layer ===" -ForegroundColor Magenta
Write-Host ""

foreach ($service in $appServices) {
    Build-Service -ServiceName $service -Current $currentStep -Total $totalServices
    Write-Host ""
    
    # Wait 2 seconds after each build to release resources
    Start-Sleep -Seconds 2
}

# ============================================
# Build complete, generate report
# ============================================
Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "   Build Complete!" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

if ($failedServices.Count -eq 0) {
    Write-Host "[OK] All services built successfully!" -ForegroundColor Green
    Write-Host ""
    Write-Host "Next step: Start all services" -ForegroundColor Yellow
    Write-Host "  Run: docker-compose up -d" -ForegroundColor Gray
} else {
    Write-Host "[WARNING] Some services failed to build:" -ForegroundColor Red
    foreach ($service in $failedServices) {
        Write-Host "  - $service" -ForegroundColor Red
    }
    Write-Host ""
    Write-Host "Please check error logs and rebuild failed services:" -ForegroundColor Yellow
    foreach ($service in $failedServices) {
        Write-Host "  docker-compose build $service" -ForegroundColor Gray
    }
}

Write-Host ""
Write-Host "Useful commands:" -ForegroundColor Gray
Write-Host "  Start all:     docker-compose up -d" -ForegroundColor Gray
Write-Host "  Check status:  docker-compose ps" -ForegroundColor Gray
Write-Host "  View logs:     docker-compose logs -f [service-name]" -ForegroundColor Gray
Write-Host "  Stop all:      docker-compose down" -ForegroundColor Gray
Write-Host ""

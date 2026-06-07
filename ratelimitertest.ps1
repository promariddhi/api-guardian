# test-rate-limit.ps1

Write-Host "Sending first burst of 60 requests..."

for ($i = 1; $i -le 60; $i++) {
    try {
        $response = Invoke-WebRequest `
            -Uri "http://localhost:8090/auth/login" `
            -Method GET `
            -ErrorAction Stop

        Write-Host "Request $i -> $($response.StatusCode)"
    }
    catch {
        $statusCode = $_.Exception.Response.StatusCode.value__
        Write-Host "Request $i -> $statusCode"
    }
}

Write-Host ""
Write-Host "Sleeping for 2 seconds..."
Start-Sleep -Seconds 2

Write-Host ""
Write-Host "Sending another 20 requests..."

for ($i = 1; $i -le 20; $i++) {
    try {
        $response = Invoke-WebRequest `
            -Uri "http://localhost:8090/auth/login" `
            -Method GET `
            -ErrorAction Stop

        Write-Host "Post-sleep Request $i -> $($response.StatusCode)"
    }
    catch {
        $statusCode = $_.Exception.Response.StatusCode.value__
        Write-Host "Post-sleep Request $i -> $statusCode"
    }
}
$tempDir = Join-Path ([IO.Path]::GetTempPath()) "LimonRunTest"
if (-not (Test-Path $tempDir)) { New-Item -ItemType Directory -Path $tempDir | Out-Null }
Copy-Item ".\LimonWhiskas.exe" $tempDir -Force

$proc = Start-Process -FilePath (Join-Path $tempDir "LimonWhiskas.exe") -WorkingDirectory $tempDir -PassThru
Start-Sleep -Seconds 3
if ($proc.HasExited) {
    Write-Host "Process exited with code: $($proc.ExitCode)"
} else {
    Write-Host "Process is running successfully! PID: $($proc.Id)"
    Stop-Process -Id $proc.Id -Force
}

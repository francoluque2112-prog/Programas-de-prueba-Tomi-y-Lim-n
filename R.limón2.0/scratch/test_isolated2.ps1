$tempTestDir = [System.IO.Path]::Combine([System.IO.Path]::GetTempPath(), "LimonIsolatedTest2")
if (-not (Test-Path $tempTestDir)) { New-Item -ItemType Directory -Path $tempTestDir | Out-Null }
Copy-Item ".\LimonWhiskas.exe" (Join-Path $tempTestDir "LimonWhiskas.exe") -Force

[System.IO.Directory]::SetCurrentDirectory($tempTestDir)

$asm = [System.Reflection.Assembly]::LoadFile((Join-Path $tempTestDir "LimonWhiskas.exe"))
$t = $asm.GetType("LimonWhiskasGame.EmbeddedResourceManager")
Write-Host "Calling EmbeddedResourceManager.ExtractAll() with CurrentDirectory = $tempTestDir..."
$mExtract = $t.GetMethod("ExtractAll", [System.Reflection.BindingFlags]"Public,Static")
$mExtract.Invoke($null, @())

$mPath = $t.GetMethod("GetExtractedFilePath", [System.Reflection.BindingFlags]"Public,Static")
$mImg = $t.GetMethod("GetImage", [System.Reflection.BindingFlags]"Public,Static")

foreach ($name in @("cazador.jpg", "glitch.jpg", "error.jpg", "Limon.png", "LimonF.png", "Whiskas.ico", "Whiskas.png", "CD-Limon.ico", "Whua.mp3", "F1.mp3", "F2.mp3", "Limon.A.mp3")) {
    $p = $mPath.Invoke($null, @($name))
    $img = $mImg.Invoke($null, @($name))
    $imgInfo = if ($img -ne $null) { "Img: $($img.Width)x$($img.Height)" } else { "Img: null" }
    Write-Host "$name -> Path: $p | $imgInfo"
}

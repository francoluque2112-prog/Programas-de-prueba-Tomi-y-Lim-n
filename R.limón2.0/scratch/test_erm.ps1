$asm = [System.Reflection.Assembly]::LoadFile((Resolve-Path '.\LimonWhiskas.exe'))
$t = $asm.GetType("LimonWhiskasGame.EmbeddedResourceManager")
if ($t -ne $null) {
    Write-Host "Calling EmbeddedResourceManager.ExtractAll()..."
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
} else {
    Write-Host "Type LimonWhiskasGame.EmbeddedResourceManager NOT FOUND!"
}

$asm = [System.Reflection.Assembly]::LoadFile((Resolve-Path '.\LimonWhiskas.exe'))
$names = $asm.GetManifestResourceNames()
Write-Host "Total resources in EXE: $($names.Count)"
foreach ($n in $names) {
    $s = $asm.GetManifestResourceStream($n)
    if ($s -ne $null) {
        Write-Host "OK: $n length=$($s.Length)"
        $s.Close()
    } else {
        Write-Host "FAIL: $n"
    }
}

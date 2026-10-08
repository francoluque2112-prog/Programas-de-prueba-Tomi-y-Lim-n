Add-Type -TypeDefinition @"
using System;
using System.Runtime.InteropServices;
using System.Text;
public class MciTest2 {
    [DllImport("winmm.dll", CharSet = CharSet.Auto)]
    public static extern long mciSendString(string cmd, StringBuilder ret, int len, IntPtr cb);
    public static long Exec(string cmd) {
        StringBuilder sb = new StringBuilder(260);
        return mciSendString(cmd, sb, sb.Capacity, IntPtr.Zero);
    }
}
"@

$tempDir = [System.IO.Path]::Combine([System.IO.Path]::GetTempPath(), "LimonWhiskasResources")
$p = Join-Path $tempDir "Whua.mp3"

Write-Host "Test 1 (quoted with type mpegvideo):" ([MciTest2]::Exec("open `"$p`" type mpegvideo alias s1"))
Write-Host "Play 1:" ([MciTest2]::Exec("play s1 from 0"))
[MciTest2]::Exec("close s1")

Write-Host "Test 2 (quoted without type):" ([MciTest2]::Exec("open `"$p`" alias s2"))
Write-Host "Play 2:" ([MciTest2]::Exec("play s2 from 0"))
[MciTest2]::Exec("close s2")

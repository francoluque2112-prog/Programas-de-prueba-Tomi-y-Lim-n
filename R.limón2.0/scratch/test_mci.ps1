$code = @"
using System;
using System.IO;
using System.Runtime.InteropServices;
using System.Text;

public class MciTest
{
    [DllImport("winmm.dll", CharSet = CharSet.Auto)]
    public static extern long mciSendString(string strCommand, StringBuilder strReturn, int iReturnLength, IntPtr hwndCallback);

    [DllImport("winmm.dll", CharSet = CharSet.Auto)]
    public static extern int mciGetErrorString(long fdwError, StringBuilder lpszErrorText, int cchErrorText);

    public static string MciCommand(string cmd)
    {
        StringBuilder ret = new StringBuilder(260);
        long err = mciSendString(cmd, ret, ret.Capacity, IntPtr.Zero);
        if (err != 0)
        {
            StringBuilder errStr = new StringBuilder(260);
            mciGetErrorString(err, errStr, errStr.Capacity);
            return "ERROR " + err + ": " + errStr.ToString();
        }
        return "OK (" + ret.ToString() + ")";
    }
}
"@

Add-Type -TypeDefinition $code -Language CSharp

$asm = [System.Reflection.Assembly]::LoadFile((Resolve-Path '.\LimonWhiskas.exe'))
$tempDir = [System.IO.Path]::Combine([System.IO.Path]::GetTempPath(), "LimonWhiskasResources")
if (-not (Test-Path $tempDir)) { New-Item -ItemType Directory -Path $tempDir | Out-Null }

foreach ($r in @("Whua.mp3", "F1.mp3", "F2.mp3", "Limon.A.mp3")) {
    $s = $asm.GetManifestResourceStream($r)
    $targetPath = [System.IO.Path]::Combine($tempDir, $r)
    $fs = [System.IO.File]::Create($targetPath)
    $s.CopyTo($fs)
    $fs.Close()
    $s.Close()
    Write-Host "Extracted $r -> $targetPath (bytes: $( (Get-Item $targetPath).Length ))"

    $cmdOpen = "open `"$targetPath`" type mpegvideo alias test_$r"
    $resOpen = [MciTest]::MciCommand($cmdOpen)
    Write-Host "MCI OPEN: $resOpen"
    
    $cmdPlay = "play test_$r from 0"
    $resPlay = [MciTest]::MciCommand($cmdPlay)
    Write-Host "MCI PLAY: $resPlay"

    [MciTest]::MciCommand("close test_$r")
}

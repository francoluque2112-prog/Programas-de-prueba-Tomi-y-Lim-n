@echo off
echo Compilando LimonWhiskas.exe autonomo e independiente...
C:\Windows\Microsoft.NET\Framework64\v4.0.30319\csc.exe /target:winexe /win32icon:"assets\CD-Limon.ico" /out:LimonWhiskas.exe /resource:"assets\cazador.jpg",cazador.jpg /resource:"assets\glitch.jpg",glitch.jpg /resource:"assets\error.jpg",error.jpg /resource:"assets\Limon.png",Limon.png /resource:"assets\LimonF.png",LimonF.png /resource:"assets\Whiskas.ico",Whiskas.ico /resource:"assets\Whiskas.png",Whiskas.png /resource:"assets\CD-Limon.ico",CD-Limon.ico /resource:"Audios\Whua.mp3",Whua.mp3 /resource:"Audios\F1.mp3",F1.mp3 /resource:"Audios\F2.mp3",F2.mp3 /resource:"Audios\Limon.A.mp3",Limon.A.mp3 "src\Program.cs" /r:Microsoft.CSharp.dll /r:System.Core.dll /r:System.Windows.Forms.dll /r:System.Drawing.dll /r:System.dll
if %errorlevel% equ 0 (
    echo [OK] LimonWhiskas.exe compilado exitosamente (100%% Autonomo).
) else (
    echo [ERROR] Error al compilar.
)

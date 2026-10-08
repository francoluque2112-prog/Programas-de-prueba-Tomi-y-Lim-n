package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"unsafe"
)

var (
	procMciSendStringW = winmmDLL.NewProc("mciSendStringW")
)

func mciCommand(cmd string) error {
	pCmd, err := syscall.UTF16PtrFromString(cmd)
	if err != nil {
		return err
	}
	r, _, _ := procMciSendStringW.Call(
		uintptr(unsafe.Pointer(pCmd)),
		0,
		0,
		0,
	)
	if r != 0 {
		return fmt.Errorf("MCI error %d for cmd: %s", r, cmd)
	}
	return nil
}

func PlayIntroMusic() {
	cwd, _ := os.Getwd()
	introPath := filepath.Join(cwd, "Audios", "Intro.mp3")
	if _, err := os.Stat(introPath); os.IsNotExist(err) {
		introPath = filepath.Join(cwd, "Audios", "intro.mp3")
	}
	_ = mciCommand("close intro")
	_ = mciCommand(fmt.Sprintf(`open "%s" type mpegvideo alias intro`, introPath))
	_ = mciCommand("play intro")
}

func GetIntroDuration() float64 {
	buf := make([]uint16, 128)
	pCmd, _ := syscall.UTF16PtrFromString("status intro length")
	r, _, _ := procMciSendStringW.Call(
		uintptr(unsafe.Pointer(pCmd)),
		uintptr(unsafe.Pointer(&buf[0])),
		128,
		0,
	)
	if r == 0 {
		str := syscall.UTF16ToString(buf)
		var ms int
		fmt.Sscanf(str, "%d", &ms)
		if ms > 0 {
			return float64(ms) / 1000.0
		}
	}
	return 5.0
}

func StopIntroMusic() {
	_ = mciCommand("stop intro")
	_ = mciCommand("close intro")
}

func PlayChaseMusic() {
	cwd, _ := os.Getwd()
	chasePath := filepath.Join(cwd, "Audios", "Chase.mp3")
	if _, err := os.Stat(chasePath); os.IsNotExist(err) {
		chasePath = filepath.Join(cwd, "Audios", "chase.mp3")
	}
	_ = mciCommand("close chase")
	_ = mciCommand(fmt.Sprintf(`open "%s" type mpegvideo alias chase`, chasePath))
	_ = mciCommand("play chase repeat")
}

func StopChaseMusic() {
	_ = mciCommand("stop chase")
	_ = mciCommand("close chase")
}

func StopAllMusic() {
	StopIntroMusic()
	StopChaseMusic()
}

// PlayTVVideo plays Whua.mp4 inside a borderless WPF window positioned at the TV screen
func PlayTVVideo(x, y, w, h int, onFinished func()) *exec.Cmd {
	cwd, _ := os.Getwd()
	videoPath := filepath.Join(cwd, "Videos", "Whua.mp4")
	scriptPath := filepath.Join(cwd, "src", "play_video.ps1")

	cmd := exec.Command(
		"powershell.exe",
		"-WindowStyle", "Hidden",
		"-ExecutionPolicy", "Bypass",
		"-File", scriptPath,
		"-x", fmt.Sprintf("%d", x),
		"-y", fmt.Sprintf("%d", y),
		"-w", fmt.Sprintf("%d", w),
		"-h", fmt.Sprintf("%d", h),
		"-videoPath", videoPath,
	)

	go func() {
		_ = cmd.Run()
		if onFinished != nil {
			onFinished()
		}
	}()

	return cmd
}

package main

import (
	"fmt"
	"runtime"
	"syscall"
	"time"
	"unsafe"
)

var (
	user32   = syscall.NewLazyDLL("user32.dll")
	gdi32    = syscall.NewLazyDLL("gdi32.dll")
	kernel32 = syscall.NewLazyDLL("kernel32.dll")

	procRegisterClassExW     = user32.NewProc("RegisterClassExW")
	procCreateWindowExW      = user32.NewProc("CreateWindowExW")
	procDefWindowProcW       = user32.NewProc("DefWindowProcW")
	procShowWindow           = user32.NewProc("ShowWindow")
	procUpdateWindow         = user32.NewProc("UpdateWindow")
	procGetMessageW          = user32.NewProc("GetMessageW")
	procPeekMessageW         = user32.NewProc("PeekMessageW")
	procTranslateMessage     = user32.NewProc("TranslateMessage")
	procDispatchMessageW     = user32.NewProc("DispatchMessageW")
	procPostQuitMessage      = user32.NewProc("PostQuitMessage")
	procGetDC                = user32.NewProc("GetDC")
	procReleaseDC            = user32.NewProc("ReleaseDC")
	procGetSystemMetrics     = user32.NewProc("GetSystemMetrics")
	procUpdateLayeredWindow  = user32.NewProc("UpdateLayeredWindow")
	procGetCursorPos         = user32.NewProc("GetCursorPos")
	procGetAsyncKeyState      = user32.NewProc("GetAsyncKeyState")
	procDestroyWindow        = user32.NewProc("DestroyWindow")
	procSetWindowPos         = user32.NewProc("SetWindowPos")

	procCreateCompatibleDC   = gdi32.NewProc("CreateCompatibleDC")
	procDeleteDC             = gdi32.NewProc("DeleteDC")
	procCreateDIBSection     = gdi32.NewProc("CreateDIBSection")
	procDeleteObject         = gdi32.NewProc("DeleteObject")
	procSelectObject         = gdi32.NewProc("SelectObject")

	procGetModuleHandleW     = kernel32.NewProc("GetModuleHandleW")

	shell32             = syscall.NewLazyDLL("shell32.dll")
	procDragAcceptFiles = shell32.NewProc("DragAcceptFiles")
	procDragQueryFileW  = shell32.NewProc("DragQueryFileW")
	procDragQueryPoint  = shell32.NewProc("DragQueryPoint")
	procDragFinish      = shell32.NewProc("DragFinish")
)

const (
	WS_POPUP          = 0x80000000
	WS_VISIBLE        = 0x10000000
	WS_EX_TOPMOST     = 0x00000008
	WS_EX_LAYERED     = 0x00080000
	WS_EX_TOOLWINDOW  = 0x00000080
	WS_EX_ACCEPTFILES = 0x00000010

	SM_CXSCREEN = 0
	SM_CYSCREEN = 1

	ULW_ALPHA = 0x00000002
	AC_SRC_OVER = 0x00
	AC_SRC_ALPHA = 0x01

	WM_DESTROY     = 0x0002
	WM_KEYDOWN     = 0x0100
	WM_KEYUP       = 0x0101
	WM_LBUTTONDOWN = 0x0201
	WM_LBUTTONUP   = 0x0202
	WM_MOUSEMOVE   = 0x0200
	WM_DROPFILES   = 0x0233

	VK_ESCAPE = 0x1B
	VK_R      = 0x52
	VK_O      = 0x4F
	VK_LBUTTON = 0x01

	PM_REMOVE = 0x0001

	DIB_RGB_COLORS = 0
	BI_RGB         = 0
)

type WNDCLASSEXW struct {
	CbSize        uint32
	Style         uint32
	LpfnWndProc   uintptr
	CbClsExtra    int32
	CbWndExtra    int32
	HInstance     uintptr
	HIcon         uintptr
	HCursor       uintptr
	HbrBackground uintptr
	LpszMenuName  *uint16
	LpszClassName *uint16
	HIconSm       uintptr
}

type POINT struct {
	X, Y int32
}

type SIZE struct {
	CX, CY int32
}

type BLENDFUNCTION struct {
	BlendOp             byte
	BlendFlags          byte
	SourceConstantAlpha byte
	AlphaFormat         byte
}

type BITMAPINFOHEADER struct {
	BiSize          uint32
	BiWidth         int32
	BiHeight        int32
	BiPlanes        uint16
	BiBitCount      uint16
	BiCompression   uint32
	BiSizeImage     uint32
	BiXPelsPerMeter int32
	BiYPelsPerMeter int32
	BiClrUsed       uint32
	BiClrImportant  uint32
}

type MSG struct {
	Hwnd    uintptr
	Message uint32
	WParam  uintptr
	LParam  uintptr
	Time    uint32
	Pt      POINT
}

var (
	globalHwnd uintptr
	globalGame *Game
)

func wndProc(hwnd uintptr, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case WM_DESTROY:
		procPostQuitMessage.Call(0)
		return 0
	case WM_DROPFILES:
		hDrop := wParam
		var pt POINT
		procDragQueryPoint.Call(hDrop, uintptr(unsafe.Pointer(&pt)))

		countVal, _, _ := procDragQueryFileW.Call(hDrop, 0xFFFFFFFF, 0, 0)
		count := int(countVal)

		for i := 0; i < count; i++ {
			buf := make([]uint16, 512)
			procDragQueryFileW.Call(hDrop, uintptr(i), uintptr(unsafe.Pointer(&buf[0])), 512)
			filePath := syscall.UTF16ToString(buf)
			if globalGame != nil {
				globalGame.OnFileDrop(filePath, int(pt.X), int(pt.Y))
			}
		}
		procDragFinish.Call(hDrop)
		return 0
	}
	r, _, _ := procDefWindowProcW.Call(hwnd, uintptr(msg), wParam, lParam)
	return r
}

func RunTransparentGame(game *Game) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	hInst, _, _ := procGetModuleHandleW.Call(0)
	className, _ := syscall.UTF16PtrFromString("TomiGameWindowClass")
	windowName, _ := syscall.UTF16PtrFromString("Tomi el Perro")

	var wc WNDCLASSEXW
	wc.CbSize = uint32(unsafe.Sizeof(wc))
	wc.Style = 0
	wc.LpfnWndProc = syscall.NewCallback(wndProc)
	wc.HInstance = hInst
	wc.LpszClassName = className

	procRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc)))

	screenW := game.ScreenW
	screenH := game.ScreenH

	exStyle := WS_EX_LAYERED | WS_EX_TOPMOST | WS_EX_TOOLWINDOW
	style := WS_POPUP | WS_VISIBLE

	hwnd, _, err := procCreateWindowExW.Call(
		uintptr(exStyle),
		uintptr(unsafe.Pointer(className)),
		uintptr(unsafe.Pointer(windowName)),
		uintptr(style),
		0, 0,
		uintptr(screenW),
		uintptr(screenH),
		0, 0,
		hInst,
		0,
	)

	if hwnd == 0 {
		return fmt.Errorf("CreateWindowEx failed: %v", err)
	}
	globalHwnd = hwnd
	globalGame = game

	// Enable Drag and Drop for files from Desktop/Explorer
	procDragAcceptFiles.Call(hwnd, 1)

	screenDC, _, _ := procGetDC.Call(0)
	memDC, _, _ := procCreateCompatibleDC.Call(screenDC)

	var bmi BITMAPINFOHEADER
	bmi.BiSize = uint32(unsafe.Sizeof(bmi))
	bmi.BiWidth = int32(screenW)
	bmi.BiHeight = -int32(screenH) // Top-down DIB
	bmi.BiPlanes = 1
	bmi.BiBitCount = 32
	bmi.BiCompression = BI_RGB

	var bitsPtr unsafe.Pointer
	hBitmap, _, _ := procCreateDIBSection.Call(
		memDC,
		uintptr(unsafe.Pointer(&bmi)),
		DIB_RGB_COLORS,
		uintptr(unsafe.Pointer(&bitsPtr)),
		0,
		0,
	)

	procSelectObject.Call(memDC, hBitmap)

	fb := NewFramebuffer(screenW, screenH)

	procShowWindow.Call(hwnd, 5) // SW_SHOW
	procUpdateWindow.Call(hwnd)

	ticker := time.NewTicker(16 * time.Millisecond) // ~60 FPS
	defer ticker.Stop()

	var msg MSG
	lastTime := time.Now()

	for {
		// Process Windows messages
		for {
			hasMsg, _, _ := procPeekMessageW.Call(
				uintptr(unsafe.Pointer(&msg)),
				0,
				0,
				0,
				PM_REMOVE,
			)
			if hasMsg == 0 {
				break
			}
			if msg.Message == WM_DESTROY || (msg.Message == 0x0012) { // WM_QUIT
				return nil
			}
			procTranslateMessage.Call(uintptr(unsafe.Pointer(&msg)))
			procDispatchMessageW.Call(uintptr(unsafe.Pointer(&msg)))
		}

		<-ticker.C
		now := time.Now()
		dt := now.Sub(lastTime).Seconds()
		if dt > 0.05 {
			dt = 0.05
		}
		lastTime = now

		// Keyboard controls: ESC to quit, R to restart, O to open Desktop folder
		escState, _, _ := procGetAsyncKeyState.Call(VK_ESCAPE)
		if (escState & 0x8000) != 0 {
			break
		}

		rState, _, _ := procGetAsyncKeyState.Call(VK_R)
		if (rState & 0x0001) != 0 {
			game.Reset()
		}

		oState, _, _ := procGetAsyncKeyState.Call(VK_O)
		if (oState & 0x0001) != 0 {
			OpenDesktopFolder()
		}

		// Mouse cursor position & left button
		var pt POINT
		procGetCursorPos.Call(uintptr(unsafe.Pointer(&pt)))
		lBtnState, _, _ := procGetAsyncKeyState.Call(VK_LBUTTON)
		isMouseDown := (lBtnState & 0x8000) != 0

		// Update game logic
		game.Update(dt, int(pt.X), int(pt.Y), isMouseDown)
		if game.ShouldQuit {
			break
		}

		// Render frame
		game.Render(fb)

		// Copy framebuffer pixels into DIB Section memory
		copyMemory(bitsPtr, unsafe.Pointer(&fb.Pixels[0]), len(fb.Pixels))

		// Update Layered Window on Desktop
		ptSrc := POINT{0, 0}
		ptDst := POINT{0, 0}
		sizeWnd := SIZE{int32(screenW), int32(screenH)}
		blend := BLENDFUNCTION{
			BlendOp:             AC_SRC_OVER,
			BlendFlags:          0,
			SourceConstantAlpha: 255,
			AlphaFormat:         AC_SRC_ALPHA,
		}

		procUpdateLayeredWindow.Call(
			hwnd,
			screenDC,
			uintptr(unsafe.Pointer(&ptDst)),
			uintptr(unsafe.Pointer(&sizeWnd)),
			memDC,
			uintptr(unsafe.Pointer(&ptSrc)),
			0,
			uintptr(unsafe.Pointer(&blend)),
			ULW_ALPHA,
		)
	}

	// Cleanup
	procDeleteObject.Call(hBitmap)
	procDeleteDC.Call(memDC)
	procReleaseDC.Call(0, screenDC)
	procDestroyWindow.Call(hwnd)

	return nil
}

func copyMemory(dst, src unsafe.Pointer, length int) {
	dstSlice := unsafe.Slice((*byte)(dst), length)
	srcSlice := unsafe.Slice((*byte)(src), length)
	copy(dstSlice, srcSlice)
}

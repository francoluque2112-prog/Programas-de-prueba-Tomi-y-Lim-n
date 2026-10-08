package main

import (
	"image"
	"image/color"
	_ "image/png"
	"math"
	"os"
)

// Sprite represents a 32-bit BGRA premultiplied image ready for fast blending
type Sprite struct {
	W, H   int
	Pixels []byte // len = W * H * 4 (B, G, R, A)
}

// Framebuffer represents the screen drawing canvas (BGRA premultiplied)
type Framebuffer struct {
	W, H   int
	Pitch  int
	Pixels []byte
}

func NewFramebuffer(w, h int) *Framebuffer {
	return &Framebuffer{
		W:      w,
		H:      h,
		Pitch:  w * 4,
		Pixels: make([]byte, w*h*4),
	}
}

func (fb *Framebuffer) Clear() {
	for i := range fb.Pixels {
		fb.Pixels[i] = 0
	}
}

// LoadSprite loads a PNG file, automatically crops transparent borders, and premultiplies alpha
func LoadSprite(path string) (*Sprite, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		return nil, err
	}

	bounds := img.Bounds()
	minX, minY, maxX, maxY := bounds.Dx(), bounds.Dy(), 0, 0
	hasPixel := false

	for y := 0; y < bounds.Dy(); y++ {
		for x := 0; x < bounds.Dx(); x++ {
			_, _, _, a := img.At(bounds.Min.X+x, bounds.Min.Y+y).RGBA()
			if a > 2000 { // threshold for non-transparent
				hasPixel = true
				if x < minX {
					minX = x
				}
				if x > maxX {
					maxX = x
				}
				if y < minY {
					minY = y
				}
				if y > maxY {
					maxY = y
				}
			}
		}
	}

	if !hasPixel {
		minX, minY, maxX, maxY = 0, 0, bounds.Dx()-1, bounds.Dy()-1
	}

	cropW := maxX - minX + 1
	cropH := maxY - minY + 1
	sprite := &Sprite{
		W:      cropW,
		H:      cropH,
		Pixels: make([]byte, cropW*cropH*4),
	}

	for y := 0; y < cropH; y++ {
		for x := 0; x < cropW; x++ {
			c := img.At(bounds.Min.X+minX+x, bounds.Min.Y+minY+y)
			r, g, b, a := c.RGBA() // 0..65535
			a8 := byte(a >> 8)
			r8 := byte(r >> 8)
			g8 := byte(g >> 8)
			b8 := byte(b >> 8)

			// Premultiply
			alphaF := float64(a8) / 255.0
			pb := byte(float64(b8) * alphaF)
			pg := byte(float64(g8) * alphaF)
			pr := byte(float64(r8) * alphaF)

			idx := (y*cropW + x) * 4
			sprite.Pixels[idx+0] = pb
			sprite.Pixels[idx+1] = pg
			sprite.Pixels[idx+2] = pr
			sprite.Pixels[idx+3] = a8
		}
	}

	return sprite, nil
}

// Scale resizes a sprite to target width and height using bilinear interpolation
func (s *Sprite) Scale(targetW, targetH int) *Sprite {
	if s == nil || targetW <= 0 || targetH <= 0 {
		return nil
	}
	out := &Sprite{
		W:      targetW,
		H:      targetH,
		Pixels: make([]byte, targetW*targetH*4),
	}

	xRatio := float64(s.W-1) / float64(targetW)
	yRatio := float64(s.H-1) / float64(targetH)

	for y := 0; y < targetH; y++ {
		srcY := float64(y) * yRatio
		yL := int(srcY)
		yH := yL + 1
		if yH >= s.H {
			yH = s.H - 1
		}
		yWeight := srcY - float64(yL)

		for x := 0; x < targetW; x++ {
			srcX := float64(x) * xRatio
			xL := int(srcX)
			xH := xL + 1
			if xH >= s.W {
				xH = s.W - 1
			}
			xWeight := srcX - float64(xL)

			idx00 := (yL*s.W + xL) * 4
			idx10 := (yL*s.W + xH) * 4
			idx01 := (yH*s.W + xL) * 4
			idx11 := (yH*s.W + xH) * 4

			outIdx := (y*targetW + x) * 4
			for c := 0; c < 4; c++ {
				top := float64(s.Pixels[idx00+c])*(1.0-xWeight) + float64(s.Pixels[idx10+c])*xWeight
				bottom := float64(s.Pixels[idx01+c])*(1.0-xWeight) + float64(s.Pixels[idx11+c])*xWeight
				val := top*(1.0-yWeight) + bottom*yWeight
				if val > 255.0 {
					val = 255.0
				}
				out.Pixels[outIdx+c] = byte(val)
			}
		}
	}

	return out
}

// FlipH creates a horizontally flipped copy of the sprite
func (s *Sprite) FlipH() *Sprite {
	if s == nil {
		return nil
	}
	out := &Sprite{
		W:      s.W,
		H:      s.H,
		Pixels: make([]byte, len(s.Pixels)),
	}
	for y := 0; y < s.H; y++ {
		for x := 0; x < s.W; x++ {
			srcIdx := (y*s.W + x) * 4
			dstIdx := (y*s.W + (s.W - 1 - x)) * 4
			out.Pixels[dstIdx+0] = s.Pixels[srcIdx+0]
			out.Pixels[dstIdx+1] = s.Pixels[srcIdx+1]
			out.Pixels[dstIdx+2] = s.Pixels[srcIdx+2]
			out.Pixels[dstIdx+3] = s.Pixels[srcIdx+3]
		}
	}
	return out
}

// DrawSprite draws a sprite onto the framebuffer at (dstX, dstY) with optional alpha multiplier (0..1)
func (fb *Framebuffer) DrawSprite(s *Sprite, dstX, dstY int, alphaMult float64) {
	if s == nil || alphaMult <= 0.001 {
		return
	}

	minX := dstX
	if minX < 0 {
		minX = 0
	}
	maxX := dstX + s.W
	if maxX > fb.W {
		maxX = fb.W
	}

	minY := dstY
	if minY < 0 {
		minY = 0
	}
	maxY := dstY + s.H
	if maxY > fb.H {
		maxY = fb.H
	}

	if minX >= maxX || minY >= maxY {
		return
	}

	for y := minY; y < maxY; y++ {
		srcY := y - dstY
		fbRow := y * fb.Pitch
		sRow := srcY * s.W * 4

		for x := minX; x < maxX; x++ {
			srcX := x - dstX
			sIdx := sRow + srcX*4
			fbIdx := fbRow + x*4

			sa := float64(s.Pixels[sIdx+3]) / 255.0 * alphaMult
			if sa <= 0.001 {
				continue
			}

			sb := float64(s.Pixels[sIdx+0]) * alphaMult
			sg := float64(s.Pixels[sIdx+1]) * alphaMult
			sr := float64(s.Pixels[sIdx+2]) * alphaMult

			invA := 1.0 - sa

			db := float64(fb.Pixels[fbIdx+0])
			dg := float64(fb.Pixels[fbIdx+1])
			dr := float64(fb.Pixels[fbIdx+2])
			da := float64(fb.Pixels[fbIdx+3])

			fb.Pixels[fbIdx+0] = byte(sb + db*invA)
			fb.Pixels[fbIdx+1] = byte(sg + dg*invA)
			fb.Pixels[fbIdx+2] = byte(sr + dr*invA)
			fb.Pixels[fbIdx+3] = byte(sa*255.0 + da*invA)
		}
	}
}

// DrawFilledRect draws a filled rectangle with RGBA color
func (fb *Framebuffer) DrawFilledRect(rx, ry, rw, rh int, c color.RGBA) {
	if c.A == 0 || rw <= 0 || rh <= 0 {
		return
	}

	minX := rx
	if minX < 0 {
		minX = 0
	}
	maxX := rx + rw
	if maxX > fb.W {
		maxX = fb.W
	}

	minY := ry
	if minY < 0 {
		minY = 0
	}
	maxY := ry + rh
	if maxY > fb.H {
		maxY = fb.H
	}

	if minX >= maxX || minY >= maxY {
		return
	}

	sa := float64(c.A) / 255.0
	sr := float64(c.R) * sa
	sg := float64(c.G) * sa
	sb := float64(c.B) * sa
	invA := 1.0 - sa

	for y := minY; y < maxY; y++ {
		fbRow := y * fb.Pitch
		for x := minX; x < maxX; x++ {
			fbIdx := fbRow + x*4

			db := float64(fb.Pixels[fbIdx+0])
			dg := float64(fb.Pixels[fbIdx+1])
			dr := float64(fb.Pixels[fbIdx+2])
			da := float64(fb.Pixels[fbIdx+3])

			fb.Pixels[fbIdx+0] = byte(sb + db*invA)
			fb.Pixels[fbIdx+1] = byte(sg + dg*invA)
			fb.Pixels[fbIdx+2] = byte(sr + dr*invA)
			fb.Pixels[fbIdx+3] = byte(sa*255.0 + da*invA)
		}
	}
}

// DrawRoundedBox draws a styled rounded card with border and shadow
func (fb *Framebuffer) DrawRoundedBox(x, y, w, h, radius int, bg color.RGBA, border color.RGBA, borderWidth int) {
	// Simple shadow
	shadowColor := color.RGBA{0, 0, 0, 80}
	fb.DrawFilledRect(x+4, y+4, w, h, shadowColor)

	// Background
	fb.DrawFilledRect(x, y, w, h, bg)

	// Border
	if borderWidth > 0 && border.A > 0 {
		fb.DrawFilledRect(x, y, w, borderWidth, border)
		fb.DrawFilledRect(x, y+h-borderWidth, w, borderWidth, border)
		fb.DrawFilledRect(x, y, borderWidth, h, border)
		fb.DrawFilledRect(x+w-borderWidth, y, borderWidth, h, border)
	}
}

// DrawCircle draws an anti-aliased filled circle
func (fb *Framebuffer) DrawCircle(cx, cy, r int, c color.RGBA) {
	if r <= 0 || c.A == 0 {
		return
	}

	minX := cx - r - 1
	if minX < 0 { minX = 0 }
	maxX := cx + r + 1
	if maxX > fb.W { maxX = fb.W }

	minY := cy - r - 1
	if minY < 0 { minY = 0 }
	maxY := cy + r + 1
	if maxY > fb.H { maxY = fb.H }

	rF := float64(r)
	baseA := float64(c.A) / 255.0

	for y := minY; y < maxY; y++ {
		dy := float64(y) - float64(cy)
		fbRow := y * fb.Pitch
		for x := minX; x < maxX; x++ {
			dx := float64(x) - float64(cx)
			dist := math.Sqrt(dx*dx + dy*dy)
			if dist <= rF {
				alpha := baseA
				if dist > rF-1.0 {
					alpha *= (rF - dist)
				}
				if alpha <= 0.001 {
					continue
				}

				sr := float64(c.R) * alpha
				sg := float64(c.G) * alpha
				sb := float64(c.B) * alpha
				invA := 1.0 - alpha

				fbIdx := fbRow + x*4
				db := float64(fb.Pixels[fbIdx+0])
				dg := float64(fb.Pixels[fbIdx+1])
				dr := float64(fb.Pixels[fbIdx+2])
				da := float64(fb.Pixels[fbIdx+3])

				fb.Pixels[fbIdx+0] = byte(sb + db*invA)
				fb.Pixels[fbIdx+1] = byte(sg + dg*invA)
				fb.Pixels[fbIdx+2] = byte(sr + dr*invA)
				fb.Pixels[fbIdx+3] = byte(alpha*255.0 + da*invA)
			}
		}
	}
}

// DrawPawPrint draws a cute dog footprint (main pad + 4 toe pads)
func (fb *Framebuffer) DrawPawPrint(cx, cy int, size int, c color.RGBA) {
	mainR := size / 2
	toeR := size / 5
	fb.DrawCircle(cx, cy+toeR, mainR, c)

	// 4 Toes
	angles := []float64{-math.Pi * 0.65, -math.Pi * 0.4, -math.Pi * 0.15, 0.1}
	dist := float64(mainR + toeR + 1)
	for _, ang := range angles {
		tx := cx + int(dist*math.Cos(ang))
		ty := cy + int(dist*math.Sin(ang)) - toeR
		fb.DrawCircle(tx, ty, toeR, c)
	}
}

// DrawBone draws a cartoon bone token at (cx, cy)
func (fb *Framebuffer) DrawBone(cx, cy int, scale float64, hovered bool) {
	r := int(7.0 * scale)

	boneColor := color.RGBA{250, 245, 235, 240}
	shadowColor := color.RGBA{190, 180, 160, 220}
	if hovered {
		boneColor = color.RGBA{255, 255, 210, 255}
		// Glow ring
		glowColor := color.RGBA{255, 215, 0, 90}
		fb.DrawCircle(cx, cy, int(30.0*scale), glowColor)
	}

	// Bone main shaft
	shaftW := int(26.0 * scale)
	shaftH := int(12.0 * scale)
	fb.DrawFilledRect(cx-shaftW/2, cy-shaftH/2, shaftW, shaftH, boneColor)
	fb.DrawFilledRect(cx-shaftW/2, cy+shaftH/4, shaftW, shaftH/4, shadowColor)

	// Bone knobs (left 2 circles, right 2 circles)
	offX := shaftW / 2
	offY := int(5.0 * scale)

	fb.DrawCircle(cx-offX, cy-offY, r, boneColor)
	fb.DrawCircle(cx-offX, cy+offY, r, boneColor)
	fb.DrawCircle(cx+offX, cy-offY, r, boneColor)
	fb.DrawCircle(cx+offX, cy+offY, r, boneColor)

	// Knobs shadows
	fb.DrawCircle(cx-offX, cy+offY+1, r-2, shadowColor)
	fb.DrawCircle(cx+offX, cy+offY+1, r-2, shadowColor)
}

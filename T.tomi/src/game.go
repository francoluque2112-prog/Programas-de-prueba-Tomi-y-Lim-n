package main

import (
	"fmt"
	"image/color"
	"math"
	"math/rand"
	"path/filepath"
	"strings"
	"sync/atomic"
)

type GameState int

const (
	StateSleeping GameState = iota // Tomi is inside TV
	StateEmerging                  // Tomi is jumping out of TV
	StateChasing                   // Normal Tomi chasing with 5s delay
	StateRetreatingToTV            // Tomi backing up into TV (intro.mp3)
	StatePlayingVideo              // Video Whua.mp4 playing inside TV screen
	StateHijackEmerging            // Hijack Tomi emerges from TV (Chase.mp3)
	StateHijackChasing             // Hijack Tomi high-speed glitch chase
	StateCaught                    // Player caught by Tomi or Htomi
	StateWon                       // All 5 bones delivered to Cuna
)

type Point struct {
	X, Y float64
}

type TimedPoint struct {
	Time float64
	X, Y float64
}

type FloatingParticle struct {
	X, Y   float64
	Vx, Vy float64
	Alpha  float64
	Text   string
	Color  color.RGBA
}

type Game struct {
	ScreenW, ScreenH int
	State            GameState
	StateTimer       float64
	ChaseTime        float64

	// Sprites
	SpriteTV        *Sprite
	SpriteTomi      *Sprite
	SpriteTomiFlip  *Sprite
	SpriteHtomi     *Sprite
	SpriteHtomiFlip *Sprite
	SpriteCuna      *Sprite

	// TV info
	TVX, TVY float64
	TVW, TVH int

	// Cuna info
	CunaX, CunaY float64
	CunaW, CunaH int

	// Tomi info
	TomiX, TomiY       float64
	TomiFacingLeft     bool
	TomiSpeed          float64
	TomiWalkPhase      float64
	TomiIsMoving       bool
	TomiScaleW         int
	TomiScaleH         int
	IsHijacked         bool
	TomiPathHistory    []Point

	// Hijack Tomi glitch / speed variance timers
	HijackPauseTimer float64
	HijackBurstTimer float64
	HijackSpeed      float64

	// Video playback synchronization
	VideoPlaying  int32
	VideoFinished int32
	IntroDuration float64

	// Timed mouse history for delayed replay
	MouseHistory     []TimedPoint
	MouseWaypoints   []Point
	MouseMovingTimer float64
	LiveMouseSpeed   float64
	LiveMouseX       int
	LiveMouseY       int
	LastMouseX       int
	LastMouseY       int
	IsMouseDown      bool
	WasMouseDown     bool

	// Bones count delivered via file drag
	BonesDelivered int
	DeliveredMap   map[int]bool

	// Visual effects
	Particles []*FloatingParticle

	// UI Buttons
	BtnRestartRect [4]int // x, y, w, h
	BtnQuitRect    [4]int // x, y, w, h
	ShouldQuit     bool

	// Desktop folders
	DesktopFolders []string
}

func NewGame(screenW, screenH int, spriteTV, spriteTomi, spriteHtomi, spriteCuna *Sprite) *Game {
	g := &Game{
		ScreenW:      screenW,
		ScreenH:      screenH,
		SpriteTV:     spriteTV,
		SpriteTomi:   spriteTomi,
		SpriteHtomi:  spriteHtomi,
		SpriteCuna:   spriteCuna,
		DeliveredMap: make(map[int]bool),
	}

	if spriteTomi != nil {
		g.SpriteTomiFlip = spriteTomi.FlipH()
	}
	if spriteHtomi != nil {
		g.SpriteHtomiFlip = spriteHtomi.FlipH()
	}

	// TV scale & placement at top center
	g.TVW = 340
	g.TVH = int(float64(g.TVW) * float64(spriteTV.H) / float64(spriteTV.W))
	g.TVX = float64(screenW/2 - g.TVW/2)
	g.TVY = 30

	// Cuna scale & placement at bottom right
	g.CunaW = 300
	g.CunaH = int(float64(g.CunaW) * float64(spriteCuna.H) / float64(spriteCuna.W))
	g.CunaX = float64(screenW - g.CunaW - 60)
	g.CunaY = float64(screenH - g.CunaH - 60)

	// Tomi scale
	g.TomiScaleW = 120
	g.TomiScaleH = int(float64(g.TomiScaleW) * float64(spriteTomi.H) / float64(spriteTomi.W))

	// Scale base sprites
	g.SpriteTV = spriteTV.Scale(g.TVW, g.TVH)
	g.SpriteCuna = spriteCuna.Scale(g.CunaW, g.CunaH)
	g.SpriteTomi = spriteTomi.Scale(g.TomiScaleW, g.TomiScaleH)
	g.SpriteTomiFlip = g.SpriteTomi.FlipH()

	if spriteHtomi != nil {
		g.SpriteHtomi = spriteHtomi.Scale(g.TomiScaleW, g.TomiScaleH)
		g.SpriteHtomiFlip = g.SpriteHtomi.FlipH()
	}

	g.Reset()
	return g
}

func (g *Game) Reset() {
	StopAllMusic()
	g.State = StateSleeping
	g.StateTimer = 0
	g.ChaseTime = 0
	g.IsHijacked = false
	tvStartX := g.TVX + float64(g.TVW)/2 - float64(g.TomiScaleW)/2
	tvStartY := g.TVY + float64(g.TVH)/2 - float64(g.TomiScaleH)/2 + 10
	g.TomiX = tvStartX
	g.TomiY = tvStartY
	g.TomiPathHistory = []Point{{X: tvStartX, Y: tvStartY}}
	g.TomiSpeed = 9.5
	g.HijackSpeed = 15.0
	g.HijackPauseTimer = 0
	g.HijackBurstTimer = 1.0
	g.TomiFacingLeft = false
	g.TomiWalkPhase = 0
	g.TomiIsMoving = false

	atomic.StoreInt32(&g.VideoPlaying, 0)
	atomic.StoreInt32(&g.VideoFinished, 0)

	g.MouseHistory = make([]TimedPoint, 0, 1000)
	g.MouseWaypoints = make([]Point, 0, 1000)
	g.MouseMovingTimer = 0
	g.LiveMouseSpeed = 0
	g.BonesDelivered = 0
	g.DeliveredMap = make(map[int]bool)
	g.Particles = make([]*FloatingParticle, 0)
}

func (g *Game) Update(dt float64, mouseX, mouseY int, isMouseDown bool) {
	// Detect live mouse movement and calculate speed
	mdx := float64(mouseX - g.LastMouseX)
	mdy := float64(mouseY - g.LastMouseY)
	mouseDist := math.Sqrt(mdx*mdx + mdy*mdy)

	if dt > 0.0001 {
		instantSpeed := mouseDist / dt
		if mouseDist > 0.8 {
			g.LiveMouseSpeed = g.LiveMouseSpeed*0.4 + instantSpeed*0.6
			g.MouseMovingTimer = 0.15
		} else {
			g.LiveMouseSpeed = g.LiveMouseSpeed * 0.3
			if g.MouseMovingTimer > 0 {
				g.MouseMovingTimer -= dt
			}
		}
	}
	g.LastMouseX = mouseX
	g.LastMouseY = mouseY

	g.LiveMouseX = mouseX
	g.LiveMouseY = mouseY
	g.IsMouseDown = isMouseDown
	g.StateTimer += dt

	mouseJustPressed := isMouseDown && !g.WasMouseDown
	g.WasMouseDown = isMouseDown

	// Update particles
	for i := len(g.Particles) - 1; i >= 0; i-- {
		p := g.Particles[i]
		p.X += p.Vx * dt
		p.Y += p.Vy * dt
		p.Alpha -= dt * 0.5
		if p.Alpha <= 0 {
			g.Particles = append(g.Particles[:i], g.Particles[i+1:]...)
		}
	}

	switch g.State {
	case StateSleeping:
		tvCenterX := g.TVX + float64(g.TVW)/2
		tvCenterY := g.TVY + float64(g.TVH)/2
		dx := float64(mouseX) - tvCenterX
		dy := float64(mouseY) - tvCenterY
		dist := math.Sqrt(dx*dx + dy*dy)

		if dist < 170.0 {
			g.State = StateEmerging
			g.StateTimer = 0
			PlayBarkSound()
			g.Particles = append(g.Particles, &FloatingParticle{
				X:     tvCenterX,
				Y:     tvCenterY - 40,
				Vx:    0,
				Vy:    -30,
				Alpha: 1.0,
				Text:  "¡GUAU! (!)",
				Color: color.RGBA{255, 220, 0, 255},
			})
		}

	case StateEmerging:
		jumpProgress := g.StateTimer / 0.5
		if jumpProgress >= 1.0 {
			g.State = StateChasing
			g.StateTimer = 0
			g.ChaseTime = 0
			g.MouseHistory = []TimedPoint{{Time: 0, X: float64(mouseX), Y: float64(mouseY)}}
			g.MouseWaypoints = []Point{{X: float64(mouseX), Y: float64(mouseY)}}
		} else {
			tvCenterX := g.TVX + float64(g.TVW)/2 - float64(g.TomiScaleW)/2
			tvBottomY := g.TVY + float64(g.TVH) - 20
			g.TomiX = tvCenterX
			g.TomiY = g.TVY + 40 + (tvBottomY-(g.TVY+40))*jumpProgress
		}
		// Record point
		g.recordTomiStep()

	case StateChasing:
		g.ChaseTime += dt

		// Record live mouse position to history and waypoints
		g.MouseHistory = append(g.MouseHistory, TimedPoint{
			Time: g.ChaseTime,
			X:    float64(mouseX),
			Y:    float64(mouseY),
		})

		cutoff := g.ChaseTime - 12.0
		for len(g.MouseHistory) > 1 && g.MouseHistory[1].Time < cutoff {
			g.MouseHistory = g.MouseHistory[1:]
		}

		if len(g.MouseWaypoints) == 0 {
			g.MouseWaypoints = append(g.MouseWaypoints, Point{X: float64(mouseX), Y: float64(mouseY)})
		} else {
			lastW := g.MouseWaypoints[len(g.MouseWaypoints)-1]
			wdx := float64(mouseX) - lastW.X
			wdy := float64(mouseY) - lastW.Y
			if wdx*wdx+wdy*wdy >= 25.0 {
				g.MouseWaypoints = append(g.MouseWaypoints, Point{X: float64(mouseX), Y: float64(mouseY)})
			}
		}

		// "dependiendo la velocidad que muevas el mouse tomi normal se mueve a esa velocidad"
		if g.MouseMovingTimer > 0 && g.LiveMouseSpeed > 5.0 {
			var target Point
			if len(g.MouseWaypoints) > 0 {
				target = g.MouseWaypoints[0]
			} else {
				target = Point{X: float64(mouseX), Y: float64(mouseY)}
			}

			tomiCenterX := g.TomiX + float64(g.TomiScaleW)/2
			tomiCenterY := g.TomiY + float64(g.TomiScaleH)*0.75

			tdx := target.X - tomiCenterX
			tdy := target.Y - tomiCenterY
			distToTarget := math.Sqrt(tdx*tdx + tdy*tdy)

			// Step proportional to mouse speed
			step := g.LiveMouseSpeed * dt * 1.05
			if step < 2.5 {
				step = 2.5
			}
			if step > 30.0 {
				step = 30.0
			}

			if distToTarget <= step {
				g.TomiX = target.X - float64(g.TomiScaleW)/2
				g.TomiY = target.Y - float64(g.TomiScaleH)*0.75
				if len(g.MouseWaypoints) > 0 {
					g.MouseWaypoints = g.MouseWaypoints[1:]
				}
			} else {
				nx := tdx / distToTarget
				ny := tdy / distToTarget
				g.TomiX += nx * step
				g.TomiY += ny * step

				if nx < -0.1 {
					g.TomiFacingLeft = true
				} else if nx > 0.1 {
					g.TomiFacingLeft = false
				}
			}

			g.TomiIsMoving = true
			walkSpeed := g.LiveMouseSpeed / 30.0
			if walkSpeed < 2.0 {
				walkSpeed = 2.0
			}
			g.TomiWalkPhase += dt * walkSpeed
			g.recordTomiStep()
		} else {
			g.TomiIsMoving = false
		}

		// Collision with live mouse
		tomiCatchX := g.TomiX + float64(g.TomiScaleW)/2
		tomiCatchY := g.TomiY + float64(g.TomiScaleH)*0.6
		cdx := float64(mouseX) - tomiCatchX
		cdy := float64(mouseY) - tomiCatchY
		if math.Sqrt(cdx*cdx+cdy*cdy) < 46.0 {
			g.State = StateCaught
			g.StateTimer = 0
			PlayLoseSound()
		}

	case StateRetreatingToTV:
		// Tomi retraces his exact past path backwards into the TV, accelerating each second!
		t := g.StateTimer
		// Speed grows each second of Intro.mp3
		currentStep := (7.0 + t*8.0) * 60.0 * dt

		// If time reached Intro.mp3 duration or path is finished, enter TV
		if (g.IntroDuration > 0 && t >= g.IntroDuration-0.1) || len(g.TomiPathHistory) == 0 {
			// Tomi is now inside the TV! Stop intro music and start video Whua.mp4
			StopIntroMusic()
			g.State = StatePlayingVideo
			g.StateTimer = 0
			atomic.StoreInt32(&g.VideoPlaying, 1)
			atomic.StoreInt32(&g.VideoFinished, 0)

			// Place inside TV
			g.TomiX = g.TVX + float64(g.TVW)/2 - float64(g.TomiScaleW)/2
			g.TomiY = g.TVY + float64(g.TVH)/2 - float64(g.TomiScaleH)/2 + 10

			// Calculate TV screen coordinates
			screenX := int(g.TVX + float64(g.TVW)*0.09)
			screenY := int(g.TVY + float64(g.TVH)*0.33)
			screenW := int(float64(g.TVW) * 0.68)
			screenH := int(float64(g.TVH) * 0.55)

			PlayTVVideo(screenX, screenY, screenW, screenH, func() {
				atomic.StoreInt32(&g.VideoFinished, 1)
			})
		} else {
			target := g.TomiPathHistory[len(g.TomiPathHistory)-1]
			rdx := target.X - g.TomiX
			rdy := target.Y - g.TomiY
			rDist := math.Sqrt(rdx*rdx + rdy*rdy)

			if rDist <= currentStep || rDist < 3.0 {
				g.TomiX = target.X
				g.TomiY = target.Y
				g.TomiPathHistory = g.TomiPathHistory[:len(g.TomiPathHistory)-1]
			} else {
				g.TomiX += (rdx / rDist) * currentStep
				g.TomiY += (rdy / rDist) * currentStep
				// Walking backward: facing opposite direction
				g.TomiFacingLeft = rdx > 0
				g.TomiWalkPhase += dt * (14.0 + t*12.0)
			}
		}

	case StatePlayingVideo:
		// Waiting for Whua.mp4 to complete
		if atomic.LoadInt32(&g.VideoFinished) == 1 || g.StateTimer > 25.0 {
			// Video finished! Hijack Tomi emerges with Chase.mp3!
			g.State = StateHijackEmerging
			g.StateTimer = 0
			g.IsHijacked = true
			PlayChaseMusic()

			g.Particles = append(g.Particles, &FloatingParticle{
				X:     g.TVX + float64(g.TVW)/2,
				Y:     g.TVY + float64(g.TVH)/2 - 30,
				Vx:    0,
				Vy:    -40,
				Alpha: 1.0,
				Text:  "⚠️ ¡¡HIJACK TOMI HA DESPERTADO!! ⚠️",
				Color: color.RGBA{255, 240, 0, 255},
			})
		}

	case StateHijackEmerging:
		jumpProgress := g.StateTimer / 0.6
		if jumpProgress >= 1.0 {
			g.State = StateHijackChasing
			g.StateTimer = 0
			g.ChaseTime = 0
			g.MouseHistory = []TimedPoint{{Time: 0, X: float64(mouseX), Y: float64(mouseY)}}
			g.HijackSpeed = 15.0
			g.HijackPauseTimer = 0
			g.HijackBurstTimer = 0.8
		} else {
			tvCenterX := g.TVX + float64(g.TVW)/2 - float64(g.TomiScaleW)/2
			tvBottomY := g.TVY + float64(g.TVH) - 20
			g.TomiX = tvCenterX
			g.TomiY = g.TVY + 40 + (tvBottomY-(g.TVY+40))*jumpProgress
		}

	case StateHijackChasing:
		g.ChaseTime += dt

		// Record live mouse position
		g.MouseHistory = append(g.MouseHistory, TimedPoint{
			Time: g.ChaseTime,
			X:    float64(mouseX),
			Y:    float64(mouseY),
		})

		cutoff := g.ChaseTime - 12.0
		for len(g.MouseHistory) > 1 && g.MouseHistory[1].Time < cutoff {
			g.MouseHistory = g.MouseHistory[1:]
		}

		// Random pause and speed variance mechanics for Hijack Tomi
		if g.HijackPauseTimer > 0 {
			g.HijackPauseTimer -= dt
			g.TomiIsMoving = false
			// Glitch spark while paused
			if rand.Float64() < 0.3 {
				g.Particles = append(g.Particles, &FloatingParticle{
					X:     g.TomiX + float64(g.TomiScaleW)/2 + (rand.Float64()-0.5)*30,
					Y:     g.TomiY + float64(g.TomiScaleH)/2 + (rand.Float64()-0.5)*30,
					Vx:    (rand.Float64() - 0.5) * 40,
					Vy:    (rand.Float64() - 0.5) * 40,
					Alpha: 0.8,
					Text:  "⚡",
					Color: color.RGBA{255, 255, 0, 255},
				})
			}
		} else {
			g.HijackBurstTimer -= dt
			if g.HijackBurstTimer <= 0 {
				// Random twitch: pause briefly or surge speed!
				if rand.Float64() < 0.4 {
					g.HijackPauseTimer = 0.15 + rand.Float64()*0.3
				} else {
					g.HijackSpeed = 13.0 + rand.Float64()*9.0 // 13 to 22 px/tick!
				}
				g.HijackBurstTimer = 0.6 + rand.Float64()*1.2
			}

			// Delayed mouse replay (e.g. 4s delay), but DOES NOT stop when mouse is still!
			targetTime := g.ChaseTime - 4.0
			var targetX, targetY float64
			if targetTime >= 0 && len(g.MouseHistory) > 0 {
				targetX, targetY, _ = g.getInterpolatedPoint(targetTime)
			} else {
				targetX = float64(mouseX)
				targetY = float64(mouseY)
			}

			tomiCenterX := g.TomiX + float64(g.TomiScaleW)/2
			tomiCenterY := g.TomiY + float64(g.TomiScaleH)*0.75

			tdx := targetX - tomiCenterX
			tdy := targetY - tomiCenterY
			distToTarget := math.Sqrt(tdx*tdx + tdy*tdy)

			if distToTarget > 2.0 {
				g.TomiIsMoving = true
				step := g.HijackSpeed * 60.0 * dt
				if distToTarget <= step {
					g.TomiX = targetX - float64(g.TomiScaleW)/2
					g.TomiY = targetY - float64(g.TomiScaleH)*0.75
				} else {
					nx := tdx / distToTarget
					ny := tdy / distToTarget
					g.TomiX += nx * step
					g.TomiY += ny * step

					if nx < -0.1 {
						g.TomiFacingLeft = true
					} else if nx > 0.1 {
						g.TomiFacingLeft = false
					}
				}
				g.TomiWalkPhase += dt * 18.0
			}
		}

		// Collision check with live mouse
		tomiCatchX := g.TomiX + float64(g.TomiScaleW)/2
		tomiCatchY := g.TomiY + float64(g.TomiScaleH)*0.6
		cdx := float64(mouseX) - tomiCatchX
		cdy := float64(mouseY) - tomiCatchY
		if math.Sqrt(cdx*cdx+cdy*cdy) < 48.0 {
			StopChaseMusic()
			g.State = StateCaught
			g.StateTimer = 0
			PlayLoseSound()
		}

	case StateWon:
		StopAllMusic()
		cunaCenterX := g.CunaX + float64(g.CunaW)/2 - float64(g.TomiScaleW)/2
		cunaCenterY := g.CunaY + float64(g.CunaH)/2 - float64(g.TomiScaleH)/2 - 10
		wdx := cunaCenterX - g.TomiX
		wdy := cunaCenterY - g.TomiY
		wDist := math.Sqrt(wdx*wdx + wdy*wdy)

		if wDist > 4.0 {
			step := 6.0 * 60.0 * dt
			g.TomiX += (wdx / wDist) * step
			g.TomiY += (wdy / wDist) * step
			g.TomiFacingLeft = wdx < 0
			g.TomiWalkPhase += dt * 8.0
		}

		if rand.Float64() < 0.08 {
			g.Particles = append(g.Particles, &FloatingParticle{
				X:     g.CunaX + float64(rand.Intn(g.CunaW)),
				Y:     g.CunaY + float64(rand.Intn(g.CunaH/2)),
				Vx:    (rand.Float64() - 0.5) * 20,
				Vy:    -40 - rand.Float64()*30,
				Alpha: 1.0,
				Text:  "♥",
				Color: color.RGBA{255, 80, 120, 255},
			})
		}

		if mouseJustPressed {
			if g.isPointInRect(mouseX, mouseY, g.BtnRestartRect) {
				g.Reset()
			} else if g.isPointInRect(mouseX, mouseY, g.BtnQuitRect) {
				g.ShouldQuit = true
			}
		}

	case StateCaught:
		StopAllMusic()
		if mouseJustPressed {
			if g.isPointInRect(mouseX, mouseY, g.BtnRestartRect) {
				g.Reset()
			} else if g.isPointInRect(mouseX, mouseY, g.BtnQuitRect) {
				g.ShouldQuit = true
			}
		}
	}
}

func (g *Game) recordTomiStep() {
	if len(g.TomiPathHistory) == 0 {
		g.TomiPathHistory = append(g.TomiPathHistory, Point{X: g.TomiX, Y: g.TomiY})
		return
	}
	last := g.TomiPathHistory[len(g.TomiPathHistory)-1]
	dx := g.TomiX - last.X
	dy := g.TomiY - last.Y
	if dx*dx+dy*dy >= 9.0 {
		g.TomiPathHistory = append(g.TomiPathHistory, Point{X: g.TomiX, Y: g.TomiY})
	}
}

func (g *Game) getInterpolatedPoint(targetTime float64) (float64, float64, bool) {
	if len(g.MouseHistory) == 0 {
		return float64(g.LiveMouseX), float64(g.LiveMouseY), false
	}
	if targetTime <= g.MouseHistory[0].Time {
		return g.MouseHistory[0].X, g.MouseHistory[0].Y, false
	}
	last := g.MouseHistory[len(g.MouseHistory)-1]
	if targetTime >= last.Time {
		return last.X, last.Y, false
	}

	for i := 0; i < len(g.MouseHistory)-1; i++ {
		p1 := g.MouseHistory[i]
		p2 := g.MouseHistory[i+1]
		if p1.Time <= targetTime && targetTime <= p2.Time {
			tSpan := p2.Time - p1.Time
			if tSpan <= 0.0001 {
				return p1.X, p1.Y, false
			}
			factor := (targetTime - p1.Time) / tSpan
			x := p1.X + (p2.X-p1.X)*factor
			y := p1.Y + (p2.Y-p1.Y)*factor

			mdx := p2.X - p1.X
			mdy := p2.Y - p1.Y
			dist := math.Sqrt(mdx*mdx + mdy*mdy)
			wasMoving := dist > 1.5

			return x, y, wasMoving
		}
	}

	return last.X, last.Y, false
}

func (g *Game) isPointInRect(x, y int, r [4]int) bool {
	return x >= r[0] && x <= r[0]+r[2] && y >= r[1] && y <= r[1]+r[3]
}

func (g *Game) Render(fb *Framebuffer) {
	fb.Clear()

	// 1. Draw Cuna (Asset Cuna.png)
	fb.DrawSprite(g.SpriteCuna, int(g.CunaX), int(g.CunaY), 1.0)
	cunaLabel := fmt.Sprintf("Cuna de Tomi [%d/5 🦴]", g.BonesDelivered)
	cunaCol := color.RGBA{255, 255, 255, 220}
	fb.DrawText(int(g.CunaX)+g.CunaW/2-len(cunaLabel)*4, int(g.CunaY)+g.CunaH+8, cunaLabel, 1, cunaCol, true)

	// Draw bones inside Cuna
	for i := 0; i < g.BonesDelivered; i++ {
		bx := int(g.CunaX) + 50 + i*42
		by := int(g.CunaY) + int(float64(g.CunaH)*0.65)
		fb.DrawBone(bx, by, 0.7, false)
	}

	// 2. Draw Tomi or Htomi
	currentSpr := g.SpriteTomi
	currentFlip := g.SpriteTomiFlip
	if g.IsHijacked && g.SpriteHtomi != nil {
		currentSpr = g.SpriteHtomi
		currentFlip = g.SpriteHtomiFlip
	}

	if g.State == StateSleeping {
		screenCenterX := int(g.TVX) + g.TVW/2
		screenCenterY := int(g.TVY) + g.TVH/2 + 8
		fb.DrawSprite(g.SpriteTomi, screenCenterX-g.TomiScaleW/2, screenCenterY-g.TomiScaleH/2, 0.5)
		fb.DrawText(screenCenterX+20, screenCenterY-40, "Zzz...", 2, color.RGBA{200, 220, 255, 220}, true)
	} else if g.State != StatePlayingVideo {
		bounceY := 0
		if g.TomiIsMoving || g.State == StateWon || g.State == StateRetreatingToTV {
			bounceY = int(math.Sin(g.TomiWalkPhase) * 4.0)
		}

		spr := currentSpr
		if g.TomiFacingLeft {
			spr = currentFlip
		}
		fb.DrawSprite(spr, int(g.TomiX), int(g.TomiY)+bounceY, 1.0)
	}

	// 3. Draw TV (Asset TV antiguo.png)
	fb.DrawSprite(g.SpriteTV, int(g.TVX), int(g.TVY), 1.0)

	// 4. Draw Floating Particles
	for _, p := range g.Particles {
		pCol := p.Color
		pCol.A = byte(p.Alpha * 255.0)
		fb.DrawText(int(p.X)-len(p.Text)*4, int(p.Y), p.Text, 1, pCol, true)
	}

	// 5. Draw Top HUD Banner
	g.drawHUD(fb)

	// 6. Draw Modal Windows for Game Over / Victory
	if g.State == StateCaught {
		g.drawGameOverModal(fb)
	} else if g.State == StateWon {
		g.drawVictoryModal(fb)
	}
}

func (g *Game) drawHUD(fb *Framebuffer) {
	bannerW := 740
	bannerH := 50
	bannerX := g.ScreenW/2 - bannerW/2
	bannerY := 12

	bgCol := color.RGBA{20, 25, 35, 215}
	borderCol := color.RGBA{255, 180, 50, 240}
	if g.IsHijacked {
		borderCol = color.RGBA{255, 50, 50, 255}
		bgCol = color.RGBA{35, 15, 15, 230}
	}
	fb.DrawRoundedBox(bannerX, bannerY, bannerW, bannerH, 8, bgCol, borderCol, 2)

	title := fmt.Sprintf("HUESITOS ENTREGADOS A LA CUNA: %d / 5", g.BonesDelivered)
	if g.IsHijacked {
		title = fmt.Sprintf("⚠️ ¡¡ESQUIVA A HIJACK TOMI!! ⚠️  HUESITOS: %d / 5", g.BonesDelivered)
	}
	fb.DrawText(bannerX+18, bannerY+8, title, 2, color.RGBA{255, 220, 80, 255}, true)

	hint := "Arrastra archivos a la Cuna | Arrastra LimonWhiskas.exe a la TV | [O] Abrir Escritorio | [R] Reiniciar | [ESC] Salir"
	fb.DrawText(bannerX+18, bannerY+30, hint, 1, color.RGBA{210, 225, 240, 230}, true)

	if g.State == StateSleeping {
		wakeHint := "¡Acerca el mouse a la TV antigua para despertar a Tomi!"
		fb.DrawText(int(g.TVX)+g.TVW/2-len(wakeHint)*4, int(g.TVY)+g.TVH+10, wakeHint, 1, color.RGBA{255, 255, 100, 240}, true)
	} else if g.State == StateChasing {
		if g.ChaseTime < 5.0 {
			delayHint := fmt.Sprintf("¡Tomi comenzara a seguir el recorrido en %.1fs!", 5.0-g.ChaseTime)
			fb.DrawText(int(g.TVX)+g.TVW/2-len(delayHint)*4, int(g.TVY)+g.TVH+10, delayHint, 1, color.RGBA{255, 200, 50, 240}, true)
		}
	} else if g.State == StateRetreatingToTV {
		hint := "Tomi esta retrocediendo a la TV..."
		fb.DrawText(int(g.TVX)+g.TVW/2-len(hint)*4, int(g.TVY)+g.TVH+10, hint, 1, color.RGBA{255, 180, 80, 240}, true)
	} else if g.State == StatePlayingVideo {
		hint := "Reproduciendo video en la TV..."
		fb.DrawText(int(g.TVX)+g.TVW/2-len(hint)*4, int(g.TVY)+g.TVH+10, hint, 1, color.RGBA{255, 80, 80, 240}, true)
	}
}

func (g *Game) drawGameOverModal(fb *Framebuffer) {
	modalW := 620
	modalH := 240
	modalX := g.ScreenW/2 - modalW/2
	modalY := g.ScreenH/2 - modalH/2

	bgCol := color.RGBA{30, 15, 15, 245}
	borderCol := color.RGBA{255, 60, 60, 255}
	fb.DrawRoundedBox(modalX, modalY, modalW, modalH, 12, bgCol, borderCol, 3)

	header := "¡TOMI TE ATRAPO!"
	if g.IsHijacked {
		header = "¡¡HIJACK TOMI TE COMIO EL MOUSE!!"
	}
	fb.DrawText(modalX+modalW/2-len(header)*8, modalY+24, header, 2, color.RGBA{255, 80, 80, 255}, true)

	msg := "\"no me diste los huesitos por eso me comi tu mouse\""
	fb.DrawText(modalX+modalW/2-len(msg)*4, modalY+75, msg, 1, color.RGBA{255, 230, 230, 255}, true)

	submsg := fmt.Sprintf("Lograste llevar %d de 5 huesitos a la cuna.", g.BonesDelivered)
	fb.DrawText(modalX+modalW/2-len(submsg)*4, modalY+105, submsg, 1, color.RGBA{200, 200, 200, 240}, true)

	btnW := 180
	btnH := 40
	btnY := modalY + 160

	btn1X := modalX + modalW/2 - btnW - 15
	g.BtnRestartRect = [4]int{btn1X, btnY, btnW, btnH}
	h1 := g.isPointInRect(g.LiveMouseX, g.LiveMouseY, g.BtnRestartRect)
	btn1Bg := color.RGBA{50, 140, 60, 240}
	if h1 {
		btn1Bg = color.RGBA{70, 190, 80, 255}
	}
	fb.DrawRoundedBox(btn1X, btnY, btnW, btnH, 6, btn1Bg, color.RGBA{255, 255, 255, 220}, 2)
	b1Text := "Reintentar [R]"
	fb.DrawText(btn1X+btnW/2-len(b1Text)*4, btnY+14, b1Text, 1, color.RGBA{255, 255, 255, 255}, true)

	btn2X := modalX + modalW/2 + 15
	g.BtnQuitRect = [4]int{btn2X, btnY, btnW, btnH}
	h2 := g.isPointInRect(g.LiveMouseX, g.LiveMouseY, g.BtnQuitRect)
	btn2Bg := color.RGBA{140, 40, 40, 240}
	if h2 {
		btn2Bg = color.RGBA{190, 50, 50, 255}
	}
	fb.DrawRoundedBox(btn2X, btnY, btnW, btnH, 6, btn2Bg, color.RGBA{255, 255, 255, 220}, 2)
	b2Text := "Salir [ESC]"
	fb.DrawText(btn2X+btnW/2-len(b2Text)*4, btnY+14, b2Text, 1, color.RGBA{255, 255, 255, 255}, true)
}

func (g *Game) drawVictoryModal(fb *Framebuffer) {
	modalW := 640
	modalH := 240
	modalX := g.ScreenW/2 - modalW/2
	modalY := g.ScreenH/2 - modalH/2

	bgCol := color.RGBA{15, 30, 20, 245}
	borderCol := color.RGBA{80, 230, 100, 255}
	fb.DrawRoundedBox(modalX, modalY, modalW, modalH, 12, bgCol, borderCol, 3)

	header := "¡GANASTE! :D"
	fb.DrawText(modalX+modalW/2-len(header)*8, modalY+24, header, 2, color.RGBA{100, 255, 120, 255}, true)

	msg := "¡Tomi esta feliz con todos sus huesitos en su cuna!"
	if g.IsHijacked {
		msg = "¡Lograste salvar a Tomi y calmarlo en su cuna!"
	}
	fb.DrawText(modalX+modalW/2-len(msg)*4, modalY+75, msg, 1, color.RGBA{230, 255, 230, 255}, true)

	submsg := "¡Completaste el desafio esquivando el recorrido del perro!"
	fb.DrawText(modalX+modalW/2-len(submsg)*4, modalY+105, submsg, 1, color.RGBA{200, 230, 200, 240}, true)

	btnW := 190
	btnH := 40
	btnY := modalY + 160

	btn1X := modalX + modalW/2 - btnW - 15
	g.BtnRestartRect = [4]int{btn1X, btnY, btnW, btnH}
	h1 := g.isPointInRect(g.LiveMouseX, g.LiveMouseY, g.BtnRestartRect)
	btn1Bg := color.RGBA{40, 140, 160, 240}
	if h1 {
		btn1Bg = color.RGBA{50, 180, 210, 255}
	}
	fb.DrawRoundedBox(btn1X, btnY, btnW, btnH, 6, btn1Bg, color.RGBA{255, 255, 255, 220}, 2)
	b1Text := "Jugar de nuevo [R]"
	fb.DrawText(btn1X+btnW/2-len(b1Text)*4, btnY+14, b1Text, 1, color.RGBA{255, 255, 255, 255}, true)

	btn2X := modalX + modalW/2 + 15
	g.BtnQuitRect = [4]int{btn2X, btnY, btnW, btnH}
	h2 := g.isPointInRect(g.LiveMouseX, g.LiveMouseY, g.BtnQuitRect)
	btn2Bg := color.RGBA{120, 50, 60, 240}
	if h2 {
		btn2Bg = color.RGBA{160, 60, 70, 255}
	}
	fb.DrawRoundedBox(btn2X, btnY, btnW, btnH, 6, btn2Bg, color.RGBA{255, 255, 255, 220}, 2)
	b2Text := "Salir [ESC]"
	fb.DrawText(btn2X+btnW/2-len(b2Text)*4, btnY+14, b2Text, 1, color.RGBA{255, 255, 255, 255}, true)
}

// OnFileDrop handles real Windows Explorer files dragged and dropped onto the game overlay
func (g *Game) OnFileDrop(filePath string, dropX, dropY int) {
	fileName := filepath.Base(filePath)
	lowerName := strings.ToLower(fileName)

	// 1. Check if dropped file is LimonWhiskas.exe onto the TV
	if strings.Contains(lowerName, "limonwhiskas") || (strings.HasSuffix(lowerName, ".exe") && strings.Contains(lowerName, "limon")) {
		tvCenterX := g.TVX + float64(g.TVW)/2
		tvCenterY := g.TVY + float64(g.TVH)/2
		dropDx := float64(dropX) - tvCenterX
		dropDy := float64(dropY) - tvCenterY

		isOverTV := math.Abs(dropDx) < float64(g.TVW)*0.6 && math.Abs(dropDy) < float64(g.TVH)*0.6

		if isOverTV {
			// Trigger LimonWhiskas retreat into TV with intro.mp3
			if g.State == StateChasing || g.State == StateSleeping {
				g.State = StateRetreatingToTV
				g.StateTimer = 0
				PlayIntroMusic()
				g.IntroDuration = GetIntroDuration()
				if g.IntroDuration <= 0 {
					g.IntroDuration = 5.0
				}

				g.Particles = append(g.Particles, &FloatingParticle{
					X:     tvCenterX,
					Y:     g.TVY + 10,
					Vx:    0,
					Vy:    -35,
					Alpha: 1.0,
					Text:  "🍋 ¡LimonWhiskas.exe insertado en la TV!",
					Color: color.RGBA{255, 220, 0, 255},
				})
			}
			return
		} else {
			g.Particles = append(g.Particles, &FloatingParticle{
				X:     float64(dropX),
				Y:     float64(dropY) - 20,
				Vx:    0,
				Vy:    -30,
				Alpha: 1.0,
				Text:  "¡Arrastra LimonWhiskas.exe a la TV!",
				Color: color.RGBA{255, 200, 50, 255},
			})
			return
		}
	}

	// 2. Check if dropped file is a bone file for the Cuna
	if g.State == StateSleeping {
		g.State = StateEmerging
		g.StateTimer = 0
		PlayBarkSound()
	}

	cunaCenterX := g.CunaX + float64(g.CunaW)/2
	cunaCenterY := g.CunaY + float64(g.CunaH)/2
	dropDx := float64(dropX) - cunaCenterX
	dropDy := float64(dropY) - cunaCenterY

	isOverCuna := math.Abs(dropDx) < float64(g.CunaW)*0.65 && math.Abs(dropDy) < float64(g.CunaH)*0.65

	if !isOverCuna {
		g.Particles = append(g.Particles, &FloatingParticle{
			X:     float64(dropX),
			Y:     float64(dropY) - 20,
			Vx:    0,
			Vy:    -30,
			Alpha: 1.0,
			Text:  "¡Arrastra los huesos a la CUNA!",
			Color: color.RGBA{255, 100, 100, 255},
		})
		return
	}

	boneID := 0
	if strings.Contains(fileName, "1") || strings.Contains(filePath, "1") {
		boneID = 1
	} else if strings.Contains(fileName, "2") || strings.Contains(filePath, "2") {
		boneID = 2
	} else if strings.Contains(fileName, "3") || strings.Contains(filePath, "3") {
		boneID = 3
	} else if strings.Contains(fileName, "4") || strings.Contains(filePath, "4") {
		boneID = 4
	} else if strings.Contains(fileName, "5") || strings.Contains(filePath, "5") {
		boneID = 5
	} else {
		for i := 1; i <= 5; i++ {
			if !g.DeliveredMap[i] {
				boneID = i
				break
			}
		}
	}

	if boneID >= 1 && boneID <= 5 {
		if g.DeliveredMap[boneID] {
			g.Particles = append(g.Particles, &FloatingParticle{
				X:     cunaCenterX,
				Y:     g.CunaY + 20,
				Vx:    0,
				Vy:    -30,
				Alpha: 1.0,
				Text:  fmt.Sprintf("¡Huesito #%d ya estaba en la Cuna!", boneID),
				Color: color.RGBA{255, 220, 80, 255},
			})
			return
		}

		g.DeliveredMap[boneID] = true
		g.BonesDelivered++
		PlayCunaSound()

		g.Particles = append(g.Particles, &FloatingParticle{
			X:     cunaCenterX,
			Y:     g.CunaY + 15,
			Vx:    0,
			Vy:    -35,
			Alpha: 1.0,
			Text:  fmt.Sprintf("🦴 ¡Archivo Huesito #%d entregado! (%d/5)", boneID, g.BonesDelivered),
			Color: color.RGBA{100, 255, 120, 255},
		})

		if g.BonesDelivered >= 5 {
			StopAllMusic()
			g.State = StateWon
			g.StateTimer = 0
			PlayWinSound()
		}
	}
}

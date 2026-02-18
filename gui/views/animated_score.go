//go:build gui

package views

import (
	"fmt"
	"image/color"
	"math"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"
)

// ScoreRing is an animated circular progress indicator for the performance score
type ScoreRing struct {
	widget.BaseWidget
	mu           sync.Mutex
	score        int
	displayScore float64
	animating    bool
	text         *canvas.Text
	ring         *canvas.Circle
}

// NewScoreRing creates a new animated score ring widget
func NewScoreRing() *ScoreRing {
	s := &ScoreRing{
		score:        0,
		displayScore: 0,
	}
	s.ExtendBaseWidget(s)
	return s
}

// SetScore sets the target score and animates to it
func (s *ScoreRing) SetScore(score int) {
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}
	s.mu.Lock()
	s.score = score
	alreadyRunning := s.animating
	s.mu.Unlock()
	if !alreadyRunning {
		go s.animateTo(score)
	}
}

// GetScore returns the current target score
func (s *ScoreRing) GetScore() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.score
}

func (s *ScoreRing) animateTo(target int) {
	s.mu.Lock()
	s.animating = true
	s.mu.Unlock()
	defer func() {
		s.mu.Lock()
		s.animating = false
		s.mu.Unlock()
	}()

	targetFloat := float64(target)
	for {
		s.mu.Lock()
		cur := s.displayScore
		s.mu.Unlock()
		if math.Abs(cur-targetFloat) < 0.5 {
			break
		}
		s.mu.Lock()
		if s.displayScore < targetFloat {
			s.displayScore = math.Min(s.displayScore+0.8, targetFloat)
		} else {
			s.displayScore = math.Max(s.displayScore-0.8, targetFloat)
		}
		s.mu.Unlock()
		s.Refresh()
		time.Sleep(15 * time.Millisecond)
	}
	s.mu.Lock()
	s.displayScore = targetFloat
	s.mu.Unlock()
	s.Refresh()
}

// CreateRenderer creates the widget renderer
func (s *ScoreRing) CreateRenderer() fyne.WidgetRenderer {
	s.mu.Lock()
	scoreColor := s.getColorForScore()
	s.mu.Unlock()

	s.ring = canvas.NewCircle(scoreColor)
	s.ring.StrokeWidth = 10
	s.ring.StrokeColor = scoreColor
	s.ring.FillColor = color.Transparent

	s.text = canvas.NewText("--", color.White)
	s.text.TextSize = 48
	s.text.TextStyle = fyne.TextStyle{Bold: true}
	s.text.Alignment = fyne.TextAlignCenter

	return &scoreRingRenderer{
		ring:  s.ring,
		text:  s.text,
		score: s,
	}
}

// getColorForScore returns the color for the current displayScore.
// Caller must hold s.mu.
func (s *ScoreRing) getColorForScore() color.Color {
	score := int(s.displayScore)
	switch {
	case score >= 80:
		return color.RGBA{R: 50, G: 205, B: 50, A: 255} // Green
	case score >= 61:
		return color.RGBA{R: 255, G: 215, B: 0, A: 255} // Yellow
	case score >= 31:
		return color.RGBA{R: 255, G: 140, B: 0, A: 255} // Orange
	default:
		return color.RGBA{R: 220, G: 30, B: 30, A: 255} // Red
	}
}

type scoreRingRenderer struct {
	ring  *canvas.Circle
	text  *canvas.Text
	score *ScoreRing
}

func (r *scoreRingRenderer) Layout(size fyne.Size) {
	r.ring.Resize(size)
	r.ring.Move(fyne.NewPos(0, 0))

	textSize := r.text.MinSize()
	r.text.Move(fyne.NewPos(
		(size.Width-textSize.Width)/2,
		(size.Height-textSize.Height)/2,
	))
	r.text.Resize(textSize)
}

func (r *scoreRingRenderer) MinSize() fyne.Size {
	return fyne.NewSize(200, 200)
}

func (r *scoreRingRenderer) Refresh() {
	// Snapshot fields under lock before using them in render calls
	r.score.mu.Lock()
	ds := r.score.displayScore
	scoreColor := r.score.getColorForScore()
	r.score.mu.Unlock()

	// Update text
	if ds > 0 {
		r.text.Text = fmt.Sprintf("%d", int(ds))
	} else {
		r.text.Text = "--"
	}

	// Update color based on current score
	r.ring.StrokeColor = scoreColor
	r.text.Color = scoreColor

	// Update stroke to create "filling" effect
	// The ring appears to fill as the score increases
	percentage := ds / 100.0
	if percentage > 0 {
		r.ring.StrokeWidth = float32(10 + (percentage * 5)) // Thicker stroke as score increases
	}

	canvas.Refresh(r.ring)
	canvas.Refresh(r.text)
}

func (r *scoreRingRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.ring, r.text}
}

func (r *scoreRingRenderer) Destroy() {}

//go:build gui

package views

import (
	"fmt"
	"image/color"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"
)

// ScoreRing is an animated circular progress indicator for the performance score
type ScoreRing struct {
	widget.BaseWidget
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

	s.score = score
	if !s.animating {
		go s.animateTo(score)
	}
}

// GetScore returns the current target score
func (s *ScoreRing) GetScore() int {
	return s.score
}

func (s *ScoreRing) animateTo(target int) {
	s.animating = true
	defer func() { s.animating = false }()

	targetFloat := float64(target)
	step := 0.8

	for s.displayScore < targetFloat-0.5 || s.displayScore > targetFloat+0.5 {
		if s.displayScore < targetFloat {
			s.displayScore += step
			if s.displayScore > targetFloat {
				s.displayScore = targetFloat
			}
		} else {
			s.displayScore -= step
			if s.displayScore < targetFloat {
				s.displayScore = targetFloat
			}
		}

		s.Refresh()
		time.Sleep(15 * time.Millisecond)
	}

	s.displayScore = targetFloat
	s.Refresh()
}

// CreateRenderer creates the widget renderer
func (s *ScoreRing) CreateRenderer() fyne.WidgetRenderer {
	s.ring = canvas.NewCircle(s.getColorForScore())
	s.ring.StrokeWidth = 10
	s.ring.StrokeColor = s.getColorForScore()
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
	// Update text
	if r.score.displayScore > 0 {
		r.text.Text = fmt.Sprintf("%d", int(r.score.displayScore))
	} else {
		r.text.Text = "--"
	}

	// Update color based on current score
	scoreColor := r.score.getColorForScore()
	r.ring.StrokeColor = scoreColor
	r.text.Color = scoreColor

	// Update stroke to create "filling" effect
	// The ring appears to fill as the score increases
	percentage := r.score.displayScore / 100.0
	if percentage > 0 {
		r.ring.StrokeWidth = 10 + (percentage * 5) // Thicker stroke as score increases
	}

	canvas.Refresh(r.ring)
	canvas.Refresh(r.text)
}

func (r *scoreRingRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.ring, r.text}
}

func (r *scoreRingRenderer) Destroy() {}

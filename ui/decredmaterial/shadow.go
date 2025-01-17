package decredmaterial

import (
	"image"
	"image/color"
	"math"

	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
)

type Shadow struct {
	radius  float32
	surface color.NRGBA

	ambientColor  color.NRGBA
	penumbraColor color.NRGBA
	umbraColor    color.NRGBA
}

const (
	shadowElevation = 3
)

func (t *Theme) Shadow() *Shadow {
	return &Shadow{
		radius:        4,
		surface:       t.Color.Surface,
		ambientColor:  color.NRGBA{A: 0x1},
		penumbraColor: color.NRGBA{A: 0x5},
		umbraColor:    color.NRGBA{A: 0x10},
	}
}

func (t *Theme) TransparentShadow(radius float32) *Shadow {
	s := t.Shadow()
	s.surface = color.NRGBA{}
	s.radius = radius
	return s
}

func (s *Shadow) Layout(gtx C, w layout.Widget) D {
	return layout.Stack{}.Layout(gtx,
		layout.Expanded(func(gtx C) D {
			s.layout(gtx)
			surface := clip.UniformRRect(f32.Rectangle{Max: layout.FPt(gtx.Constraints.Min)}, s.radius)
			paint.FillShape(gtx.Ops, s.surface, surface.Op(gtx.Ops))
			return D{Size: gtx.Constraints.Min}
		}),
		layout.Stacked(w),
	)
}

func (s *Shadow) layout(gtx C) D {
	sz := gtx.Constraints.Min
	rr := float32(gtx.Px(unit.Dp(shadowElevation)))

	r := f32.Rect(0, 0, float32(sz.X), float32(sz.Y))
	s.layoutShadow(gtx, r, rr)

	return layout.Dimensions{Size: sz}
}

func (s *Shadow) layoutShadow(gtx layout.Context, r f32.Rectangle, rr float32) {
	offset := pxf(gtx.Metric, unit.Dp(shadowElevation))

	ambient := r
	gradientBox(gtx.Ops, ambient, rr, offset/2, s.ambientColor)

	penumbra := r.Add(f32.Pt(0, offset/2))
	gradientBox(gtx.Ops, penumbra, rr, offset, s.penumbraColor)

	umbra := outset(penumbra, -offset/2)
	gradientBox(gtx.Ops, umbra, rr/4, offset/2, s.umbraColor)
}

func gradientBox(ops *op.Ops, r f32.Rectangle, rr, spread float32, col color.NRGBA) {
	paint.FillShape(ops, col, clip.RRect{
		Rect: outset(r, spread),
		SE:   rr + spread, SW: rr + spread, NW: rr + spread, NE: rr + spread,
	}.Op(ops))
}

func round(r f32.Rectangle) f32.Rectangle {
	return f32.Rectangle{
		Min: f32.Point{
			X: float32(math.Round(float64(r.Min.X))),
			Y: float32(math.Round(float64(r.Min.Y))),
		},
		Max: f32.Point{
			X: float32(math.Round(float64(r.Max.X))),
			Y: float32(math.Round(float64(r.Max.Y))),
		},
	}
}

func outset(r f32.Rectangle, rr float32) f32.Rectangle {
	r.Min.X -= rr
	r.Min.Y -= rr
	r.Max.X += rr
	r.Max.Y += rr
	return r
}

func pxf(c unit.Metric, v unit.Value) float32 {
	switch v.U {
	case unit.UnitPx:
		return v.V
	case unit.UnitDp:
		s := c.PxPerDp
		if s == 0 {
			s = 1
		}
		return s * v.V
	case unit.UnitSp:
		s := c.PxPerSp
		if s == 0 {
			s = 1
		}
		return s * v.V
	default:
		panic("unknown unit")
	}
}

func topLeft(r image.Rectangle) image.Point     { return r.Min }
func topRight(r image.Rectangle) image.Point    { return image.Point{X: r.Max.X, Y: r.Min.Y} }
func bottomRight(r image.Rectangle) image.Point { return r.Max }
func bottomLeft(r image.Rectangle) image.Point  { return image.Point{X: r.Min.X, Y: r.Max.Y} }

package canvas

import (
	"exp/draw"
	"image"
	//	"fmt"
	"math"
	"freetype-go.googlecode.com/hg/freetype/raster"
)

// Box creates a rectangular image of the given size, filled with the given colour,
// with a border-size border of colour borderCol.
//
func Box(width, height int, col image.Image, border int, borderCol image.Image) image.Image {
	img := image.NewRGBA(width, height)
	if border < 0 {
		border = 0
	}
	r := draw.Rect(0, 0, width, height)
	draw.Draw(img, r.Inset(border), col, draw.ZP)
	draw.Border(img, r, border, borderCol, draw.ZP)
	return img
}

// An ImageItem is an Item that uses an image
// to draw itself. It is intended to be used as a building
// block for other Items.
type ImageItem struct {
	r      draw.Rectangle
	img    image.Image
	opaque bool
}

func (obj *ImageItem) Draw(dst *image.RGBA, clip draw.Rectangle) {
	dr := obj.r.Clip(clip)
	sp := dr.Min.Sub(obj.r.Min)
	op := draw.Over
	if obj.opaque {
		op = draw.Src
	}
	draw.DrawMask(dst, dr, obj.img, sp, nil, draw.ZP, op)
}

func (obj *ImageItem) SetContainer(c *Canvas) {
}

func (obj *ImageItem) Opaque() bool {
	return obj.opaque
}

func (obj *ImageItem) Bbox() draw.Rectangle {
	return obj.r
}

func (obj *ImageItem) HitTest(p draw.Point) bool {
	return p.In(obj.r)
}

// An Image represents an rectangular (but possibly
// transparent) image.
//
type Image struct {
	Item
	item   ImageItem // access to the fields of the ImageItem
	canvas *Canvas
}

// Image returns a new Image which will be drawn using img,
// with p giving the coordinate of the image's top left corner.
//
func NewImage(img image.Image, opaque bool, p draw.Point) *Image {
	obj := new(Image)
	obj.Item = &obj.item
	obj.item.r = draw.Rectangle{p, p.Add(draw.Pt(img.Width(), img.Height()))}
	obj.item.img = img
	obj.item.opaque = opaque
	return obj
}

func (obj *Image) SetContainer(c *Canvas) {
	obj.canvas = c
}

// SetMinPoint moves the image's upper left corner to p.
//
func (obj *Image) SetMinPoint(p draw.Point) {
	if p.Eq(obj.item.r.Min) {
		return
	}
	obj.canvas.Atomically(func(flush FlushFunc) {
		r := obj.item.r
		obj.item.r = r.Add(p.Sub(r.Min))
		flush(r, false, &obj.item)
		flush(obj.item.r, false, &obj.item)
	})
}

func (obj *Image) Delete() {
	obj.canvas.Delete(&obj.item)
}

// A Polygon represents a filled polygon.
//
type Polygon struct {
	Item
	raster RasterItem
	canvas *Canvas
	points []raster.Point
}

// Polygon returns a new PolyObject, using col for its fill colour, and
// using points for its vertices.
//
func NewPolygon(col image.Color, points []draw.Point) *Polygon {
	obj := new(Polygon)
	rpoints := make([]raster.Point, len(points))
	for i, p := range points {
		rpoints[i] = pixel2fixPoint(p)
	}
	obj.raster.SetColor(col)
	obj.points = rpoints
	obj.Item = &obj.raster
	return obj
}

func (obj *Polygon) SetContainer(c *Canvas) {
	obj.canvas = c
	if c != nil {
		obj.raster.SetBounds(c.Width(), c.Height())
		obj.rasterize()
	}
}

func (obj *Polygon) Delete() {
	obj.canvas.Delete(obj)
}

func (obj *Polygon) Move(delta draw.Point) {
	obj.canvas.Atomically(func(flush FlushFunc) {
		r := obj.raster.Bbox()
		rdelta := pixel2fixPoint(delta)
		for i := range obj.points {
			p := &obj.points[i]
			p.X += rdelta.X
			p.Y += rdelta.Y
		}
		obj.rasterize()
		flush(r, false, &obj.raster)
		flush(obj.raster.Bbox(), false, &obj.raster)
	})
}

func (obj *Polygon) rasterize() {
	obj.raster.Clear()
	if len(obj.points) > 0 {
		obj.raster.Start(obj.points[0])
		for _, p := range obj.points[1:] {
			obj.raster.Add1(p)
		}
		obj.raster.Add1(obj.points[0])
	}
	obj.raster.CalcBbox()
}

// A line object represents a single straight line.
type Line struct {
	Item
	raster RasterItem
	canvas *Canvas
	p0, p1 raster.Point
	width  raster.Fixed
}

// Line returns a new Line, coloured with col, from p0 to p1,
// of the given width.
//
func NewLine(col image.Color, p0, p1 draw.Point, width float) *Line {
	obj := new(Line)
	obj.p0 = pixel2fixPoint(p0)
	obj.p1 = pixel2fixPoint(p1)
	obj.width = float2fix(width)
	obj.raster.SetColor(col)
	obj.Item = &obj.raster
	obj.rasterize()
	return obj
}

func (obj *Line) SetContainer(c *Canvas) {
	obj.canvas = c
	if c != nil {
		obj.raster.SetBounds(c.Width(), c.Height())
		obj.rasterize()
	}
}

func (obj *Line) rasterize() {
	obj.raster.Clear()
	sin, cos := isincos2(obj.p1.X-obj.p0.X, obj.p1.Y-obj.p0.Y)
	dx := (cos * obj.width) / (2 * fixScale)
	dy := (sin * obj.width) / (2 * fixScale)
	q := raster.Point{
		obj.p0.X + fixScale/2 - sin/2,
		obj.p0.Y + fixScale/2 - cos/2,
	}
	p0 := raster.Point{q.X - dx, q.Y + dy}
	obj.raster.Start(p0)
	obj.raster.Add1(raster.Point{q.X + dx, q.Y - dy})

	q = raster.Point{
		obj.p1.X + fixScale/2 + sin/2,
		obj.p1.Y + fixScale/2 + cos/2,
	}
	obj.raster.Add1(raster.Point{q.X + dx, q.Y - dy})
	obj.raster.Add1(raster.Point{q.X - dx, q.Y + dy})
	obj.raster.Add1(p0)
	obj.raster.CalcBbox()
}

func (obj *Line) Move(delta draw.Point) {
	p0 := fix2pixelPoint(obj.p0)
	p1 := fix2pixelPoint(obj.p1)
	obj.SetEndPoints(p0.Add(delta), p1.Add(delta))
}

// SetEndPoints changes the end coordinates of the Line.
//
func (obj *Line) SetEndPoints(p0, p1 draw.Point) {
	obj.canvas.Atomically(func(flush FlushFunc) {
		r := obj.raster.Bbox()
		obj.p0 = pixel2fixPoint(p0)
		obj.p1 = pixel2fixPoint(p1)
		obj.rasterize()
		flush(r, false, &obj.raster)
		flush(obj.raster.Bbox(), false, &obj.raster)
	})
}

func (obj *Line) Delete() {
	obj.canvas.Delete(&obj.raster)
}

// SetColor changes the colour of the line
//
func (obj *Line) SetColor(col image.Color) {
	obj.canvas.Atomically(func(flush FlushFunc) {
		obj.raster.SetColor(col)
		flush(obj.raster.Bbox(), false, &obj.raster)
	})
}


const (
	fixBits  = 8
	fixScale = 1 << fixBits // matches raster.Fixed
)

func float2fix(f float) raster.Fixed {
	return raster.Fixed(f*fixScale + 0.5)
}

func int2fix(i int) raster.Fixed {
	return raster.Fixed(i << fixBits)
}

func fix2int(i raster.Fixed) int {
	return int((i + fixScale/2) >> fixBits)
}

func pixel2fixPoint(p draw.Point) raster.Point {
	return raster.Point{raster.Fixed(p.X << fixBits), raster.Fixed(p.Y << fixBits)}
}

func fix2pixelPoint(p raster.Point) draw.Point {
	return draw.Point{int((p.X + fixScale/2) >> fixBits), int((p.Y + fixScale/2) >> fixBits)}
}

// could do it in fixed point, but what's 0.5us between friends?
func isincos2(x, y raster.Fixed) (isin, icos raster.Fixed) {
	sin, cos := math.Sincos(math.Atan2(fixed2float(x), fixed2float(y)))
	isin = float2fixed(sin)
	icos = float2fixed(cos)
	return
}

func float2fixed(f float64) raster.Fixed {
	if f < 0 {
		return raster.Fixed(f*256 + 0.5)
	}
	return raster.Fixed(f*256 - 0.5)
}

func fixed2float(f raster.Fixed) float64 {
	return float64(f) / 256
}
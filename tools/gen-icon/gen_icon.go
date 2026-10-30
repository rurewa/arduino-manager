// Генератор иконки для Arduino-Manager.
//
// Запуск:
//   go run gen_icon.go
//
// Результат: Icon.png (256x256, RGBA) рядом с исполняемым файлом.
package main

import (
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
)

const (
	finalSize = 256 // итоговый размер
	ss        = 4   // суперсэмплинг (рисуем в 4x и даунскейлим)
	drawSize  = finalSize * ss
)

func main() {
	img := image.NewRGBA(image.Rect(0, 0, drawSize, drawSize))

	teal := color.RGBA{R: 0x00, G: 0x97, B: 0x9D, A: 0xFF} // Arduino teal
	white := color.RGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF}

	// 1. Скруглённый квадрат-фон
	fillRoundedRect(img, 0, 0, 256, 256, 56, teal)

	// 2. Ножки чипа (рисуем ДО корпуса, чтобы перекрытие спряталось под ним)
	pinWid := 14.0
	pinLen := 22.0
	overlap := 12.0
	for _, c := range []float64{90, 128, 166} {
		fillRect(img, c-pinWid/2, 68-pinLen, pinWid, pinLen+overlap, white)       // верх
		fillRect(img, c-pinWid/2, 188-overlap, pinWid, pinLen+overlap, white)     // низ
		fillRect(img, 68-pinLen, c-pinWid/2, pinLen+overlap, pinWid, white)       // лево
		fillRect(img, 188-overlap, c-pinWid/2, pinLen+overlap, pinWid, white)     // право
	}

	// 3. Корпус чипа
	fillRoundedRect(img, 68, 68, 120, 120, 20, white)

	// 4. Внутренняя teal-область
	fillRoundedRect(img, 88, 88, 80, 80, 10, teal)

	// 5. Белая стрелка «вниз» (символ установки)
	fillRect(img, 118, 98, 20, 34, white) // древко
	fillTriangle(img,
		[2]float64{102, 126}, // левый угол основания
		[2]float64{154, 126}, // правый угол основания
		[2]float64{128, 158}, // вершина
		white,
	)

	// Даунскейл
	final := downscale(img, finalSize)

	f, err := os.Create("Icon.png")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	if err := png.Encode(f, final); err != nil {
		panic(err)
	}
	println("Wrote Icon.png (256x256)")
}

// --- Примитивы (все координаты — в «итоговом» 256-пространстве) ---

func fillRect(img *image.RGBA, x, y, w, h float64, c color.RGBA) {
	x0 := int(x * ss)
	y0 := int(y * ss)
	x1 := int((x + w) * ss)
	y1 := int((y + h) * ss)
	b := img.Bounds()
	for py := y0; py < y1; py++ {
		for px := x0; px < x1; px++ {
			if px >= b.Min.X && py >= b.Min.Y && px < b.Max.X && py < b.Max.Y {
				img.SetRGBA(px, py, c)
			}
		}
	}
}

func fillRoundedRect(img *image.RGBA, x, y, w, h, r float64, c color.RGBA) {
	// Центральные полосы
	fillRect(img, x+r, y, w-2*r, h, c)
	fillRect(img, x, y+r, w, h-2*r, c)
	// Четыре угла-круга
	for _, cc := range [][2]float64{
		{x + r, y + r},
		{x + w - r, y + r},
		{x + r, y + h - r},
		{x + w - r, y + h - r},
	} {
		fillCircle(img, cc[0], cc[1], r, c)
	}
}

func fillCircle(img *image.RGBA, cx, cy, r float64, c color.RGBA) {
	cx *= ss
	cy *= ss
	r *= ss
	r2 := r * r
	x0 := int(cx - r)
	y0 := int(cy - r)
	x1 := int(cx+r) + 1
	y1 := int(cy+r) + 1
	b := img.Bounds()
	for py := y0; py < y1; py++ {
		for px := x0; px < x1; px++ {
			dx := float64(px) + 0.5 - cx
			dy := float64(py) + 0.5 - cy
			if dx*dx+dy*dy <= r2 &&
				px >= b.Min.X && py >= b.Min.Y && px < b.Max.X && py < b.Max.Y {
				img.SetRGBA(px, py, c)
			}
		}
	}
}

func fillTriangle(img *image.RGBA, a, b, c [2]float64, col color.RGBA) {
	a[0] *= ss; a[1] *= ss
	b[0] *= ss; b[1] *= ss
	c[0] *= ss; c[1] *= ss
	x0 := int(math.Min(a[0], math.Min(b[0], c[0])))
	x1 := int(math.Max(a[0], math.Max(b[0], c[0]))) + 1
	y0 := int(math.Min(a[1], math.Min(b[1], c[1])))
	y1 := int(math.Max(a[1], math.Max(b[1], c[1]))) + 1
	bounds := img.Bounds()
	for py := y0; py < y1; py++ {
		for px := x0; px < x1; px++ {
			if pointInTriangle(float64(px)+0.5, float64(py)+0.5, a, b, c) &&
				px >= bounds.Min.X && py >= bounds.Min.Y &&
				px < bounds.Max.X && py < bounds.Max.Y {
				img.SetRGBA(px, py, col)
			}
		}
	}
}

func pointInTriangle(px, py float64, a, b, c [2]float64) bool {
	d1 := (px-b[0])*(a[1]-b[1]) - (a[0]-b[0])*(py-b[1])
	d2 := (px-c[0])*(b[1]-c[1]) - (b[0]-c[0])*(py-c[1])
	d3 := (px-a[0])*(c[1]-a[1]) - (c[0]-a[0])*(py-a[1])
	hasNeg := d1 < 0 || d2 < 0 || d3 < 0
	hasPos := d1 > 0 || d2 > 0 || d3 > 0
	return !(hasNeg && hasPos)
}

// downscale — усреднение factor×factor с корректным un-premultiply.
func downscale(src *image.RGBA, n int) *image.RGBA {
	dst := image.NewRGBA(image.Rect(0, 0, n, n))
	factor := src.Bounds().Dx() / n
	cnt := uint64(factor * factor)
	for y := 0; y < n; y++ {
		for x := 0; x < n; x++ {
			var rP, gP, bP, aP uint64
			for dy := 0; dy < factor; dy++ {
				for dx := 0; dx < factor; dx++ {
					r0, g0, b0, a0 := src.At(x*factor+dx, y*factor+dy).RGBA()
					rP += uint64(r0)
					gP += uint64(g0)
					bP += uint64(b0)
					aP += uint64(a0)
				}
			}
			rP /= cnt
			gP /= cnt
			bP /= cnt
			aP /= cnt
			var rU, gU, bU uint64
			if aP > 0 {
				rU = rP * 0xFFFF / aP
				gU = gP * 0xFFFF / aP
				bU = bP * 0xFFFF / aP
			}
			if rU > 0xFFFF { rU = 0xFFFF }
			if gU > 0xFFFF { gU = 0xFFFF }
			if bU > 0xFFFF { bU = 0xFFFF }
			dst.SetRGBA(x, y, color.RGBA{
				R: uint8(rU >> 8),
				G: uint8(gU >> 8),
				B: uint8(bU >> 8),
				A: uint8(aP >> 8),
			})
		}
	}
	return dst
}
package main

import (
	"fmt"
	"image"
	"image/color"
	"time"

	"github.com/llgcode/draw2d"
	"github.com/llgcode/draw2d/draw2dimg"
)

func savePrStatImage(reviewers []ReviewerStat, prCount int, fileName string) error {
	if verbose {
		fmt.Println("Generating image ", fileName)
	}

	nReviewers := len(reviewers)
	w := 1000.0

	// dedicate pixel for header, then per row and then footer
	h := 50.0 + 20.0*float64(nReviewers) + 20.0
	// Initialize the graphic context on an RGBA image
	dest := image.NewRGBA(image.Rect(0, 0, int(w), int(h)))
	gc := draw2dimg.NewGraphicContext(dest)

	// Font stuff setup
	draw2d.SetFontFolder(".")

	// Draw the border and title/header
	gc.SetFontSize(14)
	drawRect(gc, 0, 0, w, h, image.Black, image.White)
	str := fmt.Sprintf("Reviewer Stats for %v pull requests", prCount)
	gc.SetFillColor(color.Black)
	gc.FillStringAt(str, 10, 20)

	y := 60.0
	rightX := 300.0 // Right aligning all text to be here
	gap := 10.0     // gap between name and bar chart
	maxNameLen := 30

	maxBarWidth := (w - 10) - (rightX + gap)

	for _, reviewer := range reviewers {
		// trim or use the name if it fits
		str := reviewer.Name
		if len(str) > maxNameLen {
			str = str[:maxNameLen]
		}
		// find the width of the name and right align and print
		gc.SetFontSize(12)
		l, _, r, _ := gc.GetStringBounds(str)
		strW := r - l
		strH := 15.0

		textColor := color.RGBA{50, 50, 50, 0xff}
		x := (rightX - strW)
		gc.SetFillColor(textColor)
		gc.FillStringAt(str, x, y)

		// Draw the bar
		width := (maxBarWidth / float64(prCount)) * float64(reviewer.Count)
		x = rightX + gap
		barFillCol := color.RGBA{100, 100, 100, 0xff}
		drawRect(gc, x, y-strH, width, strH, barFillCol, barFillCol)
		drawRect(gc, x, y-strH, maxBarWidth, strH, color.Black, color.Transparent)

		val := fmt.Sprintf("%v", reviewer.Count)
		gc.SetFillColor(textColor)
		gc.SetFontSize(10)
		gc.FillStringAt(val, x+width+10.0, y-2)

		y += 20
	}

	// Write the footer justified to right margin
	str = fmt.Sprintf("Generated on %v", time.Now().Format("01-02-2006 15:04:05"))
	gc.SetFontSize(8)
	gc.SetFillColor(color.Black)
	l, _, r, _ := gc.GetStringBounds(str)
	sw := r - l // width of the text
	gc.FillStringAt(str, w-sw-10, h-10)

	err := draw2dimg.SaveToPngFile(fileName, dest)
	if err != nil {
		return err
	}

	if verbose {
		fmt.Println("Generated", fileName)
	}

	return nil
}

func drawRect(gc *draw2dimg.GraphicContext, x, y, w, h float64, lineColor, fillColor color.Color) {
	gc.SetStrokeColor(lineColor)
	gc.SetFillColor(fillColor)
	gc.SetLineWidth(2)
	gc.MoveTo(x, y)
	gc.LineTo(x+w, y)
	gc.LineTo(x+w, y+h)
	gc.LineTo(x, y+h)
	gc.LineTo(x, y)
	gc.Close()
	gc.FillStroke()
}

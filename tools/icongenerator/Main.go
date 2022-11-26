package main

import (
	"bytes"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"text/template"
)

// Copyright (c) Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//

/**
 * We represent charge by a circle segment (cf https://en.wikipedia.org/wiki/Circular_segment, though
 * we have the segment at the bottom of the circle.)
 *
 * The segment is constructed so that:
 *
 *     (area of segment)/(area of the circle) ~ charge percentage.
 *
 * This array (indexed from 0 to 100 inclusive) contains the calculated angles, so if you want a segment
 * that covers (say) 37% of the circle (ie. represents a battery charged 37%), you use the angle at
 * position 37 in the array.
 */


type TemplateData struct {
	Stroke      string
	CircleFill  string
	UseMarker   bool
	StrokeWidth int
	StartX      int
	EndX        int
	StartY      int
	BigArchflag int
}

const twoPi = 2*math.Pi

// cf. https://en.wikipedia.org/wiki/Circular_segment 
// We use the notation defined on that wiki page, except we place the segment at the 
// bottom of the circle rather than the top.
// When segment covers an angle Ψ ( Ψ ∈ [0,2π]), it starts at an angle 
// 3π/2 - Ψ/2 +  and runs counterclockwise through  3π/2 + Ψ/2 
// 
// The area of the segment is (Ψ - sin(Ψ))/2 
// The starting point of the segment is  cos(3π/2 - Ψ/2), sin(3π/2 - Ψ/2) = -sin(Ψ/2), -cos(Ψ/2) 
// The ending point is                   cos(3π/2 + Ψ/2), sin(3π/2 + Ψ/2) =  sin(Ψ/2), -cos(Ψ/2) 
//
func calculateAngel(percentage int) float64 {
	if percentage <= 0 {
		return 0
	} else if percentage >= 100 {
		return 2*math.Pi 
	} else {
		var minPsi float64 = 0
		var maxPsi float64 = twoPi
		for maxPsi - minPsi > 0.00001 {
			var tmpPsi = (maxPsi + minPsi)/2 
			var tmpArea = (tmpPsi - math.Sin(tmpPsi))/twoPi
		    var tmpPercentage =  int(100*tmpArea)
			if tmpPercentage > percentage {
				maxPsi = tmpPsi
			} else {
				minPsi = tmpPsi
			}
		}
		return maxPsi
	}
}


func buildSvg(percentage int, charging bool) string {
	var svgTemplate = `
		<svg viewBox='0 0 100 100'  xmlns='http://www.w3.org/2000/svg'>
			<g  stroke='{{.Stroke}}' stroke-width='{{.StrokeWidth}}'> 
				{{if .UseMarker}}    
				<path d='M {{.StartX}} {{.StartY}}  A 40 40 0 {{.BigArchflag}} 0 {{.EndX}} {{.StartY}}' fill='darkgray' stroke='none' />
				{{end}} 
            	<circle cx="50" cy="50" r="40" fill="{{.CircleFill}}"/>
			</g>
		</svg>`

	var angle = calculateAngel(percentage)
	const segmentRadius float64 = 40
	// cf. above
	var startX = -math.Sin(angle/2) 
	var startY = -math.Cos(angle/2)

	var bigArchFlag = 0
	if angle > math.Pi {
		bigArchFlag = 1
	}

	var stroke = "black"
	if percentage < 15 && !charging {
		stroke = "red"
	}

	var strokeWidth = 6
	if charging {
		strokeWidth = 12
	}


	var templateData = TemplateData{
		Stroke:      stroke,
		CircleFill:  "none",
		UseMarker:   true,
		StrokeWidth: strokeWidth,
		StartX:      50 + int(math.Round(40*startX)),
		EndX:        50 - int(math.Round(40*startX)),
		StartY:      50 - int(math.Round(40*startY)),
		BigArchflag: bigArchFlag,
	}

	if templateData.StartX >= 50 && percentage > 50 {
		templateData.UseMarker = false
		templateData.CircleFill = "darkgray"
	}

	if t, err := template.New("").Parse(strings.TrimSpace(svgTemplate)); err != nil {
		panic(err)
	} else {
		var collector = bytes.Buffer{}
		t.ExecuteTemplate(&collector, "", templateData)
		return string(collector.Bytes())
	}
}

const usage = `
Usage:

    batteryicon <percentage> <charging>

will produce a battery icon. percentage must be an integer in the range 0..100, and charging must be 'true' or 'false'

	batteryicon --help 

will produce this text`

func main() {
	var errMessage = ""
	if len(os.Args) == 2 && "--help" == os.Args[1] {
		fmt.Println(usage)
	} else if len(os.Args) == 3 {
		if percentage, err := strconv.Atoi(os.Args[1]); err != nil {
			errMessage = err.Error()
		} else if percentage < 0 || percentage > 100 {
			errMessage = "percentage must be in range 0..100"
		} else if os.Args[2] != "true" && os.Args[2] != "false" {
			errMessage = "charging must be 'true' or 'false'"
		} else {
			charging := os.Args[2] == "true"
			fmt.Print(buildSvg(percentage, charging))
		}
	} else {
		errMessage = "Wrong number of arguments"
	}

	if errMessage != "" {
		fmt.Fprintln(os.Stderr, errMessage)
		fmt.Fprintln(os.Stderr, usage)
		os.Exit(1)
	}

}

/*export let battery = (displayDevice) => {
    let { Percentage, State } = displayDevice ? displayDevice : { Percentage: -1, State: "Unknown" }
    if (State === "Unknown") {
        Percentage = 0
    }
    let angle = Percentage >= 0 && Percentage <= 100 ? angles[Math.round(Percentage)] : 0;
    const segmentRadius = 46;
    // We place the segment at the bottom of the circle, from (3PI/2 - angle/2) to (3PI/2 + angle/2)
    let startX = segmentRadius * Math.cos(3 * Math.PI / 2 - angle / 2);
    let startY = -segmentRadius * Math.sin(3 * Math.PI / 2 - angle / 2);
    let bigArchFlag = angle > Math.PI ? 1 : 0;

    let strokeColor = State === "Unknown" || (State !== "Charging" && Percentage < 10) ? "red" : "black";
    let fillColor = State === "Unknown" || (State !== "Charging" && Percentage < 10) ? "red" : "darkgray";
    let strokeWidth = strokeColor === "red" || State === "Charging" || State === "Fully charged" ? 14 : 6;
    const circleRadius = segmentRadius + strokeWidth/2
    let alertText = State === "Unknown" ? '?' : (State !== "Charging" && Percentage < 10) ? '!' : "";
    let title = State + " " + Math.round(Percentage) + "%";

    let marker = Math.round(Percentage) === 100 ?
        circle({cx:0, cy:0, r:46, fill:fillColor, stroke:"none"}) :
        path({d:`M ${startX} ${startY} A ${segmentRadius} ${segmentRadius} 0 ${bigArchFlag} 0 ${-startX} ${startY}`, fill:fillColor, stroke:"none"});

    return div({title: title},
        svg({xmlns:"http://www.w3.org/2000/svg", width:"16px", height:"16px", viewBox:"-60 -65 120 125"},
            circle({cx:"0", cy:"0", r:circleRadius, fill:"white", stroke:strokeColor, strokeWidth:strokeWidth}),
            marker,
            text({textAnchor:"middle", x:"0", y:"30", style:{ fontWeight: "bold", fontSize: "90px", fill: "red" }}, alertText)
        )
    )

}*/

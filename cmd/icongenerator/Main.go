package main

// Copyright (c) Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//

import (
	"bytes"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"text/template"
)

const twoPi = 2*math.Pi


/**
 * We represent charge by a circle segment (cf https://en.wikipedia.org/wiki/Circular_segment, though
 * we have the segment at the bottom of the circle rather than the top.
 *
 * The segment is constructed so that:
 *
 *     (area of segment)/(area of the circle) ~ charge percentage.
 *
 * When segment covers an angle Ψ ( Ψ ∈ [0,2π]), it starts at an angle 
 *
 * 	   3π/2 - Ψ/2 
 *
 * and runs counterclockwise to  
 *
 *      3π/2 + Ψ/2 
 * 
 * The area of the segment is 
 * 
 *      (Ψ - sin(Ψ))/2 
 * 
 * The starting point of the segment is  
 *
 *      cos(3π/2 - Ψ/2), sin(3π/2 - Ψ/2) = -sin(Ψ/2), -cos(Ψ/2) 
 *
 * The ending point is                   
 *
 *      cos(3π/2 + Ψ/2), sin(3π/2 + Ψ/2) =  sin(Ψ/2), -cos(Ψ/2) 
 *
 */
func calculateAngel(percentage int) float64 {
	// We find the angle by binary search
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
	    <svg viewBox="-100 -100 200 200"  xmlns="http://www.w3.org/2000/svg">
			<g stroke="{{.stroke}}"> 
				<circle r="80" fill="white" stroke="none"/>
				{{if .full}}    
	            	<circle r="80" fill="darkgray" stroke="none"/>
				{{ else }}
					<path d="M {{.startX}} {{.startY}}  A 80 80 0 {{.bigarchflag}} 0 {{.endX}} {{.startY}}" fill="darkgray" stroke="none"/>
				{{end}} 
				{{if .charging}}
					<circle r="80" stroke-width="20" fill="none"/>
					<circle r="94" stroke="white" stroke-width="8" fill="none"/>
				{{else}}	
					<circle r="80" stroke-width="10" fill="none"/>
					<circle r="89" stroke="white" stroke-width="8" fill="none"/>
				{{end}}	
			</g>
		</svg>`

	var angle = calculateAngel(percentage)
	var startX = -math.Sin(angle/2) 
	var startY = -math.Cos(angle/2)

	var data = map[string]any{
		"startX": int(math.Round(80*startX)),
		"endX": -int(math.Round(80*startX)),
		"startY": -int(math.Round(80*startY)),
		"full": percentage > 98,
		"charging": charging,
		"stroke": "black",
		"bigarchflag": 0,
	}
	
	if percentage < 15 && !charging {
		data["stroke"] = "red"
	}
	
	if angle > math.Pi {
		data["bigarchflag"] = 1
	}
	
	if t, err := template.New("").Parse(strings.TrimSpace(svgTemplate)); err != nil {
		panic(err)
	} else {
		var collector = bytes.Buffer{}
		t.ExecuteTemplate(&collector, "", data)
		return string(collector.Bytes())
	}
}

func main() {
	var errMessage = ""
	if len(os.Args) == 3 {
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
		fmt.Fprintln(os.Stderr, "Usage: icongenerator <percentage> <charging>", errMessage)
		os.Exit(1)
	}

}



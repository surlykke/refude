package main

import (
	"fmt"
	"math"
)

const precision = 0.01

func angle2area(angle float64) float64 {
	return (angle - math.Sin(angle))/2
}

func area2angle(area float64) float64 {
	var min = float64(0)
	var max  = 2*math.Pi;

	for {
		var minval = angle2area(min)
		var maxval = angle2area(max)

		if minval > area || maxval < area {
			panic(fmt.Sprintf("minval: %f, maxval: %f, area: %f", minval, maxval, area))
		}

		var midpoint = (min + max)/2
		if (minval > maxval - precision) {
			return midpoint
		}
		var midval = angle2area(midpoint)

		if midval > area {
			max = midpoint
		} else {
			min = midpoint
		}
	}

}

func main() {
	var angles = make([]float64, 101)
	angles[0] = 0
	for i := 1;  i < 100 ; i++ {
		var area = math.Pi*float64(i)/100.0
		angles[i] = area2angle(area)
	}
	angles[100] = 2*math.Pi
	fmt.Println(angles)
}

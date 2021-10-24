// Copyright (c) Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
/*
 * Copyright (c) Christian Surlykke
 *
 * This file is part of the refude project.
 * It is distributed under the GPL v2 license.
 * Please refer to the LICENSE file for a copy of the license.
 */

import { div } from "./elements.js";


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
const angles = [0, 0.7117670855789375, 0.9081166264282996, 1.0676506283684062, 1.1658253987930873, 1.2640001692177685, 1.3621749396424492, 1.4296700943094176, 1.5033011721279281, 1.564660403643354, 1.6260196351587797, 1.6873788666742053, 1.7364662518865457, 1.7978254834019716, 1.846912868614312, 1.8960002538266523, 1.932815792735908, 1.9819031779482486, 2.0309905631605893, 2.067806102069844, 2.1168934872821845, 2.1537090261914402, 2.190524565100695, 2.239611950313036, 2.2733595276465204, 2.3101750665557756, 2.346990605465031, 2.383806144374286, 2.4206216832835414, 2.457437222192797, 2.4881168379505105, 2.524932376859766, 2.561747915769021, 2.5924275315267336, 2.629243070435989, 2.6599226861937018, 2.696738225102957, 2.72741784086067, 2.758097456618383, 2.7949129955276386, 2.825592611285351, 2.856272227043064, 2.8869518428007765, 2.923767381710032, 2.9544469974677448, 2.9851266132254572, 3.01580622898317, 3.046485844740883, 3.0771654604985965, 3.107845076256309, 3.1446606151655643, 3.175340230923277, 3.2060198466809897, 3.236699462438703, 3.267379078196416, 3.298058693954129, 3.3287383097118415, 3.3594179254695544, 3.3962334643788097, 3.426913080136522, 3.457592695894235, 3.4882723116519476, 3.5250878505612033, 3.5557674663189163, 3.586447082076629, 3.6232626209858845, 3.6539422367435974, 3.6907577756528527, 3.721437391410565, 3.7582529303198204, 3.7950684692290757, 3.825748084986789, 3.862563623896045, 3.8993791628053, 3.9361947017145553, 3.973010240623811, 4.009825779533067, 4.043573356866551, 4.092660742078891, 4.129476280988146, 4.166291819897402, 4.215379205109742, 4.252194744018997, 4.301282129231337, 4.350369514443678, 4.387185053352933, 4.436272438565274, 4.485359823777614, 4.546719055293041, 4.595806440505381, 4.657165672020807, 4.718524903536232, 4.779884135051658, 4.853515212870168, 4.9210103675371375, 5.019185137961818, 5.117359908386499, 5.21553467881118, 5.375068680751286, 5.571418221600649, 6.283185307179586];

export const svg = (props, ...children) => React.createElement('svg', props, ...children)
export const circle = (props, ...children) => React.createElement('circle', props, ...children)
export const path = (props, ...children) => React.createElement('path', props, ...children)
export const text = (props, ...children) => React.createElement('text', props, ...children)


export let battery = (displayDevice) => {
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
        svg({xmlns:"http://www.w3.org/2000/svg", width:"20px", height:"20px", viewBox:"-60 -65 120 125"},
            circle({cx:"0", cy:"0", r:circleRadius, fill:"white", stroke:strokeColor, strokeWidth:strokeWidth}),
            marker,
            text({textAnchor:"middle", x:"0", y:"30", style:{ fontWeight: "bold", fontSize: "90px", fill: "red" }}, alertText)
        )
    )

}



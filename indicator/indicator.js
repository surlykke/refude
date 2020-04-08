// Copyright (c) Christian Surlykke
//
// This file is part of the refude project. 
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
import React from 'react';

import {ipcRenderer } from 'electron'


export class Indicator extends React.Component {

    constructor(props) {
        super(props)
        this.state = {resource: null}
        
        ipcRenderer.on("screens", (evt, displays) => {
            this.getScreens(displays)
        })
        
        ipcRenderer.on("resource", (evt, resource) => {
            this.setState({ resource: resource })
        })
    }


    getScreens = (displays) => {
        let bx1, by1, bx2, by2;
        this.screens = [];
        displays.forEach((dsp, i) => {
            this.screens.push({
                x: Math.round(dsp.bounds.x * dsp.scaleFactor),
                y: Math.round(dsp.bounds.y * dsp.scaleFactor),
                w: Math.round(dsp.bounds.width * dsp.scaleFactor),
                h: Math.round(dsp.bounds.height * dsp.scaleFactor)
            });
            let x1 = this.screens[i].x;
            let y1 = this.screens[i].y;
            let x2 = this.screens[i].x + this.screens[i].w;
            let y2 = this.screens[i].y + this.screens[i].h;
            if (i === 0) {
                bx1 = x1;
                by1 = y1;
                bx2 = x2;
                by2 = y2;
            } else {
                bx1 = Math.min(bx1, x1);
                by1 = Math.min(by1, y1);
                bx2 = Math.max(bx2, x2);
                by2 = Math.max(by2, y2);
            }
        });
        this.display = { x: bx1, y: by1, w: bx2 - bx1, h: by2 - by1 };
    };

    render = () => {
        let res = this.state.resource
        if (res && res.Type === "window") {
            let { X, Y, W, H } = res.Data
            let screenShotUrl = "http://localhost:7938" + res.Self + "/screenshot?downscale=3"
            let viewBox = `${this.display.x - 3} ${this.display.y - 3} ${this.display.w + 6} ${this.display.h + 6}`;
            let rects = this.screens.map((scr, i) => <rect key={`screenRect_${i}`} x={scr.x} y={scr.y} width={scr.w} height={scr.h} fill="lightgrey" />);
            //rects.push(<rect key="winRect" x={window.X} y={window.Y} width={window.W} height={window.H} fill="grey" />);
            rects.push(<image key="winRect" x={X} y={Y} width={W} height={H} xlinkHref={screenShotUrl} />);
            return <svg key="windows" xmlns="http://www.w3.org/2000/svg" width="calc(100% - 16px)" style={{ margin: "8px" }} viewBox={viewBox}>
                {rects}
            </svg>;
        } else {
            return null
        }
    }
}


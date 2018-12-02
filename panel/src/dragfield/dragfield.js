// Copyright (c) 2015, 2016, 2017 Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
import React from 'react'
import {SCREEN, storePosition} from "../common/utils";

let win = nw.Window.get();

class DragField extends React.Component {

    onMouseDown = (event) => {
        this.move = {
            startX: win.x,
            startY: win.y,
            width: win.width,
            height: win.height,
            mouseDownX: win.x + event.clientX,
            mouseDownY: win.y + event.clientY
        };
        console.log("move:", this.move);
        document.addEventListener('mouseup', this.onMouseUp)
        document.addEventListener('mousemove', this.onMouseMove)
        console.log(SCREEN.screens);
        requestAnimationFrame(this.doMove);
    };

    onMouseMove = (event) => {
        let tmp = {
            x: this.move.startX + win.x + event.clientX - this.move.mouseDownX,
            y: this.move.startY + win.y + event.clientY - this.move.mouseDownY,
            width: this.move.width,
            height: this.move.height
        };

        // Snap to nearby edge
        const snapRadius = 10;
        SCREEN.screens.map(s => s.bounds).forEach(b => {
            if (Math.abs(b.x - tmp.x) < snapRadius) {
                tmp.x = b.x;
            } else if (Math.abs(tmp.x + tmp.width - b.x - b.width) < snapRadius) {
                tmp.x = b.x + b.width - tmp.width;
            }
            console.log("Compare", tmp.y, "and", b.y);
            if (Math.abs(b.y - tmp.y) < snapRadius) {
                console.log("Snapping", tmp.y, "->", b.y);
                tmp.y = b.y;
            } else if (Math.abs(tmp.y + tmp.height - b.y - b.height) < snapRadius) {
                tmp.y = b.y + b.height - tmp.height;
            }
        });

        this.moveTo = tmp;
        console.log("moveTo:", this.moveTo);
    };

    onMouseUp = (event) => {
        console.log("mouseup:", event.clientX, event.clientY);
        document.removeEventListener('mouseup', this.onMouseUp);
        document.removeEventListener('mousemove', this.onMouseMove);
        this.move = undefined;
        setTimeout(storePosition, 10);
    };


    doMove = () => {
        if (this.moveTo) {
            console.log("doMove");
            [win.x, win.y, win.width, win.height] = [this.moveTo.x, this.moveTo.y, this.moveTo.width, this.moveTo.height];
            this.moveTo = undefined;
        }

        if (this.move) {
            requestAnimationFrame(this.doMove)
        }
    };

    render = () =>
        <div style={this.props.style} onMouseDown={this.onMouseDown}>
        <svg width="20" height="20" viewBox="-50 -50 100 100" xmlns="http://www.w3.org/2000/svg">
            <defs>
                <marker id="arrow" viewBox="0 0 12 12" refX="3" refY="6" markerWidth="5" markerHeight="5" orient="auto-start-reverse">
                    <path d="M 0 0 L 6 6 L 0 12 z" />
                </marker>
            </defs>

            <line x1="-35" y1="0" x2="35" y2="0" fill="none" stroke="black" stroke-width="8" marker-start="url(#arrow)" marker-end="url(#arrow)"  />
            <line x1="0" y1="-35" x2="0" y2="35"  fill="none" stroke="black"  stroke-width="8" marker-start="url(#arrow)" marker-end="url(#arrow)"  />

        </svg>
        </div>



}

export {DragField}

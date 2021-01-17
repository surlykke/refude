// Copyright (c) Christian Surlykke
//
// This file is part of the refude project. 
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
import React from 'react';
import ReactDOM from 'react-dom'

import { ipcRenderer } from 'electron'
import { getUrl, findLink, path2Url, addParam } from '../common/monitor';

import "../common/common.css"

let nocache = 1

export class Indicator extends React.Component {

    constructor(props) {
        super(props)
        this.state = {}

        ipcRenderer.on("screens", (evt, displays) => {
            this.getScreens(displays)
        })

        ipcRenderer.on("linkSelected", (evt, link) => {
            if (link && link.profile == "/profile/window") {
                getUrl(link.href,
                    resp => {
                        let screenshotLink = findLink(resp.data, "related", "/profile/window-screenshot")
                        if (screenshotLink) {
                            this.setState({
                                url: screenshotLink.href,
                                geometry: resp.data.Geometry
                            })
                        } else {
                            this.setState({url: undefined})
                        }
                    },
                    err => this.setState({ url: undefined }))
            } else {
                this.setState({ url: undefined })
            }
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
        let {url, geometry} = this.state
        if (url && geometry) {
    
            let {X,Y,W,H} = geometry
            let screenShotUrl = path2Url(addParam(addParam(url, "downscale", "3"), nocache, nocache++))
            let viewBox = `${this.display.x - 3} ${this.display.y - 3} ${this.display.w + 6} ${this.display.h + 6}`;
            let rects = this.screens.map((scr, i) => <rect key={`screenRect_${i}`} x={scr.x} y={scr.y} width={scr.w} height={scr.h} stroke="black" fill="white" />);
            rects.push(<image key="winRect" x={X} y={Y} width={W} height={H} xlinkHref={screenShotUrl} />);
            rects.push(<rect x={X} y={Y} width={W} height={H} stroke="black" fill="none" />)

            return <>
                <div className="topbar" />
                <svg key="windows" xmlns="http://www.w3.org/2000/svg" width="calc(100% - 16px)" style={{ margin: "8px" }} viewBox={viewBox}>
                    {rects}
                </svg>
            </>
        } else {
            return null
        }
    }
}

ReactDOM.render(<Indicator />, document.getElementById('indicator'))

import React from 'react';
import {subscribe, SCREEN} from "../common/utils";

export class Indicator extends React.Component {

    constructor(props) {
        //devtools();
        super(props);
        this.state = {};
        this.getScreenLayout();
    }

    selectorShown = 0

    componentDidMount = () => {
        subscribe("windowSelected", w => {
            this.setState({window: w});
        });
        subscribe("selectorShown", ss => this.selectorShown = ss)
    }

    getScreenLayout = () => {
        let bx1, by1, bx2, by2;
        let screens = [];
        SCREEN.screens.forEach((scr, i) => {
            screens.push({
                x: Math.round(scr.bounds.x * scr.scaleFactor),
                y: Math.round(scr.bounds.y * scr.scaleFactor),
                w: Math.round(scr.bounds.width * scr.scaleFactor),
                h: Math.round(scr.bounds.height * scr.scaleFactor)
            });
            let x1 = screens[i].x;
            let y1 = screens[i].y;
            let x2 = screens[i].x + screens[i].w;
            let y2 = screens[i].y + screens[i].h;
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
        this.display = {x: bx1, y: by1, w: bx2 - bx1, h: by2 - by1};
        this.screens = screens;
    };


    render = () => {
        let {window} = this.state;
        if (window) {
            let viewBox = `${this.display.x - 3} ${this.display.y - 3} ${this.display.w + 6} ${this.display.h + 6}`;
            let rects = this.screens.map((scr, i) => <rect key={`screenRect_${i}`} x={scr.x} y={scr.y} width={scr.w} height={scr.h} fill="lightgrey"/>);
            //rects.push(<rect key="winRect" x={window.X} y={window.Y} width={window.W} height={window.H} fill="grey" />);
            let {X, Y, W, H} = window
            let imgUrl=`http://localhost:7938${window._self}/screenshot?downscale=3&${this.selectorShown}`
            rects.push(<image key="winRect" x={X} y={Y} width={W} height={H} xlinkHref={imgUrl}/>);
            return <svg key="windows" xmlns="http://www.w3.org/2000/svg" width="calc(100% - 16px)" style={{margin: "8px"}} viewBox={viewBox}>
                {rects}
            </svg>;
        } else {
            return null;
        }
    }
}


import React from 'react';
import {render} from 'react-dom'
import {NW, WIN, devtools} from "../../common/nw";

class Indicator extends React.Component {

    constructor(props) {
        super(props);
        this.state = {};
        WIN.on('focus', () => {
            WIN.window.opener.postMessage("focus", '*')
        });

        WIN.window.addEventListener("message", (message) => {
            console.log("message:", message);
            if (message.data.screen) {
                this.setState({screen: message.data.screen});
            } else if (message.data.highlight) {
                this.setState({highlight: message.data.highlight})
            }
        });
    }

    render = () => {
        let {screen, highlight} = this.state;
        if (screen && highlight) {
            let k = 12;
            let [x, y, w, h] = [highlight.X + k, highlight.Y + k, highlight.W - 2*k, highlight.H - 2*k];
            console.log("Drawing: ", [x,y,w,h].join(' '));
            return <div>
                <svg viewBox={screen.join(" ")}>
                    <g>
                        <polyline
                            points={[x, y, x + w, y, x + w, y + h, x, y + h, x, y].join(" ")}
                            style={{
                                fill: "none",
                                strokeWidth: k,
                                stroke: "rgb(0,0,0)"
                            }}/>
                    </g> :
                </svg>
            </div>
        } else {
            return null
        }
    }
}

render(
    <Indicator/>,
    document.getElementById("root")
);

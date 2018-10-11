import React from 'react';
import {render} from 'react-dom'
import {NW, WIN, devtools} from "../../common/nw";

class Indicator extends React.Component {

    constructor(props) {
        //devtools();
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
            let {X, Y, W, H} = highlight;
            let strokeWidth = 5;
            let rectStyle = {
                fill: "blue",
                fillOpacity: "0.1",
                stroke: "black",
                strokeWidth: strokeWidth,
                strokeOpacity: "0.9"
            };
            return <div>
                <svg viewBox={screen.join(" ")}>
                    <rect x={X + strokeWidth} y={Y + strokeWidth} width={W - 2*strokeWidth - 2} height={H - 2*strokeWidth - 2} style={rectStyle}/>
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

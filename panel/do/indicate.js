import {NW, WIN, SCREEN} from "../../common/nw";

let indicatorWindows;

export let showSelectedWindow = (win) => {
    if (win) {
        let highlightMsg = {highlight: {Name: win.Name, X: win.Geometry.X, Y: win.Geometry.Y, W: win.Geometry.W, H: win.Geometry.H}};
        if (!indicatorWindows) {
            indicatorWindows = [];
            SCREEN.screens.forEach(screen => {
                NW.Window.open(
                    "do/indicator.html",
                    {"frame": false, "always_on_top": true, "transparent": true, "focus": false},
                    iWin => {
                        console.log("screen:", screen);
                        iWin.moveTo(screen.bounds.x, screen.bounds.y);
                        let w = Math.round(screen.bounds.width);
                        let h = Math.round(screen.bounds.height);
                        iWin.resizeTo(w, h);
                        indicatorWindows.push(iWin);

                        let scaledBounds = [
                            Math.round(screen.scaleFactor*screen.bounds.x),
                            Math.round(screen.scaleFactor*screen.bounds.y),
                            Math.round(screen.scaleFactor*screen.bounds.width),
                            Math.round(screen.scaleFactor*screen.bounds.height)
                        ];

                        iWin.on('loaded', () => {
                            iWin.window.postMessage({screen: scaledBounds}, '*');
                            iWin.window.postMessage(highlightMsg, '*');
                            WIN.focus()
                        });
                        iWin.window.addEventListener('focus', () => {
                            WIN.focus()
                        });
                    });
            });
        } else {
            indicatorWindows.forEach(iWin => iWin.window.postMessage(highlightMsg, '*'));
        }
    } else {
        if (indicatorWindows) {
            indicatorWindows.forEach(indicatorWindow => {
                indicatorWindow.resizeTo(0, 0);
                indicatorWindow.close()
            });
            indicatorWindows = undefined;
        }
    }
};


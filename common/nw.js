import {doGetIfNoneMatch} from "./http";
// -------------------- NW stuff ---------------------
export let NW = window.require('nw.gui');
export let WIN = NW.Window.get();

export let nwHide = () => {
    WIN.hide();
};

export let devtools = () => {
    WIN.showDevTools();
};

export let nwSetup = (onOpen) => {
    NW.App.on("open", (args) => {
        WIN.show();
        onOpen && onOpen(args.split(/\s+/));
    })
};

let displayEtag = null;

export let watchPos = () => {
    WIN.on('move', (x, y) => {
        if (displayEtag) {
            localStorage.setItem(displayEtag + ".x", x);
            localStorage.setItem(displayEtag + ".y", y);
        }
    });
};

export let adjustPos = () => {
    doGetIfNoneMatch("wm-service", "/display", displayEtag).then(
        resp => {
            if (resp.headers && resp.headers.etag) {
                let x = localStorage.getItem(resp.headers.etag + ".x");
                let y = localStorage.getItem(resp.headers.etag + ".y");
                if (x && y) {
                    WIN.moveTo(parseInt(x), parseInt(y));
                }
                displayEtag = resp.headers.etag;
            }
        },
        resp => {}
    );
};


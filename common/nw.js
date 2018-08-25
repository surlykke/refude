// Copyright (c) 2015, 2016, 2017 Christian Surlykke
//
// This file is part of the refude project. 
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
import {doGetIfNoneMatch} from "./http";
// -------------------- NW stuff ---------------------
export let NW = window.require('nw.gui');
export let WIN = NW.Window.get();
let SCREEN = NW.Screen;
SCREEN.Init();

export let devtools = () => {
    WIN.showDevTools();
};

let displayEtag, storingSceduled, windowIsShown;

let load = () => {
    let [x, y, w, h] = [
        parseInt(localStorage.getItem(displayEtag + ".x")),
        parseInt(localStorage.getItem(displayEtag + ".y")),
        parseInt(localStorage.getItem(displayEtag + ".w")),
        parseInt(localStorage.getItem(displayEtag + ".h"))
    ];
    if (!isNaN(x) && !isNaN(y) && !isNaN(w) && !isNaN(h)) {
        [WIN.x, WIN.y, WIN.width, WIN.height] = [x, y, w, h];
    }
};

let sceduleStoring = () => {
    if (!storingSceduled) {
        storingSceduled = true;
        setTimeout(() => {
            if (displayEtag) {
                localStorage.setItem(displayEtag + ".x", WIN.x);
                localStorage.setItem(displayEtag + ".y", WIN.y);
                localStorage.setItem(displayEtag + ".w", WIN.width);
                localStorage.setItem(displayEtag + ".h", WIN.height);
            }
            storingSceduled = undefined;
        }, 1000);
    }
};

export let watchWindowPositionAndSize = () => {
    WIN.on('move', () => {
        sceduleStoring();
    });

    WIN.on('resize', () => {
        sceduleStoring();
    });
};

export let watchScreenChanges = () => {
    SCREEN.on("displayBoundsChanged", () => {
        setTimeout(() => doGetIfNoneMatch("wm-service", "/display", displayEtag).then((resp) => {
            displayEtag = resp.headers.etag;
            load();
        }), 1000);
    });
};


export let showWindowIfHidden = () => {
    if (!windowIsShown) {
        windowIsShown = true;
        doGetIfNoneMatch("wm-service", "/display", displayEtag).then(
            (resp) => {
                displayEtag = resp.headers.etag;
                load();
                WIN.show();
            },
            (resp) => {
                WIN.show();
            });
    }
};

export let hideWindow = () => {
    windowIsShown = undefined;
    WIN.hide();
};





// Copyright (c) 2015, 2016, 2017 Christian Surlykke
//
// This file is part of the refude project. 
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//

// -------------------- NW stuff ---------------------
export let Utils = window.require('nw.gui');
export let WIN = Utils.Window.get();
export let SCREEN = Utils.Screen;
SCREEN.Init();

export let devtools = () => {
    WIN.showDevTools();
};

export let screenId = () => {
    let id = "geometry"
    SCREEN.screens.forEach(screen => {
        id = id + '-' + screen.id + '-' + screen.bounds.x + '-' + screen.bounds.y + '-' + screen.bounds.width + '-' + screen.bounds.height
    });
    return id;
};

let storedPosition = {x: 0, y: 0}

let loadPosition = () => {
    let id = screenId();
    let str = localStorage.getItem(id);
    if (str) {
        let geometry = JSON.parse(str);

        if (geometry) {
            let [x, y, w, h] = [
                parseInt(geometry.x),
                parseInt(geometry.y),
                parseInt(geometry.w),
                parseInt(geometry.h)
            ];
            if (!isNaN(x) && !isNaN(y) && !isNaN(w) && !isNaN(h)) {
                storedPosition.x = x
                storedPosition.y = y
                WIN.x = storedPosition.x
                WIN.y = storedPosition.y
            }
        }
    }
};

export let managePosition = () => {
    loadPosition()
    let aboutToLoad
    SCREEN.on("displayBoundsChanged", () => {
        if (!aboutToLoad) {
            aboutToLoad = true;
            setTimeout(() => {
                loadPosition();
                aboutToLoad = undefined;
            },
                1000
            );
        }
    });

    let checkPosition = () => {
        if (WIN.x !== storedPosition.x || WIN.y !== storedPosition.y) {
            let id = screenId();
            let value = JSON.stringify({ x: WIN.x, y: WIN.y, w: WIN.width, h: WIN.height })
            localStorage.setItem(id, value);
        }

        setTimeout(checkPosition, 5000)
    }

    setTimeout(checkPosition, 10000)
}

export let manageZoom = () => {
    let zoomLevels = [25, 33, 50, 67, 75, 80, 90, 100, 110, 125, 150, 175, 200, 250, 300, 400, 500] // The ones that chromium have
                    //-7  -6  -5  -4  -3  -2  -1    0    1    2    3    4    5    6    7    8    9
    let currentZoom = Math.round(localStorage.bodyZoom || 0)
    if (currentZoom > 9) currentZoom = 9
    if (currentZoom < -7) currentZoom = -7
    document.body.style.zoom = 1.0*zoomLevels[currentZoom + 7]/100

    let zoom = (up) => {
        console.log("Into zoom, up:", up, "currentZoom:", currentZoom)
        if (up && currentZoom < 9) currentZoom++
        else if (!up && currentZoom > -7) currentZoom--
        console.log("now currentZoom:", currentZoom)
        localStorage.bodyZoom = currentZoom
        document.body.style.zoom = 1.0*zoomLevels[currentZoom + 7]/100
        publish("componentUpdated")
    }

    window.addEventListener("keydown", function (e) {
        if (e.ctrlKey && !e.shiftKey && !e.altKey && !e.metaKey && "+" === e.key) {
            zoom(true)
        } else if (e.ctrlKey && !e.shiftKey && !e.altKey && !e.metaKey && "-" === e.key) {
            zoom(false)
        }
    });
}


export let applicationRank = (app, lowercaseTerm) => {
    let tmp;
    if ((tmp = app.Name.toLowerCase().indexOf(lowercaseTerm)) > -1) {
        return -tmp;
    } else if (app.Comment && (tmp = app.Comment.toLowerCase().indexOf(lowercaseTerm)) > -1) {
        return -tmp - 100;
    } else {
        return 1;
    }
};

const PUBSUB = (() => {
    let subscriptions = {}
    return {
        subscribe: (topic, fn) => {
            subscriptions[topic] = subscriptions[topic] || [];
            subscriptions[topic].push(fn);
        },
        publish: (topic, obj) => {
            let subscribers = subscriptions[topic] || [];
            subscribers.forEach(fn => fn(obj));
        }
    };
})();

export let publish = (topic, data) => PUBSUB.publish(topic, data);
export let subscribe = (topic, fn) => PUBSUB.subscribe(topic, fn);


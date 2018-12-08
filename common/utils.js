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

export let loadPosition = () => {
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
                [WIN.x, WIN.y] = [x, y];
            }
        }
    }
};

export let storePosition = () => {
    let id = screenId();
    let value = JSON.stringify({x: WIN.x, y: WIN.y, w: WIN.width, h: WIN.height})
    localStorage.setItem(id, value);
};


let aboutToLoad
export let watchScreenChanges = () => {
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
};

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
            console.log("Publish", topic);
            let subscribers = subscriptions[topic] ||  [];
            subscribers.forEach(fn => fn(obj));
        }
    };
})();

export let publish = (topic, data) => PUBSUB.publish(topic, data);
export let subscribe = (topic, fn) => PUBSUB.subscribe(topic, fn);


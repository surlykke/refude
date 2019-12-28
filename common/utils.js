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

let savePosition = () => {
    if (WIN.x !== storedPosition.x || WIN.y !== storedPosition.y) {
        let id = screenId();
        let value = JSON.stringify({ x: WIN.x, y: WIN.y, w: WIN.width, h: WIN.height })
        localStorage.setItem(id, value);
        storedPosition.x = WIN.x
        storedPosition.y = WIN.y
    }
}

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

    subscribe("selectorShown", savePosition)    
}



export let manageZoom = () => {
    let zoom = [0.25, 0.33, 0.50, 0.67, 0.75, 0.80, 0.90, 1.00, 1.10, 1.25, 1.50, 1.75, 2.00, 2.50, 3.00, 4.00, 5.00] // The ones that chromium have

    let normalize = level => Number.isInteger(level) ? Math.max(0, Math.min(zoom.length - 1, level)) : 7

    let setZoomLevel = adjustment => {
        localStorage.zoomLevel = normalize(Number.parseInt(localStorage.zoomLevel) + adjustment)
        document.body.style.zoom = zoom[localStorage.zoomLevel]
        publish("componentUpdated")
    }
    
    setZoomLevel(0)

    window.addEventListener("keydown", function (e) {
        if (e.ctrlKey && !e.shiftKey && !e.altKey && !e.metaKey && "+" === e.key) {
            setZoomLevel(1)
        } else if (e.ctrlKey && !e.shiftKey && !e.altKey && !e.metaKey && "-" === e.key) {
            setZoomLevel(-1)
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


export let rank = (item, lowercaseTerm) => {

    if (lowercaseTerm === "" && !item.matchEmpty) {
        return -1;
    }

    let tmp = item.name.toLowerCase().indexOf(lowercaseTerm)
    if (tmp > -1) {
        return tmp
    }

    // More fluffy search - eg. pwr matches PowerOff or nvim matches neovim 
    // When we reach here we know lowercaseTerm is not empty
    if (item.matchFluffy) {
        let j = 0;
        for (let i = 0; i < item.name.length; i++) {
            if (item.name[i].toLocaleLowerCase() === lowercaseTerm[j]) {
                j++;
            }
            if (j >= lowercaseTerm.length) {
                return 100 + i
            }

        }
    }

    if (item.comment) {
        tmp = item.comment.toLowerCase().indexOf(lowercaseTerm)
        if (tmp > -1) {
            return 200 + tmp
        }
    }
    return -1
};

export let filterAndSort = (itemlist, lowercaseTerm) => {
    return itemlist.filter(i => {
        i.rank = rank(i, lowercaseTerm);
        return i.rank > -1;
    }).sort((i1, i2) => i1.rank - i2.rank);
}
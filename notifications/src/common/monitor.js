// Copyright (c) 2015, 2016, 2017 Christian Surlykke
//
// This file is part of the refude project. 
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
import {doGetIfNoneMatch} from '../common/http'

let monitorResources = (service, mimetype, onUpdated) => {
    let linksEtag = null;
    let itemMap = new Map();

    let update = () => {
        let items = [];
        for (let [k, v] of itemMap) {
            if (v.item) {
                items.push(v.item);
            }
        }
        onUpdated(items);
    }

    let getItems = () => {
        for (let [path, pair] of itemMap) {
            doGetIfNoneMatch(service, path, pair.etag).then((o) => {
                itemMap.set(path, {etag: o.headers["etag"], item: o.json});
                update();
            }, o => {
                if (o.status === 404) {
                    itemMap.delete(path);
                    update();
                }
            });
        }
    };

    let getLinks = () => {
        doGetIfNoneMatch(service, "/links", linksEtag).then(
            (o) => {
                linksEtag = o.headers["etag"];
                (o.json[mimetype] || []).forEach(path => {
                    itemMap.set(path, {})
                });
                getItems();
            },
            (o) => {
                getItems();
            });
    };

    let run = () => {
        getLinks();
        setTimeout(run, 1000);
    };
    run();
};

export {monitorResources}

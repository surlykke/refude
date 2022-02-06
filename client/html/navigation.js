// Copyright (c) Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
import { doPost, doDelete } from "./utils.js";

export const getLinks = () => Array.from(document.querySelectorAll(".link"));
export const getSelectedLink = () => getLinks().find(l => l.classList.contains('selected'));
export const selectLink = link => {
    let list = getLinks();
    link = link || list[0];
    list.forEach(l => l === link ? l.classList.add('selected') : l.classList.remove('selected'));
};

export let preferred

export let setPreferred = href => {
    console.log("setPreferred:", href) 
    preferred = href
}

export let selectPreferred = () => {
    selectLink(document.querySelector(`[href="${preferred}"]`))
}

export const activateSelected = (then) => {
    console.log("Activate selected")
    let selectedLink = getSelectedLink();
    if (selectedLink) {
        then = then || (() => { });
        doPost(selectedLink.href).then(response => response.ok && then());
    }
};

export const deleteSelected = (then) => {
    let selectedLink = getSelectedLink();
    if (selectedLink) {
        then = then || (() => { });
        doDelete(selectedLink.href).then(response => response.ok && then())
    }
};

export const move = up => {
    let list = getLinks();
    let currentIndex = list.findIndex(ld => ld.classList.contains('selected'));
    if (list.length < 2 || currentIndex < 0) {
        selectLink(list[0]);
    } else {
        let newLink = list[(currentIndex + list.length + (up ? -1 : 1)) % list.length];
        setPreferred(newLink.href)
        selectLink(newLink)
    }
};

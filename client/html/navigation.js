import { doPost, doDelete, getJson, linkHref } from "./utils.js";

export let navigator;

export const setNavigation = n => navigator = n;
export const getLinkDivs = () => Array.from(document.querySelectorAll(".link"));
export const getSelectedLinkDiv = () => getLinkDivs().find(l => l.classList.contains('selected'));
export const getSelectedAnchor = () => getSelectedLinkDiv()?.lastChild;
export const select = linkDiv => {
    let list = getLinkDivs();
    linkDiv = linkDiv || list[0];
    list.forEach(l => l === linkDiv ? l.classList.add('selected') : l.classList.remove('selected'));
};

export const activateSelected = (then) => {
    let selectedAnchor = getSelectedAnchor();
    console.log("Into activate", selectedAnchor, then);
    if (selectedAnchor) {
        then = then || (() => { });
        if (selectedAnchor.rel === "org.refude.defaultaction" || selectedAnchor.rel === "org.refude.action") {
            doPost(selectedAnchor.href).then(response => response.ok && then());
        } else if (selectedAnchor.rel === "org.refude.delete") {
            doDelete(selectedAnchor.href).then(response => response.ok && then());
        } else if (selectedAnchor.rel === "related") {
            getJson(selectedAnchor.href, json => {
                let defaultActionLink = linkHref(json, "org.refude.defaultaction");
                if (defaultActionLink) {
                    doPost(defaultActionLink).then(response => response.ok && then());
                }
            });
        }
    }
};

export const deleteSelected = (then) => {
    let selectedAnchor = getSelectedAnchor();
    if (selectedAnchor) {
        if (selectedAnchor.rel === "org.refude.delete") {
            doDelete(selectedAnchor.href).then(response => response.ok && then());
        } else if (selectedAnchor.rel === "related") {
            getJson(selectedAnchor.href, json => {
                let deleteLink = linkHref(json, "org.refude.delete");
                if (deleteLink) {
                    doDelete(deleteLink).then(response => response.ok && then())
                }
            });
        }
    }
};

export const selectActivateAndDismiss = (element) => {
    select(element)
    activateSelected(navigator.dismiss)
}

export const move = up => {
    console.log("move");
    let list = getLinkDivs();
    let currentIndex = list.findIndex(ld => ld.classList.contains('selected'));
    console.log("currentIndex:", currentIndex, "list.length:", list.length);
    if (list.length < 2 || currentIndex < 0) {
        select(list[0]);
    } else {
        select(list[(currentIndex + list.length + (up ? -1 : 1)) % list.length]);
    }
};

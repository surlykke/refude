import { doPost, doDelete, getJson, linkHref } from "./utils.js";

export let navigator;

export const getLinkDivs = () => Array.from(document.querySelectorAll(".link"));
export const getSelectedLinkDiv = () => getLinkDivs().find(l => l.classList.contains('selected'));
export const getSelectedAnchor = () => getSelectedLinkDiv()?.lastChild;
export const selectDiv = linkDiv => {
    let list = getLinkDivs();
    linkDiv = linkDiv || list[0];
    list.forEach(l => l === linkDiv ? l.classList.add('selected') : l.classList.remove('selected'));
};

export let preferred

export let setPreferred = href => preferred = href

export let selectPreferred = () => {
    let div 
    if (preferred) {
        let a = document.querySelector(`[href="${preferred}"]`)
        if (a && a.parentElement.classList.contains('link')) {
            div = a.parentElement
        }
    }
    selectDiv(div)
}

let findPreferred = () => {

}

export const activateSelected = (then) => {
    let selectedAnchor = getSelectedAnchor();
    if (selectedAnchor) {
        then = then || (() => { });
        doPost(selectedAnchor.href).then(response => response.ok && then());
    }
};

export const deleteSelected = (then) => {
    let selectedAnchor = getSelectedAnchor();
    if (selectedAnchor) {
        then = then || (() => { });
        doDelete(selectedAnchor.href).then(response => response.ok && then())
    }
};

export const selectActivateAndDismiss = (element) => {
    selectDiv(element)
    activateSelected(navigator.dismiss)
}

export const move = up => {
    let list = getLinkDivs();
    let currentIndex = list.findIndex(ld => ld.classList.contains('selected'));
    if (list.length < 2 || currentIndex < 0) {
        selectDiv(list[0]);
    } else {
        let newDiv = list[(currentIndex + list.length + (up ? -1 : 1)) % list.length];
        setPreferred(newDiv.lastChild?.href)
        selectDiv(newDiv)
    }
};

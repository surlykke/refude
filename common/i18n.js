// Copyright (c) 2015, 2016, 2017 Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
export let translations = {};

let strings = [
    "Open windows",
    "Applications",
    "Leave",
    "Open files of type <b>%0</b>with <b>%1</b>?",
    "Just once",
    "Always",
    "Cancel",
    "Open &nbsp;<b>%0</b>&nbsp;with:",
    "Applications that handle %0",
    "Other applications"
];


// ---- per country -------------

translations["da"] = {
    "Open windows": "Åbne vinduer",
    "Applications": "Applikationer",
    "Leave": "Sluk",
    "Open files of type <b>%0</b>with <b>%1</b>?":
        (args) => `Åbn filer af typen <b>${args[0]}</b> med <b>${args[1]}</b>?`,
    "Just once": "Kun nu",
    "Always": "Altid",
    "Cancel": "Annuller",
    "Open &nbsp;<b>%0</b>&nbsp;with:":
        (args) => `Åbn <b>${args[0]}</b> med:`,
    "Applications that handle %0":
        (args) => `Applikationer som håndterer ${args[0]}`,
    "Other applications": "Andre applikationer"
};



// Copyright (c) 2015, 2016, 2017 Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
import {translations} from "./i18n";

export let T = (string, ...args) => {
    let locale = navigator.language;
    let translation = translations[locale];
    if (!translation) {
        let underscorePos = locale.indexOf("_");
        if (underscorePos > -1) {
            translation = translations[locale.substring(0, underscorePos)];
        }
    }
    let translatedString;
    if (translation) {
        let translator = translation[string];
        if (translator) {
            if (typeof translator === "string") {
                translatedString = translator;
            } else {
                translatedString = translator(args)
            }
        }
    }
    if (translatedString) {
        return translatedString;
    } else {
        args.forEach((arg, i) => {
            let regexp = new RegExp(`%${i}`, 'g');
            string = string.replace(regexp, arg)
        });
        return string;
    }
};

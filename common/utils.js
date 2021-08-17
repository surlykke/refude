// Copyright (c) Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.

export let linkHref = (res, rel) => {
    rel = rel || "self"
    return res._links.find(l => l.rel === rel)?.href
}
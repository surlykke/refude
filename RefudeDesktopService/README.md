# RefudeDesktopService

RefudeDesktopServices is a part of the [RefudeServices](http://github.com/surlykke/RefudeServices) project. It exposes information about installed applications and mimetypes.

## Standard Resources 

`/notify`: Server-sent event stream

## Applications

Application resource paths are all of the form 

`/applications/<appname>.desktop`

For example you can get data about firefox (assuming its installed) with:

`RefudeGET desktop-service /applications/firefox.desktop` 

which might produce (after formatting):

```json
{
    "Type": "Application",
    "Version": "1.0",
    "Name": "Firefox",
    "GenericName": "Web Browser",
    "NoDisplay": false,
    "Comment": "Browse the Web",
    "IconName": "firefox",
    "Hidden": false,
    "OnlyShowIn": [

    ],
    "NotShowIn": [

    ],
    "Terminal": false,
    "Mimetypes": [
        "text/html",
        "text/xml",
        "application/xhtml+xml",
        "application/vnd.mozilla.xul+xml",
        "text/mml",
        "x-scheme-handler/http",
        "x-scheme-handler/https"
    ],
    "Categories": [
        "Network",
        "WebBrowser"
    ],
    "Implements": [

    ],
    "Keywords": [
        "web",
        "browser",
        "internet"
    ],
    "StartupNotify": false,
    "Actions": {
        "_default": {
            "Comment": "Browse the Web",
            "Name": "Firefox",
            "Exec": "/usr/lib/firefox/firefox %u",
            "IconName": "firefox",
            "IconUrl": ""
        },
        "new-private-window": {
            "Comment": "Firefox",
            "Name": "Desktop Action new-private-window",
            "Exec": "/usr/lib/firefox/firefox --private-window %u",
            "IconName": "firefox",
            "IconUrl": ""
        },
        "new-window": {
            "Comment": "Firefox",
            "Name": "Desktop Action new-window",
            "Exec": "/usr/lib/firefox/firefox --new-window %u",
            "IconName": "firefox",
            "IconUrl": ""
        }
    },
    "Id": "firefox.desktop",
    "RelevanceHint": 0
}
```

This is mostly a json representations of [desktop entry keys](https://specifications.freedesktop.org/desktop-entry-spec/latest/ar01s05.html), 

Differences are: 

*Actions* are collected into to a map. There will always be an action with id '\_default' korresponding to the default action of the desktopfile.
If the desktopentry has [additional application actions](https://specifications.freedesktop.org/desktop-entry-spec/latest/ar01s10.html), those 
actions will be present in the map with the action identifier as key.

*Mimetypes* are named in plural and will reflect what associations are added/removed in various mimetypes.list files.


### Collecting desktop files

Desktopfiles are collected, and corrected according to the rules set out in 
[Desktop Entry Specification](https://specifications.freedesktop.org/desktop-entry-spec/latest/index.html) and 
[Association between MIME types and applications](https://specifications.freedesktop.org/mime-apps-spec/latest/)

Very briefly this means that *more local takes presedence*. I.e. if 'foo.desktop' is found both in /usr/share/applications 
and ~/.local/share/applications the latter will hide the former. If 'foo.desktop' in /usr/share/application associates foo to a mimetype, that 
association may be removed in, say, ~/.config/mimetypes.list. 

### Launching

An application may be launched by doing a `POST` against its resource. For example doing:

```
RefudePOST desktop-service /applications/firefox.desktop
```

will launch firefox.

Additional actions may be accessed by adding the action identifier to the url POSTED against. Eg:

```
RefudePOST desktop-service /applications/firefox.desktop?action=new-private-window

```

which will evoke the 'new-window-action' defined in firefox.desktop

If no action is given, '_default' is assumed.

## Mimetypes

Mimetype resources have paths of form `/mimetypes/<type>/<subtype>`
		
For example a resource for 'text/html' may be retrieved with:

```
RefudeGET desktop-service /mimetypes/text/html
```

and will look like:

```json
{
    "Id": "text/html",
    "Comment": "HTML document",
    "Acronym": "HTML",
    "ExpandedAcronym": "HyperText Markup Language",
    "Aliases": [

    ],
    "Globs": [
        "*.html",
        "*.htm"
    ],
    "SubClassOf": [
        "text/plain"
    ],
    "IconName": "text-html",
    "GenericIcon": "text-x-generic",
    "AssociatedApplications": [
        "chromium.desktop",
        "firefox.desktop"
    ],
    "DefaultApplications": [
        "chromium.desktop"
    ]
}
```
The two fields 'AssociatedApplications' and 'DefaultApplications' reflect what's declared in installed desktop files along with
the various 'mimelist.apps' files.

The remaining fields are extracted from the mimetype database in /usr/share/mime, /usr/local/share/mime and ~/.local/share/mime

'DefaultApplication' may be changed by POST'ing against the resource. For example to set gvim as the default application for text/html, do:

```shell
RefudePOST desktop-service /mimetypes/text/html?defaultApp=gvim.desktop
```

## Scheme-handlers

Mimetypes of form `x-scheme-handler/<protocol>` are a kind of pseudo-mimetypes, used to associate applications to *protocols*.

RefudeDesktopService supports these as well, so you could set firefox to handle the 'http' protocol by doing:

```shell
RefudePOST desktop-service /mimetypes/x-scheme-handler/http?defaultApp=firefox.desktop
```


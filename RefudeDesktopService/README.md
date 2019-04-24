# RefudeServices

RefudeServices are a set restful of services for the desktop. 

These services provide resources, that can be accessed through http over tcp or 
a unix-domain-socket. 

The unix domain socket is located at 
```$XDG_RUNTIME_DIR/org.refude.desktop-service```, or if ```XDG_RUNTIME_DIR``` 
is not defined, at ```/tmp/org.refude.desktop-service```. 

You may connect over TCP at ```localhost:7938```.

For example, assuming you have firefox installed, you can get the firefox 
resource with the curl command:

```
curl http://localhost:7938/application/firefox
```
which will give (something like):

```json
{
    "_self": "/application/firefox",
    "_links": [
        {
            "href": "/application/firefox",
            "rel": "self"
        }
    ],
    "_actions": {
        "default": {
            "Description": "Launch",
            "IconName": "firefox"
        },
        "new-private-window": {
            "Description": "Open a New Private Window",
            "IconName": "firefox"
        },
        "new-window": {
            "Description": "Åbn et nyt vindue",
            "IconName": "firefox"
        }
    },
    "Type": "Application",
    "Version": "1.0",
    "Name": "Firefox - internetbrowser",
    "GenericName": "Webbrowser",
    "NoDisplay": false,
    "Comment": "Surf på internettet",
    "IconName": "firefox",
    "Hidden": false,
    "OnlyShowIn": [],
    "NotShowIn": [],
    "Exec": "firefox %u",
    "Terminal": false,
    "Categories": [
        "GNOME",
        "GTK",
        "Network",
        "WebBrowser"
    ],
    "Implements": [],
    "Keywords": null,
    "StartupNotify": true,
    "DesktopActions": {
        "new-private-window": {
            "Name": "Open a New Private Window",
            "Exec": "firefox -private-window",
            "IconName": "firefox"
        },
        "new-window": {
            "Name": "Åbn et nyt vindue",
            "Exec": "firefox -new-window",
            "IconName": "firefox"
        }
    },
    "Id": "firefox",
    "Mimetypes": [
        "application/x-xpinstall",
        "video/webm",
        "image/gif",
        "application/rdf+xml",
        "text/html",
        "image/png",
        "application/xml",
        "application/rss+xml",
        "application/xhtml+xml",
        "image/jpeg"
    ]
}

```


```
curl --unix $XDG_RUNTIME_DIR/org.refude.desktop-service http://localhost/application/firefox
```

RefudeServices also comes with a command line client - ```refuc``` - allowing 
you to do:

```
refuc /application/firefox
```

### Service structure

#### Generic resources


#### /doc

Serves this document

ou'll get a list of all 'some-service's resources with a path beginning with `/foo/baa/`


## Clients

Of course playing around with refuc, curl and such is great fun, but the real purpose of RefudeServices is to 
serve as infrastructure for desktop environments and applications.

There is only one known RefudeServices-client: [refude](https://github.com/surlykke/refude). Go check that out.

## Installing RefudeServices

RefudeServices are know to run with Go 1.8. If you have go installed and $GOPATH set up, you should be able to do:

```
go get github.com/surlykke/RefudeServices
```

and then navigate to `$GOPATH/src/github.com/surlykke/RefudeServices` and do 

```
./install.sh
```

`install.sh` will do:

* Build the services and applications and install them to `$GOPATH/bin`
* Install the scripts `RefudeGET`, `RefudePOST` and `runRefude.sh` to 
  `~/.local/bin`
* Install `RefudeGET`/`RefudePOST` completion-scripts for bash, zsh and fish to 
  various directories beneath `~/.local`
* Make a soft-link from `~/.local/bin/xdg-open` to `$GOPATH/bin/RefudeXdgOpen`. 
  This effectively shadows whatever  xdg-open you may have already installed. 
  If you do not want that, remove the link.
 
If you don't want all this done, you'll have to look inside `install.sh` and do the parts you want manually.

## Running RefudeServices

If you've done the installation described above, you can do

```
runRefude.sh
```

You may want to call runRefude.sh from some startupscript, so it runs whenever you login to your computer.


## Troubleshooting

If services get hung or misfire or something

```
runRefude.sh --restart
```

may be useful. 

You are also welcome to file bug-reports, obviously.


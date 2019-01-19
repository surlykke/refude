# RefudeServices

RefudeServices are a set restful of services for the desktop. 

These services provide resources, that can be accessed through http over unix-domain-sockets. 

The sockets reside under $XDG_RUNTIME_DIR. For example, RefudeDesktopService sits on
$XDG_RUNTIME_DIR/org.refude.desktop-service. 

Assuming you have firefox installed, and assuming $XDG_RUNTIME_DIR points to /run/user/1000, you can get data about firefox with the curl command:

```
curl --unix /run/user/1000/org.refude.desktop-service http://localhost/applications/firefox.desktop
```

RefudeServices also comes with a script - RefudeGET, which is a small wrapper for curl, allowing you to do:

```
RefudeGET desktop-service /applications/firefox.desktop
```

Via RefudeProxy it is also possible to access the resources over http/TCP on localhost:7938. Here a resource path is prefixed by the last part of the 
unix socket name, so to get data about firefox, you'd do:

curl http://localhost:7938/desktop-service/applications/firefox.desktop

The services are (all sockets reside in $XDG_RUNTIME_DIR):

* [RefudeDesktopService](RefudeDesktopService/README.md) on org.refude.desktop-services: Installed applications and mimetypes.
* [RefudeWmService](RefudeDesktopService/windows/README.md) on org.refude.wm-service: Window manager functionality.
* [RefudeIconService](RefudeIconService/README.md) on org.refude.icon-service: Installed icons and themes.
* [RefudePowerService](RefudePowerService/README.md) on org.refude.power-service: Battery status and session control.
* [RefudeNotificationService](RefudeNotificationService/README.md) on org.refude.notifications-service: Desktop notifications.
* [RefudeStatusNotifierService](RefudeStatusNotifierService/README.md) on org.refude.statusnotifier-service: Desktop status notifications. Socket

In addition to the services there are two applications:

* [RefudeProxy](RefudeProxy/README.md): Which proxies the unix socket to TCP, allowing acces via TCP on http://localhost:7938
* [RefudeXdgOpen](RefudeXdgOpen/README.md): A stand-alone application meant to serve as a drop-in replacement for freedesktops xdg-open. 

### Service structure

#### Generic resources

All services will have these resources:

##### /ping

Can be used to check if the service is alive. Will respond to a Http GET with Status 200 Ok.

##### /doc

A markdown document, documenting the resource

##### /notify 
A server-sent event stream notifying about changes to the services resources. It will emit 3 types of events:

* resource-added
* resource-removed
* resource-updated

Eg: If you install application foo whith desktop file foo.desktop, you should se an event like
  
```
event:resource-added
data:applications/foo.desktop
```

in the stream. 

Here applications/foo.desktop is the path of the newly created resource *relative to the path of the notify resource*

#### Hierachical service structure

Resource-urls are organized in a (somewhat) hierachical structure, with a concept of directories: If you do:

```
RefudeGET some-service /foo/baa/
```

you'll get a list of all 'some-service's resources with a path beginning with `/foo/baa/`


## Clients

Of course playing around with RefudeGET and curl and such is great fun, but the real purpose of RefudeServices is to 
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
* Install the scripts `RefudeGET`, `RefudePOST` and `runRefude.sh` to `~/.local/bin`
* Install `RefudeGET`/`RefudePOST` completion-scripts for bash, zsh and fish to various directories beneath `~/.local`
* Make a soft-link from `~/.local/bin/xdg-open` to `$GOPATH/bin/RefudeXdgOpen`. This effectively replaces the xdg-open 
  from freedesktop.org that you may have installed with `RefudeXdgOpen`.
 
If you don't want all this done, you'll have to look inside `install.sh` and do the parts you want manually.

## Running RefudeServices

If you've done the installation described above, you can do

```
runRefude.sh
```

You may want to call runRefude.sh from some startupscript, so it runs whenever you login to your computer.

I have this line in ~/.xprofile:

```
runRefude.sh > ~/tmp/refude.log 2>~/tmp/refude.err

```

(The redirections are not strictly necessary, but convenient when weird stuff happens)

## Troubleshooting

If services get hung or misfire or something

```
runRefude.sh --restart
```

may be useful. 

You are also welcome to file bug-reports, obviously.

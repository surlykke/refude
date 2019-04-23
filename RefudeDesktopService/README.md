# RefudeServices

RefudeServices are a set restful of services for the desktop. 

These services provide resources, that can be accessed through http over tcp or 
a unix-domain-socket. 

The unix domain socket is located at 
```$XDG_RUNTIME_DIR/org.refude.desktop-service```, or if ```XDG_RUNTIME_DIR``` 
is not defined, at ```/tmp/org.refude.desktop-service```. 

You may connect over TCP at ```localhost:7938```.

For example, assuming you have firefox installed, you can get data about firefox 
with the curl command:

```
curl http://localhost:7938/applications/firefox
```
or if you prefer unix domain sockets:

curl --unix $XDG_RUNTIME_DIR/org.refude.desktop-service http://localhost/applications/firefox


RefudeServices also comes with a command line client - ```refude``` - allowing 
you to do:

```
refude /applications/firefox
```

In addition to the services there are two applications:

* [RefudeProxy](RefudeProxy/README.md): Which proxies the unix socket to TCP, allowing acces via TCP on http://localhost:7938
* [RefudeXdgOpen](RefudeXdgOpen/README.md): A stand-alone application meant to serve as a drop-in replacement for freedesktops xdg-open. 

### Service structure

#### Generic resources


Can be used to check if the service is alive. Will respond to a Http GET with Status 200 Ok.

##### /doc

Serves this document

ou'll get a list of all 'some-service's resources with a path beginning with `/foo/baa/`


## Clients

Of course playing around with refude, curl and such is great fun, but the real purpose of RefudeServices is to 
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


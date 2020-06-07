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
curl http://localhost:7938/application/firefox.desktop
```

or, using the domain socket:

```
curl --unix $XDG_RUNTIME_DIR/org.refude.desktop-service http://localhost/application/firefox.desktop
```

RefudeServices also comes with a command line client - ```refuc``` - allowing 
you to do:

```
refuc /application/firefox.desktop
```

### Service structure

#### Generic resources


#### [/doc](#doc)

Serves this document


## Clients

Of course playing around with refuc, curl and such is great fun, but the real purpose of RefudeServices is to 
serve as infrastructure for desktop environments and applications.

There is only one known RefudeServices-client: [refude](https://github.com/surlykke/refude). Go check that out.

## Installing RefudeServices

RefudeServices is known to run with Go 1.13. If you have go installed, you should be able to do:

```
    cd to-where-you-want-to-check-out
    git clone https://github.com/surlykke/RefudeServices
    cd RefudeServices
    ./install.sh
```

`install.sh` will do:

* Invoke `go install` to build and install executables `refuc` and `RefudeDesktopService`. Do `go help install` for further information. First time `go install` is invoked, it will pull necessary dependencies from the net, so you'll need an internet connection then.
* Copy this file to `$XDG_DATA_HOME/RefudeServices`
* Add a few refude icons to the hicolor icon theme
* Copy completion-scripts for bash, and fish to 
  various directories beneath `$XDG_DATA_HOME`. The completion scripts are for `refuc`
 
If you don't want all this done, you'll have to look inside `install.sh` and do the parts you want manually. In fact, have a look inside install.sh anyway, to see what's going on.

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


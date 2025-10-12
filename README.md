# refude

refude is a window switcher and app launcher. With it, you can:

* Switch between windows
* Jump to tabs in your browser (providing your browser is chromium or a derivative thereof )
* Launch applications
* Open files in your home catalog, your download catalog and a few other

refude also comes with 
    * an implementation of the dbus notification protocol, and can show notifications on screen.


## Install 

1. Ensure that:
  - Go is installed. A version recent enough that it supports modules. 1.13.xx seems to work.
  - git, libmagic and libx11 are installed (libmagic and libx11 with development files)
  
  On Ubuntu 20.04 you may do this with:
  ```
  sudo apt install golang git libmagic-dev libx11-dev
  ```
1. cd /to/where/you/want/to/build
1. git clone https://github.com/surlykke/refude
1. cd refude
1. ./install.sh
  
Ensure that $HOME/.local/bin is in your PATH, then you can do:

```
runRefude.sh
``` 

To check everything is up and running, you can do
```
refuc /window/
```
which should produce a http-response containing json describing open windows on your desktop.

If not, then perhaps try:
```
refude-server
```
directly, which may yield useful errormessages.


## Troubleshooting

If services get hung or misfire or something

```
runRefude.sh --restart
```

may be useful. 

You are also welcome to file bug-reports, obviously.

## Usage 

With refude-server running, in your browser, go to: 

```
http://localhost:7938/desktop
```

It should show a list of your currently open windows

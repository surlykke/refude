# RefudeServices

RefudeServices are a set restful of services for the desktop.  

The main purpose of RefudeServices is to serve as a backend for [refude](https://github.com/surlykke/refude).

## Install 

1. Ensure that:
  - Go is installed. A version recent enough that it supports modules. 1.13.xx seems to work.
  - git, libmagic and libx11 are installed (libmagic and libx11 with development files)
  
  On Ubuntu 20.04 you may do this with:
  ```
  sudo apt install golang git libmagic-dev libx11-dev
  ```
1. cd /to/where/you/want/to/build
1. git clone https://github.com/surlykke/RefudeServices
1. cd RefudeServices
1. ./install.sh
  
Ensure that $HOME/.local/bin is in your PATH, then you can do:

```
runRefude.sh
``` 

To check everything is up and running, you can do
```
refuc /windows
```
which should produce a http-response containing json describing open windows on your desktop.

If not, then perhaps try:
```
RefudeServices
```
directly, which may yield useful errormessages.


## Troubleshooting

If services get hung or misfire or something

```
runRefude.sh --restart
```

may be useful. 

You are also welcome to file bug-reports, obviously.


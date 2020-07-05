# refude - some tools for your desktop

## What is it?

refude constists of:

* A panel. When launched, refude will open a (small) panel, floating above ordinary windows. It contains:
  * A clock
  * A statusnotification area. Applications supporting the [statusnotification standard](https://www.freedesktop.org/wiki/Specifications/StatusNotifierItem/) can show icons here.
  * A Battery indicator.
* A window switcher/application launcher. With this you can:
  * Switch between open windows
  * Launch installed applications
  * Open files from various directories
  * Perform actions on recently received notifications
* OSD notifications, somewhat inspired by how [ubuntu does that](https://wiki.ubuntu.com/NotifyOSD)

I use openbox as my desktop environment, and run refude inside that. It _should_ be possible to run refude on top of any EWMH compliant window manager, but I haven't really tried it.

Currently refude won't work with wayland, as it needs EWMH, and there is, to my knowledge, 
no equivalent of that in the wayland world.

## The name

Originally, the name 'refude' was meant as an abreviation of **re**st**fu**l **d**esktop **e**nvironment, 
as the accompanying project 'RefudeServices' was ment to be a set of restful services. 
These, services, however, have turned out not restful in the 
[strict sense](https://www.ics.uci.edu/~fielding/pubs/dissertation/top.htm), 
so now 'refude' is just a name.

## Installing

To install:

1. Install and ensure [RefudeServices](https://github.com/surlykke/RefudeServices) are running.
1. Install node...
1. cd to /where/you/want/to/build/refude 
1. git clone https://github.com/surlykke/refude.git
1. cd refude
1. ./install.sh

If all went well you should have an executable 'refude' in ~/.local/bin. Also there, you'll find 'refude.sh'
which acts as a wrapper for refude. 

From a terminal, run:
```
refude.sh
```
and the refude panel should show.

run ```refude.sh``` again, and the app switcher should show.  

run ```refude.sh``` again, and the selection should move down one.

run ```refude.sh up``` and the selection should move up one.

Launching refude from a terminal is somewhat tedious, so you'll wan't to bind it to some keys. 
I have, in ```.config/openbox/rc.xml```:

```
    <keybind key="S-Super_L">
      <action name="Execute">
        <command>refude.sh up</command>
      </action>
    </keybind>
    <keybind key="Super_L">
      <action name="Execute">
        <command>refude.sh</command>
      </action>
    </keybind>

```

so when I hit the windows key, refude launches/opens the switcher/moves the selection down, and SHIFT+windows key moves the selection up.






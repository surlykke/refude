# Refude - Restful Desktop Environment

## What is it?

Refude is (going to be) a suite of applications that offer generic desktop funtionality. 

Right now the first one is being developed:

- **RefudeDefaultApp** - Set default applications for your desktop.

Future additions could be:

- A program starter
- An application menu
- A notification daemon
- A battery monitor
- A leaving application (like power-off, hibernate, suspend, restart, logout... -ish)

## What's the idea?

To implement common desktop functionality in a way thats useful irrespective of what desktop you run.

The idea of Refude is *not* to create a classical desktop: There wont be a desktop with widgets or a panel. Focus is on functionality that can be used in different desktop setups. 
I hope in particular that Refude's applications could be particularly useful if you run a lightweight setup with just a window manager (Fluxbox, Openbox,...).

## No really - what's the idea?

Well - I'm also the author of [RestfulIpc](https://github.com/surlykke/RestFulIpc) - 
a project to allow processes to communicate with each other via a restful protocol. 

Refude works together with the companion project 
[Refude Services](https://github.com/surlykke/refude-services). 
Refude-services offers restful services that handles application logic and interacts 
with the underlying system. Refude-services are restful and based on RestfulIpc.

Refude's application(s) call Refude-services to make stuff happen.

## What does the applications do?

### RefudeDefaultApps 

This application will show a list of known mimetypes, and allow you to set a default application for each mimetype, choosing from the applications you have installed on your system.


It works with mimetypes and desktopfiles as defined by freedesktop.org (FDO) 
(see [desktop entry specification](https://specifications.freedesktop.org/desktop-entry-spec/latest/index.html) 
and [mimetype associations](https://specifications.freedesktop.org/mime-apps-spec/latest/) and references therein). 

(*I plan to add implementations of xdg-open and xdg-mime that (unlike the implementations from FDO !) adhere to these standards. With that in place, you can open a document from your file manager or browser or some other program, and provided that program uses xdg-open to open the document, it should open in the application you chose.*)

## How is it implemented?

Refude written in angularjs as packaged chromium apps.

## Why packaged apps?

Partly because I wanted to learn angular, and partly because angularjs makes it possible to create rich UI's with very few lines of code. 

Also, the fact that chromium packaged apps are sandboxed and don't allow acces to the local file system, enforces the dicipline of keeping logic in the services.

## Install and run

You need to have the 'DesktopService' from Refude-services installed and running - see [Refude-services](https://github.com/surlykke/refude-services).

Somewhere in your filesystem where you'd like to install, do: 
```bash
git clone https://github.com/surlykke/refude.git
```
Then install by running the app in chrome:
```bash
chromium --silent-launch --load-and-launch-app=/path../to../refude/refudedefaultapps
```
This installs RefudeDefaultApps in chrome.

After this, RefudeDefaultApps should show up in your application menu or your programstarter, whatever you use, and you can start it from there

As you can see, this is very much work in progress. Right now (march 2016) search is not working, and mimetype icons neither. Hopefuly it will be sorted out soon.

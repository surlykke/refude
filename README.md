# Refude - Restful Desktop Environment

## What is it?

Refude is a desktop environment - or will be, sometime in the future - sorta. 
The idea is to create applications that offer functionality which is useful in a 
desktop environment. Examples of such functionality could be: Application menu, application starter, battery monitor, desktop notifications...

The idea of Refude is *not* to create a classical desktop: There wont be a desktop with widgets or a panel. Focus is on functionality that can be used in different desktop setups. For example you might want to use Refude applications if you run a lightweight setup with just a window manager (like Fluxbox, Openbox or something like that).

Right now I work on an an application to set default applications. 
It works with mimetypes and desktopfiles as defined by freedesktop.org 
(see [desktop entry specification](https://specifications.freedesktop.org/desktop-entry-spec/latest/index.html) 
and [mimetype associations](https://specifications.freedesktop.org/mime-apps-spec/latest/) and references therein). 

The application will show a list of known mimetypes, and allow you to set a default application for each mimetype, choosing from the applications you have installed on your system.

## How is it implemented?

Refude works together with the companion project [Refude Services](https://github.com/surlykke/refude-services). Refude-services offers restful services, handles application logic and interacts with the underlying system.

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
Then run the app in chrome:
```bash
chromium --silent-launch --load-and-launch-app=/path../to../refude/default-applications/
```


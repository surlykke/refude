Browser identification
======================
There is, afaik, no reliable way for a browser-extension to detect which
browser it is running in.
Therefore, for this extension to work, there needs to be, in this directory
a file called browserId.js, defining the desktop-application-id of the browser the
extension is loaded into.
So, before loading into, say, Chrome, do:

echo "export const browserId=google-chrome" > browserId.js

- or if loading into Firefox:

echo "export const browserId=firefox_firefox" > browserId.js

and so on...


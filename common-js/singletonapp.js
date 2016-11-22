const http = require('http')
const fs = require('fs')
const path = require('path')
const windowManager = require('./createwin.js')

let server;
let win;

let handleRequest = function (request, response) {
	try {
		win.restore();
		win.show();
		win.focus();
		response.writeHead(200);
		response.end();
	} 
	catch (err) {
		console.log(err);
	}
}

exports.singletonApp = function(socketPath, dirPath, windowOpts) {

	req = http.get({socketPath: socketPath, method: "POST"}, (response) => {
        response.resume();
        console.log("Already running. exit...")
        process.exit(0)
    }).on("error", (error) => {
        try {
            fs.unlinkSync(socketPath);
        }
        catch (err) {
        }

		server = http.createServer(handleRequest)
        server.listen(socketPath)
   
		let dirName = path.basename(dirPath)
		win = windowManager.createWin(dirName, windowOpts);
		win.on("closed", () => { win = null; });
		win.setMenu(null);
		win.loadURL(`file://${dirPath}/index.html`);
 	});

}




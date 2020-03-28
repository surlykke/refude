// Copyright (c) Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
import React from 'react';
import { T } from "../common/translate";
import { ItemList, makeModelAndController } from "../common/itemlist"
import { applicationRank, subscribe, devtools } from "../common/utils";
import { getUrl, postUrl, patchUrl, iconUrl } from "../common/monitor"

let gui = window.require('nw.gui');
let filePath = gui.App.argv[0];
let fileName = filePath.substring(filePath.lastIndexOf('/') + 1);
let mimetypeId = gui.App.argv[1];
let mimetype

const desktopapp = "application/vnd.org.refude.desktopapplication+json";

export default class AppChooser extends React.Component {
    constructor(props) {
        super(props);
        devtools();
        [this.model, this.controller] = makeModelAndController();
        this.mimetypeIdList = [mimetypeId];
        this.mimetypeComment = {}
        this.appMap = {};
        this.fetchMimetypes(0);
    }

    fetchMimetypes = (pos) => {
        if (pos < this.mimetypeIdList.length) {
            getUrl(`/mimetype/${this.mimetypeIdList[pos]}`, resp => {
                let mt = resp.data
                if (pos === 0) {
                    mimetype = mt
                }
                this.mimetypeComment[this.mimetypeIdList[pos]] = mt.Comment;
                mt.SubClassOf.filter(m => !this.mimetypeIdList.includes(m)).forEach(m => this.mimetypeIdList.push(m));
                this.fetchMimetypes(pos + 1)
            });
        } else {
            this.fetchApps()
        }
    };

    fetchApps = () => {
        let appTakesArg = a => a.Exec && (a.Exec.toLowerCase().indexOf("%f") > -1 || a.Exec.toLowerCase().indexOf("%u") > -1)

        this.mimetypeIdList.forEach(mimetypeId => this.controller.itemMap.set(mimetypeId, []));
        this.controller.itemMap.set("other", [])
        
        getUrl("/applications", resp => {
            let place = item => {
                for (let mimetypeId of this.mimetypeIdList) {
                    if (item.app.Mimetypes.includes(mimetypeId)) {
                        this.controller.itemMap.get(mimetypeId).push(item);
                        return
                    }
                }
                this.controller.itemMap.get("other").push(item);
            }

            resp.data.filter(a => appTakesArg(a)).forEach(app => {
                let item = {
                    url: app._self,
                    name: app.Name,
                    comment: app.Comment || '',
                    image: iconUrl(app.IconName),
                    app: app,
                    matchEmpty: true
                }

                place(item);
            });
            this.controller.update();
        });
    };

    keyDown = (event) => {
        let { key, ctrlKey, shiftKey, altKey, metaKey } = event;

        if (key === "Tab" && !ctrlKey && shiftKey && !altKey && !metaKey) this.controller.move(false);
        else if (key === "Tab" && !ctrlKey && !shiftKey && !altKey && !metaKey) this.controller.move(true);
        else if (key === "ArrowUp" && !ctrlKey && !shiftKey && !altKey && !metaKey) this.controller.move(false);
        else if (key === "ArrowDown" && !ctrlKey && !shiftKey && !altKey && !metaKey) this.controller.move(true);
        else if (key === "Enter" && !ctrlKey && !shiftKey && !altKey && !metaKey) this.activate();
        else if (key === " " && !ctrlKey && !shiftKey && !altKey && !metaKey) this.activate();
        else if (key === "Escape" && !ctrlKey && !shiftKey && !altKey && !metaKey) this.close();
        else {
            return;
        }
        event.preventDefault();
    };

    activate = (item) => {
        if (!item) {
            item = this.model.items[this.model.selectedIndex()];
        }
        if (item) {
            postUrl(item.url, response => { this.close() });
        }
    }

    launch = (app, always) => {
        console.log("Launching", app.Name)
        if (always) {
            patchUrl(mimetype._self, { DefaultApp: app.Id })

        }

        postUrl(app._self + "?arg=" + encodeURIComponent(filePath), resp => {
            gui.App.quit();
        });
    };

    cancel = () => {
        this.setState({ selected: undefined });
    };

    cancelOnEscape = (event) => {
        if (event.key === 'Escape') {
            this.cancel();
        }
    };

    render = () => {

        let headingStyle = {
            padding: "0.3em"
        };

        let outerStyle = {
            maxWidth: "300px",
            maxHeight: "300px",
            display: "flex",
            flexFlow: "column",
            paddingLeft: "0.3em",
        };

        let searchBoxStyle = {
            boxSizing: "border-box",
            paddingRight: "5px",
            width: "calc(100% - 16px)",
            marginTop: "4px",
            marginBottom: "6px",
        };

        let inputStyle = {
            width: "100%",
            height: "36px",
            borderRadius: "5px",
            outlineStyle: "none",
        };



        return <div onKeyDown={this.keyDown} style={outerStyle}>
            <div key="heading" style={headingStyle}>
                <span dangerouslySetInnerHTML={{ __html: T("Open &nbsp;<b>%0</b>&nbsp;with:", fileName) }} />
            </div>
            <div style={searchBoxStyle}>
                <input id="input"
                    style={inputStyle}
                    type="search"
                    onChange={e => this.controller.setTerm(e.target.value.toLowerCase())}
                    autoComplete="off"
                    autoFocus />
            </div>
            <ItemList key="itemlist" model={this.model} onClick={this.controller.select} onDoubleClick={this.activate} />
        </div>

    }
}

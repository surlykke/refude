// Copyright (c) 2015, 2016, 2017 Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
import React from 'react';
import {render} from 'react-dom';
import {doGet, doPost, doSearch} from '../common/http'
import {T} from "../common/translate";
import {ItemList} from "../common/itemlist"
import {PopUp} from "./popup";
import {WIN, watchWindowPositionAndSize, showWindowIfHidden, devtools} from "../common/nw";
import {TitleBar} from "../common/titlebar";

let gui = window.require('nw.gui');
let filePath = gui.App.argv[0];
let fileName = filePath.substring(filePath.lastIndexOf('/') + 1);
let mimetypeId = gui.App.argv[1];
let isUrl = mimetypeId.startsWith("x-scheme-handler");
const desktopapp = "application/vnd.org.refude.desktopapplication+json";

class AppChooser extends React.Component {
    constructor(props) {
        //devtools();
        super(props);
        this.mimeMap = new Map();
        this.apps = [];
        this.state = {
            items: new Map()
        };
        this.fetch([mimetypeId], 0);
        WIN.on('loaded', () => {
            showWindowIfHidden();
            watchWindowPositionAndSize();
        });
    }

    fetch = (queued, pos) => {
        if (pos < queued.length) {
            doGet({service: "desktop-service", path: `/mimetypes/${queued[pos]}`}).then(
                (resp) => {
                    queued.push(...resp.json.SubClassOf.filter(sub => !queued.includes(sub)));
                    this.mimeMap.set(queued[pos], resp.json);
                    this.fetch(queued, pos + 1);
                },
                (resp) => {
                    this.fetch(queued, pos + 1);
                }
            )
        } else {
            let appMap = new Map();
            doSearch("desktop-service", desktopapp, "r.Exec ~i '%f' or r.Exec ~i '%u'").then(
                resp => {
                    let foundMatches;
                    let apps = resp.json;
                    for (let [mimetypeId, mimetype] of this.mimeMap) {
                        let group = T("Applications that handle %0", mimetype.Comment);
                        appMap.set(group, []);
                        apps = apps.filter(app => {
                            if (app.Mimetypes.includes(mimetypeId)) {
                                foundMatches = true;
                                appMap.get(group).push(app);
                                return false;
                            } else {
                                return true;
                            }
                        });
                    }
                    let other = foundMatches ? T("Other applications") : "";
                    appMap.set(other, []);
                    apps.forEach(app => appMap.get(other).push(app));
                    this.setState({items: appMap});
                },
                resp => {
                    console.log("error resp:", resp);
                }
            )
        }
    };

    select = (item) => {
    };

    execute = (item) => {
        if (this.mimeMap.get(mimetypeId)) {
            this.setState({selected: item});
        } else {
            this.launch(item);
        }
    };

    dismiss = () => {
        gui.App.quit();
    };

    launch = (app, always) => {
        if (always) {
            doPost(this.mimeMap.get(mimetypeId), {defaultApp: app.Id}).then(
                (resp) => {
                    doPost(app, {arg: filePath}).then(resp => {
                        gui.App.quit();
                    });
                },
                (resp) => {
                    doPost(app, {arg: filePath}).then(resp => {
                        gui.App.quit();
                    });
                }
            );
        } else {
            doPost(app, {arg: filePath}).then(resp => {
                gui.App.quit();
            });
        }
    };

    cancel = () => {
        this.setState({selected: undefined});
    };

    cancelOnEscape = (event) => {
        if (event.key === 'Escape') {
            this.cancel();
        }
    };

    render = () => {
        let style = {
            display: "flex",
            flexFlow: "column",
            height: "100%"
        };

        let headingStyle = {
            padding: "0.3em"
        };


        let buttonBarStyle = {
            position: "absolute",
            bottom: "1em",
            right: "1em",
        };

        let buttonStyle = {
            backgroundColor: "white",
            borderRadius: "5px",
            border: "black solid 2px",
            marginLeft: "0.8em",
            height: "2em",
            boxShadow: "1px 1px 1px #888888",
        };


        return <div style={style}>
            {this.state.selected &&
            <PopUp key="popup" dismiss={this.cancelOnEscape}>
                <span dangerouslySetInnerHTML={{__html: T("Open files of type <b>%0</b> with <b>%1</b>?",
                                                          this.mimeMap.get(mimetypeId).Comment,
                                                          this.state.selected.Name)}}/>
                <div style={buttonBarStyle}>
                    <button style={buttonStyle} onClick={() => this.launch(this.state.selected, false)} autoFocus>{T("Just once")}</button>
                    <button style={buttonStyle} onClick={() => this.launch(this.state.selected, true)}>{T("Always")}</button>
                    <button style={buttonStyle} onClick={this.cancel}>{T("Cancel")}</button>
                </div>
            </PopUp>}

            <TitleBar key="titlebar"/>
            <div key="heading" style={headingStyle}>
                <span dangerouslySetInnerHTML={{__html: T("Open &nbsp;<b>%0</b>&nbsp;with:", fileName)}}/>
            </div>
            <ItemList key="itemlist"
                      items={this.state.items}
                      select={this.select}
                      execute={this.execute}
                      onDismiss={this.dismiss}
                      disabled={this.state.selected}/>
        </div>
    }
}

render(<AppChooser/>, document.getElementById('root'));

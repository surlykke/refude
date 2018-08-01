// Copyright (c) 2015, 2016, 2017 Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
import React from 'react';
import {render} from 'react-dom';
import {doGet, doPost, doSearch} from '../common/http'
import {PopUp} from "./popup";
import {ItemList, linkItems} from "../common/itemlist"
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
            items: []
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
            doSearch("desktop-service", desktopapp, "r.Exec ~i '%f' or r.Exec ~i '%u'").then(
                resp => {
                    let apps = resp.json;
                    for (let [mimetypeId, mimetype] of this.mimeMap) {
                        let remains = [];
                        apps.forEach(app => {
                            if (app.Mimetypes.includes(mimetypeId)) {
                                app.__group = `Applications that handle ${mimetype.Comment}`
                                this.apps.push(app);
                            } else {
                                remains.push(app);
                            }
                        });
                        apps = remains;
                    }
                    ;
                    apps.forEach(app => app.__group = "Other applications");
                    this.apps.push(...apps);
                    this.filter("");
                },
                resp => {
                    console.log("error resp:", resp);
                }
            )
        }
    };

    filter = (term) => {
        term = term.toUpperCase();
        let filteredApps = this.apps.filter(app => app.Name.toUpperCase().includes(term));
        linkItems(filteredApps);
        this.setState({items: filteredApps});
    };

    select = (item) => {
    };

    execute = (item) => {
        this.setState({selected: item})
    };

    dismiss = () => {
        gui.App.quit();
    }

    launch = (always) => {
        if (always) {
            doPost(this.mimeMap.get(mimetypeId), {defaultApp: this.state.selected.Id}).then(
                (resp) => {
                    doPost(this.state.selected, {arg: filePath}).then(resp => {
                        gui.App.quit();
                    });
                },
                (resp) => {
                    doPost(this.state.selected, {arg: filePath}).then(resp => {
                        gui.App.quit();
                    });
                }
            );
        } else {
            doPost(this.state.selected, {arg: filePath}).then(resp => {
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
        }

        return [
            this.state.selected &&
            <PopUp key="popup" dismiss={this.cancelOnEscape}>
                Open files of type <b>{this.mimeMap.get(mimetypeId).Comment}</b><br/>
                with <b>{this.state.selected.Name}</b>?
                <div style={buttonBarStyle}>
                    <button style={buttonStyle} onClick={() => this.launch(false)} autoFocus>Just once</button>
                    <button style={buttonStyle} onClick={() => this.launch(true)}>Always</button>
                    <button style={buttonStyle} onClick={this.cancel}>Cancel</button>
                </div>
            </PopUp>
            ,
            <TitleBar key="titlebar"/>,
            <div key="heading" style={headingStyle}>
                Open &nbsp;<b>{fileName}</b>&nbsp;with:
            </div>,
            <ItemList key="itemlist" items={this.state.items}
                      onTermChange={this.filter}
                      select={this.select}
                      execute={this.execute}
                      onDismiss={this.dismiss}
                      disabled={this.state.selected}/>
        ];
    }
}

render(<AppChooser/>, document.getElementById('root'));

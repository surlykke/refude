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
import {WIN, devtools, applicationRank, subscribe} from "../common/utils";

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
        this.mimetypeIdList = [mimetypeId];
        this.mimetypeComment = {}
        this.appMap = {};
        this.state = {items: []};
        this.term = ""
        this.fetchMimetypes(0);
    }

    componentDidMount = () => {
        subscribe("dismiss", () => gui.App.quit());
        subscribe("termChanged", (term) => {
            this.term = term.toLowerCase();
            this.filterAndSort();
        });
        subscribe("itemLaunched", (item) => this.setState({selected: item.app}));
    }

    fetchMimetypes = (pos) => {
        if (pos < this.mimetypeIdList.length) {
            doGet({service: "desktop-service", path: `/mimetypes/${this.mimetypeIdList[pos]}`}).then(resp => {
                    this.mimetypeComment[this.mimetypeIdList[pos]] = resp.json.Comment;
                    resp.json.SubClassOf.filter(m => !this.mimetypeIdList.includes(m)).forEach(m => this.mimetypeIdList.push(m));
                    this.fetchMimetypes(pos + 1);
                },
                (resp) => {
                    this.fetchMimetypes(pos + 1);
                }
            )
        } else {
            this.mimetypeIdList.push('other');
            console.log("mimetypeIdList:", this.mimetypeIdList);
            console.log("mimetypeComment:", this.mimetypeComment);
            this.fetchApps()
        }
    };

    fetchApps = () => {
        doSearch("desktop-service", desktopapp, "r.Exec ~i '%f' or r.Exec ~i '%u'").then(
            resp => {
                let items = [];
                let foundApps = [];
                for (let mimetypeId of this.mimetypeIdList) {
                    this.appMap[mimetypeId] = resp.json.filter(app => !foundApps.includes(app) && (mimetypeId === 'other' || app.Mimetypes.includes(mimetypeId)))
                    foundApps.push(...this.appMap[mimetypeId]);
                }
                this.filterAndSort();
            },
            resp => {
                console.log("error resp:", resp);
            }
        )
    };


    filterAndSort = () => {
        console.log("Sorting:", this.term);
        let items = [];
        for (let mimetypeId of this.mimetypeIdList) {
            this.appMap[mimetypeId].forEach(app => {
                app.__rank = applicationRank(app, this.term);
            });
            this.appMap[mimetypeId].filter(app => app.__rank < 1).sort((a1, a2) => a1.__rank - a2.__rank).forEach(app => {
                items.push({
                    group: mimetypeId === 'other' ? T("Other applications") : T("Applications that handle " + this.mimetypeComment[mimetypeId]),
                    url: app._self,
                    description: app.Name + (app.Comment ? ' - ' + app.Comment : ''),
                    iconName: app.IconName,
                    app: app
                })
            })
        }
        console.log("Set items:", items);
        this.setState({items: items});
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

        let overlayStyle = {
            position: "fixed",
            top: "0px",
            left: "0px",
            width: "100%",
            height: "100%",
            zAxis: "10",
            backgroundColor: "rgba(255, 255, 255, 0.7)",
        };

        let popupStyle = {
            position: "absolute",
            top: "3em",
            left: "20%",
            padding: "16px",
            height: "8em",
            width: "calc(60% - 36px)",
            backgroundColor: "rgba(255, 255, 255, 1)",
            borderRadius: "5px",
            border: "solid black 2px"
        };

        let comment = this.mimetypeComment[mimetypeId];
        let appName = this.state.selected ? this.state.selected.Name : "";

        return <div style={style}>
            {this.state.selected &&
            <div style={overlayStyle} onKeyDown={this.props.dismiss}>
                <div style={popupStyle}>
                    <span dangerouslySetInnerHTML={{__html: T("Open files of type <b>%0</b> with <b>%1</b>?", comment, appName)}}/>
                    <div style={buttonBarStyle}>
                        <button style={buttonStyle} onClick={() => this.launch(this.state.selected, false)} autoFocus>{T("Just once")}</button>
                        <button style={buttonStyle} onClick={() => this.launch(this.state.selected, true)}>{T("Always")}</button>
                        <button style={buttonStyle} onClick={this.cancel}>{T("Cancel")}</button>
                    </div>
                </div>
            </div>}

            <div key="heading" style={headingStyle}>
                <span dangerouslySetInnerHTML={{__html: T("Open &nbsp;<b>%0</b>&nbsp;with:", fileName)}}/>
            </div>
            <ItemList key="itemlist" items={this.state.items} disabled={this.state.selected}/>
        </div>
    }
}

render(<AppChooser/>, document.getElementById('root'));

// Copyright (c) 2015, 2016, 2017 Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
import React from 'react';
import {render} from 'react-dom';
import {doGet, doPost, iconServiceUrl} from '../common/utils'
import {ItemList} from "../common/itemlist"
import {Item} from "../common/item"
import {SearchBox} from "../common/searchbox"

let gui = window.require('nw.gui')
let appArgument = gui.App.argv[0]
let mimetypeId = gui.App.argv[1]
let isUrl = mimetypeId.startsWith("x-scheme-handler")

class AppChooser extends React.Component {
	constructor(props) {
		super(props)
		this.mimetypeIds = []
		this.mimetypes = new Map()
		this.state = {
			iconUrl: iconServiceUrl(["unknown"], 32),
			comment: mimetypeId,
			apps: [],
			searchTerm: "",
		}
	}

	componentDidMount() {
        // Get mimetypes, starting with the given recursing through SubClassOf relation
        // Once all is gotten, fetch apps and order according to what mimetypes they support
	    let helper = (pos) => {
            if (pos >= this.mimetypeIds.length) {
                this.update(this.state.searchTerm)
            } else {
                let id = this.mimetypeIds[pos];
                doGet("desktop-service", "/mimetypes/" + id).then(
                   mimetype => {
                       mimetype.IconUrl = iconServiceUrl([mimetype.IconName, mimetype.GenericIcon]);
                       this.mimetypes[id] = mimetype;
                       console.log("Consider mimetype:", mimetype);
                       mimetype.SubClassOf.filter(subId => this.mimetypeIds.indexOf(subId) < 0).forEach(this.mimetypeIds.push);
				       if (id === mimetypeId) {
                           this.setState({iconUrl: mimetype.IconUrl, comment: mimetype.Comment});
				       }
				       helper(pos + 1);
                   }
               )
            }
        };
	    this.mimetypeIds.push(mimetypeId);
	    helper(0);
	}

    update = (searchTerm)	=> {
	    doGet("desktop-service", "/search", {q: `Name ~i '${searchTerm}'`}).then( apps => {
	        apps = apps.filter(app => app.Exec.toUpperCase().includes("%F") || app.Exec.toUpperCase().includes("%U"))
	        apps.forEach(app => {
	            app.__order = this.mimetypeIds.length;
                app.group = "Other applications"
                for (let i = 0; i < this.mimetypeIds.length; i++) {
                    if (app.Mimetypes.includes(this.mimetypeIds[i])) {
                        app.__order = i;
                        app.group = "Applications that handle: '" + this.mimetypes[mimetypeId].Comment + "'";
                        break;
                    }
                }
            });

            apps.sort((app1, app2) => app1.__order - app2.__order);
            console.log("Setting apps:", apps);
            let selected = this.state.selected
            if (apps.length > 0) {
                if (! apps.find(item => item.Self === selected)) {
                    selected = apps[0].Self;
                }
            } else {
                selected = undefined;
            }
            console.log("selected:", selected);
            this.setState({selected: selected});
            this.setState({apps: apps});
	    });
    };

	select = (self) => {
		console.log("I select, self:", self);
		this.userHasMadeSelection = true
		this.setState({selected: self})
	};

	run = (self) => {
		this.select(self)
		let app = this.state.apps.find(app => app.Self === self)
		if (!app) return
        if (this.state.useAsDefault) {
		    let mimetype = this.mimetypes[mimetypeId];
		    doPost(this.mimetypes[mimetypeId], {defaultApp: app.Id});
        }
        doPost(app, {arg: appArgument}).then(resp => {
            gui.App.quit();
        });
	};

	onKeyDown = (event) => {
		console.log("onKeyDown, event:", event);
		let {key, ctrlKey, shiftKey, altKey, metaKey} = event
		if      (key === "Tab" && !ctrlKey &&  shiftKey && !altKey && !metaKey) this.move(false)
		else if (key === "Tab" && !ctrlKey && !shiftKey && !altKey && !metaKey) this.move(true)
		else if (key === "ArrowUp" && !ctrlKey && !shiftKey && !altKey && !metaKey) this.move(false)
		else if (key === "ArrowDown" && !ctrlKey && !shiftKey && !altKey && !metaKey) this.move(true)
		else if (key === "Enter" && !ctrlKey && !shiftKey && !altKey && !metaKey) this.run(this.state.selected)
		else if (key === " " && !ctrlKey && !shiftKey && !altKey && !metaKey) this.run(this.state.selected)
		else if (key === "Escape" && !ctrlKey && !shiftKey && !altKey && !metaKey) gui.App.quit()
		else return;
		//event.stopPropagation()
	};

	move = (down) => {
		let {apps, selected} = this.state;
		console.log("move, selected:", selected, ", apps:", apps);
		let index = apps.findIndex(app => app.Self === selected);
        console.log("index:", index);

		if (index > -1) {
			index = (index + apps.length + (down ? 1 : -1)) % apps.length;
			console.log(" - now:", index, ", selecting:", apps[index])
			this.select(apps[index].Self)
		}
	}

	onTermChange = (event) => {
		this.setState({searchTerm: event.target.value});
        this.update(event.target.value);
	}

	render = () => {
		let {iconUrl, comment, apps, selected, confirm} = this.state
		console.log("Render: selected:", selected);
		let styles = {
			content: {
				position: "relative",
				display: "flex",
				flexDirection: "column",
				boxSizing: "border-box",
				width: "100%",
				height: "100%",
				padding: "8px",
				//margin: "8px 8px 0px 8px",
			},
			heading: {
				marginBottom: "8px",
			},
			item: {
				marginBottom: "8px",
			},
			searchBox: {
				width: "calc(100% - 16px)",
				marginBottom: "3px",
			},
			list: {
				flex: "1",
				paddingBottom: "80px",
			},
			useAsDefault: {
				boxSizing: "border-box",
				width: "calc(100% - 16px)",
				paddingTop: "12px",
				paddingBottom: "8px",
				display: "flex",
				borderRadius: "5px",
				backgroundColor: "rgba(245,245,245,1)",
			},
		}

		let item = {
			IconUrl: iconUrl,
			Name: appArgument,
			Comment: comment
		}

		return (
			<div style={styles.content} onKeyDown={this.onKeyDown}>
				<div style={styles.heading}>Select an application to open:</div>
				<Item item={item} style={styles.item}/>
				<SearchBox style={styles.searchBox} onChange={this.onTermChange} searchTerm={this.state.searchTerm}/>
				<ItemList style={styles.list} items={apps} selectedSelf={selected} select={this.select} execute={this.run}/>
				<div style={{height: "8px"}}/>
				{	this.state.selected &&
					<div style={styles.useAsDefault}>
						<input id="checkbox"
							   type="checkbox"
							   value={this.state.useAsDefault}
							   onChange={(evt) => {this.setState({useAsDefault: evt.target.checked})}}/>
						<label htmlFor="checkbox" accessKey="M">
							Always use <em>{this.state.selected.Name}</em> to open {isUrl ? "urls" : "files"} of this type?
						</label>
					</div>
				}
			</div>
		)
	}
}

render(
	<AppChooser appArgument={gui.App.argv[0]} mimetypeId={gui.App.argv[1]}/>,
	document.getElementById('root')
);

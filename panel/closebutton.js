// Copyright (c) Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
import React from 'react'
import { ipcRenderer } from 'electron'

let quit = () => {
	ipcRenderer.send("panelClose")
}

export class CloseButton extends React.Component {
	render = () => 
		<div onClick={quit} className="clickable" title="Close refude">
			<svg width="20" height="20" viewBox="0 0 100 100">
				<g fillOpacity="0" strokeWidth="12" stroke="black">
					<line x1="15" y1="15" x2="85" y2="85"/>
					<line x1="15" y1="85" x2="85" y2="15"/>
				</g>
			</svg>
		</div>
}


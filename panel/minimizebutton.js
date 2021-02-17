// Copyright (c) Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
import React from 'react'
import { ipcRenderer } from 'electron'

let quit = () => {
	ipcRenderer.send("panelMinimize")
}

export class MinimizeButton extends React.Component {
	render = () =>
		<div onClick={quit} className="clickable" title="Hide panel for 5 seconds">
			<svg width="20" height="20" viewBox="0 0 100 100">
				<path d="M 0 30 C 30 10, 60 0, 55 90" stroke="black" strokeWidth="12" fill="none" />
				<path d="M 20 60 L 55 90 L 90 60" stroke="black" strokeWidth="14" fill="none" />
			</svg>
		</div>
}


// Copyright (c) Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
import React from 'react'
import './Panel.css'

class DragField extends React.Component {

    render = () => {
        return <div className="plugin drag">
            <svg width="14" height="14" viewBox="-50 -50 100 100" xmlns="http://www.w3.org/2000/svg">
                <defs>
                    <marker id="arrow" viewBox="0 0 12 12" refX="3" refY="6" markerWidth="5" markerHeight="5" orient="auto-start-reverse">
                        <path d="M 0 0 L 6 6 L 0 12 z" />
                    </marker>
                </defs>
                <line x1="-35" y1="0" x2="35" y2="0" fill="none" stroke="black" strokeWidth="8" markerStart="url(#arrow)" markerEnd="url(#arrow)" />
                <line x1="0" y1="-35" x2="0" y2="35" fill="none" stroke="black" strokeWidth="8" markerStart="url(#arrow)" markerEnd="url(#arrow)" />
            </svg>
        </div >
    }


}

export { DragField }

// Copyright (c) Christian Surlykke
//
// This file is part of the refude project. 
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
import React from 'react'

export let Flash = ({ flash }) => {
    return flash ? 
        <div id="flashDiv" className="flash">
            <div id="iconDiv" className="flash-icon">
                <img width="100%" height="100%" src={flash.icon} alt="" />
            </div>
            <div id="messageDiv" className="flash-message">
                <>
                    <div className="flash-title">
                        {flash.title}
                    </div>
                    <div className="flash-body">
                        {flash.comment}
                    </div>
                </>

            </div>
        </div> :
        null
}

import React from 'react'
import ReactDOM from 'react-dom'
import { Indicator } from './indicator'

console.log("Render Indicator into ", document.getElementById('indicator'))
ReactDOM.render(<Indicator/>, document.getElementById('indicator'))

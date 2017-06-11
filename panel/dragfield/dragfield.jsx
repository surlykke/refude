import React from 'react'

const circleCenters = [{x:30, y:20}, {x:30, y:50}, {x:30, y:80},
                       {x:70, y:20}, {x:70, y:50}, {x:70, y:80}]

class DragField extends React.Component {

	constructor(props) {
		super(props)
	}

	render = () =>
		<div className="panel-plugin dragfield">
			<svg viewBox="0 0 100 100" >
				<g fillOpacity="1">
					{circleCenters.map(c => (
						<circle key={"" + c.x + c.y} cx={c.x} cy={c.y} r="10" stroke="black" strokeWidth="3" fill="dark-gray" />
					))}
				</g>
			</svg>
		</div>
}

export {DragField}

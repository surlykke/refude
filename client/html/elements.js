// Copyright (c) Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.

export const div = (props, ...children) => React.createElement('div', props, ...children);
export const a = (props, ...children) => React.createElement('a', props, ...children);
export const span = (props, ...children) => React.createElement('span', props, ...children);
export const input = (props, ...children) => React.createElement('input', props, ...children);
export const hr = () => React.createElement('hr');
export const p = (props, ...children) => React.createElement('p', props, ...children);
export const img = (props, ...children) => React.createElement('img', props, ...children);
export const table = (props, ...children) => React.createElement('table', props, ...children);
export const tbody = (props, ...children) => React.createElement('tbody', props, ...children);
export const tr = (props, ...children) => React.createElement('tr', props, ...children);
export const td = (props, ...children) => React.createElement('td', props, ...children);
export const materialIcon = (name) => React.createElement('i', { className: 'material-icons', style: { color: 'light-grey' } }, name);
export const frag = (...children) => React.createElement(React.Fragment, null, ...children);

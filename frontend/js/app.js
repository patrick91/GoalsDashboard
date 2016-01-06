import React from 'react';
import ReactDOM from 'react-dom';

import FontFaceObserver from 'fontfaceobserver';

import Steps from './components/Steps';


import '../css/main.css';


// Observer loading of Source Sans Pro
const openSansObserver = new FontFaceObserver('Source Sans Pro', {});

// When Open Sans is loaded, add the js-open-sans-loaded class to the body
openSansObserver.check().then(() => {
    document.body.classList.add('js-source-sans-pro-loaded');
}, () => {
    document.body.classList.remove('js-source-sans-pro-loaded');
});


ReactDOM.render(
    <Steps />,
    document.querySelector('#app')
);

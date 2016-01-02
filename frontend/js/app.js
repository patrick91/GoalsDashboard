import React from 'react';
import ReactDOM from 'react-dom';

import FontFaceObserver from 'fontfaceobserver';


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
    <h1>Hello Goals</h1>,
    document.querySelector('#app')
);

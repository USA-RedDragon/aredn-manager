export function getWebsocketURI() {
  const loc = window.location;
  let newURI;
  if (loc.protocol === 'https:') {
    newURI = 'wss:';
  } else {
    newURI = 'ws:';
  }
  // nodejs development
  if (window.location.port == 5173) {
    // Change port to 3333
    newURI += '//' + loc.hostname + ':3333';
  } else {
    newURI += '//' + loc.host;
  }
  newURI += '/ws';
  console.log('Websocket URI: "' + newURI + '"');
  return newURI;
}

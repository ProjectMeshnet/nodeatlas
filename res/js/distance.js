var loc1;
var loc2;

function initDistance() {
	map.on('popupopen', onDistanceClick);
}

function onDistanceClick(e) {
	var loc = e.popup._source.getLatLng();
	if (typeof loc1 == 'undefined') {
		loc1 = loc;
	} else if (loc == loc1) {
		// TODO: create a real error message with
		//       the nice popup thing that has yet
		//       to be written yet.
		// You already picked this node.
		alert('You have already picked this node.');
	} else if (typeof loc2 == 'undefined') {
		loc2 = loc;
		map.off('popupopen', onDistanceClick);
	} 
}

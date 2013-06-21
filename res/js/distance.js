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
		alertDistance(findDistance(loc1, loc2));
	} 
}

function findDistance(loc1, loc2) {	
	var lat1 = loc1.lat;
	var lat2 = loc2.lat;
	var lon1 = loc1.lng;
	var lon2 = loc2.lng;
		
	var R = 6371; // km
	var dLat = (lat2-lat1).toRad();
	var dLon = (lon2-lon1).toRad();
	lat1 = lat1.toRad();
	lat2 = lat2.toRad();
		
	var a = Math.sin(dLat/2) * Math.sin(dLat/2) +
			Math.sin(dLon/2) * Math.sin(dLon/2) * Math.cos(lat1) * Math.cos(lat2); 
	var c = 2 * Math.atan2(Math.sqrt(a), Math.sqrt(1-a)); 
	var d = R * c;
	return d;
}

function alertDistance(distance) {
	alert(distance+'km');
}

Number.prototype.toRad = function() {
	return this * Math.PI / 180;
}
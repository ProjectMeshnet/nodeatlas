var loc1;
var loc2;

function initDistance() {
    map.on('popupopen', onDistanceClick);
}

function onDistanceClick(e) {
    var loc = e.popup._source.getLatLng();
    var popup = L.popup();
    popup.setLatLng(loc);
    popup.setContent('You have already picked this node.');
    if (typeof loc1 == 'undefined') {
	loc1 = loc;
    } else if (loc == loc1) {
	popup.addTo(map);
    } else if (typeof loc2 == 'undefined') {
	map.removeLayer(popup);
	loc2 = loc;
	map.off('popupopen', onDistanceClick);
	drawLine([loc1, loc2], loc1.distanceTo(loc2));
	loc1 = undefined;
	loc2 = undefined;
    }
}

function drawLine(points, distance) {
    var line = new L.Polyline(points, {
	color: '#000',
	weight: 8,
	opacity: 0.5,
	smoothFactor: 1
    });
    
    var popup = L.popup();
    popup.setLatLng(line.getBounds().getCenter());
    popup.setContent(distance.toMiles() + ' miles<br/>' + distance.toKilometers() + ' km');
    
    line.on('click', function() {
	map.removeLayer(this);
	map.removeLayer(popup);
    });
    
    line.on('mouseover', function() {
	popup.addTo(map);
    });
    
    line.addTo(map);
    popup.addTo(map);
}

Number.prototype.toMiles = function() {
    return parseFloat((this/1609.344).toFixed(2));
}

Number.prototype.toKilometers = function() {
    return parseFloat((this/1000).toFixed(2));
}
$(document).ready(function() {

	addJSFiles();
	
	$.ajaxSetup({cache:true});
	
	$('#addme').tooltip();
	$('#distance').tooltip();
	
	// this handles loading the child nodes now
	loadChildMaps();
		
	$('#search').keyup(function() {
		// TODO display these in some fashion
		var searchResults = search(nodes, $(this).val());
	});
	
	mapview();

	map.addControl(new L.Control.Scale());
	
	// If you're at /verify/xxx
	var key = verifying();
	if (key != '') {
		verifyNode(key);
	}
});

$(window).resize(function () {
	$('#map').css('height', ($(window).height()));
}).resize();

function geoLocate() {
	map.locate({setView: true, maxZoom: 17});
	map.on('locationfound', function() {
		map.off('locationfound');
	});
}

function loadChildMaps() {
	$.getJSON("/api/child_maps", function(response) {
		for (i in response.data) {
			mapObj = response.data[i]
			cachedMaps[mapObj.ID] = {
				"name": mapObj.Name,
				"hostname": mapObj.Hostname
			}
		}
		addNodes();
	});
}

function mapview() {
	// If the URL fragment is provided, it should be of the form
	// 'zoom/lat/lon'. If it is, use it to set the view.
	var view = window.location.hash.replace('#', '').split('/');
	if (view.length > 0) {
		map.setView([view[1], view[2]], view[0]);
	}
}

function nodexxx(node) {
	var path = window.location.pathname.split('/');
	if (path[1] != "node") return false;
	return (path[2] == node);
}

function verifying() {
	var path = window.location.pathname.split('/');
	if (path[1] != "verify") return '';
	else return path[2];
}

function onMapClick(e) {
	var markerLocation = new L.LatLng(e.latlng.lat, e.latlng.lng);
	var marker = new L.Marker(markerLocation, {icon: newUserIcon});
	newUser.clearLayers();
	newUser.addLayer(marker);
	if ($('#inputform').length > 0) {
		// only update lat/lng
		$('#latitude').val(e.latlng.lat.toFixed(6));
		$('#longitude').val(e.latlng.lng.toFixed(6));
	}
	else $('#wrap').append(getForm(e.latlng.lat.toFixed(6), e.latlng.lng.toFixed(6)));
	
	$('.node').remove();
	
	$('#inputform').fadeIn(500);
	
	$('#name').focus();
}

function addJSFiles() {
	var html = '<script type="text/javascript" src="/js/distance.js"></script>';
	html += '<script type="text/javascript" src="/js/icon.js"></script>';
	html += '<script type="text/javascript" src="/js/status.js"></script>';
	html += '<script type="text/javascript" src="/js/common.js"></script>';
	html += '<script type="text/javascript" src="/js/search.js"></script>';
	html += '<script type="text/javascript" src="/js/captcha.js"></script>';
	html += '<script type="text/javascript" src="/js/node.js"></script>';
	html += '<script type="text/javascript" src="/js/verify.js"></script>';
	html += '<script type="text/javascript" src="/js/form.js"></script>';
	html += '<script type="text/javascript" src="/js/layers.js"></script>';
	$('head').append(html);
}

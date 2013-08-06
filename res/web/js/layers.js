function addNodes() {
	$.ajax({
		type: "GET",
		url: "/api/all?geojson",
		dataType:"json",
		success: addLayers
	});
}

function addLayers(response) {
		allL(jQuery.extend(true, {}, response));
	
		activeL(jQuery.extend(true, {}, response));
		potentialL(jQuery.extend(true, {}, response));
		
		residentialL(jQuery.extend(true, {}, response));
		vpsL(jQuery.extend(true, {}, response));
		
		wirelessL(jQuery.extend(true, {}, response));
		internetL(jQuery.extend(true, {}, response));
		wiredL(jQuery.extend(true, {}, response));
		
		// Disable all layers on start except for
		// All Layers which shows everything
		allNodes();
}

function createMarker(feature, latlng) {
	var html = '<div class="node">';
	html +=  '<h4>'+feature.properties.OwnerName+'</h4><h4>';
	if (feature.properties.SourceID) {
		html += '<a href="'+cachedMaps[feature.properties.SourceID].hostname+'/node/'+feature.id+'" class="btn btn-mini btn-info" id="sendMessage">Message</a>';
	} else {
		html += '<button class="btn btn-mini btn-info" id="sendMessage">Message</button>';
	}
	html += '&nbsp;<button class="btn btn-mini btn-success" id="edit">Edit</button>';
	html += '&nbsp;<button class="btn btn-mini btn-warning" id="delete">Delete</button></h4>';
	html += '<div class="text-center"><a href="/node/'+feature.id+'" class="btn btn-small btn-primary">'+feature.id+'</a></div><hr>';
	
if (feature.properties.SourceID) {
	html += '<div class="property">Source</div>';
		if (cachedMaps[feature.properties.SourceID] != null) {
			sourceMap = cachedMaps[feature.properties.SourceID];
			html += '<div class="more">Retrieved from <a href="'+sourceMap.hostname+'/node/'+feature.id+'">'+(sourceMap.name ? sourceMap.name : 'another map')+'</a>.</div>';
		} else {
			html += '<div class="more">Retrieved from another map.</div>';
		}
	}
	
	if (feature.properties.Details) {
		html += '<div class="property">Details</div><div class="more">'+feature.properties.Details+'</div>';
	}
	if (feature.properties.Contact) {
		html += '<div class="property">Contact</div><div class="more">'+feature.properties.Contact+'</div>';
	}
	if (feature.properties.PGP) {
		html += '<div class="property">PGP</div><div class="more">'+(feature.properties.PGP).toUpperCase()+'</div>';
	}
	
	html += '</div>';
	
	var p = L.popup();
	p.setLatLng(latlng);
	p.setContent(html);
	
	
	// Use the status to set an appropriate icon and effects.
	var icon = inactiveNodeIcon;
	if (feature.properties.Status & STATUS_ACTIVE > 0) {
		icon = activeNodeIcon;
	}
		
	// If it's a VPS, show the VPS icon instead of the active/inactive icon
	if (!(feature.properties.Status & STATUS_PHYSICAL)) {
		icon = VPSIcon;
	}
	
	// Create the Marker with options set above.
	var m = L.marker(latlng, {icon: icon}).bindPopup(html);
	
	// If we have /node/xxx then center the map on it
	if (nodexxx(feature.id)) {
		map.setView(latlng, 8);
		nodeInfoClick(html, true);
	}
		
	return m;
}

function allL(response) {
	L.geoJson(response.data, {
		pointToLayer: createMarker
	}).addTo(all).on('click', nodeInfoClick);
}

function activeL(response) {
	var res = response.data, dats = [];
	for (var i in res.features) {
		var stat = res.features[i].properties.Status;
		if ((stat&STATUS_ACTIVE) > 0) dats[dats.length] = res.features[i];
	}
	res.features = dats;
	L.geoJson(res, {
		pointToLayer: createMarker
	}).addTo(active).on('click', nodeInfoClick);
}

function potentialL(response) {
	var res = response.data, dats = [];
	for (var i in res.features) {
		var stat = res.features[i].properties.Status;
		if ((stat&STATUS_ACTIVE) <= 0) dats[dats.length] = res.features[i];
	}
	res.features = dats;
	L.geoJson(res, {
		pointToLayer: createMarker
	}).addTo(potential).on('click', nodeInfoClick);
}

function wirelessL(response) {
	var res = response.data, dats = [];
	for (var i in res.features) {
		var stat = res.features[i].properties.Status;
		if ((stat&STATUS_WIRELESS) > 0) dats[dats.length] = res.features[i];
	}
	res.features = dats;
	L.geoJson(res, {
		pointToLayer: createMarker
	}).addTo(wireless).on('click', nodeInfoClick);
}

function residentialL(response) {
	var res = response.data, dats = [];
	for (var i in res.features) {
		var stat = res.features[i].properties.Status;
		if ((stat&STATUS_PHYSICAL) > 0) dats[dats.length] = res.features[i];
	}
	res.features = dats;
	L.geoJson(res, {
		pointToLayer: createMarker
	}).addTo(residential).on('click', nodeInfoClick);
}

function vpsL(response) {
	var res = response.data, dats = [];
	for (var i in res.features) {
		var stat = res.features[i].properties.Status;
		if ((stat&STATUS_PHYSICAL) <= 0) dats[dats.length] = res.features[i];
	}
	res.features = dats;
	L.geoJson(res, {
		pointToLayer: createMarker
	}).addTo(vps).on('click', nodeInfoClick);
}

function internetL(response) {
	var res = response.data, dats = [];
	for (var i in res.features) {
		var stat = res.features[i].properties.Status;
		if ((stat&STATUS_INTERNET) > 0) dats[dats.length] = res.features[i];
	}
	res.features = dats;
	L.geoJson(res, {
		pointToLayer: createMarker
	}).addTo(internet).on('click', nodeInfoClick);
}

function wiredL(response) {
	var res = response.data, dats = [];
	for (var i in res.features) {
		var stat = res.features[i].properties.Status;
		if ((stat&STATUS_WIRED) > 0) dats[dats.length] = res.features[i];
	}
	res.features = dats;
	L.geoJson(res, {
		pointToLayer: createMarker
	}).addTo(wired).on('click', nodeInfoClick);
}

function allNodes() {
	if ($('#all_l').hasClass('disabled')) {
		map.addLayer(all);
		$('#all_l').removeClass('disabled');
		map.removeLayer(active);
		$('#active_l').addClass('disabled');
		map.removeLayer(potential);
		$('#potential_l').addClass('disabled');
		map.removeLayer(residential);
		$('#residential_l').addClass('disabled');
		map.removeLayer(vps);
		$('#vps_l').addClass('disabled');
		map.removeLayer(internet);
		$('#internet_l').addClass('disabled');
		map.removeLayer(wireless);
		$('#wireless_l').addClass('disabled');
		map.removeLayer(wired);
		$('#wired_l').addClass('disabled');
	} else {
		map.removeLayer(all);
		$('#all_l').addClass('disabled');
	}
}

function activeNodes() {
	if ($('#active_l').hasClass('disabled')) {
		map.addLayer(active);
		$('#active_l').removeClass('disabled');
	} else {
		map.removeLayer(active);
		$('#active_l').addClass('disabled');
	}
}

function potentialNodes() {
	if ($('#potential_l').hasClass('disabled')) {
		map.addLayer(potential);
		$('#potential_l').removeClass('disabled');
	} else {
		map.removeLayer(potential);
		$('#potential_l').addClass('disabled');
	}
}

function residentialNodes() {
	if ($('#residential_l').hasClass('disabled')) {
		map.addLayer(residential);
		$('#residential_l').removeClass('disabled');
	} else {
		map.removeLayer(residential);
		$('#residential_l').addClass('disabled');
	}
}

function vpsNodes() {
	if ($('#vps_l').hasClass('disabled')) {
		map.addLayer(vps);
		$('#vps_l').removeClass('disabled');
	} else {
		map.removeLayer(vps);
		$('#vps_l').addClass('disabled');
	}
}

function internetNodes() {
	if ($('#internet_l').hasClass('disabled')) {
		map.addLayer(internet);
		$('#internet_l').removeClass('disabled');
	} else {
		map.removeLayer(internet);
		$('#internet_l').addClass('disabled');
	}
}

function wirelessNodes() {
	if ($('#wireless_l').hasClass('disabled')) {
		map.addLayer(wireless);
		$('#wireless_l').removeClass('disabled');
	} else {
		map.removeLayer(wireless);
		$('#wireless_l').addClass('disabled');
	}
}

function wiredNodes() {
	if ($('#wired_l').hasClass('disabled')) {
		map.addLayer(wired);
		$('#wired_l').removeClass('disabled');
	} else {
		map.removeLayer(wired);
		$('#wired_l').addClass('disabled');
	}
}
function addNodes() {
	$.getJSON("/api/all?geojson", function (response) {
		// Active nodes
		var res = response.data, dats = [];
		for (var i in res.features) {
			var stat = res.features[i].properties.Status;
			if ((stat&STATUS_ACTIVE) > 0) dats[dats.length] = res.features[i];
		}
		res.features = dats;
		L.geoJson(res, {
			pointToLayer: createMarker
		}).addTo(active).on('click', nodeInfoClick);
		
		// Potential nodes
		res = response.data, dats = [];
		for (var i in res.features) {
			var stat = res.features[i].properties.Status;
			if ((stat&STATUS_ACTIVE) <= 0) dats[dats.length] = res.features[i];
		}
		res.features = dats;
		L.geoJson(res, {
			pointToLayer: createMarker
		}).addTo(potential).on('click', nodeInfoClick);
		
		// Residential nodes
		res = response.data, dats = [];
		for (var i in res.features) {
			var stat = res.features[i].properties.Status;
			if ((stat&STATUS_PHYSICAL) > 0) { alert('hi'); dats[dats.length] = res.features[i]; }
		}
		res.features = dats;
		L.geoJson(res, {
			pointToLayer: createMarker
		}).addTo(residential).on('click', nodeInfoClick);
		
		// VPS nodes
		res = response.data, dats = [];
		for (var i in res.features) {
			var stat = res.features[i].properties.Status;
			if ((stat&STATUS_PHYSICAL) <= 0) dats[dats.length] = res.features[i];
		}
		res.features = dats;
		L.geoJson(res, {
			pointToLayer: createMarker
		}).addTo(vps).on('click', nodeInfoClick);
		
		// Internet nodes
		res = response.data, dats = [];
		for (var i in res.features) {
			var stat = res.features[i].properties.Status;
			if ((stat&STATUS_INTERNET) > 0) dats[dats.length] = res.features[i];
		}
		res.features = dats;
		L.geoJson(res, {
			pointToLayer: createMarker
		}).addTo(internet).on('click', nodeInfoClick);
		
		// Wireless nodes
		res = response.data, dats = [];
		for (var i in res.features) {
			var stat = res.features[i].properties.Status;
			if ((stat&STATUS_WIRELESS) > 0) dats[dats.length] = res.features[i];
		}
		res.features = dats;
		L.geoJson(res, {
			pointToLayer: createMarker
		}).addTo(wireless).on('click', nodeInfoClick);
		
		// Wired (eth) nodes
		res = response.data, dats = [];
		for (var i in res.features) {
			var stat = res.features[i].properties.Status;
			if ((stat&STATUS_WIRED) > 0) dats[dats.length] = res.features[i];
		}
		res.features = dats;
		L.geoJson(res, {
			pointToLayer: createMarker
		}).addTo(wired).on('click', nodeInfoClick);
		
	});
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

function activeNodes() {
	if ($('#active_l').hasClass('.disabled')) {
		map.addLayer(active);
		$('#active_l').removeClass('.disabled');
	} else {
		map.removeLayer(active);
		$('#active_l').addClass('.disabled');
	}
}

function potentialNodes() {
	
}

function residentialNodes() {
	
}

function vpsNodes() {
	
}

function internetNodes() {
	
}

function wirelessNodes() {
	if ($('#wireles_l').hasClass('.disabled')) {
		map.addLayer(wireless);
		$('#wireless_l').removeClass('.disabled');
	} else {
		map.removeLayer(wireless);
		$('#wireless_l').addClass('.disabled');
	}
}

function wiredNodes() {
	
}
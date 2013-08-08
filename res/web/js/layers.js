var nodes = [];
var statuses = [];

function addNodes() {
	$.ajax({
		type: "GET",
		url: "/api/all?geojson",
		dataType:"json",
		success: addLayers
	});
}

function addLayers(response) {
	// When we load the page, we want to add all of 
	// the layers to just the basic "all" layers
	L.geoJson(response.data, {
		pointToLayer: createMarker
	}).addTo(all).on('click', nodeInfoClick);
	// Now we also want to create the two arrays that
	// we have allocated at the top of the file.
	// `nodes` will contain an array of the [object Object] nodes
	// while `statuses` will contain an array of the int32
	// statuses for each corrisponding node.
	for (i in response.data.features) {
		nodes[i] = jQuery.extend(true, {}, response.data.features[i]);
		statuses[i] = nodes[i].properties.Status;
	}
}

function filterLayer() {
	var filter = getFilter();
	temp.clearLayers();
	var matches = [];
	var notFilter = getNotFilter();
	for (var i = 0; i < nodes.length; i++) {
		if ((filter&statuses[i]) == filter) {
			if (typeof notFilter == 'undefined') {
				matches[matches.length] = nodes[i];
			} else if ((notFilter|statuses[i]) == notFilter) {
				matches[matches.length] = nodes[i];
			}
		}
	}
	
	L.geoJson(matches, {
		pointToLayer: createMarker
	}).addTo(temp).on('click', nodeInfoClick);
}

function getNotFilter() {
	var notFilter;
	if (!($('#potential_l').hasClass('disabled'))) {
		if (!($('#vps_l').hasClass('disabled'))) {
			// Both Potential and VPS
			notFilter = ~01111111;
			notFilter ^= STATUS_PHYSICAL;
		} else {
			// Only potential
			notFilter = ~0;
			notFilter ^= STATUS_ACTIVE;
		}
	} else if (!($('#vps_l').hasClass('disabled'))) {
		// Only VPS
		notFilter = ~01111110;
		notFilter ^= STATUS_PHYSICAL;
	} 
	
	return notFilter;
}

function getFilter() {
	var active = 0, residential = 0, internet = 0, wireless = 0, wired = 0;
		
	if (!($("#active_l").hasClass('disabled'))) active = STATUS_ACTIVE;
	if (!($("#residential_l").hasClass('disabled'))) residential = STATUS_PHYSICAL;
	if (!($("#internet_l").hasClass('disabled'))) internet = STATUS_INTERNET;
	if (!($("#wireless_l").hasClass('disabled'))) wireless = STATUS_WIRELESS;
	if (!($("#wired_l").hasClass('disabled'))) wired = STATUS_WIRED;
	
	return (active|residential|internet|wireless|wired);
}

function onOff() {
	if ($('#all_l').hasClass('disabled')) {
		// Stuff on the on/off
		$('#all_l').removeClass('disabled');
		$('#all_l').html('On');
		// Other Stuff
		$('#layer_1, #active_l, #potential_l').removeClass('hidden');
	} else {
		// Stuff on the on/off
		$('#all_l').addClass('disabled');
		$('#all_l').html('Off');
		// Other Stuff
		$('#layer_1, #active_l, #potential_l').addClass('hidden disabled');
		$('#layer_2, #residential_l, #vps_l').addClass('hidden disabled');
		$('#layer_3, #internet_l, #wireless_l, #wired_l').addClass('hidden disabled');
		// Reset Filter
		map.removeLayer(temp);
		map.addLayer(all);
	}
}

function activeNodes() {
	map.removeLayer(all);
	if ($('#active_l').hasClass('disabled')) {
		if (!($('#potential_l').hasClass('disabled'))) {
			// If potential is already set, we want to
			// change it from potential to active, so
			// first we change some UI stuff.
			$('#potential_l').addClass('disabled');
		}
		$('#active_l').removeClass('disabled');
		$('#layer_2, #residential_l, #vps_l').removeClass('hidden');
		// Tell the filter to look for active nodes
		// and ignore potential nodes.
		filterLayer();
	} else {
		// Active is already set; do nothing.
	}
}

function potentialNodes() {
	map.removeLayer(all);
	if ($('#potential_l').hasClass('disabled')) {
		if (!($('#active_l').hasClass('disabled'))) {
			// If active is already set, we want to
			// change it from active to potential, so
			// first we change some UI stuff.
			$('#active_l').addClass('disabled');
		}
		$('#potential_l').removeClass('disabled');
		$('#layer_2, #residential_l, #vps_l').removeClass('hidden');
		// Tell the filter to look for potential nodes
		// and ignore potential nodes.
		filterLayer();
	} else {
		// Active is already set; do nothing.
	}
}

function residentialNodes() {
	if ($('#residential_l').hasClass('disabled')) {
		if (!($('#vps_l').hasClass('disabled'))) {
			// If residential is already set, we want to
			// change it from residential to vps, so
			// first we change some UI stuff.
			$('#vps_l').addClass('disabled');
		}
		$('#residential_l').removeClass('disabled');
		$('#layer_3, #internet_l, #wireless_l, #wired_l').removeClass('hidden');
		// Tell the filter to look for active nodes
		// and ignore potential nodes.
		filterLayer();
	} else {
		// Residential is already set; do nothing.
	}
}

function vpsNodes() {
	if ($('#vps_l').hasClass('disabled')) {
		if (!($('#residential_l').hasClass('disabled'))) {
			$('#residential_l').addClass('disabled');
		}
		$('#vps_l').removeClass('disabled');
		$('#layer_3, #internet_l, #wireless_l, #wired_l').removeClass('hidden');
		// Tell the filter to look for active nodes
		// and ignore potential nodes.
		filterLayer();
	} else {
		// Residential is already set; do nothing.
	}
}

function internetNodes() {
	if ($('#internet_l').hasClass('disabled')) {
		$('#internet_l').removeClass('disabled');
	} else {
		$('#internet_l').addClass('disabled');
	}
	filterLayer();
}

function wirelessNodes() {
	if ($('#wireless_l').hasClass('disabled')) {
		$('#wireless_l').removeClass('disabled');
	} else {
		$('#wireless_l').addClass('disabled');
	}
	filterLayer();
}

function wiredNodes() {
	if ($('#wired_l').hasClass('disabled')) {
		$('#wired_l').removeClass('disabled');
	} else {
		$('#wired_l').addClass('disabled');
	}
	filterLayer();
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
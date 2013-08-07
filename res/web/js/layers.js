var nodes = [];
var statuses = [];
var DEFAULT_FILTER;
// TODO: SET DEFAULT FILTER
var filter = DEFAULT_FILTER;

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
		statuses[i] = bit32Status(nodes[i].properties.Status);
	}
}

function bit32Status(s) {
	var status = '';
	// We want to take the regular status and turn it
	// into a binary number. The regular status is an
	// unsigned 32 int. For the most recent version of
	// the chart, view GitHub issue #56, otherwise you can
	// read the chart below. This will be updated accordingly
	//
	// https://github.com/ProjectMeshnet/nodeatlas/issues/56
	//
	//   _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _
	//  | <<    |       1       |       0        |
	//  |_ _ _ _|_ _ _ _ _ _ _ _|_ _ _ _ _ _ _ _ |
	//  | 0     | active        | planned        |
	//  | 1-6   |-----------reserved-------------|
	//  | 7     | physical      | vps            |
	//  | 8     | internet      | no internet    |
	//  | 9     | wireless      | no wireless    |
	//  | 10    | wired(eth)    | no wired(eth)  |
	//  | 11-15 |-----------reserved-------------|
	//  | 16-23 |-----------reserved-------------|
	//  | 24    | pingable      | down           |
	//  | 25-31 |-----------reserved-------------|
	//  |_ _ _ _|_ _ _ _ _ _ _ _|_ _ _ _ _ _ _ _ |
	// 
	
	status += ~~((s&STATUS_ACTIVE)>0);    // 0
	status += '000000';                   // 1-6 are reserved
	status += ~~((s&STATUS_PHYSICAL)>0);  // 7
	status += ~~((s&STATUS_INTERNET)>0);  // 8
	status += ~~((s&STATUS_WIRELESS)>0);  // 9
	status += ~~((s&STATUS_WIRED)>0);     // 10
	status += '00000';                    // 11-15 are reserved
	status += '00000000';                 // 16-23 are reserved
	status += ~~((s&STATUS_PINGABLE)>0);  // 24
	status += '0000000';                  // 25-31 are reserved
		
	return status;
}

function filterLayer() {
	
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
		$('#layer_1').addClass('hidden');
		$('#active_l').addClass('hidden');
		$('#potential_l').addClass('hidden');
		$('#layer_2').addClass('hidden');
		$('#residential_l').addClass('hidden');
		$('#vps_l').addClass('hidden');
		$('#layer_3').addClass('hidden');
		$('#internet_l').addClass('hidden');
		$('#wireless_l').addClass('hidden');
		$('#wired_l').addClass('hidden');
		// Reset Filter
		filter = DEFAULT_FILTER;
	}
}

function activeNodes() {
	if ($('#active_l').hasClass('disabled')) {
		if (!($('#potential_l').hasClass('disabled'))) {
			// If potential is already set, we want to
			// change it from potential to active, so
			// first we change some UI stuff.
			$('#potential_l').addClass('disabled');
		}
		$('#active_l').removeClass('disabled');
		// Tell the filter to look for active nodes
		// and ignore potential nodes.
	} else {
		
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
	
}

function wiredNodes() {
	
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
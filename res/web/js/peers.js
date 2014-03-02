function getConnections() {
    $.getJSON("/api/all_peers", function(data) {
	drawConnections(data);
    });
}

function drawMeshLink(points) {
    var line = new L.Polyline(points, {
        color: '#008',
        weight: 2,
        opacity: 0.2,
        smoothFactor: 1,
		clickable: false
    });

    line.addTo(map);
}
	
function drawConnections(connections) {
	// Get only the data section from the API call.
    var peerConns = connections["data"];

	// Loop through each item in the list of connections.
    for (var i = 0; i < peerConns.length; i++) {
		// Get nodeA and nodeB explicitly, to minimize the number of
		// lookups.
		var nodeA = nodesById[peerConns[i]["A"]];
		var nodeB = nodesById[peerConns[i]["B"]];
		
		// If both nodes are present, then pass their parameters to
		// drawMeshLink.
		if (nodeA && nodeB) {
			drawMeshLink(
				[L.latLng(nodeA.geometry.coordinates[1],
						  nodeA.geometry.coordinates[0]),
				 L.latLng(nodeB.geometry.coordinates[1],
						  nodeB.geometry.coordinates[0])]);
		}
    }
}

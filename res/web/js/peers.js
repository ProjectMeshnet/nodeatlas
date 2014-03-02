function getConnections() {
    $.getJSON("/api/all_peers", function(data) {
	drawConnections(data);
    });
}

function drawMeshLink(points, distance) {
    var line = new L.Polyline(points, {
        color: '#008',
        weight: 2,
        opacity: 0.2,
        smoothFactor: 1
    });

    line.on('click', function() {
        map.removeLayer(this);
    });

    line.addTo(map);
}
	
function drawConnections(connections) {
    var peerConns = [];
    for (var i = 0; i < connections["data"].length; i++) {
	var peerData = connections["data"][i];
	var peerA = peerData["Source"];
	for (var j = 0; j < peerData["Destinations"].length; j++) {
	    var peerB = peerData["Destinations"][j];
	    var cmp = peerA.localeCompare(peerB);
	    if (cmp < 0 && $.inArray([peerA, peerB], peerConns) < 0) {
		peerConns.push([peerA, peerB]);
	    } else if (cmp > 0 && $.inArray([peerB, peerA], peerConns) < 0) {
		peerConns.push([peerB, peerA]);
	    } // otherwise they're already connected or they're the same node
	}
    }

    for (var i = 0; i < peerConns.length; i++) {
	var nodeA = nodesById[peerConns[i][0]];
	var nodeB = nodesById[peerConns[i][1]];
	if (nodeA && nodeB) {
	    drawMeshLink(
		[L.latLng(nodeA.geometry.coordinates[1],
			  nodeA.geometry.coordinates[0]),
		 L.latLng(nodeB.geometry.coordinates[1],
			  nodeB.geometry.coordinates[0])],
		0
	    );
	}
    }
}

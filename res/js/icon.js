var NodeIcon = L.Icon.extend({
    options: {
        shadowUrl: '/res/img/marker-shadow.png',
    }
});

var activeNodeIcon = new NodeIcon({
	iconUrl: '/res/img/marker-icon.png',
	iconSize: [25, 41],
	iconAnchor: [15, 35],
	popupAnchor: [-1, -25]
});

var	inactiveNodeIcon = new NodeIcon({
	iconUrl: '/res/img/inactive-marker.png',
	iconSize: [20, 41],
	iconAnchor: [10, 35]
});

var newUserIcon = new NodeIcon({
	iconUrl: '/res/img/marker-icon_gray.png',
	iconSize: [25, 41],
	iconAnchor: [15, 35],
	popupAnchor: [-1, -25]
});
var NodeIcon = L.Icon.extend({
    options: {
        shadowUrl: '/img/marker-shadow.png',
    }
});

var activeNodeIcon = new NodeIcon({
	iconUrl: '/img/marker-icon.png',
	iconSize: [25, 41],
	iconAnchor: [15, 35],
	popupAnchor: [-1, -25]
});

var	inactiveNodeIcon = new NodeIcon({
	iconUrl: '/img/marker-icon_light_gray.png',
	iconSize: [25, 41],
	iconAnchor: [15, 35],
	popupAnchor: [-1, -25]
});

var newUserIcon = new NodeIcon({
	iconUrl: '/img/marker-icon_gray.png',
	iconSize: [25, 41],
	iconAnchor: [15, 35],
	popupAnchor: [-1, -25]
});

var VPSIcon = new NodeIcon({ 
	iconUrl: '/img/marker-icon_light_gray.png',
	iconSize: [25, 41],
	iconAnchor: [15, 35],
	popupAnchor: [-1, -25]
});
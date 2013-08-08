var NodeIcon = L.Icon.extend({
    options: {
        shadowUrl: '/img/shadow.png',
		iconSize: [25, 41],
		iconAnchor: [15, 35],
		popupAnchor: [-1, -25]
    }
});

var activeNodeIcon = new NodeIcon({
	iconUrl: '/img/node.png'
});

var	inactiveNodeIcon = new NodeIcon({
	iconUrl: '/img/inactive.png'
});

var newUserIcon = new NodeIcon({
	iconUrl: '/img/newUser.png'
});

var VPSIcon = new NodeIcon({ 
	iconUrl: '/img/vps.png'
});
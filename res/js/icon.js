var NodeIcon = L.Icon.extend({
    options: {
        shadowUrl: '/res/img/marker-shadow.png',
    }
});

var activeNodeIcon = new NodeIcon({
	iconUrl: '/res/img/marker-icon.png'
});
var	inactiveNodeIcon = new NodeIcon({
	iconUrl: '/res/img/inactive-marker.png',
	shadowAnchor: [2, 0]
});

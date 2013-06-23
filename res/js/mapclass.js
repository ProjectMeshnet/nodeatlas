var NodeMarker = L.Marker.Extend({
	icon: nodeIcon()
};

nodeMarker = function(latlng, options) {
	return new NodeMarker{latlng, options}
};

var NodeIcon = L.Icon.extend({
    options: {
		iconUrl: "/res/img/marker-icon.png",
        shadowUrl: '/res/img/marker-shadow.png',
        iconSize:     [38, 95],
        shadowSize:   [50, 64],
        iconAnchor:   [22, 94],
        shadowAnchor: [4, 62],
        popupAnchor:  [-3, -76]
    }
});

nodeIcon = function(options) {
	return new NodeIcon{options}
};

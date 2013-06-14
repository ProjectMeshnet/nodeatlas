var map, newUser, cloudmade, cloudmadeAttribution, cloudmadeUrl, users, firstLoad;

firstLoad = true;

cloudmadeAttribution = 'Map data &copy; 2011 OpenStreetMap contributors, Imagery &copy; 2011  CloudMade'; 
cloudmadeUrl = 'http://{s}.tile.cloudmade.com/64022233ce4f40c6acb6473ccdaec5b3/{styleId}/256/{z}/{x}/{y}.png',

users = new L.FeatureGroup();
users = new L.MarkerClusterGroup({spiderfyOnMaxZoom: true, showCoverageOnHover: false, zoomToBoundsOnClick: true});
newUser = new L.LayerGroup();

// GeoJSON layer
//geojson = new L.GeoJSON();

cloudmade = L.tileLayer(cloudmadeUrl, {styleId: 22677, attribution: cloudmadeAttribution});

map = new L.Map('map', {
    center: new L.LatLng(39.90973623453719, -93.69140625),
    zoom: 6,
    layers: [cloudmade, users, newUser]
});

// GeoLocation Control
function geoLocate() {
    map.locate({setView: true, maxZoom: 14});
}
var geolocControl = new L.control({
    position: 'topright'
});
geolocControl.onAdd = function (map) {
    var div = L.DomUtil.create('div', 'leaflet-control-zoom leaflet-control');
    div.innerHTML = '<a class="leaflet-control-geoloc" href="#" onclick="geoLocate(); return false;" title="My location"></a>';
    return div;
};

map.addControl(geolocControl);
map.addControl(new L.Control.Scale());

//map.locate({setView: true, maxZoom: 3});

$(document).ready(function() {
    $.ajaxSetup({cache:true});
    $('#map').css('height', ($(window).height() - 40));

    // Populate the map with nodes from /api/all.
    addNodes();
});

$(window).resize(function () {
    $('#map').css('height', ($(window).height() - 40));
}).resize();

function geoLocate() {
    map.locate({setView: true, maxZoom: 17});
}

function initRegistration() {
    map.addEventListener('click', onMapClick);
    $('#map').css('cursor', 'crosshair');
    return false;
}

function cancelRegistration() {
    newUser.clearLayers();
    $('#map').css('cursor', '');
    map.removeEventListener('click', onMapClick);
}

function addNodes() {
	$.getJSON("/api/all?geojson", function (response) {
		// TODO: Check for errors here (response.error)
		L.geoJSON(response.data, {
			onEachFeature: onEachNode
		}).addTo(map);
	});
}

function onEachNode(feature, layer) {
    // If the Feature properties include popupContent, display it.
    if (feature.properties && feature.properties.popupContent) {
        layer.bindPopup(feature.properties.popupContent);
    }
}

//not working yet as we're still finalizing the api - ds
function insertUser() {
    $("#loading-mask").show();
    $("#loading").show();
    var address = $("#address").val();
    var name = $("#name").val();
    var latitude = $("#latitude").val();
    var longitude = $("#longitude").val();
    if (name.length == 0) {
        alert("Name is required!");
        return false;
    }
    if (email.length == 0) {
        alert("Email is required!");
        return false;
    }
    var dataString = 'name='+ name + '&email=' + email + '&address=' + addr + '&latitude=' + latitude + '&longitude=' + longitude;
    $.ajax({
        type: "POST",
        url: "api/add",
        data: dataString,
        success: function() {
            cancelRegistration();
            users.clearLayers();
            getUsers();
            $("#loading-mask").hide();
            $("#loading").hide();
            $('#insertSuccessModal').modal('show');
        }
    });
    return false;
}

function onMapClick(e) {
    var markerLocation = new L.LatLng(e.latlng.lat, e.latlng.lng);
    var marker = new L.Marker(markerLocation);
    newUser.clearLayers();
    newUser.addLayer(marker);
    var form =  '<form id="inputform" enctype="multipart/form-data" class="well">'+
        '<label><strong>Name:</strong> <i>marker title</i></label>'+
        '<input type="text" class="input-medium" placeholder="Required" id="name" name="name" />'+
        '<label><strong>Email:</strong> <i>never shared</i></label>'+
        '<input type="text" class="input-medium" placeholder="Required" id="email" name="email" />'+
        '<label><strong>IPv6:</strong></label>'+
        '<input type="text" class="input-medium" id="addr" name="addr" placeholder="Optional" />'+
        '<input style="display: none;" type="text" id="lat" name="lat" value="'+e.latlng.lat.toFixed(6)+'" />'+
        '<input style="display: none;" type="text" id="lng" name="lng" value="'+e.latlng.lng.toFixed(6)+'" /><br><br>'+
        '<div class="row-fluid">'+
        '<div class="span6" style="text-align:center;"><button type="button" class="btn btn-small" onclick="cancelRegistration()">Cancel</button></div>'+
        '<div class="span6" style="text-align:center;"><button type="button" class="btn btn-small btn-primary" onclick="insertUser()">Submit</button></div>'+
        '</div>'+
        '</form>';
    marker.bindPopup(form).openPopup();
}

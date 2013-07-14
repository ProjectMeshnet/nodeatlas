function verifyNode() {
	// verifyNode submits POST /api/verify with the ID being the last
	// element of the current URL, (e.g. /verify/012345).
	var path = window.location.pathname.split( '/' );
	var id = path[ path.length - 1 ];

	// Once the ID has been retrieved, POST it to /api/verify.
	$.ajax({
		type: "GET",
		url: "/api/verify",
		data: { "id": id },
		success: function() {
			$('#status').text("Successful").hide().fadeIn(1000);
		},
		error: function(response) {
			$('#status').text("Unsuccessful").hide().fadeIn(1000);
		}
	});
}

verifyNode()

function verifyNode(key) {
	// verifyNode submits POST /api/verify with the ID being the last
	// element of the current URL, (e.g. /verify/012345).

	// Once the ID has been retrieved, POST it to /api/verify.
	$.ajax({
		type: "GET",
		url: "/api/verify",
		data: { "id": key },
		success: function() {
			//$('#status').text("Successful").hide().fadeIn(1000);
			alert('success');
		},
		error: function(response) {
			//$('#status').text("Unsuccessful").hide().fadeIn(1000);
			alert('error');
		}
	});
}
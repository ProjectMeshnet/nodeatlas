function verifyNode(id) {
	// verifyNode submits GET /api/verify with the ID being the last
	// element of the current URL, (e.g. /verify/012345). Once the ID
	// has been retrieved, GET it to /api/verify.
	$.ajax({
		type: "GET",
		url: "/api/verify",
		data: { "id": id },	
		success: function() {
			var success = '<div class="alert alert-success" id="alert"><strong>Success!</strong>&nbsp;';
			success += 'node verified';
			$('#wrap').append(success);
			setTimeout(function() {
				$('#alert').fadeOut(500, function() {
					$('#alert').remove();
					window.location.replace('/');
				});
			}, 1000);
		},
		error: function(data) {
			var error = '<div class="alert alert-danger" id="alert"><strong>Error:</strong>&nbsp;';
			error += JSON.parse(data.responseText).error+'</div>';
			$('#wrap').append(error);
			setTimeout(function() {
				$('#alert').fadeOut(500, function() {
					$('#alert').remove();
				});
			}, 3000);
		}
	});
}

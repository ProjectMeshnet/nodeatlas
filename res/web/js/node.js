function initRegistration() {
	$('#map').css('cursor', 'crosshair');
	map.addEventListener('click', onMapClick);
	return false;
}

function cancelRegistration() {
	newUser.clearLayers();
	$('#map').css('cursor', '');
	$('#inputform').fadeOut(500);
	map.removeEventListener('click', onMapClick);
}

function insertUser() {
	var address = $("#address").val();
	var name = $("#name").val();
	var email = $("#email").val();
	var details = $("#details").val();
	
	var pgp = $("#pgp").val();
	var contact = $("#contact").val();
	
	// ^[a-zA-Z0-9]{8}{0,2}$
	// TODO ^ USE REGEX TO CHECK PGP
	
	var latitude = $("#latitude").val();
	var longitude = $("#longitude").val();
	
	var active = 0, residential = 0, internet = 0, wireless = 0, wired = 0;
	
	if ($("#active").is(':checked')) {
		active = STATUS_ACTIVE;
	}
	
	if ($("#residential").is(':checked')) {
		residential = STATUS_PHYSICAL;
	}
	
	if ($("#internet").is(':checked')) {
		internet = STATUS_INTERNET;
	}
	
	if ($("#wireless").is(':checked')) {
		wireless = STATUS_INTERNET;
	}
	
	if ($("#wired").is(':checked')) {
		wired = STATUS_WIRED;
	}
	
	var status = active|residential|internet|wireless|wired;

	if (name.length == 0) {
		alert("Name is required!");
		return false;
	}
	
	if (email.length == 0) {
		alert("Email is required!");
		return false;
	}
	
	var dataObject = {
		'name': name,
		'email': email,
		'address': address,
		'latitude': latitude,
		'longitude': longitude,
		'status': status,
		'contact': contact,
		'details': details,
		'pgp': pgp
	};
	
	$.ajax({
		type: "POST",
		url: "/api/node",
		data: dataObject,
		success: function(response) {
			if (response.error == null) {
				cancelRegistration();
				nodelayer.clearLayers();
				addNodes();
				if (response.data == 'node registered')
					$('#insertSuccessModalNoVerify').modal('show');
				else
					$('#insertSuccessModal').modal('show');
			} else {
				if (response.error == 'addressInvalid') {
					alert("Unable to create node; the address you have entered is invalid.");
				} else {
					alert("Unable to create node: " + response.error);
				}
			}
		},
		error: function(data) {
			alert("Something went wrong! If you try again, it might work. If it doesn't, contact the admin from the About Page.");
			$("#loading-mask").hide();
			$("#loading").hide();
		}
	});
	return false;
}

function getForm(lat, lng) {
	var form =  '<form id="inputform" enctype="multipart/form-data">';
	form += '<div class="tabby">';
		form += '<div class="tab" id="one">';
			form += '<label><strong>Name</strong> <span class="desc">Marker title</span></label>';
			form += '<input type="text" class="input-medium" placeholder="Required" id="name" name="name" maxlength="255" />';
			form += '<label><strong>Email</strong> <span class="desc">Never shared</span></label>';
			form += '<input type="email" class="input-medium" placeholder="Required" id="email" name="email" />';
			form += '<label><strong>Address</strong> <span class="desc">Network-specific IP</span></label>';
			form += '<input type="text" class="input-medium" id="address" name="address" placeholder="Required" maxlength="39"/>';
			form += '<label><strong>Details</strong></label>';
			form += '<input type="text" class="input-medium" placeholder="Home, Work, ..." id="details" maxlength="255"/><br/>';
			form += '<div class="row"><div class="col col-lg-6" style="text-align:center;">';
			form += '<button class="btn btn-small" href="#" onclick="cancelRegistration(); return false;">Cancel</button></div>';
			form += '<div class="col col-lg-6" style="text-align:center;">';
			form += '<button class="btn btn-small btn-primary" href="#" onclick="next(2); return false;">Next</button></div></div>';
		form += '</div>';
		form += '<div class="tab" id="two">';
			form += '<p><strong>Status</strong><br><small>Check all that apply.</small></p>';
			form += '<label>';
					form += '<input type="checkbox" id="active"> ';
					form += 'Active node';
				form += '</label><br/>';
				form += '<label>';
					form += '<input type="checkbox" id="residential"> ';
					form += 'Residential node';
				form += '</label><br/><br/>';
			form += '<div class="row"><div class="col col-lg-6" style="text-align:center;">';
			form += '<button class="btn btn-small btn-primary" href="#" onclick="next(1); return false;">Back</button></div>';
			form += '<div class="col col-lg-6" style="text-align:center;">';
			form += '<button class="btn btn-small btn-primary" href="#" onclick="next(3); return false;">Next</button></div></div>';
		form += '</div>';
		form += '<div class="tab" id="three">';
			form += '<p><strong>Status</strong><br/><small>Check all that apply.</small></p>';
				form += '<label>';
					form += '<input type="checkbox" id="internet"> ';
					form += 'Internet access';
				form += '</label><br/>';
				form += '<label>';
					form += '<input type="checkbox" id="wireless"> ';
					form += 'Wireless access';
				form += '</label><br/>';
				form += '<label>';
					form += '<input type="checkbox" id="wired"> ';
					form += 'Wired (eth) access';
				form += '</label><br/><br/>';
			form += '<div class="row">';
			form += '<div class="col col-lg-6" style="text-align:center;">';
			form += '<button class="btn btn-small btn-primary" href="#" onclick="next(2); return false;">Back</button></div>';
			form += '<div class="col col-lg-6" style="text-align:center;">';
			form += '<button class="btn btn-small btn-primary" href="#" onclick="next(4); return false;">Next</button></div></div>';
		form += '</div>';
		form += '<div class="tab" id="four">';
			form += '<p><strong>PGP</strong><br/><small>8 digit or 16 digit.</small></p>';
			form += '<input type="text" class="input-medium" placeholder="CAFEBABE" id="pgp" name="pgp" maxlength="16" />';
			form += '<label><strong>Contact</strong></label>';
			form += '<textarea class="contact" id="contact" placeholder="XMPP username, Reddit username, ..." maxlength="255"></textarea><br/>';
			form += '<input style="display: none;" type="text" id="latitude" name="latitude" value="'+lat+'"/>';
			form += '<input style="display: none;" type="text" id="longitude" name="longitude" value="'+lng+'"/>';
			form += '<div class="row"><div class="col col-lg-6" style="text-align:center;">';
			form += '<button class="btn btn-small btn-primary" href="#" onclick="next(3); return false;">Back</button></div>';
			form += '<div class="col col-lg-6" style="text-align:center;">';
			form += '<button class="btn btn-small btn-info" href="#" onclick="insertUser(); return false;">Submit</button></div></div>';
		form += '</div>';
	form += '</div>';
	form += '</form>';
	return form;
}

function next(which) {
	if (which == 1) {
		$('#one').css('display', 'block');
		$('#two').css('display', 'none');
		$('#three').css('display', 'none');
		$('#four').css('display', 'none');
	} else if (which == 2) {
		$('#one').css('display', 'none');
		$('#two').css('display', 'block');
		$('#three').css('display', 'none');
		$('#four').css('display', 'none');
	} else if (which == 3) {
		$('#one').css('display', 'none');
		$('#two').css('display', 'none');
		$('#three').css('display', 'block');
		$('#four').css('display', 'none');
	} else if (which == 4) {
		$('#one').css('display', 'none');
		$('#two').css('display', 'none');
		$('#three').css('display', 'none');
		$('#four').css('display', 'block');
	}
}

function nodeInfoClick(e, on) {
	var html;
	$('.node').remove();
	$('#messageCreate').remove();
	if (!on) e.layer.closePopup();
	if (on) html = e;
	else html = e.layer._popup._content;
	$('#wrap').append(html);
	$('.node').hide(); 
	$('.node').fadeIn(500);
	var name = html.substring(html.indexOf('<h4>')+4, html.indexOf('&nbsp;'));
	ipv6 = html.substring(html.indexOf('a href')+14);
	ipv6 = ipv6.substring(0, ipv6.indexOf('"'));
	// EDIT NODE
	$('#edit').bind('click', function() {
		$('.node').fadeOut(500, function() {
			$('.node').remove();
			$('#wrap').append(getForm(e.layer.getLatLng().lat, e.layer.getLatLng().lng));
			// Now we want to set
			$.getJSON('/api/node?address='+ipv6, function(response) {
				$('#name').val(response.data.OwnerName);
				$('#email').prop('disabled', 'disabled');
				$('#email').val('Can\'t change.');
				$('#').val();
				$('#').val();
				$('#').val();
			});
			
			
			$('#inputform').fadeIn(500);
			$('#name').focus();
		});
	});
	// SEND MESSAGE
	$('#sendMessage').bind('click', function(e) {
		$('.node').fadeOut(500, function() {
			$('.node').remove();
			var form = createMessageForm(name, ipv6);
			$('#wrap').append(form);
			$('#cancelmessage').bind('click', function(e) {
				$('#messageCreate').fadeOut(500, function() {
					$('#messageCreate').remove();
				});
			});
			$('#nextpagesubmit').bind('click', function(e) {
				loadCAPTCHA($('#captcha_img'));
				$('#cancel').bind('click', function(e) {
					$('#messageCreate').fadeOut(500, function() {
						$('#messageCreate').remove();
					});
			});
			});
			$('#sendmessage').bind('click', function(e) {
				$('#messageCreate').fadeOut(500, function() {
					var from = $('#from').val();
					var address = $('#address').val();
					var subject = $('#subject').val();
					var message = $('#message').val();
					var captcha = $('#captcha').val();
					var key = $('#captcha_img').attr('src');
					key = key.substring(9, key.length-4);
					captcha = key + ':' + captcha;
					var msg = {
						'from': from,
						'address': address,
						'subject': subject,
						'message': message,
						'captcha': captcha
					};
					$.ajax({
						type: "POST",
						url: "/api/message",
						data: msg,
						success: function(response) {
							var success = '<div class="alert alert-success" id="alert"><strong>Success!</strong>&nbsp;';
							success += 'message sent';
							$('#wrap').append(success);
							setTimeout(function() {
								$('#alert').fadeOut(500, function() {
									$('#alert').remove();
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
					$('#messageCreate').remove();
				});
			});
			$('#messageCreate').hide();
			$('#messageCreate').fadeIn(500);
		});
	});
}

function createMessageForm(name, ipv6) {
	var html = '<div id="messageCreate">';
		html += '<div id="one">';
			html += '<h6>Send Message to '+name+'</h6>';
			html += '<input type="email" placeholder="Your Email" id="from" required>';
			html += '<input type="text" id="address" value="'+ipv6+'" disabled class="hidden">';
			html += '<br/><input type="text" placeholder="Subject" id="subject" required>';
			html += '<br/><textarea placeholder="Body" id="message" required></textarea>';
			html += '<br/><br/><div class="row"><div class="col col-lg-6" style="text-align:center;">';
			html += '<input type="reset" id="cancelmessage" class="btn btn-small" value="Cancel Message"></div>';
			html += '<div class="col col-lg-6" style="text-align:center;">';
			html += '<input type="submit" id="nextpagesubmit" class="btn btn-small btn-primary" value="Next Page" onclick="next(2); return false;"></div></div>';
		html += '</div><div id="two">';
			html += '<img id="captcha_img">';
			html += '<br/><input type="text" placeholder="Captcha" id="captcha" required>';
			html += '<br/><div class="row"><div class="col col-lg-6" style="text-align:center;">';
			html += '<input type="reset" id="cancel" class="btn btn-small" value="Cancel Message"></div>';
			html += '<div class="col col-lg-6" style="text-align:center;">';
			html += '<input type="submit" id="sendmessage" class="btn btn-small btn-primary" value="Send Message"></div></div>';
	html += '</div></div>';
	return html;
}

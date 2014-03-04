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

function addError(fadewhat, err) {
    $(fadewhat).fadeOut(500, function(){
	var error = '<div class="alert alert-danger" id="alert"><strong>Error:</strong>&nbsp;';
	error += err+'</div>';
	$('#wrap').append(error);
	setTimeout(function() {
	    $('#alert').fadeOut(500, function() {
		$('#alert').remove();
		$(fadewhat).fadeIn(500);
	    });
	}, 3000);
    });
}

function updateData(cmd) {

  var c = (cmd == 'wget') ? 'wget -qO-' : 'curl -s';
  var d = (cmd == 'wget') ? ' --post-data ' : ' -d ';

  $.getJSON('/api/token', function(token){
      var address = $("#address").val();
      var lat = $("#latitude").val();
      var lon = $("#longitude").val();
      var name = $("#name").val();
      var email = $("#email").val();
      var contact = $("#contact").val();
      var details = $("#details").val()
      var pgp = $("#pgp").val()
      var status = getSTATUS();
      token = token.data;

      var data = c + d + '"address=' + address + '"' + d + '"latitude=' + lat + '"' +
                 d + '"longitude=' + lon + '"' + d + '"name=' + name + '"' + d + '"email=' + email + '"' +
                 d + '"contact=' + contact + '"' + d + '"details=' + details + '"' +
                 d + '"pgp=' + pgp + '"' + d + '"status=' + status + '"' + d + '"token=' + token + '"' + ' "' +
                 window.location.protocol + '//' + window.location.host + '/api/node"';

      $('#'+cmd).val(data);
  });
}

function insertUser() {
    var address = $("#address").val();
    var name = $("#name").val();
    var email = $("#email").val();

    if (name.length == 0) {
	addError('#inputform', 'a name is required');
	return false;
    }

    if (email.length == 0) {
	addError('#inputform', 'an email is required');
	return false;
    }

    if (address.length == 0) {
	addError('#inputform', 'an address is required');
	return false;
    }

    $('#inputform').fadeOut(500, function() {
	$.getJSON('/api/token', function(token){
	    var dataObject = {
		'name': name,
		'email': email,
		'address': address,
		'latitude': $("#latitude").val(),
		'longitude': $("#longitude").val(),
		'status': getSTATUS(),
		'contact': $("#contact").val(),
		'details': $("#details").val(),
		'pgp': $("#pgp").val(),
		'token': token.data
	    };
	    $.ajax({
		type: "POST",
		url: "/api/node",
		data: dataObject,
		success: function(response) {
		    cancelRegistration();
		    var success = '<div class="alert alert-success" id="alert"><strong>Success!</strong>&nbsp;';
		    if (sendmail) {
			success += 'Please check your email for a verification link!';
		    } else {
			success += 'Your node has been added!';
		    }
		    $('#wrap').append(success);
		    setTimeout(function() {
			$('#alert').fadeOut(500, function() {
			    $('#alert').remove();
			    all.clearLayers();
			    addNodes();
			});
		    }, 3000);
		},
		error: function(data) {
		    var error = '<div class="alert alert-danger" id="alert"><strong>Error:</strong>&nbsp;';
		    error += JSON.parse(data.responseText).error+'</div>';
		    $('#wrap').append(error);
		    setTimeout(function() {
			$('#alert').fadeOut(500, function() {
			    $('#alert').remove();
			    $('#inputform').fadeIn(500);
			});
		    }, 3000);
		}
	    });
	});
    });
    return false;
}

function nodeInfoClick(e, on) {
    var m = $.extend(true, {}, e);
    var html;
    $('#inputform').remove();
    $('.node').remove();
    $('#messageCreate').remove();
    if (!on) e.layer.closePopup();
    if (on) html = e;
    else html = e.layer._popup._content;
    $('#wrap').append(html);
    $('.node').hide();
    $('.node').fadeIn(500);
    var name = html.substring(html.indexOf('<h4>')+4, html.indexOf('</h4>'));
    ipv6 = html.substring(html.indexOf('a href')+14);
    ipv6 = ipv6.substring(0, ipv6.indexOf('"'));

    // CLOSE NODE
    $('#closeNode').bind('click', function() {
	$('.node').fadeOut(500, function() {
	    $('.node').remove();
	});
    });

    // EDIT NODE
    $('#edit').bind('click', function() {
	edit(e, ipv6, m);
    });

    // SEND MESSAGE
    $('#sendMessage').bind('click', function() {
	message(name, ipv6);
    });
}

function getSTATUS() {
    var active = 0, residential = 0, internet = 0, wireless = 0, wired = 0;

    if ($("#active").is(':checked')) active = STATUS_ACTIVE;
    if ($("#residential").is(':checked')) residential = STATUS_PHYSICAL;
    if ($("#internet").is(':checked')) internet = STATUS_INTERNET;
    if ($("#wireless").is(':checked')) wireless = STATUS_WIRELESS;
    if ($("#wired").is(':checked')) wired = STATUS_WIRED;

    return (active|residential|internet|wireless|wired);
}

function message(name, ipv6) {
    $('.node').fadeOut(500, function() {
	$('.node').remove();
	var form = createMessageForm(name, ipv6);
	$('#wrap').append(form);
	$('#from').focus();
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
		$.getJSON('/api/token', function(token){
		    var msg = {
			'from': from,
			'address': address,
			'subject': subject,
			'message': message,
			'captcha': captcha,
			'token': token.data
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
				    $('#messageCreate').remove();
				});
			    }, 3000);
			},
			error: function(data) {
			    var error = '<div class="alert alert-danger" id="alert"><strong>Error:</strong>&nbsp;';
			    error += JSON.parse(data.responseText).error+'</div>';
			    $('#wrap').append(error);
			    setTimeout(function() {
				$('#alert').fadeOut(500, function() {
				    $('#alert').remove();
				    $('#messageCreate').fadeIn(500);
				});
			    }, 3000);
			}
		    });
		});
	    });
	});
	$('#messageCreate').hide();
	$('#messageCreate').fadeIn(500);
    });
}

function edit(e, ipv6, m) {
    $('.node').fadeOut(500, function() {
	$('.node').remove();
	$('#wrap').append(getForm(e.layer.getLatLng().lat, e.layer.getLatLng().lng));
	$('#submitatlas').prop('onclick', '');

	// Now we want to set shit that is already there.
	$.getJSON('/api/node?address='+ipv6, function(response) {
	    $('#name').val(response.data.OwnerName);
	    $('#email').prop('disabled', 'disabled');
	    $('#email').val('Can\'t change');
	    $('#address').val(response.data.Addr);
	    $('#address').prop('disabled', 'disabled');
	    $('#details').val(response.data.Details);
	    $('#pgp').val(response.data.PGP);
	    $('#contact').val(response.data.Contact);

	    var STATUS = response.data.Status;

	    if ((STATUS&STATUS_ACTIVE) > 0) $('#active').prop('checked', true);
	    else $('#active').prop('checked', false);

	    if ((STATUS&STATUS_PHYSICAL) > 0) $('#residential').prop('checked', true);
	    else $('#residential').prop('checked', false);

	    if ((STATUS&STATUS_INTERNET) > 0) $('#internet').prop('checked', true);
	    else $('#internet').prop('checked', false);

	    if ((STATUS&STATUS_WIRELESS) > 0) $('#wireless').prop('checked', true);
	    else $('#wireless').prop('checked', false);

	    if ((STATUS&STATUS_WIRED) > 0) $('#wired').prop('checked', true);
	    else $('#wired').prop('checked', false);

	    // Click submit
	    $('#submitatlas').bind('click', function() {
		$('#inputform').fadeOut(500);
		$.getJSON('/api/token', function(token){
		    var data = {
			'name': $("#name").val(),
			'address': $("#address").val(),
			'latitude': $("#latitude").val(),
			'longitude': $("#longitude").val(),
			'status': getSTATUS(),
			'contact': $("#contact").val(),
			'details': $("#details").val(),
			'pgp': $("#pgp").val(),
			'token': token.data
		    };
		    $.ajax({
			type: "POST",
			url: "/api/update_node",
			data: data,
			success: function(response) {
			    repNDone();
			    all.clearLayers();
			    addNodes();
			    var success = '<div class="alert alert-success" id="alert"><strong>Success!</strong>&nbsp;';
			    success += 'node updated';
			    $('#wrap').append(success);
			    setTimeout(function() {
				$('#alert').fadeOut(500, function() {
				    $('#alert').remove();
				});
			    }, 3000);
			},
			error: function(data) {
			    var error = '<div class="alert alert-danger" id="alert"><strong>Error:</strong>&nbsp;';
			    error += JSON.parse(data.responseText).error+'</div>';
			    $('#wrap').append(error);
			    setTimeout(function() {
				$('#alert').fadeOut(500, function() {
				    $('#alert').remove();
				    $('#inputform').fadeIn(500);
				});
			    }, 3000);
			}
		    });
		});
	    });
	});
	$('#inputform').fadeIn(500);
	$('#name').focus();

	// DELETE NODE
        $('#delete').bind('click', function() {
            deleteNode(ipv6);
        });

	// REPOSITON NODE
	$('#reposition').bind('click', function() {
	    repositionNode(m);
	});
    });
}

function repositionNode(m) {
    var newM = new L.Marker(m.layer._popup._source.getLatLng(), {icon: newUserIcon});
    newUser.clearLayers();
    newUser.addLayer(newM);
    map.removeLayer(m.layer);
    $('#map').css('cursor', 'crosshair');
    map.addEventListener('click', repNAdd);
}

function repNDone() {
    newUser.clearLayers();
    $('#map').css('cursor', '');
    $('#inputform').fadeOut(500);
    map.removeEventListener('click', repNAdd);
}

function repNAdd(e) {
    var newM = new L.Marker((new L.LatLng(e.latlng.lat, e.latlng.lng)), {icon: newUserIcon});
    newUser.clearLayers();
    newUser.addLayer(newM);
    $('#latitude').val(e.latlng.lat);
    $('#longitude').val(e.latlng.lng);
}

function deleteNode(ipv6) {
    $('#delete').prop('id', 'reallydelete');
    $('#reallydelete').removeClass('btn-warning');
    $('#reallydelete').addClass('btn-danger');
    $('#reallydelete').html('Are you sure?');
    $('#reallydelete').bind('click', function() {
	$('#inputform').fadeOut(500, function() {
	    $('#inputform').remove();
	    $.getJSON('/api/token', function(token){
		$.ajax({
		    type: "POST",
		    url: "/api/delete_node",
		    data: {address: ipv6, token: token.data},
		    success: function(response) {
			all.clearLayers();
			addNodes();
			var success = '<div class="alert alert-success" id="alert"><strong>Success!</strong>&nbsp;';
			success += 'node deleted';
			$('#wrap').append(success);
			setTimeout(function() {
			    $('#alert').fadeOut(500, function() {
				$('#alert').remove();
			    });
			}, 3000);
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
	    });
	});
    });
}

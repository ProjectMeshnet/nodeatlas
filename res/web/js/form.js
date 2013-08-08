function createMessageForm(name, ipv6) {
	var html = '<div id="messageCreate">';
		html += '<div id="one">';
			html += '<h6>Send Message to '+name+'</h6>';
			html += '<input type="email" placeholder="Your Email" id="from" maxlength="255">';
			html += '<input type="text" id="address" value="'+ipv6+'" disabled class="hidden">';
			html += '<br/><input type="text" placeholder="Subject" id="subject"  maxlength="255">';
			html += '<br/><textarea placeholder="Body" id="message"  maxlength="999"></textarea>';
			html += '<br/><br/><div class="row"><div class="col col-lg-6" style="text-align:center;">';
			html += '<input type="reset" id="cancelmessage" class="btn btn-small" value="Cancel Message"></div>';
			html += '<div class="col col-lg-6" style="text-align:center;">';
			html += '<input type="submit" id="nextpagesubmit" class="btn btn-small btn-primary" value="Next Page" onclick="next(2); return false;"></div></div>';
		html += '</div><div id="two">';
			html += '<img id="captcha_img">';
			html += '<br/><input type="text" placeholder="Captcha" id="captcha"  maxlength="255">';
			html += '<br/><div class="row"><div class="col col-lg-6" style="text-align:center;">';
			html += '<input type="reset" onclick="next(1); return false;" class="btn btn-small" value="Back"></div>';
			html += '<div class="col col-lg-6" style="text-align:center;">';
			html += '<input type="submit" id="sendmessage" class="btn btn-small btn-primary" value="Send Message"></div></div>';
	html += '</div></div>';
	return html;
}

function getForm(lat, lng) {
	$('#inputform').remove();
	var form =  '<div id="inputform">';
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
			form += '<button class="btn btn-small btn-info" onclick="insertUser(); return false;" id="submitatlas">Submit</button></div></div>';
		form += '</div>';
	form += '</div>';
	form += '</div>';
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
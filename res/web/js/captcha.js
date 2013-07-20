function loadCAPTCHA(element) {
	$.getJSON("/api/key",
		function(response) {
			if (response.data.length > 0) {
				element.data('captchakey', response.data)
				element.attr('src', '/captcha/'+response.data+'.png')
			}
		}
	)
}

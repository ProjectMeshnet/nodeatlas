// Common.js is a file that's mainly on every
// html page, and things that should happen on
// every page, javascript wise, should happen
// here so we don't reuse the same code over
// and over.

$(document).ready(function() {

	fixNavbarBrand();

});

function fixNavbarBrand() {
	// Function to check the height of the navbar
	// If it's too big, then decrease the font size
	// on the navbar-brand until we get a nice same
	// height on all navbars
	for (;;) {
		var wrong = $('.navbar-brand').css('height');
		var correct = $('.navbar-nav').css('height');
		wrong = parseFloat(wrong.substring(0, wrong.length-2));
		correct = parseFloat(correct.substring(0, correct.length-2));
		if (correct >= wrong) return;
		var size = $('.navbar-brand').css('font-size');
		size = parseFloat(size.substring(0, size.length-2));
		$('.navbar-brand').css('font-size', (--size) + 'px');		
	}
}

function hide(x) {
	$('#bringnavbarback').remove();
	$('.navbar').animate({top: '-500px'}, 1000);
	$('#wrap').append('<div id="bringnavbarback">Show</div>');
	$('#bringnavbarback').animate({top: '0px'}, 500);
	$('.leaflet-top').animate({top: '0px'}, 500);
	$('#bringnavbarback').bind('click', function() {
		$('.navbar').animate({top: '0px'}, 500);
		$('.leaflet-top').animate({top: '45px'}, 500);
		$('#bringnavbarback').animate({top: '-500px'}, 1000);
	});
}
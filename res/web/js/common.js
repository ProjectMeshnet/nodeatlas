// Common.js is a file that's mainly on every html page, and things
// that should happen on every page, javascript wise, should happen
// here so we don't reuse the same code over and over.

$(document).ready(function() {
    
    fixNavbarBrand();
    if (readonly) {
	addDBWarning();
    }
    if (global) {
	GlobalMap();
    }
    
});

function fixNavbarBrand() {
    // Function to check the height of the navbar if it's too big,
    // then decrease the font size on the navbar-brand until we get a
    // nice same height on all navbars
    if ($('.container').css('max-width') != 'none') for (;;) {
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
    $('.navbar').fadeOut(500, function() {
	$('#wrap').append('<div id="bringnavbarback">Show</div>');
	$('#bringnavbarback').fadeIn(500, function() {
	    $('#bringnavbarback').bind('click', function() {
		$('#bringnavbarback').fadeOut(500, function() {
		    $('.navbar').fadeIn(500);
		});
	    });
	});
    });
}

// If `.Database.VerifyDisabled` is set to true then we want to add a
// warning at the top
function addDBWarning() {
    var warning = '<div class="alert alert-danger" id="alert-left">Database is in read only mode.</div>';
    $('#wrap').append(warning);
}


// If the map is a Global Map, we have to do some things differently
// than if it was a normal map.
function GlobalMap() {
    // Create meshlocal button
    var ml = '<li><a href="/meshlocals/">MeshLocals</a></li>';
    $('.nav navbar-nav').append(ml);
}
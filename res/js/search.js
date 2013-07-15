// crawls the given list of nodes for query
// and returns the nodes that 
function search(nodes, query) {
	query = $.trim(query);
	var results = new Array();
	for (var i in nodes) {
		var curNode = nodes[i];
		// relevance is how relevant the result is crazy right
		curNode.relevance = 0;
		
		// highlights is a map of property names to:
		// arrays of 2-element arrays representing start and stop highlight points
		// booleans, with true meaning it is all highlighted and false meaning no match
		curNode.highlights = {};
		if (matchNode(curNode, query)) {
			results.push(curNode);
		}
	}
	results.sort(function(a, b) {
		return b.relevance - a.relevance;
	});
	return results;
}

function matchNode(node, search) {
	var rel = (matchAddr(node, search) +
			   matchDetails(node, search)
			  ) / 2.0;
	node.relevance = rel;
	return rel;
}

function matchAddr(node, search) {
	var match = search.indexOf(node.id) > -1;
	if (match) {
		window.location = '/node/' + node.id;
	}
	return match ? 1.0 : 0.0;
}

function matchDetails(node, search) {
	if (!node.properties.Details)
		return 0.0;

	node.highlights.Details = false;

	var desc = node.properties.Details;
	var terms = search.split(" ");
	var searchSpace = desc.split(" ").length;
	var matches = 0;

	for (var t in terms) {
		var index = desc.indexOf(terms[t]);
		if (index > -1) {
			if (!node.highlights.Details) {
				node.highlights.Details = new Array();
			}
			node.highlights.Details.push([index, index + terms[t].length - 1]);
			matches++;
		}
	}
	// may want to calculate this better/differently?
	if ((terms * searchSpace) == 0)
		return 0.0;
	var res = 1.0 * matches * matches / (terms * searchSpace);
	return res;
}
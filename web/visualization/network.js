'use strict'

var width = 1000,
    height = 1000,
    links = [],
    nodes = [],
    nodeset = new Set(),
    linkset = new Set(),
    lastTotal = 0;

var simulation = d3.forceSimulation()
    .force("link", d3.forceLink().id(function(d) { return d.id; }).strength(.2))
    .force("charge", d3.forceManyBody().strength(-100))
    .force("r", d3.forceRadial(Math.min(width, height)/3, width/2, height/2).strength(.2));
    
var draw = () => {

  var color = d3.scaleOrdinal(d3.schemeCategory20);

  d3.json("convos.json", function(error, convos) {
    if (error) throw error;

    Object.entries(convos).forEach((pair) => {
      if (!nodeset.has(pair[1].Source)) {
        nodeset.add(pair[1].Source);
        nodes.push({id: pair[1].Source, group: 0})
      }
      if (!nodeset.has(pair[1].Destination)) {
        nodeset.add(pair[1].Destination);
        nodes.push({id: pair[1].Destination, group: 0})
      }
      if (!linkset.has(pair[0])) {
        linkset.add(pair[0]);
        links.push({
          id: pair[0],
          source: pair[1].Source,
          target: pair[1].Destination,
          value: -1
        });
      }
    });

    console.log(nodes.length, links.length)

    var link = d3.select("#links")
      .selectAll("line");
    link
      .data(links, (d) => d.id)
      // .exit().remove()
      .enter().append("line")
        .attr("data-id", (d) => d.id)
        .attr("stroke-width", function(d) { return Math.sqrt(d.value); });

    var node = d3.select("#nodes")
      .selectAll("circle");
    node
      .data(nodes, (d) => d.id)
      // .exit().remove()
      .enter().append("circle")
        .attr("r", 10)
        .attr("data-id", (d) => d.id)
        .attr("fill", function(d) { return color(d.group); })
        .call(d3.drag()
            .on("start", dragstarted)
            .on("drag", dragged)
            .on("end", dragended));

    node.append("title")
        .text(function(d) { return d.id; });

    simulation
        .nodes(nodes)
        .on("tick", ticked);

    simulation.force("link")
        .links(links);

    if (nodes.length + links.length > lastTotal) {
      simulation.alpha(1).restart();
      lastTotal = nodes.length + links.length;
    }

    function ticked() {
      link
          .attr("x1", function(d) { return d.source.x; })
          .attr("y1", function(d) { return d.source.y; })
          .attr("x2", function(d) { return d.target.x; })
          .attr("y2", function(d) { return d.target.y; });

      node
          .attr("cx", function(d) { return d.x; })
          .attr("cy", function(d) { return d.y; });
    }


  });

  function dragstarted(d) {
    if (!d3.event.active) simulation.alphaTarget(0.3).restart();
    d.fx = d.x;
    d.fy = d.y;
  }

  function dragged(d) {
    d.fx = d3.event.x;
    d.fy = d3.event.y;
  }

  function dragended(d) {
    if (!d3.event.active) simulation.alphaTarget(0);
    d.fx = null;
    d.fy = null;
  }
}
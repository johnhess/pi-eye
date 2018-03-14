'use strict'
/**
Utilities to render the bezier connection histogram.
*/

const HEIGHT = 42;
const RADIUS = 5;
const RADIUS_STROKE = RADIUS + 1;

// TODO: Better detection, possibly at backend.
let isInternal = (host) => {
  return (host.indexOf("192") == 0 || host.indexOf("224") == 0 || host.indexOf("172") == 0 || host.indexOf('.local') != -1);
};

let renderHists = (hists) => {
  const hostTraffic = getHostsWithTraffic(hists);
  sparks(hostTraffic);
  beziers(hists);
}

let pushOrStart = (key, element, map) => {
  if (!map.has(key)) {
    map.set(key, []);
  }
  map.set(key, map.get(key).concat([element]));
};

let sumConversations = (hosts) => {
  const summedHosts = new Map;
  hosts.forEach((trafficSets, host) => {
    const summedTraffic = [];
    trafficSets.forEach((trafficSet) => {
      // the earliest, longest histograms are guaranteed to come first
      const startingSummedChunk = summedTraffic.findIndex(
        (summedChunk) => trafficSet[0].Timestamp == summedChunk.Timestamp
      );
      trafficSet.forEach((chunk, index) => {
        const sumIndex = startingSummedChunk + index;
        if (startingSummedChunk == -1 || sumIndex >= summedTraffic.length) {
          summedTraffic.push(chunk);
        } else {
          summedTraffic[sumIndex].Count += chunk.Count;
        }
      });
    });
    // Ignore empty traffic histories
    const totalTraffic = summedTraffic.reduce((sum, chunk) => sum + chunk.Count, 0) 
    if (totalTraffic) {
      summedHosts.set(host, summedTraffic);
    }
  });
  return summedHosts;
};

let getHostsWithTraffic = (hists) => {
  // host: [[...], [...]] (inner arrays are traffic for each matching)
  const internalHosts = new Map();
  const externalHosts = new Map();

  hists.forEach((convo) => {
    if (isInternal(convo.Source) && !isInternal(convo.Destination)) {
      pushOrStart(convo.Source, convo.Traffic, internalHosts)
      pushOrStart(convo.Destination, convo.Traffic, externalHosts)
    } else if (!isInternal(convo.Source) && isInternal(convo.Destination)) {
      pushOrStart(convo.Source, convo.Traffic, externalHosts)
      pushOrStart(convo.Destination, convo.Traffic, internalHosts)
    } else {
      // unclear conversation. don't plot.
    }
  });

  return {
    internal: sumConversations(internalHosts),
    external: sumConversations(externalHosts)
  }
};

let addHist = (name, traffic, maxTraf, minTime, elemSelector, faviconUrl) => {
  const jElem = $(elemSelector);
  const labelElem = $(elemSelector + '-labels');
  const width = jElem.width();
  const x = d3.scale.linear().range([0, width]);
  const y = d3.scale.linear().range([HEIGHT / 2, 0]);
  const faviconOffset = HEIGHT / 2 + 2;
  const line = d3.svg.line()
   .interpolate("basis")
   .x((d) => x(d.Timestamp))
   .y((d) => y(Math.log(d.Count + 1)));

  // Domain might be better defined as global max and min
  x.domain([minTime, traffic[traffic.length - 1].Timestamp]);
  y.domain([0, Math.log(maxTraf + 1)]);
  let elem = d3.select(elemSelector)
    .append('svg')
  let svg = elem
    .attr('width', width)
    .attr('height', HEIGHT)
    .attr('data-host', name)
    .append('g');
  svg.append('path')
    .datum(traffic)
    .attr('class', 'sparkline')
    .attr('d', line);
  svg.append('text')
    .attr('x', 18)
    .attr('y', faviconOffset + 12)
    .attr('class', 'device-label')
    .html(name);
  svg.append('rect')
    .attr('y', faviconOffset)
    .attr('height', 16)
    .attr('width', 16)
    .attr('class', 'host-favicon-rect')
  svg.append('image')
    .attr('y', faviconOffset)
    .attr('class', 'host-favicon')
    // Google runs a public favicon pulling service
    .attr('href', faviconUrl || `https:\/\/www.google.com/s2/favicons?domain=${name}`);
}

let sparks = (hostTraffic) => {
  // Find global traffic max
  let maxTraf = 0;
  const firstTrafficChunk = hostTraffic.internal.values().next().value;
  let minTime = firstTrafficChunk ? firstTrafficChunk[0].Timestamp : 0;
  hostTraffic.internal.forEach((traffic) => {
    traffic.forEach((chunk) => {
      maxTraf = Math.max(maxTraf, chunk.Count);
    });
  });
  hostTraffic.external.forEach((traffic) => {
    traffic.forEach((chunk) => {
      maxTraf = Math.max(maxTraf, chunk.Count);
    });
  });

  // add histograms for each host
  $('#local').empty();
  hostTraffic.internal.forEach((traffic, host, map) => {
    // TODO: Credit Gregor on the landing page.
    addHist(host, traffic, maxTraf, minTime, '#local', '/gregor_cresnar_device.png');
  });
  $('#external').empty();
  hostTraffic.external.forEach((traffic, host, map) => {
    addHist(host, traffic, maxTraf, minTime, '#external');
  })
}

let beziers = (hists) => {
  $('#beziers').empty();
  // draw a bezier for each conversation
  hists.forEach((convo) => {
    let leftHost, rightHost;
    if (isInternal(convo.Source) && !isInternal(convo.Destination)) {
      leftHost = convo.Source;
      rightHost = convo.Destination;
    } else if (!isInternal(convo.Source) && isInternal(convo.Destination)) {
      leftHost = convo.Destination;
      rightHost = convo.Source;
    } else {
      return
    }

    // Using made up constants, gradually fade the traffic out.
    let weight = 1;
    const minOpacity = .05;
    const recentTraffic = convo.Traffic.slice(convo.Traffic.length - 10)
      .reduce((sum, chunk) => {
        weight = weight * 2;
        return sum + chunk.Count * weight;
      }, 0);
    const logAdjustedTraffic = Math.log(recentTraffic + 1)/60;
    const blue = logAdjustedTraffic ? 
      Math.min(255, parseInt(recentTraffic * 1)) : 0;
    const stroke = logAdjustedTraffic ? 
      `rgba(0, 0, ${blue}, .5)` : 
      '#BBB';

    // assume local and external have same vertical offset
    const containerOffset = $('#local').offset().top;
    const leftDiv = $(`[data-host='${leftHost}']`);
    const rightDiv = $(`[data-host='${rightHost}']`);

    if (leftDiv.offset() && rightDiv.offset()) {
        const startX = leftDiv.offset().left + leftDiv.width();
        const startY = leftDiv.offset().top + HEIGHT / 2 - containerOffset;
        const endX = rightDiv.offset().left;
        const endY = rightDiv.offset().top + HEIGHT / 2 - containerOffset;
        const bezWidth = $('#beziers').width();
        const fullWidth = bezWidth;
        const halfWidth = bezWidth / 2;
        let bez = (point) => {
          return `M ${RADIUS}, ${point[0].y} C ${halfWidth}, ${point[0].y}, ${halfWidth}, ${point[1].y}, ${fullWidth - RADIUS}, ${point[1].y}`;
        };
        const path = d3.select('#beziers')
          .append('path')
          .datum([{x: 0, y: startY}, {x: bezWidth, y: endY}])
          .attr('d', bez)
          .attr('stroke', stroke)
        const start = d3.select('#beziers')
          .append('circle')
          .attr('class', 'bezier-endpoint')
          .attr('cx', RADIUS_STROKE)
          .attr('cy', startY)
          .attr('r', RADIUS)
        const end = d3.select('#beziers')
          .append('circle')
          .attr('class', 'bezier-endpoint')
          .attr('cx', fullWidth - RADIUS_STROKE)
          .attr('cy', endY)
          .attr('r', RADIUS)
      }
  })

}
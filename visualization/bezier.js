'use strict'
/**
Utilities to render the bezier connection histogram.
*/

const HEIGHT = 16;

// TODO: Better detection, possibly at backend.
let isInternal = (host) => {
  return (host.indexOf("192") == 0 || host.indexOf("172") == 0 || host.indexOf('.local') != -1);
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
  console.log('finished summing conversations');
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
      console.log('unclear convo', convo.Source, convo.Destination)
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
  const height = HEIGHT;
  const x = d3.scale.linear().range([0, width]);
  const y = d3.scale.linear().range([height, 0]);
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
    .attr('height', height)
    .append('g');
  svg.append('path')
    .datum(traffic)
    .attr('class', 'sparkline')
    .attr('d', line);
  svg.append('text')
    .attr('x', 18)
    .attr('y', height - 5)
    .attr('class', 'device-label')
    .attr('data-host', name)
    .html(name);
  svg.append('rect')
    .attr('height', 16)
    .attr('width', 16)
    .attr('class', 'host-favicon-rect')
  svg.append('image')
    .attr('class', 'host-favicon')
    // Google runs a public favicon pulling service
    .attr('href', faviconUrl || `https:\/\/www.google.com/s2/favicons?domain=${name}`);
}

let sparks = (hostTraffic) => {
  // Find global traffic max
  let maxTraf = 0;
  let minTime = hostTraffic.internal.values().next().value[0].Timestamp;
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
      console.log('unclear convo', convo.Source, convo.Destination);
      return
    }
    const leftDiv = $(`[data-host='${leftHost}']`);
    const rightDiv = $(`[data-host='${rightHost}']`);

    if (leftDiv.offset() && rightDiv.offset()) {
        const startX = leftDiv.offset().left + leftDiv.width();
        const startY = leftDiv.offset().top + HEIGHT / 2;
        const endX = rightDiv.offset().left;
        const endY = rightDiv.offset().top + HEIGHT / 2;
    
        const bezWidth = $('#beziers').width();
        let bez = (point) => {
          const fullWidth = bezWidth;
          const halfWidth = bezWidth / 2;
          return `M 0, ${point[0].y} C ${halfWidth}, ${point[0].y}, ${halfWidth}, ${point[1].y}, ${fullWidth}, ${point[1].y}`;
        };
        const path = d3.select('#beziers')
          .append('path')
          .datum([{x: 0, y: startY}, {x: bezWidth, y: endY}])
          .attr('d', bez);
    }
  })

}
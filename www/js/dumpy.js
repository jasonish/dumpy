function setElementValue(element, value) {
  element.val(value);
  return false;
}

function setStartTime(unit, value) {
  if (unit === null) {
    $("input[name='start-time']").val(moment().format());
  } else {
    $("input[name='start-time']").val(moment().subtract(value, unit).format());
  }
  return false;
}

$(document).ready(function () {
  const params = new Proxy(new URLSearchParams(window.location.search), {
    get: (searchParams, prop) => searchParams.get(prop),
  });
  if (params.event) {
    document.getElementById("event-input").innerText = params.event;
  }

  $("#input-timezone-offset").val(moment().format("Z"));
  $("input[name='start-time']").val(moment().format());

  // Bootstrap 5 tab handling
  const triggerTabList = document.querySelectorAll("#query-tabs a");
  triggerTabList.forEach((triggerEl) => {
    triggerEl.addEventListener("click", (event) => {
      event.preventDefault();
      const tab = new bootstrap.Tab(triggerEl);
      tab.show();
      $("#query-type").val($(event.target).attr("query-type"));
    });
  });

  // If there is an event entered already, set the tab active.
  if ($("#event-input").val() != "") {
    var tabEl = document.querySelector('#query-tabs a[href="#event-tab"]');
    var tab = new bootstrap.Tab(tabEl);
    tab.show();
    $("#query-type").val($(tabEl).attr("query-type"));
  }

  // Load the known spools from the server.
  fetch("/api/spools")
    .then((response) => response.json())
    .then((spools) => {
      let selector = document.getElementById("spool-selector");
      spools.forEach((spool) => {
        const option = document.createElement("option");
        option.text = spool;
        selector.add(option);
      });
    });
});

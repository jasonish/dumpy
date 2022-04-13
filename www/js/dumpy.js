function setElementValue(element, value) {
    element.val(value);
}

function setStartTime(unit, value) {
    if (unit === null) {
        $("input[name='start-time']").val(moment().format())
    } else {
        $("input[name='start-time']").val(moment().subtract(value, unit).format())
    }
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

    $("#query-tabs a").click(function (e) {
        e.preventDefault();
        $(this).tab("show");
        $("#query-type").val($(this).attr("query-type"));
    });

    // If there is an event entered already, set the tab active.
    if ($("#event-input").val() != "") {
        var tab = $('#query-tabs a[href="#event-tab"]');
        tab.tab('show')
        $("#query-type").val(tab.attr("query-type"));
    }

    // Load the known spools from the server.
    fetch("/api/spools").then(response => response.json()).then(spools => {
        let selector = document.getElementById("spool-selector");
        spools.forEach((spool) => {
            const option = document.createElement("option");
            option.text = spool;
            selector.add(option);
        })
    })
});

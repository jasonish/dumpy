function setElementValue(element, value) {
    element.val(value);
}

function setStartTime(unit, value) {
    if (unit === null) {
        $("input[name='start-time']").val(moment().format())
    }
    else {
        $("input[name='start-time']").val(moment().subtract(unit, value).format())
    }
}

$(document).ready(function () {

    $("#input-timezone-offset").val(moment().format("Z"));
    $("input[name='start-time']").val(moment().format());

});
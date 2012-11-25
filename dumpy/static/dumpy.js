/**
 * TODO
 * - When new bootstrap comes out, see if toggle for the popover works.
 */

function resize_textarea(e) {
    /* First shrink to fit. */
    while (e.rows > 1 && e.scrollHeight < $(e).outerHeight()) {
	e.rows = e.rows - 1;
    }
    /* Then expand to fit. */
    while (e.scrollHeight > $(e).outerHeight()) {
	e.rows = e.rows + 1;
    }
}

/* Get a printable version of the users time zone offset. */
function get_timezone_offset() {
    var now = new Date();
    var offset = now.getTimezoneOffset();
    var sign;
    if (offset <= 0)
	sign = "+";
    else
	sign = "-";
    offset = Math.abs(offset);
    var hours = "0{}";
    var mins = "0{}";
    hours = hours.replace("{}", now.getTimezoneOffset() / 60);
    mins = mins.replace("{}", now.getTimezoneOffset() % 60);
    return sign + hours.substring(hours.length - 2, hours.lenth) +
	mins.substring(mins.length - 2, mins.length);
}

var timezone_offset = get_timezone_offset();

$(document).ready(function() {

    /* Form submit hook to clear any existing warnings/errors. */
    $("#query-form").submit(function(e) {
	$(".alert").alert("close");
	$("#tzoffset").val(timezone_offset);
	return true;
    });

    $("#textarea-query").bind({

	focusin: function() {
	    this.orig_placeholder = this.placeholder;
	    this.placeholder = "Filter or Event: ? for help";

	    /**
	     * Its a little tricky to select all the existing text,
	     * but this seems to do it in Firefox/Chrome/Safari.
	     * 
	     * At least I think it makes sense to auto-select the
	     * text.
	     */
	    var that = $(this);
	    window.setTimeout(function() {
		that.select();
	    }, 1);
	},

	focusout: function() {
	    // Restore the placeholder text.
	    this.placeholder = this.orig_placeholder;
	},

	keydown: function(event) {
	    if (event.keyCode == 13) {
		event.preventDefault();
		$("#query-form").submit();
	    }
	    else if (event.keyCode == 191) {
		// Display help if ? is the first character.
		if ($(this).val() == "") {
		    event.preventDefault();
		    $("#modal-help-filters").modal("show");
		}
	    }
	},

	keyup: function() {
	    resize_textarea(this);
	},
    });

    $("input.time-input").bind({

	keydown: function() {
	    if (event.keyCode == 191) {
		event.preventDefault();
		if ($(this).attr("name") == "start-time") {
		    $("#modal-help-start-time").modal("show");
		}
		else if ($(this).attr("name") == "end-time") {
		    $("#modal-help-end-time").modal("show");
		}
	    }
	},

	focusin: function() {
	    this.orig_placeholder = this.placeholder;
	    if (this.id == "start-time-input") {
		this.placeholder = "Start Time: ? for help";
	    }
	    else if (this.id == "end-time-input") {
		this.placeholder = "End Time: ? for help";
	    }
	},

	focusout: function() {
	    // Restore the placeholder text.
	    this.placeholder = this.orig_placeholder;
	},

    });

    /* On page load adjust the filter text area to catch the case
     * where the user hit back and the contents was pre-populated. */
    $("#textarea-query").each(function() {
	resize_textarea(this);
    });

    /** Event to load help contents on first view. */
    $(".modal-help").on("show", function() {
	$(".modal-body", $(this)).each(function() {
	    var content_url = $(this).attr("data-content");
	    $(this).load(content_url, function() {
		$(this).html($(this).html().replace(
			/{offset}/g, timezone_offset));
	    });
	    $(this).on("show", null);
	});
    });

    /** Whenever a modal help window is opened, focus the close button
     * for quick dismissal. */
    $(".modal-help").on("shown", function() {
	$(".btn", $(this)).focus();
    });

    /**
     * This might not be ideal behaviour, but focus the start time or
     * end time inputs after closing their help windows.  Makes for a
     * better workflow when the modal is triggered from within the
     * input box.
     */
    $("#modal-help-start-time").on("hidden", function() {
	$("#start-time-input").focus();
    });
    $("#modal-help-end-time").on("hidden", function() {
	$("#end-time-input").focus();
    });
    $("#modal-help-filters").on("hidden", function() {
	$("#textarea-query").focus();
    });
    
});

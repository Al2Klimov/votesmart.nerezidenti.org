(function() {
    var now = new Date;

    switch (now.getMonth()) {
        case 11:
            var day = now.getDate();

            if (24 <= day && day <= 31) {
                snow()
            }
            break;

        case 0:
            if (now.getDate() == 1) {
                snow()
            }
    }

    function snow() {
        var script = document.createElement("script");

        script.setAttribute("type", "text/javascript");
        script.setAttribute("src", "snowstorm-min.js");

        document.getElementsByTagName("body")[0].appendChild(script);
    }
})();

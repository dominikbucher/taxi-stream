$(function () {
    function openoverlay() {
        $('#overlay').show();
    }

    var map = L.map('map').setView([40.730610, -73.969242], 13);

    // Hydda layer.
    var Hydda = L.tileLayer('https://{s}.tile.openstreetmap.se/hydda/full/{z}/{x}/{y}.png', {
        attribution: '&copy; <a href="http://osm.org/copyright">OpenStreetMap</a> contributors'
    });

    // Other option.
    var Stamen_Toner = L.tileLayer('https://stamen-tiles-{s}.a.ssl.fastly.net/toner/{z}/{x}/{y}.{ext}', {
        attribution: 'Map tiles by <a href="http://stamen.com">Stamen Design</a>, <a href="http://creativecommons.org/licenses/by/3.0">CC BY 3.0</a> &mdash; Map data &copy; <a href="http://www.openstreetmap.org/copyright">OpenStreetMap</a>',
        subdomains: 'abcd',
        minZoom: 0,
        maxZoom: 20,
        ext: 'png'
    });
    Stamen_Toner.addTo(map);

    function getRandomColor() {
        var letters = '0123456789ABCDEF';
        var color = '#';
        for (var i = 0; i < 6; i++) {
            color += letters[Math.floor(Math.random() * 16)];
        }
        return color;
    }

    function getTaxiColor(status) {
        switch (status) {
            case "empty":
                return "#FFFFFF"; // white
            case "reserved":
                return "#FF971A"; // orange
            case "occupied":
                return "#FF2708"; // red
        }
    }

    function hexToRGB(hex, alpha) {
        var r = parseInt(hex.slice(1, 3), 16),
            g = parseInt(hex.slice(3, 5), 16),
            b = parseInt(hex.slice(5, 7), 16);

        if (alpha) {
            return "rgba(" + r + ", " + g + ", " + b + ", " + alpha + ")";
        } else {
            return "rgb(" + r + ", " + g + ", " + b + ")";
        }
    }

    function createNewTaxi(id) {
        return {
            taxiId: id,
            status: "empty",
            color: getTaxiColor("empty"),
            lon: 0.0,
            lat: 0.0,
            numOccupants: 0
        }
    }

    taxiData = {
        0: {
            taxiId: 0,
            status: "empty",
            color: getTaxiColor("empty"),
            lon: -73.969242,
            lat: 40.730610,
            numOccupants: 0
        }
    };

    function renderCircle(ctx, point, fillStyle, strokeStyle, radius) {
        ctx.fillStyle = fillStyle;
        ctx.strokeStyle = strokeStyle;
        ctx.beginPath();
        ctx.arc(point.x, point.y, radius, 0, Math.PI * 2.0, true);
        ctx.closePath();
        ctx.fill();
        ctx.stroke();
    }

    var taxiLayer = L.canvasLayer()
        .delegate({
            onDrawLayer: function (info) {
                var ctx = info.canvas.getContext('2d');
                ctx.clearRect(0, 0, info.canvas.width, info.canvas.height);
                ctx.fillStyle = "rgba(0, 0, 0, 0.2)";
                ctx.fillRect(0, 0, info.canvas.width, info.canvas.height);
                for (var id in taxiData) {
                    var data = taxiData[id];
                    var point = info.layer._map.latLngToContainerPoint([data.lat, data.lon]);
                    var color = getTaxiColor(data.status);
                    renderCircle(ctx, point, hexToRGB(color, 0.5), hexToRGB(color, 0.9), 5.0);
                }
            }
        });
    taxiLayer.addTo(map);

    var ws;
    if (window.WebSocket === undefined) {
        $("#container").append("Your browser does not support WebSockets");
        return;
    } else {
        ws = initWS();
    }

    function initWS() {
        var socket = new WebSocket("ws://" + window.location.hostname + ":8080/ws"),
            container = $("#container");
        socket.onopen = function () {
            container.append("<p>Socket is open</p>");
        };
        socket.onmessage = function (e) {
            var data = JSON.parse(e.data);
            if (!(data["taxiId"] in taxiData)) {
                taxiData[data["taxiId"]] = createNewTaxi(data["taxiId"]);
            }
            if ("lon" in data && "lat" in data) {
                taxiData[data["taxiId"]].lon = data["lon"];
                taxiData[data["taxiId"]].lat = data["lat"];
            }
            if ("numOccupants" in data) {
                taxiData[data["taxiId"]].numOccupants = data["numOccupants"];
                if (data["numOccupants"] > 0) {
                    taxiData[data["taxiId"]].status = "occupied";
                }
            }
            if ("reservationLon" in data && "reservationLat" in data) {
                taxiData[data["taxiId"]].status = "reserved";
            }
            if ("totalAmount" in data) {
                taxiData[data["taxiId"]].status = "empty";
            }
            taxiLayer.needRedraw();
        };
        socket.onclose = function () {
            container.append("<p>Socket closed</p>");
        };
        return socket;
    }

    $("#sendBtn").click(function (e) {
        e.preventDefault();
        ws.send(JSON.stringify({Num: parseInt($("#numberField").val())}));
    });
});
/*
 *
 * This file contains client side Javascript that is rendered
 * when a client hits the RelayR route with a GET request.
 *
 */

package relayR

const connectionClassScript = `

var RelayRConnection = (function() {
	var route = '%v';
	var transport = {
		WebSocket: {
			connect: function(c) {
				var s = this;
				s.socket = new WebSocket("ws://" + window.location.host + route + "/ws?connectionId=" + transport.ConnectionId);
				s.socket.onclose = function(evt) {
					setTimeout(function() {
						web.n(); // renegotiate
					}, 2000);
				};

				s.socket.onmessage = function(evt) {
					c(evt.data);
				};

				s.socket.onerror = function(evt) {
					setTimeout(function() {
						s.connect(cId, c);
					});
				};

				s.socket.onopen = function(evt) {
					RelayRConnection.r();
				};
			},
			send: function(data) {
				var s = this;
				s.socket.send(data);
			}
		}
	};
	var web = (function() {
		return {
			x: function() {
				var xd;

				if (window.XMLHttpRequest) {
					xd = new XMLHttpRequest();
				}
				else {
					xd = new ActiveXObject("Microsoft.XMLHTTP");
				}

				return xd;
			},
			g: function(u, c) {
				var s = this;

				var xd = s.x();

				xd.open('GET', u, true);

				xd.onreadystatechange = function() {
					if (xd.readyState === 4 && xd.status === 200) {
						c(xd);
					}
				};

				xd.send();
			},
			p: function(u, d, c, t, e) {
				var s = this;

				var xd = s.x();

				xd.open('POST', u, true);
				xd.setRequestHeader("Content-type", "application/" + t);

				xd.onreadystatechange = function() {
					if (xd.readyState === 4 && xd.status === 200) {
						c(xd);
					} 
				};

				xd.onerror = function() {
					e(xd);
				};

				xd.send(d);
			},
			t: function() {
				if (!!window.WebSocket) {
					return "WebSocket";
				}

				// TODO: Implement SSE Circuit
				/*if (typeof EventSource !== 'undefined') {
					return "SSE";
				}*/
			},
			n: function() {
				var s = this;
				var t = s.t();
				web.p(route + "/negotiate", JSON.stringify({ t: t }), function(result) {
					var obj = JSON.parse(result.responseText);
					transport.ConnectionId = obj.ConnectionID;
					setTimeout(function() {
						transport[t].connect(function(data) {
							var cobj = JSON.parse(data);
							var lobj = RelayR[cobj.R].client;
							var args = [];
							for (var i = 0; i < cobj.A.length; i++) {
								args.push(cobj.A[i][0]);
							}
							lobj[cobj.M].apply(lobj||window, args);
						});
					}, 0);
				}, "json",
				function(result) {
					// error .. try again
					setTimeout(function() {
						s.n();
					}, 2000);
				});
			}
		};
	})();

	return {
		ready: function(r) {
			RelayRConnection.r = r;

			web.n();
		},
		callServer: function(r, f, a) {
			transport[web.t()].send(JSON.stringify({ S: true, C: transport.ConnectionId, R: r, M: f, A: a}));
		}
	};
})();

`

const relayClassBegin = `

var RelayR = (function() {
	return {

`

const relayBegin = `

%v: {

	client: {},

	server: {

`

// {0} == function name
// {1} == Relay name
// {2} == function name
const relayMethod = `

%v: function() {
	RelayRConnection.callServer('%v', '%v', Array.prototype.slice.call(arguments));
},

`

const relayEnd = `

},

},

`

const relayClassEnd = `

	};
})();

`

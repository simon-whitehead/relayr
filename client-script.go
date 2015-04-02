/*
 *
 * This file contains client side Javascript that is rendered
 * when a client hits the RelayR route with a GET request.
 *
 */

package relayr

const connectionClassScript = `

var RelayRConnection = {};

RelayRConnection = (function() {
	var readyCalled = false;
	var web, transport;
	var route = '%v';
	transport = {
		websocket: {
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
					if (!readyCalled) {
						RelayRConnection.r();
						readyCalled = true;
					}
				};
			},
			send: function(data) {
				var s = this;
				s.socket.send(data);
			}
		},
		longpoll: {
			connect: function(c) {
				if (!readyCalled) {
					RelayRConnection.r();
					readyCalled = true;
				}
				var retry;
				retry = function() {
					web.gj(route + '/longpoll?connectionId=' + transport.ConnectionId + '&_=' + new Date().getTime(), function(data) {
						if (data.responseText) {
							var reconn = JSON.parse(data.responseText);
							if (reconn.Z) {
								web.n();
							} else {
								c(data);
								retry();
							}
						} else {
							web.n();
						}
					});
				};

				retry();
			},
			send: function(data) {
				var s = this;
				web.p(route + '/call?connectionId=' + transport.ConnectionId + '&_=' + new Date().getTime(), data, null, "json", null); 
			}
		}
	};
	web = (function() {
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
			gj: function(u, c) {
				var s = this;

				var xd = s.x();

				xd.open('GET', u, true);
				xd.setRequestHeader("Content-type", "application/json");

				xd.onreadystatechange = function() {
					if (xd.readyState === 4) {
						if (xd.status === 200) {
							c(xd);
						} 
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
					if (xd.readyState === 4) {
						if (xd.status === 200) {
							if (c) {
								c(xd);
							}
						}
					} 
				};

				xd.onerror = function() {
					if (e) {
						e(xd);
					}
				};

				xd.send(d);

				window.onbeforeunload = function() {
					delete xd;
					xd = null;
				};
			},
			t: function() {
				if (!!window.WebSocket) {
					return "websocket";
				} else {
					return "longpoll";
				}

				// TODO: Implement SSE Circuit
				/*if (typeof EventSource !== 'undefined') {
					return "SSE";
				}*/
			},
			n: function() {
				var s = this;
				var t = s.t();
				web.p(route + "/negotiate?_=" + new Date().getTime(), JSON.stringify({ t: t }), function(result) {
					var obj = JSON.parse(result.responseText);
					transport.ConnectionId = obj.ConnectionID;
					setTimeout(function() {
						transport[t].connect(function(data) {
							var cobj;
							if (data.responseText && data.status && data.responseXML) {
								cobj = JSON.parse(data.responseText);
							} else {
								if (data.responseText == "") return;
								cobj = JSON.parse(data);
							}
							var lobj = RelayR[cobj.R].client;
							var args = [];
							for (var i = 0; i < cobj.A.length; i++) {
								args.push(cobj.A[i]);
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

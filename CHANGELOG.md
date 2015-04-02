## RelayR changelog

#### v0.2.1 - v0.3.0

* FEATURE: Added Long Polling transport - allowing all browsers that support AJAX requests to work with RelayR. RelayR will test
for WebSocket support and if not found, will fall back to Long Polling.
* BUGFIX: Removed renegotiation bug.
* BUGFIX: Stop `RelayRConnection.ready` being called more than once.

----------------

#### v0.2.0 - v0.2.1

* Added `clientScript` caching variable, which allows the RelayR client-side script to be cached in a `[]byte` slice and served each time.
* Added `DisableScriptCache()` package-level function to disable the above functionality and regenerate the script each page refresh (mostly for debug purposes).

----------------

#### v0.1.0 - v0.2.0

There are are four necessary breaking changes in the update from 0.1.0 to 0.2.0 to make it more "Go-like" and consumer friendly.

* The package name now has a lowercase r; "relayr" instead of "relayR". This was one of the only gripes from the community about this package.
* You no longer need to embed `*relayr.Relay` .. but you now need to have all of your relay methods accept a `*relayr.Relay` as their first argument.
* Groups are now an object in themselves and no longer "hang off" of the `Clients` property of a Relay. The examples have been updated to reflect this.
* relayr.Exchange now implements `ServeHTTP`, which means it can now handle a route directly, instead of using `relayr.Handle` .. you can use `http.Handle`.

The rest should _mostly_ work as it did. I have completely re-written most of the underlying code for easier maintenance and better documentation. The public API now has better documentation and should make a lot more sense.


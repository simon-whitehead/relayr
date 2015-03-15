## RelayR changelog


#### v0.1.0 - v0.2.0

There are are four necessary breaking changes in the update from 0.1.0 to 0.2.0 to make it more "Go-like" and consumer friendly.

* The package name now has a lowercase r; "relayr" instead of "relayR". This was one of the only gripes from the community about this package.
* You no longer need to embed `*relayr.Relay` .. but you now need to have all of your relay methods accept a `*relayr.Relay` as their first argument.
* Groups are now an object in themselves and no longer "hang off" of the `Clients` property of a Relay. The examples have been updated to reflect this.
* relayr.Exchange now implements `ServeHTTP`, which means it can now handle a route directly, instead of using `relayr.Handle` .. you can use `http.Handle`.

The rest should _mostly_ work as it did. I have completely re-written most of the underlying code for easier maintenance and better documentation. The public API now has better documentation and should make a lot more sense.


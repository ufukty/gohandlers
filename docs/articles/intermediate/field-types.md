# 🧬 Field types

## Custom (de)serialization rules

Depending on where a request/response parameter is transfered, there needs to be special encoding & decoding process handled. Even though the rules of encoding for URL path and query parts are more or less same, Gohandlers lets you define separate methods to distinguish them.

A Go type needs to implement (de)serialization methods listed below. The exact set of methods for one particular handler is decided based on the Content-Type of bodies of requests and responses that Go type is used as a field type for its bindings.

| Method         | Description                                                     |
| -------------- | --------------------------------------------------------------- |
| `.FromRoute()` | Deserializes from **route representation** to typed value       |
| `.ToRoute()`   | Serializes typed value to place in route                        |
| `.FromQuery()` | Deserializes from **query representation** to typed value       |
| `.ToQuery()`   | Serializes typed value to place in query                        |
| `.FromForm()`  | Deserializes from `x-www-form-urlencoded` body representation   |
| `.ToForm()`    | Serializes typed value to place in `x-www-form-urlencoded` body |

## Custom validation rules

Gohandlers generated request binding validator methods only aggregate errors returned by the validators of fields. So, each request binding type field should implement a validator method in signature of `Validate() error`.

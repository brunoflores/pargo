# ParGo

Package pargo provides a Go client for the [Pardot REST API](http://developer.pardot.com).

`go get github.com/brunoflores/pargo`

Getting started:

```go
type prospect struct {
    ID    int    `json:"id"`
    Email string `json:"email"`
}
response := []prospect{} // Placeholder for the response.
pardot := pargo.NewPargo(
    pargo.UserAccount{
        Email:   "my@email.com",
        Pass:    "mypass",
        UserKey: "myclientkey",
    })
// Conventionally, each endpoint as a method of the client might
// expect a struct of options with the same name.
// Note: the method is called on the client `pardot`,
// and the options struct is accessed via the package name `pargo`.
err := pardot.QueryProspects(pargo.QueryProspects{
    Offset:      0,
    Limit:       200,
    Fields:      []string{"id", "email"},
    PlaceHolder: &response, // Pointer to placeholder.
})
if err != nil {
    // Handle error, optionally testing for custom ParGo errors:
    switch err.(type) {
    case pargo.ErrLoginFailed:
        // Invalid credentials.
    case pargo.ErrInvalidJSON:
        // Invalid request.
    default:
        // Some other error.
    }
}
// ... Use `response` slice.
```

## Running tests

To run all tests:

```
$ go test
```

To run a specific test:

```
$ go test -run=TestReuseAPIKeyUntilExpired
```

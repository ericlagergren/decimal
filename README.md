# decimal [![Build Status](https://travis-ci.org/ericlagergren/decimal.png?branch=master)](https://travis-ci.org/ericlagergren/decimal) [![GoDoc](https://godoc.org/github.com/ericlagergren/decimal?status.svg)](https://godoc.org/github.com/ericlagergren/decimal)

`decimal` implements arbitrary precision, decimal floating-point numbers, per 
the [General Decimal Arithmetic](http://speleotrove.com/decimal/) specification.

## Features

 * Useful zero values.
   The zero value of a `decimal.Big` is 0, just like `math/big`.

 * Multiple operating modes.
   Different operating modes allow you to tailor the package's behavior to your
   needs. The GDA mode strictly implements the GDA specification, while the Go
   mode implements familiar Go idioms.

 * High performance.
   `decimal` is consistently one of the fastest arbitrary-precision decimal 
   floating-point libraries, regardless of language.

 * An extensive math library.
   The `math/` subpackage implements elementary and trigonometric functions,
   continued fractions, and more.

 * A familiar, idiomatic API.
   `decimal`'s API follows `math/big`'s API, so there isn't a steep learning 
   curve.

## Installation

`go get github.com/ericlagergren/decimal`

## Documentation

[GoDoc](http://godoc.org/github.com/ericlagergren/decimal)

## Versioning

`decimal` uses Semantic Versioning. The current version is 3.3.1.

`decimal` only explicitly supports the two most recent major Go 1.X versions.

## Contributing to `decimal`

We welcome contributions to `decimal` of any kinds including bug reports, issues, feature requests, PRs to documentation, feature implementations and bug fixes.

### Code contribution

To make a contribution do the following:

* Fork the project, start new branch and make your changes
* Make test cases for the new code
* Do not forget to run `go fmt`
* Add or update documentation if you are adding new features or changing functionality
* Provide a good commit message with the reference to a closed issue if any

### Testing the code

There are 4 main types of tests:

1. Python-generated test tables. Python 3 is used to generate test tables that are parsed using the `suite` package. 
2. `dectest` tables. The `dectest` tables are similar to Python tables. They are parsed using `dectest` package. 
3. Custom tests. The custom tests are strewn thought the different *_test.go files.
4. Bug-specific  tests. They are placed inside `issues_test.go` file. 

#### Generating test data

Before running the tests test data have to be generated. 

To generate Python test tables run the following commands inside the project's directory:

```
cd testdata/pytables
./tables.py 500
cd ../..
```

To generate `dectest` data execute the command:

```
testdata/dectest/generate.bash
```

#### Running the test cases

To run the full test suite execute the following command. Please, note that the timeout for tests increased to 30 minutes. 

```
go test -timeout 30m -v
```

Some test could take a long time, use `-short` flag to run only quick tests.

```
go test -short -v
```

## License

[BSD 3-clause](https://github.com/ericlagergren/decimal/blob/master/LICENSE)

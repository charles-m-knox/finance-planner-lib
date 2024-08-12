# finance-planner-lib

A Go library that processes a list of recurring bill definitions (*transactions* or `TX` in the code) and projects expenses into the future (*results*).

This library powers my Finance Planner applications.

## Usage

This will be updated later.

## Unit testing

The purpose of this library is to confidently calculate financial operations.

Unit test coverage is a big part of facilitating this, since it's large math and data structure manipulation.

To run unit tests, execute the following in your terminal:

```bash
# will open a browser tab with the code coverage
make test
```

Currently, unit testing code coverage is at **98.0%**.

This project will not aim to hit 100% code coverage, since some of the error handling branches are not worth the effort of reproducing them in a unit test environment. The core business logic in this application is the priority, and the vast majority of it is tested.

## TODOs and Limitations

Currently the library doesn't support weekday-specific recurrence patterns like `First monday of every month`. The underlying `rrule` library *does* support these patterns. In the future, this library will aim to provide an API that allows for more direct passthrough of all possible `rrule` options so that maximum flexibility is possible. Unfortunately, because this library was originally extracted from my other financial planning projects, the goal of specifically providing an `rrule`-like API was not in scope. However, to reiterate, future scope will move towards this direction.

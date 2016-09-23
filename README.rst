****************************************
Vanilla: an extended Go standard library
****************************************

Vanilla extends the Golang builtin packages with standard-library like packages
that are commonly useful for business logic, database access, and software management.


Packages
========
clock
    extensions to the "time" package that encourage timezones and allow time to be modified in tests

crypto
    extensions to the "crypto" package to simplify common crypto operations

date
    extensions to the "time" package that deal explicitly with dates

expect
    extensions to the "testing" package with simple assertions/expectations

httpx
    extensions to the "net/http" package with better Mux and Middleware

null
    utils for nullable types that support database/sql, encoding/gob, and encoding/json

semver
    utils for Semantic Versioning

sql
    extenstions to the "database/sql" package to help with dynamic SQL

uuid
    utils for UUIDs (Universally Unique Identifiers)


Issues
======

The GitHub issues for this project are an appropriate place for bug reports,
feature requests/suggestions, and pull requests.

Please do not create issues like "I can't get it to work" or "how do I do X with Vanilla?".
For support using vanilla, feel free to contact **kevin@reflexionhealth.com**
who will respond on a best-effort basis.


License
=======

Unless otherwise specified Vanilla packages are licensed under a BSD 3-Clause license (see LICENSE_).

.. _LICENSE: LICENSE

# firetest [![GoDoc](https://godoc.org/github.com/zabawaba99/firetest?status.svg)](https://godoc.org/github.com/zabawaba99/firetest) [![Build Status](https://travis-ci.org/zabawaba99/firetest.svg?branch=master)](https://travis-ci.org/zabawaba99/firetest) [![Coverage Status](https://coveralls.io/repos/zabawaba99/firetest/badge.svg?branch=master)](https://coveralls.io/r/zabawaba99/firetest)

Firebase test server meant for use in unit tests

**Note: This project is not meant to be used as a Firebase replacement
nor to compete with Firebase. **

### Implemented

* [Basic API Usage](https://www.firebase.com/docs/rest/api/#section-api-usage)
  * POST
  * GET
  * PUT
  * PATCH
  * DELETE
* [Query parameters](https://www.firebase.com/docs/rest/api/#section-query-parameters):
  * auth

### Not Supported

* [Query parameters](https://www.firebase.com/docs/rest/api/#section-query-parameters):
  * shallow
  * print
  * format
  * download
* [Streaming](https://www.firebase.com/docs/rest/api/#section-streaming)
* [Priorities](https://www.firebase.com/docs/rest/api/#section-priorities)
* [Server Values](https://www.firebase.com/docs/rest/api/#section-server-values)
* [Security Rules](https://www.firebase.com/docs/rest/api/#section-security-rules)
* [Error Conditions](https://www.firebase.com/docs/rest/api/#section-error-conditions)

## Contributing

1. Fork it
2. Create your feature branch (`git checkout -b new-feature`)
3. Commit your changes (`git commit -am 'awesome things with tests'`)
4. Push to the branch (`git push origin new-feature`)
5. Create new Pull Request

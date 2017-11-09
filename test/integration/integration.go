// +build integration
package integration

// For integration test, we chose BDD testcase framework.

// https://github.com/onsi/ginkgo
// More details please ref: http://onsi.github.io/ginkgo/

// RMD is a restful API server, for RESTFUL assert we use httpexpect
// https://github.com/gavv/httpexpect
// Install httpexpect: $ go get gopkg.in/gavv/httpexpect.v1
// Also install other dependency:
// $ go get github.com/dgrijalva/jwt-go
// $ go get github.com/labstack/echo
// $ go get "google.golang.org/appengine"
// $ go get "gopkg.in/kataras/iris.v6"

// In ordre to distinguish unit test and integration test, we use -short as the
// option of "go test"

// There are already some unit tests in RMD source code.
// We can Converte the Existing Tests to ginkgo by:
// $ ginkgo convert github.com/your/package
// Ref: http://onsi.github.io/ginkgo/#converting-existing-tests

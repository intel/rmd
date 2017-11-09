// +build integrationHTTPS

package integrationHTTPS

import (
	. "github.com/onsi/ginkgo"
	"net/http"
)

var _ = Describe("PAMAuth", func() {

	var (
		path, username, password string
	)

	Describe("Get https requests", func() {
		Describe("Get policy", func() {

			BeforeEach(func() {
				path = "/policy"
			})

			Context("Get policy with valid Berkeley db credentials that is also authorized", func() {
				BeforeEach(func() {
					username = "user"
					password = "user1"
				})
				It("Should return 200 StatusOK", func() {
					he.GET(path).
						WithHeader("Content-Type", "application/json").
						WithBasicAuth(username, password).
						Expect().
						Status(http.StatusOK)
				})
			})

			Context("Get policy with valid Berkeley db credentials that is not authorized", func() {
				BeforeEach(func() {
					username = "test"
					password = "test1"
				})
				It("Should return 403 StatusForbidden", func() {
					he.GET(path).
						WithHeader("Content-Type", "application/json").
						WithBasicAuth(username, password).
						Expect().
						Status(http.StatusForbidden)
				})
			})

			Context("Get policy with invalid Berkeley db user", func() {
				BeforeEach(func() {
					username = "use"
					password = "user1"
				})
				It("Should return 401 StatusUnauthorized", func() {
					he.GET(path).
						WithHeader("Content-Type", "application/json").
						WithBasicAuth(username, password).
						Expect().
						Status(http.StatusUnauthorized).
						Text().
						Equal("User not known to the underlying authentication module\n")
				})
			})

			Context("Get policy with incorrect Berkeley db credentials", func() {
				BeforeEach(func() {
					username = "user"
					password = "user2"
				})
				It("Should return 401 StatusUnauthorized", func() {
					he.GET(path).
						WithHeader("Content-Type", "application/json").
						WithBasicAuth(username, password).
						Expect().
						Status(http.StatusUnauthorized).
						Text().
						Equal("Authentication failure\n")
				})
			})

			// Edit unix credentials here according to your testing platform
			/*
				Context("Get policy with valid unix credentials that is also authorized", func() {
					// Please use credentials different from those defined in Berkeley db for a consistent error message
					BeforeEach(func() {
						username = "root"
						password = "s"
					})
					It("Should return 200 StatusOK", func() {
						he.GET(path).
							WithHeader("Content-Type", "application/json").
							WithBasicAuth(username, password).
							Expect().
							Status(http.StatusOK)
					})
				})

				Context("Get policy with valid unix credentials that is not authorized", func() {
					// Please use credentials different from those defined in Berkeley db for a consistent error message
					BeforeEach(func() {
						username = "common"
						password = "common"
					})
					It("Should return 403 StatusForbidden", func() {
						he.GET(path).
							WithHeader("Content-Type", "application/json").
							WithBasicAuth(username, password).
							Expect().
							Status(http.StatusForbidden)
					})
				})
			*/

			Context("Get policy with invalid unix user", func() {
				// Please use credentials different from those defined in Berkeley db for a consistent error message
				BeforeEach(func() {
					username = "com"
					password = "common"
				})
				It("Should return 401 StatusUnauthorized", func() {
					he.GET(path).
						WithHeader("Content-Type", "application/json").
						WithBasicAuth(username, password).
						Expect().
						Status(http.StatusUnauthorized).
						Text().
						Equal("User not known to the underlying authentication module\n")
				})
			})

			Context("Get policy with incorrect unix credentials", func() {
				// Please use credentials different from those defined in Berkeley db for a consistent error message
				BeforeEach(func() {
					username = "root"
					password = "com"
				})
				It("Should return 401 StatusUnauthorized", func() {
					he.GET(path).
						WithHeader("Content-Type", "application/json").
						WithBasicAuth(username, password).
						Expect().
						Status(http.StatusUnauthorized).
						Text().
						Equal("User not known to the underlying authentication module\n")
				})
			})
		})
	})
})

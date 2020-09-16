package integration_test

import (
	"os"
	"testing"

	"github.com/yashap/crius/internal/app"

	"github.com/gin-gonic/gin"

	"github.com/franela/goblin"
	. "github.com/onsi/gomega"
	"github.com/yashap/crius/internal/integration_test/util"
)

var testDB *util.TestDB

func TestMain(m *testing.M) {
	testDB = util.NewTestDB()

	testExitCode := m.Run()

	testDB.Shutdown(true)
	os.Exit(testExitCode)
}

func Test(t *testing.T) {
	g := goblin.Goblin(t)
	RegisterFailHandler(func(m string, _ ...int) { g.Fail(m) })
	crius := app.NewCrius(app.DBConfig{
		User:     testDB.User,
		Password: testDB.Password,
		DBName:   testDB.Database,
		Port:     testDB.Port,
	})
	crius.MigrateDB()

	g.Describe("POST /services", func() {
		g.It("Should create a new service", func() {
			postBody := gin.H{
				"code": "tops",
				"name": "Teams, Organizations and Permissions Service",
				"endpoints": []gin.H{
					{
						"code": "GET /teams/{id}",
						"name": "Get team by id",
					},
				},
			}
			response := util.HttpRequest(crius.Router(), "POST", "/services", postBody)
			Expect(response.Code).To(Equal(200))
			Expect(response.Body["id"]).To(Equal(float64(1)))
		})

		g.It("Should update a service, adding an endpoint", func() {
			postBody := gin.H{
				"code": "tops",
				"name": "Teams, Organizations and Permissions Service",
				"endpoints": []gin.H{
					{
						"code": "GET /teams/{id}",
						"name": "Get team by id",
					},
					{
						"code": "DELETE /teams/{id}",
						"name": "Delete team by id",
					},
				},
			}
			response := util.HttpRequest(crius.Router(), "POST", "/services", postBody)
			Expect(response.Code).To(Equal(200))
			Expect(response.Body["id"]).To(Equal(float64(1)))
		})
	})
}

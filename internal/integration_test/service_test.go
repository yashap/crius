package integration_test

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"os"
	"testing"

	"github.com/yashap/crius/internal/app"

	"github.com/franela/goblin"
	. "github.com/onsi/gomega"
	"github.com/yashap/crius/internal/integration_test/util"
)

type fixture struct {
	app   app.Crius
	label string
}

var fixtures []fixture

func TestMain(m *testing.M) {
	pgDB := util.NewPostgresTestDB()
	defer pgDB.Shutdown(true)
	pgDB.Database.Migrate()
	fixtures = append(fixtures, fixture{app.NewCrius(pgDB.Database), "Postgres"})

	mysqlDB := util.NewMySQLTestDB()
	defer mysqlDB.Shutdown(true)
	mysqlDB.Database.Migrate()
	fixtures = append(fixtures, fixture{app.NewCrius(mysqlDB.Database), "MySQL"})

	testExitCode := m.Run()
	os.Exit(testExitCode)
}

func Test(t *testing.T) {
	g := goblin.Goblin(t)
	RegisterFailHandler(func(m string, _ ...int) { g.Fail(m) })
	for _, f := range fixtures {
		runTests(g, f.app, f.label)
	}
}

func runTests(g *goblin.G, crius app.Crius, label string) {
	g.Describe(fmt.Sprintf("When running against %s", label), func() {
		g.Describe("POST /services", func() {
			g.It("Should create a new service", func() {
				postBody := gin.H{
					"code": "tops",
					"name": "Teams, Organizations and Permissions Service",
					"endpoints": []gin.H{
						{
							"code": "GET /teams/{id}",
							"name": "Get team by id",
							"dependencies": gin.H{},
						},
					},
				}
				response := util.HttpRequest(crius.Router(), "POST", "/services", postBody)
				Expect(response.Code).To(Equal(200))
				Expect(response.Body["id"]).To(Equal(float64(1)))
				response = util.HttpRequest(crius.Router(), "GET", "/services/tops", nil)
				Expect(response.Code).To(Equal(200))
				Expect(util.JsonString(response.Body)).To(MatchJSON(util.JsonString(postBody)))
			})

			g.It("Should update a service, adding an endpoint", func() {
				postBody := gin.H{
					"code": "tops",
					"name": "Teams, Organizations and Permissions Service",
					"endpoints": []gin.H{
						{
							"code": "DELETE /teams/{id}",
							"name": "Delete team by id",
							"dependencies": gin.H{},
						},
						{
							"code": "GET /teams/{id}",
							"name": "Get team by id",
							"dependencies": gin.H{},
						},
					},
				}
				response := util.HttpRequest(crius.Router(), "POST", "/services", postBody)
				Expect(response.Code).To(Equal(200))
				Expect(response.Body["id"]).To(Equal(float64(1)))
				response = util.HttpRequest(crius.Router(), "GET", "/services/tops", nil)
				Expect(response.Code).To(Equal(200))
				Expect(util.JsonString(response.Body)).To(MatchJSON(util.JsonString(postBody)))
			})
		})
	})
}

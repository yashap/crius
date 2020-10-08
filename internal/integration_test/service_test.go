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
	os.Exit(testMain(m))
}

func testMain(m *testing.M) int {
	pgDB := util.NewPostgresTestDB()
	defer pgDB.Shutdown(true)
	pgDB.Database.Migrate()
	fixtures = append(fixtures, fixture{app.NewCrius(pgDB.Database), "Postgres"})

	mysqlDB := util.NewMySQLTestDB()
	defer mysqlDB.Shutdown(true)
	mysqlDB.Database.Migrate()
	fixtures = append(fixtures, fixture{app.NewCrius(mysqlDB.Database), "MySQL"})

	return m.Run()
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
					"code": "nhl_games",
					"name": "NHL Games Service ",
					"endpoints": []gin.H{
						{
							"code":                        "GET /games/{id}",
							"name":                        "Get an NHL game by id",
							"serviceEndpointDependencies": gin.H{},
						},
					},
				}
				response := util.HttpRequest(crius.Router(), "POST", "/services", postBody)
				Expect(response.Code).To(Equal(200))
				Expect(response.Body["id"]).To(Equal(float64(1)))
				expectServiceToEqual(crius.Router(), "nhl_games", postBody)
			})

			g.It("Should update a service, adding an endpoint", func() {
				postBody := gin.H{
					"code": "nhl_games",
					"name": "NHL Games Service ",
					"endpoints": []gin.H{
						{
							"code":                        "DELETE /games/{id}",
							"name":                        "Delete an NHL game by id",
							"serviceEndpointDependencies": gin.H{},
						},
						{
							"code":                        "GET /games/{id}",
							"name":                        "Get an NHL game by id",
							"serviceEndpointDependencies": gin.H{},
						},
					},
				}
				response := util.HttpRequest(crius.Router(), "POST", "/services", postBody)
				Expect(response.Code).To(Equal(200))
				Expect(response.Body["id"]).To(Equal(float64(1)))
				expectServiceToEqual(crius.Router(), "nhl_games", postBody)
			})

			g.It("Should update a service, removing an endpoint", func() {
				postBody := gin.H{
					"code": "nhl_games",
					"name": "NHL Games Service ",
					"endpoints": []gin.H{
						{
							"code":                        "DELETE /games/{id}",
							"name":                        "Delete an NHL game by id",
							"serviceEndpointDependencies": gin.H{},
						},
					},
				}
				response := util.HttpRequest(crius.Router(), "POST", "/services", postBody)
				Expect(response.Code).To(Equal(200))
				Expect(response.Body["id"]).To(Equal(float64(1)))
				expectServiceToEqual(crius.Router(), "nhl_games", postBody)
			})
		})
	})
}

func expectServiceToEqual(router *gin.Engine, serviceCode string, expectedService map[string]interface{}) bool {
	response := util.HttpRequest(router, "GET", "/services/"+serviceCode, nil)
	Expect(response.Code).To(Equal(200))
	return Expect(util.JsonString(response.Body)).To(MatchJSON(util.JsonString(expectedService)))
}

package main

import (
	"fmt"
	"os"

	"net/http"

	"github.com/Financial-Times/base-ft-rw-app-go/baseftrwapp"
	"github.com/Financial-Times/go-fthealth/v1a"
	"github.com/Financial-Times/neo-utils-go/neoutils"
	log "github.com/Sirupsen/logrus"
	"github.com/jawher/mow.cli"
	"github.com/jmcvetta/neoism"
	"github.com/Financial-Times/financial-instruments-rw-neo4j/financialinstruments"
)

func main() {
	app := cli.App("financial-instruments-rw-neo4j", "A RESTful API for managing financial instruments in neo4j")
	neoURL := app.String(cli.StringOpt{
		Name:   "neo-url",
		Value:  "http://neo4j:neo4j_new@localhost:7474/db/data/",
		Desc:   "neo4j endpoint URL",
		EnvVar: "NEO_URL",
	})
	graphiteTCPAddress := app.String(cli.StringOpt{
		Name:   "graphiteTCPAddress",
		Value:  "",
		Desc:   "Graphite TCP address, e.g. graphite.ft.com:2003. Leave as default if you do NOT want to output to graphite (e.g. if running locally",
		EnvVar: "GRAPHITE_ADDRESS",
	})
	graphitePrefix := app.String(cli.StringOpt{
		Name:   "graphitePrefix",
		Value:  "",
		Desc:   "Prefix to use. Should start with content, include the environment, and the host name. e.g. coco.pre-prod.roles-rw-neo4j.1 or content.test.people.rw.neo4j.ftaps58938-law1a-eu-t",
		EnvVar: "GRAPHITE_PREFIX",
	})
	port := app.Int(cli.IntOpt{
		Name:   "port",
		Value:  8080,
		Desc:   "Port to listen on",
		EnvVar: "APP_PORT",
	})
	batchSize := app.Int(cli.IntOpt{
		Name:   "batchSize",
		Value:  1024,
		Desc:   "Maximum number of statements to execute per batch",
		EnvVar: "BATCH_SIZE",
	})
	logMetrics := app.Bool(cli.BoolOpt{
		Name:   "logMetrics",
		Value:  false,
		Desc:   "Whether to log metrics. Set to true if running locally and you want metrics output",
		EnvVar: "LOG_METRICS",
	})
	env := app.String(cli.StringOpt{
		Name:  "env",
		Value: "local",
		Desc:  "environment this app is running in",
	})

	app.Action = func() {
		db, err := neoism.Connect(*neoURL)
		db.Session.Client = &http.Client{Transport: &http.Transport{MaxIdleConnsPerHost: 100}}
		if err != nil {
			log.Errorf("Could not connect to neo4j, error=[%s]\n", err)
		}
		batchRunner := neoutils.NewBatchCypherRunner(neoutils.TransactionalCypherRunner{db}, *batchSize)
		financialInstrumentsDriver := financialinstruments.NewCypherFinancialInstrumentService(batchRunner, db)
		financialInstrumentsDriver.Initialise()

		baseftrwapp.OutputMetricsIfRequired(*graphiteTCPAddress, *graphitePrefix, *logMetrics)

		services := map[string]baseftrwapp.Service{
			"financialInstruments": financialInstrumentsDriver,
		}

		var checks []v1a.Check
		for _, service := range services {
			checks = append(checks, makeCheck(service, batchRunner))
		}

		healthHandler := v1a.Handler("ft-financial-instruments_rw_neo4j ServiceModule", "Writes 'financial insturments' to Neo4j, usually as part of a bulk upload done on a schedule", checks...)
		baseftrwapp.RunServerWithConf(baseftrwapp.RWConf{
			Services:      services,
			HealthHandler: healthHandler,
			Port:          *port,
			ServiceName:   "financial-instruments-rw-neo4j",
			Env:           *env,
			EnableReqLog:  false,
		})
	}
	log.SetLevel(log.InfoLevel)
	log.Printf("Application started with args %s\n", os.Args)

	app.Run(os.Args)
}

func makeCheck(service baseftrwapp.Service, cr neoutils.CypherRunner) v1a.Check {
	return v1a.Check{
		BusinessImpact:   "Cannot read/write financial instrument via this writer",
		Name:             "Check connectivity to Neo4j - neoUrl is a parameter in hieradata for this service",
		PanicGuide:       "TODO - write panic guide",
		Severity:         1,
		TechnicalSummary: fmt.Sprintf("Cannot connect to Neo4j instance %s with at least one financial instrument loaded in it", cr),
		Checker: func() (string, error) {
			return "", service.Check()
		},
	}
}

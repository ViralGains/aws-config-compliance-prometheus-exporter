package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
        "strings"
	"time"


	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/configservice"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Compliance struct {
	ConfigRuleName string
	Compliance     string
	CapExceeded    bool
	CappedCount    int64
	Env            string
}

var (
	//nolint:gochecknoglobals
	compliance = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "aws_custom",
		Subsystem: "config",
		Name:      "compliance",
		Help:      "Number of compliance",
	},
		[]string{"config_rule_name", "compliance", "cap_exceeded", "env"},
	)
)

func main() {
        fmt.Println("AWS Config Prometheus Exporter starting on port 8083")
        fmt.Println(getEnvironment())
	interval, err := getInterval()
	if err != nil {
		log.Fatal(err)
	}

	prometheus.MustRegister(compliance)

	http.Handle("/metrics", promhttp.Handler())

	go func() {
		ticker := time.NewTicker(time.Duration(interval) * time.Second)

		// register metrics as background
		for range ticker.C {
			err := snapshot()
			if err != nil {
				log.Fatal(err)
			}
		}
	}()
	log.Fatal(http.ListenAndServe(":8083", nil))
}

func snapshot() error {
	compliance.Reset()

	Compliances, err := getcompliances()
	if err != nil {
		return fmt.Errorf("failed to get compliances: %w", err)
	}

	for _, Compliance := range Compliances {
		labels := prometheus.Labels{
			"config_rule_name": Compliance.ConfigRuleName,
			"compliance":       Compliance.Compliance,
			"cap_exceeded":     strconv.FormatBool(Compliance.CapExceeded),
			"env":       Compliance.Env,
		}
		compliance.With(labels).Set(float64(Compliance.CappedCount))
	}

	return nil
}


func getEnvironment() string {
        var envstring = os.Getenv("AWS_CONFIG_SCRAPE_PREFIX")
        if len(envstring) == 0 {
          log.Fatal("You need to specify the env variable AWS_CONFIG_SCRAPE_PREFIX")
        } 
        return envstring
}

func getInterval() (int, error) {
	const defaultAWSAPIIntervalSecond = 10
	AWSAPIInterval := os.Getenv("AWS_API_INTERVAL")
	if len(AWSAPIInterval) == 0 {
		return defaultAWSAPIIntervalSecond, nil
	}

	integerAWSAPIInterval, err := strconv.Atoi(AWSAPIInterval)
	if err != nil {
		return 0, fmt.Errorf("failed to read Datadog Config: %w", err)
	}

	return integerAWSAPIInterval, nil
}

func getcompliances() ([]Compliance, error) {
	var result []*configservice.ComplianceByConfigRule

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	svc := configservice.New(sess)
	input := &configservice.DescribeComplianceByConfigRuleInput{}
	var env = getEnvironment()
	for {
		ret, err := svc.DescribeComplianceByConfigRule(input)
		if err != nil {
			return nil, fmt.Errorf("failed to describe compliance: %w", err)
		}
                filtered := filterByEnv(env, ret.ComplianceByConfigRules)
                result = append(result, filtered...)
		// pagination
		if ret.NextToken == nil {
			break
		}
		input.NextToken = ret.NextToken
	}

	Compliances := make([]Compliance, len(result))
	for i, comp := range result {
		var CapExceeded bool
		var CappedCount int64
		if comp.Compliance.ComplianceContributorCount != nil {
			CapExceeded = *comp.Compliance.ComplianceContributorCount.CapExceeded
			CappedCount = *comp.Compliance.ComplianceContributorCount.CappedCount
		}

		Compliances[i] = Compliance{
			ConfigRuleName: *comp.ConfigRuleName,
			Compliance:     *comp.Compliance.ComplianceType,
			CapExceeded:    CapExceeded,
			CappedCount:    CappedCount,
                        Env:		env,
		}
	}

	return Compliances, nil
}

func filterByEnv(env string, rules []*configservice.ComplianceByConfigRule) []*configservice.ComplianceByConfigRule {
  var filteredSlice []*configservice.ComplianceByConfigRule
  for _, element := range rules {
    if(strings.Contains(*element.ConfigRuleName, env)){
      filteredSlice = append(filteredSlice, element)
    }
  }
  return filteredSlice
}

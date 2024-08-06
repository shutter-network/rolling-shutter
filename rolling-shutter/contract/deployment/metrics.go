package deployment

import "github.com/prometheus/client_golang/prometheus"

var metricsContractDeploymentInfo = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Namespace: "shutter",
		Subsystem: "contract",
		Name:      "deployment_info",
		Help:      "Information about the contract deployments in use",
	},
	[]string{"name", "address", "chainId", "blockNumber"})

func InitMetrics() {
	prometheus.MustRegister(metricsContractDeploymentInfo)
}

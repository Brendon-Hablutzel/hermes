package prometheus

import (
	"hermes/app/common"
	"hermes/app/types"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

var _ prometheus.Collector = &basicCollector{}

type ProjectStats struct {
	Name             string
	TotalResources   int
	HealthyResources int
}

type basicCollector struct {
	TotalResources       *prometheus.Desc
	HealthyResources     *prometheus.Desc
	FailedFetchResources *prometheus.Desc
	ResourceStatusString *prometheus.Desc

	projectDefinitions []types.ProjectDefinition
	clients            common.Clients
}

func NewBasicCollector(projectDefinitions []types.ProjectDefinition, clients common.Clients) prometheus.Collector {
	return &basicCollector{
		TotalResources: prometheus.NewDesc(
			"resources_total",
			"Number of resources",
			[]string{"project", "deployment"},
			nil,
		),
		HealthyResources: prometheus.NewDesc(
			"resources_healthy",
			"Number of healthy resources",
			[]string{"project", "deployment"},
			nil,
		),
		FailedFetchResources: prometheus.NewDesc(
			"resources_failed_fetch",
			"Number of resources whose status couldn't be fetched",
			[]string{"project", "deployment"},
			nil,
		),
		ResourceStatusString: prometheus.NewDesc(
			"resource_status",
			"Status of a resource",
			[]string{"project", "deployment", "resource", "type", "status"},
			nil,
		),
		projectDefinitions: projectDefinitions,
		clients:            clients,
	}
}

func (c *basicCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.TotalResources
	ch <- c.HealthyResources
}

// https://stackoverflow.com/questions/68887416/grafana-state-timeline-panel-with-values-states-supplied-by-label
func (c *basicCollector) Collect(ch chan<- prometheus.Metric) {
	for _, project := range c.projectDefinitions {
		for _, deployment := range project.Deployments {
			var wg sync.WaitGroup
			var mu sync.Mutex
			totalResources := len(deployment.Resources)
			healthyResources := 0
			failedFetchResources := 0
			for _, resource := range deployment.Resources {
				wg.Add(1)
				go func() {
					defer wg.Done()

					status, err := common.GetResourceStatus(&c.clients, resource)

					if err != nil {
						mu.Lock()
						failedFetchResources += 1
						mu.Unlock()
						return
					}

					healthValue := 0
					if status.IsHealthy() {
						healthValue = 1
					}

					ch <- prometheus.MustNewConstMetric(
						c.ResourceStatusString,
						prometheus.GaugeValue,
						float64(healthValue),
						project.Name,
						deployment.Name,
						resource.Name,
						string(resource.Type),
						status.GetStatusString(),
					)

					if status.IsHealthy() {
						mu.Lock()
						healthyResources += 1
						mu.Unlock()
					}

				}()
			}

			wg.Wait()

			ch <- prometheus.MustNewConstMetric(
				c.TotalResources,
				prometheus.GaugeValue,
				float64(totalResources),
				project.Name,
				deployment.Name,
			)
			ch <- prometheus.MustNewConstMetric(
				c.HealthyResources,
				prometheus.GaugeValue,
				float64(healthyResources),
				project.Name,
				deployment.Name,
			)

			// TODO: do we need a per-resource metric to show which ones failed, or will this be obvious
			// if there is no entry
			ch <- prometheus.MustNewConstMetric(
				c.FailedFetchResources,
				prometheus.GaugeValue,
				float64(failedFetchResources),
				project.Name,
				deployment.Name,
			)
		}
	}

	// log.Printf("Metrics collected - Total: %d, Healthy: %d", stats.TotalResources, stats.HealthyResources)
}

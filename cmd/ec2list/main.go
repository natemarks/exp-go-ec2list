package main

import (
	"context"
	"fmt"
	"github.com/sertvitas/exp-go-ec2list/helper"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/jedib0t/go-pretty/v6/table"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Please provide a partial cluster name as an argument.")
	}

	partialClusterName := os.Args[1]

	// Load the Shared AWS Configuration (from ~/.aws/config)
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	// Create an Amazon ECS client
	ecsClient := ecs.NewFromConfig(cfg)

	// List clusters
	clustersOutput, err := ecsClient.ListClusters(context.TODO(), &ecs.ListClustersInput{})
	if err != nil {
		log.Fatalf("failed to list clusters, %v", err)
	}

	clusterArns := []string{}
	for _, clusterArn := range clustersOutput.ClusterArns {
		if strings.Contains(clusterArn, partialClusterName) {
			clusterArns = append(clusterArns, clusterArn)
		}
	}

	if len(clusterArns) == 0 {
		fmt.Printf("No matching clusters found for filter: %s\n", partialClusterName)
		return
	}

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Cluster Name", "Task Definition", "Desired Count", "Pending Count", "Running Count", "Container Image"})

	for _, clusterArn := range clusterArns {
		clusterName := helper.ExtractName(clusterArn)

		// List services in the cluster
		servicesOutput, err := ecsClient.ListServices(context.TODO(), &ecs.ListServicesInput{
			Cluster: &clusterArn,
		})
		if err != nil {
			log.Fatalf("failed to list services for cluster %s, %v", clusterArn, err)
		}

		if len(servicesOutput.ServiceArns) == 0 {
			continue
		}

		// Describe the services to get their details
		describeServicesOutput, err := ecsClient.DescribeServices(context.TODO(), &ecs.DescribeServicesInput{
			Cluster:  &clusterArn,
			Services: servicesOutput.ServiceArns,
		})
		if err != nil {
			log.Fatalf("failed to describe services for cluster %s, %v", clusterArn, err)
		}

		for _, service := range describeServicesOutput.Services {
			for _, deployment := range service.Deployments {
				taskDefinition := deployment.TaskDefinition

				// Describe the task definition to get container definitions
				taskDefinitionOutput, err := ecsClient.DescribeTaskDefinition(context.TODO(), &ecs.DescribeTaskDefinitionInput{
					TaskDefinition: taskDefinition,
				})
				if err != nil {
					log.Fatalf("failed to describe task definition %s, %v", *taskDefinition, err)
				}

				containerImages := ""
				for _, containerDef := range taskDefinitionOutput.TaskDefinition.ContainerDefinitions {
					containerImage := helper.ExtractImageName(*containerDef.Image)
					if !strings.Contains(containerImage, "nginx") {
						if containerImages != "" {
							containerImages += ", "
						}
						containerImages += containerImage
					}
				}

				t.AppendRow(table.Row{
					clusterName,
					helper.ExtractName(*taskDefinition),
					deployment.DesiredCount,
					deployment.PendingCount,
					deployment.RunningCount,
					containerImages,
				})
			}
		}
	}

	t.Render()
}

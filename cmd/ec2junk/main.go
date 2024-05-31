package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/sertvitas/exp-go-ec2list/helper"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
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

	// Iterate over each cluster
	for _, clusterArn := range clustersOutput.ClusterArns {
		clusterName := helper.ExtractName(clusterArn)

		if strings.Contains(clusterName, partialClusterName) {
			// List tasks with desired status RUNNING
			tasksOutput, err := ecsClient.ListTasks(context.TODO(), &ecs.ListTasksInput{
				Cluster:       &clusterName,
				DesiredStatus: "RUNNING",
			})
			if err != nil {
				log.Fatalf("failed to list tasks for cluster %s, %v", clusterName, err)
			}

			fmt.Printf("Cluster: %s\n", clusterName)
			for _, taskArn := range tasksOutput.TaskArns {
				taskID := helper.ExtractName(taskArn)

				// Describe the task to get the task definition ARN
				describeTasksOutput, err := ecsClient.DescribeTasks(context.TODO(), &ecs.DescribeTasksInput{
					Cluster: &clusterName,
					Tasks:   []string{taskArn},
				})
				if err != nil {
					log.Fatalf("failed to describe task %s, %v", taskID, err)
				}
				taskDefinitionArn := describeTasksOutput.Tasks[0].TaskDefinitionArn

				// Describe the task definition to get container definitions
				taskDefinitionOutput, err := ecsClient.DescribeTaskDefinition(context.TODO(), &ecs.DescribeTaskDefinitionInput{
					TaskDefinition: taskDefinitionArn,
				})
				if err != nil {
					log.Fatalf("failed to describe task definition %s, %v", *taskDefinitionArn, err)
				}

				fmt.Printf("\tTask ID: %s\n", taskID)
				for _, containerDef := range taskDefinitionOutput.TaskDefinition.ContainerDefinitions {
					containerImage := *containerDef.Image
					if !strings.Contains(containerImage, "nginx") {
						fmt.Printf("\t\tContainer Image: %s\n", containerImage)
					}
				}
			}
		}
	}
}

//func extractName(arn string) string {
//	parts := strings.Split(arn, "/")
//	return parts[len(parts)-1]
//}

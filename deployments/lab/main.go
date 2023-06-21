package main

import (
	"context"
	"flag"
	"fmt"
	"path/filepath"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

type tenant struct {
	User            string
	Num             uint64
	ResourcesLabels map[string]string
	Resources       map[string]string
}

var ResourceTypes []string
var appsID map[string]string //appID: user

func main() {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}
        num := uint64(50)
	ResourceTypes = []string{"cpu", "memory"}
	appsID = make(map[string]string, 0)
	podsClient := clientset.CoreV1().Pods(apiv1.NamespaceDefault)
	tenants := []tenant{
		tenant{
			"user1",
			num,
			map[string]string{"cpu": "2000", "memory": "8000000000", "duration": "50"},
			map[string]string{"cpu": "2", "memory": "8G", "duration": "50"},
		},
		tenant{
			"user2",
			num,
			map[string]string{"cpu": "1000", "memory": "4000000000", "duration": "200"},
			map[string]string{"cpu": "1", "memory": "4G", "duration": "200"},
		},
		tenant{
			"user3",
			num,
			map[string]string{"cpu": "8000", "memory": "2000000000", "duration": "50"},
			map[string]string{"cpu": "8", "memory": "2G", "duration": "50"},
		},
		tenant{
			"user4",
			num,
			map[string]string{"cpu": "4000", "memory": "1000000000", "duration": "200"},
			map[string]string{"cpu": "4", "memory": "1G", "duration": "200"},
		},
	}

	for _, user := range tenants {
		for index := uint64(1); index <= user.Num; index++ {
			c1Resources := make(map[apiv1.ResourceName]resource.Quantity)
			c1Resources[apiv1.ResourceMemory] = resource.MustParse(user.Resources["memory"])
			c1Resources[apiv1.ResourceCPU] = resource.MustParse(user.Resources["cpu"])
			appID := fmt.Sprintf("app-%s-%06d", user.User, index)
			appsID[appID] = user.User
			pod := &apiv1.Pod{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Pod",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("pod-%s-%06d", user.User, index),
					Namespace: "default",
					Labels: map[string]string{
						"applicationId":                appID,
						"queue":                        "root.sandbox",
						"yunikorn.apache.org/username": user.User,
						"vcore":                        user.ResourcesLabels["cpu"],
						"memory":                       user.ResourcesLabels["memory"],
						"duration":                     user.ResourcesLabels["duration"],
					},
				},
				Spec: apiv1.PodSpec{
					SchedulerName: "yunikorn",
					RestartPolicy: "Never",
					Containers: []apiv1.Container{
						{
							Name:    "sleep",
							Image:   "alpine:latest",
							Command: []string{"sleep", user.Resources["duration"]},
							Resources: apiv1.ResourceRequirements{
								Requests: c1Resources,
								Limits:   c1Resources,
							},
						},
					},
				},
			}
			_, err := podsClient.Create(context.TODO(), pod, metav1.CreateOptions{})
			if err != nil {
				panic(err)
			}
		}
		fmt.Printf("%s has %d application\n", user.User, user.Num)
	}
}

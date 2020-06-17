// Copyright 2020 The Okteto Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package init

import (
	"context"
	"fmt"

	k8Client "github.com/okteto/okteto/pkg/k8s/client"
	okLabels "github.com/okteto/okteto/pkg/k8s/labels"
	"github.com/okteto/okteto/pkg/k8s/pods"
	"github.com/okteto/okteto/pkg/k8s/replicasets"
	"github.com/okteto/okteto/pkg/k8s/services"
	"github.com/okteto/okteto/pkg/log"
	"github.com/okteto/okteto/pkg/model"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var (
	componentLabels []string = []string{"app.kubernetes.io/component", "component", "app"}
)

//SetDevDefaultsFromDeployment sets dev defaults from a running deployment
func SetDevDefaultsFromDeployment(dev *model.Dev, d *appsv1.Deployment, container string) (*model.Dev, error) {
	ctx := context.Background()
	c, config, _, err := k8Client.GetLocal()
	if err != nil {
		return nil, err
	}

	pod, err := getRunningPod(d, container, c)
	if err != nil {
		return nil, err
	}

	dev = setUserFromPod(ctx, dev, pod, container, config, c)
	dev = setWorkDirFromPod(ctx, dev, pod, container, config, c)
	dev = setCommandFromPod(ctx, dev, pod, container, config, c)
	dev = setNameAndLabelsFromDeployment(ctx, dev, d)
	dev = setAnnotationsFromDeployment(dev, d)
	dev = setResourcesFromPod(dev, pod, container)

	dev, err = setForwardsFromPod(ctx, dev, pod, c)
	if err != nil {
		return nil, err
	}

	return dev, nil
}

func getRunningPod(d *appsv1.Deployment, container string, c *kubernetes.Clientset) (*apiv1.Pod, error) {
	rs, err := replicasets.GetReplicaSetByDeployment(d, "", c)
	if err != nil {
		return nil, err
	}
	pod, err := pods.GetPodByReplicaSet(rs, "", c)
	if err != nil {
		return nil, err
	}
	if pod.Status.Phase != apiv1.PodRunning {
		return nil, fmt.Errorf("no pod is running for deployment '%s'", d.Name)
	}
	for _, containerstatus := range pod.Status.ContainerStatuses {
		if containerstatus.Name == container && containerstatus.State.Running == nil {
			return nil, fmt.Errorf("no pod is running for deployment '%s'", d.Name)
		}
	}
	return pod, nil
}

func setUserFromPod(ctx context.Context, dev *model.Dev, pod *apiv1.Pod, container string, config *rest.Config, c *kubernetes.Clientset) *model.Dev {
	if dev.WorkDir != "" {
		return dev
	}
	userID, err := pods.GetUserByPod(ctx, pod, container, config, c)
	if err != nil {
		log.Infof("error getting user of the deployment: %s", err)
	}
	if userID != 0 {
		dev.SecurityContext = &model.SecurityContext{
			RunAsUser: &userID,
		}
	}
	return dev
}

func setWorkDirFromPod(ctx context.Context, dev *model.Dev, pod *apiv1.Pod, container string, config *rest.Config, c *kubernetes.Clientset) *model.Dev {
	if dev.WorkDir != "" {
		return dev
	}
	workdir, err := pods.GetWorkdirByPod(ctx, pod, container, config, c)
	if err != nil {
		log.Infof("error getting workdir of the deployment: %s", err)
		workdir = "/okteto"
	}
	dev.WorkDir = workdir
	return dev
}

func setCommandFromPod(ctx context.Context, dev *model.Dev, pod *apiv1.Pod, container string, config *rest.Config, c *kubernetes.Clientset) *model.Dev {
	if dev.Command.Values != nil {
		return dev
	}
	if pods.CheckIfBashIsAvailable(ctx, pod, container, config, c) {
		dev.Command.Values = []string{"bash"}
	} else {
		dev.Command.Values = []string{"sh"}
	}
	return dev
}

func setForwardsFromPod(ctx context.Context, dev *model.Dev, pod *apiv1.Pod, c *kubernetes.Clientset) (*model.Dev, error) {
	ports, err := services.GetPortsByPod(pod, c)
	if err != nil {
		return nil, err
	}
	seenPorts := map[int]bool{}
	for _, f := range dev.Forward {
		seenPorts[f.Local] = true
	}
	for _, port := range ports {
		localPort := port
		if port <= 1024 {
			localPort = port + 8000
		}
		for seenPorts[localPort] {
			localPort++
		}
		seenPorts[localPort] = true
		dev.Forward = append(
			dev.Forward,
			model.Forward{
				Local:  localPort,
				Remote: port,
			},
		)
	}
	return dev, nil
}

func setNameAndLabelsFromDeployment(ctx context.Context, dev *model.Dev, d *appsv1.Deployment) *model.Dev {
	for _, l := range componentLabels {
		component := d.Labels[l]
		if component == "" {
			continue
		}
		dev.Name = component
		dev.Labels = map[string]string{l: component}
		return dev
	}
	dev.Name = d.Name
	return dev

}

func setAnnotationsFromDeployment(dev *model.Dev, d *appsv1.Deployment) *model.Dev {
	if v := d.Annotations[okLabels.FluxAnnotation]; v != "" {
		dev.Annotations = map[string]string{"fluxcd.io/ignore": "true"}
	}
	return dev
}

func setResourcesFromPod(dev *model.Dev, pod *apiv1.Pod, container string) *model.Dev {
	for i := range pod.Spec.Containers {
		if pod.Spec.Containers[i].Name != container {
			continue
		}
		if pod.Spec.Containers[i].Resources.Limits != nil {
			dev.Resources = model.ResourceRequirements{
				Limits: model.ResourceList{
					apiv1.ResourceCPU:    resource.MustParse("1"),
					apiv1.ResourceMemory: resource.MustParse("2Gi"),
				},
			}
		}
		break
	}
	return dev
}

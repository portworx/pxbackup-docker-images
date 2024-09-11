package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/logCollector/utils"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	jobResource           = "job"
	podResource           = "pod"
	cmResource            = "configmap"
	resourceQuotaResource = "resourcequota"
	prometheusResource    = "prometheus"
	alertmanagerResource  = "alertmanager"
	statefulSetResource   = "statefulset"
	deploymentResource    = "deployment"
	networkPolicyResource = "networkpolicy"
	pvcResource           = "persistentvolumeclaim"
)

func main() {

	namespaceValue := flag.String("namespace", "", "Px-backup namespace for gathering resources, If not provided it will scan the px-backup service across all the namespaces to determine the px-backup deployed namespace")
	outputDir := flag.String("output-dir", "", "Output directory for storing the gathered resources. Provide '.' for placing the output in current directory")
	tailLines := flag.Int("tail-lines", 3000, "Number of lines to tail from the logs. Provide '-1' for getting all the lines from the logs")
	kubeconfigPath := flag.String("kubeconfig", "", "Path to the kubeconfig file. If not provided, it will use the KUBECONFIG environment variable")
	flag.Parse()

	// Load kubeconfig
	var kubeconfig string
	if *kubeconfigPath != "" {
		os.Setenv("KUBECONFIG", *kubeconfigPath)
		kubeconfig = *kubeconfigPath
	} else {
		kubeconfig = os.Getenv("KUBECONFIG")
		if len(kubeconfig) == 0 {
			fmt.Println("Set KUBECONFIG environment variable or provide the path to the kubeconfig file")
			os.Exit(1)
		}
	}
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	// Create a Kubernetes client
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	var baseDir string
	if *outputDir == "" {
		baseDir = "/tmp/pxb-diags-output"
	} else {
		baseDir = filepath.Join(*outputDir, "pxb-diags-output")
	}

	var (
		jobDescribePath         = filepath.Join(baseDir, "job/describe")
		jobSpecPath             = filepath.Join(baseDir, "job/spec")
		logsPath                = filepath.Join(baseDir, "logs")
		logsPreviousPath        = filepath.Join(baseDir, "logs-previous")
		podDescribePath         = filepath.Join(baseDir, "pod/describe")
		podSpecPath             = filepath.Join(baseDir, "pod/spec")
		resourceQuotaPath       = filepath.Join(baseDir, "resourcequota")
		networkPolicyPath       = filepath.Join(baseDir, "networkpolicy")
		cmPath                  = filepath.Join(baseDir, "configmap")
		statefulSetSpecPath     = filepath.Join(baseDir, "statefulset/spec")
		statefulSetDescribePath = filepath.Join(baseDir, "statefulset/describe")
		podListPath             = filepath.Join(baseDir, "pod_list.txt")
		prometheusListPath      = filepath.Join(baseDir, "prometheus_list.txt")
		alertmanagerListPath    = filepath.Join(baseDir, "alertmanager_list.txt")
		deploymentSpecPath      = filepath.Join(baseDir, "deployment/spec")
		deploymentDescribePath  = filepath.Join(baseDir, "deployment/describe")
		pvcSpecPath             = filepath.Join(baseDir, "pvc/spec")
		pvcDescribePath         = filepath.Join(baseDir, "pvc/describe")
	)

	directories := []string{
		jobDescribePath,
		jobSpecPath,
		logsPath,
		logsPreviousPath,
		podDescribePath,
		podSpecPath,
		networkPolicyPath,
		cmPath,
		statefulSetSpecPath,
		statefulSetDescribePath,
		deploymentSpecPath,
		deploymentDescribePath,
		resourceQuotaPath,
		pvcSpecPath,
		pvcDescribePath,
	}

	for _, dir := range directories {
		err := os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			fmt.Printf("Error creating directory %s: %v\n", dir, err)
			return
		}
	}

	var namespace string
	if *namespaceValue == "" {
		services, err := clientset.CoreV1().Services("").List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			fmt.Printf("Error listing services: %s", err.Error())
		}
		for _, service := range services.Items {
			if service.Name == "px-backup" {
				namespace = service.Namespace
				break
			}
		}
	} else {
		namespace = *namespaceValue
	}

	if namespace == "" {
		fmt.Println("Px-backup not found. Please load the px-backup deployed kubeconfig and provide the namespace")
		os.Exit(1)
	}

	// Get the list of pods in the specified namespace
	pods, err := clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}

	// Get the list of jobs in the specified namespace
	jobs, err := clientset.BatchV1().Jobs(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}

	// Get the list of config maps in the specified namespace
	cms, err := clientset.CoreV1().ConfigMaps(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}

	// Get the list of stateful sets in the specified namespace
	statefulSets, err := clientset.AppsV1().StatefulSets(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}

	// Get the list of deployments in the specified namespace
	deployments, err := clientset.AppsV1().Deployments(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}

	// Fetch the list of ResourceQuota in the namespace
	quotaList, err := clientset.CoreV1().ResourceQuotas(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}

	// Fetch the list of NetworkPolicies in the namespace
	networkPolicyList, err := clientset.NetworkingV1().NetworkPolicies(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}

	// Fetch the list of PersistentVolumeClaims in the namespace
	pvcList, err := clientset.CoreV1().PersistentVolumeClaims(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}

	var wg sync.WaitGroup

	for _, networkPolicy := range networkPolicyList.Items {
		wg.Add(1)
		go func(np networkingv1.NetworkPolicy) {
			defer wg.Done()

			err := utils.WriteResourceSpecToFile(networkPolicyPath, namespace, np.Name, networkPolicyResource)
			if err != nil {
				fmt.Printf("Error writing network policy spec to file for network policy %s: %v", np.Name, err)
			} else {
				fmt.Printf("Network Policy spec written to file for NetworkPolicy: %s\n", np.Name)
			}
		}(networkPolicy)
	}

	for _, quota := range quotaList.Items {
		wg.Add(1)
		go func(quota v1.ResourceQuota) {
			defer wg.Done()

			err := utils.WriteResourceSpecToFile(resourceQuotaPath, namespace, quota.Name, resourceQuotaResource)
			if err != nil {
				fmt.Printf("Error writing resource quota spec to file for resource quota %s: %v", quota.Name, err)
			} else {
				fmt.Printf("Resource Quota spec written to file for ResourceQuota: %s\n", quota.Name)
			}
		}(quota)
	}

	for _, cm := range cms.Items {
		wg.Add(1)
		go func(cm v1.ConfigMap) {
			defer wg.Done()

			err := utils.WriteResourceSpecToFile(cmPath, namespace, cm.Name, cmResource)
			if err != nil {
				fmt.Printf("Error writing configmap spec to file for configmap %s: %v", cm.Name, err)
			} else {
				fmt.Printf("ConfigMap spec written to file for ConfigMap: %s\n", cm.Name)
			}
		}(cm)
	}

	for _, deployment := range deployments.Items {
		wg.Add(1)
		go func(dep appsv1.Deployment) {
			defer wg.Done()

			err := utils.WriteResourceSpecToFile(deploymentSpecPath, namespace, dep.Name, deploymentResource)
			if err != nil {
				fmt.Printf("Error writing deployment spec to file for deployment %s: %v", dep.Name, err)
			} else {
				fmt.Printf("Deployment spec written to file for Deployment: %s\n", dep.Name)
			}
		}(deployment)

		wg.Add(1)
		go func(dep appsv1.Deployment) {
			defer wg.Done()

			err := utils.WriteResourceDescToFile(deploymentDescribePath, namespace, dep.Name, deploymentResource)
			if err != nil {
				fmt.Printf("Error writing deployment description to file for deployment %s: %v", dep.Name, err)
			} else {
				fmt.Printf("Deployment description written to file for Deployment: %s\n", dep.Name)
			}
		}(deployment)
	}

	for _, pvc := range pvcList.Items {
		wg.Add(1)
		go func(pvc v1.PersistentVolumeClaim) {
			defer wg.Done()

			err := utils.WriteResourceSpecToFile(pvcSpecPath, namespace, pvc.Name, pvcResource)
			if err != nil {
				fmt.Printf("Error writing pvc spec to file for pvc %s: %v", pvc.Name, err)
			} else {
				fmt.Printf("PVC spec written to file for PVC: %s\n", pvc.Name)
			}
		}(pvc)

		wg.Add(1)
		go func(pvc v1.PersistentVolumeClaim) {
			defer wg.Done()

			err := utils.WriteResourceDescToFile(pvcDescribePath, namespace, pvc.Name, pvcResource)
			if err != nil {
				fmt.Printf("Error writing pvc description to file for pvc %s: %v", pvc.Name, err)
			} else {
				fmt.Printf("PVC description written to file for PVC: %s\n", pvc.Name)
			}
		}(pvc)
	}

	for _, statefulSet := range statefulSets.Items {
		wg.Add(1)
		go func(ss appsv1.StatefulSet) {
			defer wg.Done()

			err := utils.WriteResourceSpecToFile(statefulSetSpecPath, namespace, ss.Name, statefulSetResource)
			if err != nil {
				fmt.Printf("Error writing stateful set spec to file for stateful set %s: %v", ss.Name, err)
			} else {
				fmt.Printf("Stateful set spec written to file for StatefulSet: %s\n", ss.Name)
			}
		}(statefulSet)

		wg.Add(1)
		go func(ss appsv1.StatefulSet) {
			defer wg.Done()

			err := utils.WriteResourceDescToFile(statefulSetDescribePath, namespace, ss.Name, statefulSetResource)
			if err != nil {
				fmt.Printf("Error writing stateful set description to file for stateful set %s: %v", ss.Name, err)
			} else {
				fmt.Printf("Stateful set description written to file for StatefulSet: %s\n", ss.Name)
			}
		}(statefulSet)
	}

	for _, job := range jobs.Items {
		wg.Add(1)
		go func(job batchv1.Job) {
			defer wg.Done()

			err := utils.WriteResourceSpecToFile(jobSpecPath, namespace, job.Name, jobResource)
			if err != nil {
				fmt.Printf("Error writing job spec to file for job %s: %v", job.Name, err)
			} else {
				fmt.Printf("Job spec written to file for Job: %s\n", job.Name)
			}
		}(job)

		wg.Add(1)
		go func(Job batchv1.Job) {
			defer wg.Done()

			err := utils.WriteResourceDescToFile(jobDescribePath, namespace, Job.Name, jobResource)
			if err != nil {
				fmt.Printf("Error writing job description to file for job %s: %v", Job.Name, err)
			} else {
				fmt.Printf("Job description written to file for Job: %s\n", Job.Name)
			}
		}(job)
	}

	// Geg logs for all the pods
	for _, pod := range pods.Items {

		// Get current logs for all the initContainers and containers in the pod
		for _, initContainer := range pod.Spec.InitContainers {
			wg.Add(1)
			go func(podName, containerName string) {
				defer wg.Done()

				logs, err := utils.GetPodLogs(clientset, namespace, podName, containerName, int64(*tailLines), false)
				if err != nil {
					fmt.Printf("Error getting logs for pod %s, initContainer %s: %v", podName, containerName, err)
					return
				}
				err = utils.WriteLogsToFile(logsPath, podName, containerName, logs, "init-container")
				if err != nil {
					fmt.Printf("Error writing logs to file for pod %s, initContainer %s: %v", podName, containerName, err)
				} else {
					fmt.Printf("Logs written to file for Pod: %s, InitContainer: %s\n", podName, containerName)
				}
			}(pod.Name, initContainer.Name)
		}

		for _, container := range pod.Spec.Containers {
			wg.Add(1)
			go func(podName, containerName string) {
				defer wg.Done()

				logs, err := utils.GetPodLogs(clientset, namespace, podName, containerName, int64(*tailLines), false)
				if err != nil {
					fmt.Printf("Error getting logs for pod %s, container %s: %v", podName, containerName, err)
					return
				}
				err = utils.WriteLogsToFile(logsPath, podName, containerName, logs, "container")
				if err != nil {
					fmt.Printf("Error writing logs to file for pod %s, container %s: %v", podName, containerName, err)
				} else {
					fmt.Printf("Logs written to file for Pod: %s, Container: %s\n", podName, containerName)
				}
			}(pod.Name, container.Name)
		}

		// Get previous logs for all the initContainers and containers in the pod
		for _, initContainer := range pod.Spec.InitContainers {
			wg.Add(1)
			go func(podName, containerName string) {
				defer wg.Done()

				logs, err := utils.GetPodLogs(clientset, namespace, podName, containerName, int64(*tailLines), true)
				if err != nil {
					fmt.Printf("Error getting logs for pod %s, initContainer %s: %v", podName, containerName, err)
					return
				}
				if logs == "" {
					return
				}
				err = utils.WriteLogsToFile(logsPreviousPath, podName, containerName, logs, "init-container")
				if err != nil {
					fmt.Printf("Error writing logs to file for pod %s, initContainer %s: %v", podName, containerName, err)
				} else {
					fmt.Printf("Logs written to file for Pod: %s, InitContainer: %s\n", podName, containerName)
				}
			}(pod.Name, initContainer.Name)
		}

		for _, container := range pod.Spec.Containers {
			wg.Add(1)
			go func(podName, containerName string) {
				defer wg.Done()

				logs, err := utils.GetPodLogs(clientset, namespace, podName, containerName, int64(*tailLines), true)
				if err != nil {
					fmt.Printf("Error getting logs for pod %s, container %s: %v", podName, containerName, err)
					return
				}
				if logs == "" {
					return
				}
				err = utils.WriteLogsToFile(logsPreviousPath, podName, containerName, logs, "container")
				if err != nil {
					fmt.Printf("Error writing logs to file for pod %s, container %s: %v", podName, containerName, err)
				} else {
					fmt.Printf("Logs written to file for Pod: %s, Container: %s\n", podName, containerName)
				}
			}(pod.Name, container.Name)
		}

		wg.Add(1)
		go func(pod v1.Pod) {
			defer wg.Done()

			err := utils.WriteResourceSpecToFile(podSpecPath, namespace, pod.Name, podResource)
			if err != nil {
				fmt.Printf("Error writing pod spec to file for pod %s: %v", pod.Name, err)
			} else {
				fmt.Printf("Pod spec written to file for Pod: %s\n", pod.Name)
			}
		}(pod)

		wg.Add(1)
		go func(pod v1.Pod) {
			defer wg.Done()

			err := utils.WriteResourceDescToFile(podDescribePath, namespace, pod.Name, podResource)
			if err != nil {
				fmt.Printf("Error writing pod spec to file for pod %s: %v", pod.Name, err)
			} else {
				fmt.Printf("Pod Description written to file for Pod: %s\n", pod.Name)
			}
		}(pod)
	}

	utils.WriteResourceListToFile(podResource, namespace, podListPath)
	utils.WriteResourceListToFile(prometheusResource, namespace, prometheusListPath)
	utils.WriteResourceListToFile(alertmanagerResource, namespace, alertmanagerListPath)

	wg.Wait()

	fmt.Printf("Resource gathering completed successfully and have been stored in %s\n", baseDir)
}

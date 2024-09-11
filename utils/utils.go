package utils

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

// WriteResourceSpecToFile writes the spec of a resource to a file
func WriteResourceSpecToFile(filePath, namespace, resourceName, resourceType string) error {

	cmd := exec.Command("kubectl", "get", resourceType, resourceName, "-n", namespace, "-o", "yaml")
	file, err := os.Create(filepath.Join(filePath, fmt.Sprintf("%s.yaml", resourceName)))
	if err != nil {
		return fmt.Errorf("error creating file: %v", err)
	}

	defer file.Close()

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
		return err
	}

	// Write the output to a file
	_, err = file.WriteString(out.String())
	if err != nil {
		fmt.Printf("Error writing spec of %s/%s to file: %v", resourceType, resourceName, err)
	}
	return nil
}

// WriteResourceDescToFile writes the description of a resource to a file
func WriteResourceDescToFile(filePath, namespace, resourceName, resourceType string) error {

	cmd := exec.Command("kubectl", "describe", resourceType, resourceName, "-n", namespace)
	file, err := os.Create(filepath.Join(filePath, fmt.Sprintf("%s.txt", resourceName)))
	if err != nil {
		return fmt.Errorf("error creating file: %v", err)
	}

	defer file.Close()

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
		return err
	}

	// Write the output to a file
	_, err = file.WriteString(out.String())
	if err != nil {
		fmt.Printf("Error writing description of %s/%s to file: %v", resourceType, resourceName, err)
	}
	return nil
}

// GetPodLogs fetches the logs of a specific container within a pod
func GetPodLogs(clientset *kubernetes.Clientset, namespace, podName, containerName string, tailLines int64, previous bool) (string, error) {
	podLogOpts := v1.PodLogOptions{
		Container: containerName,
	}

	if tailLines > 0 {
		podLogOpts.TailLines = &tailLines
	}

	if previous {
		podLogOpts.Previous = true
	}

	req := clientset.CoreV1().Pods(namespace).GetLogs(podName, &podLogOpts)
	logs, err := req.Stream(context.TODO())

	if err != nil && strings.Contains(err.Error(), "previous terminated container") {
		return "", nil
	}

	if err != nil {
		return "", fmt.Errorf("error in getting logs: %v", err)
	}
	defer logs.Close()

	result := ""
	buf := make([]byte, 2000)
	for {
		n, err := logs.Read(buf)
		if n == 0 {
			if err == nil {
				continue
			}
			if err.Error() == "EOF" {
				break
			}
			return "", err
		}
		result += string(buf[:n])
	}

	return result, nil
}

// WriteLogsToFile writes the logs to a file named after the pod and container or initContainer
func WriteLogsToFile(outputDir, podName, containerName, logs, containerType string) error {
	filePath := filepath.Join(outputDir, fmt.Sprintf("%s_%s_%s.log", podName, containerName, containerType))
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("error creating file: %v", err)
	}
	defer file.Close()

	_, err = file.WriteString(logs)
	if err != nil {
		return fmt.Errorf("error writing logs to file: %v", err)
	}

	return nil
}

// WriteResourceListToFile writes the list of a resource to a file
func WriteResourceListToFile(resource, namespace, filePath string) {
	var cmd *exec.Cmd
	if resource == "" {
		cmd = exec.Command("kubectl", "get", resource, "--all-namespaces")
	} else {
		cmd = exec.Command("kubectl", "get", resource, "-n", namespace)
	}
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
	}

	// Write the output to a file
	err = os.WriteFile(filePath, out.Bytes(), 0644)
	if err != nil {
		fmt.Printf("Error writing %s list to file: %v\n", resource, err)
	} else {
		fmt.Printf("Successfully wrote %s list to %s\n", resource, filePath)
	}
}

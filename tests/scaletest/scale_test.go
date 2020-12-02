/*
 * Copyright 2019-2020 VMware, Inc.
 * All Rights Reserved.
* Licensed under the Apache License, Version 2.0 (the "License");
* you may not use this file except in compliance with the License.
* You may obtain a copy of the License at
*   http://www.apache.org/licenses/LICENSE-2.0
* Unless required by applicable law or agreed to in writing, software
* distributed under the License is distributed on an "AS IS" BASIS,
* WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
* See the License for the specific language governing permissions and
* limitations under the License.
*/

package scaletest

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/avinetworks/sdk/go/clients"
	"github.com/onsi/gomega"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/tests/scaletest/lib"
)

const (
	SECURE     = "secure"
	INSECURE   = "insecure"
	MULTIHOST  = "multi-host"
	CONTROLLER = "Controller"
	KUBENODE   = "Node"
)

var (
	testbedFileName          string
	namespace                string
	appName                  string
	serviceNamePrefix        string
	ingressNamePrefix        string
	akoPodName               string
	AviClients               []*clients.AviClient
	numGoRoutines            int
	listOfServicesCreated    []string
	ingressesCreated         []string
	ingressesDeleted         []string
	ingressHostNames         []string
	ingressSecureHostNames   []string
	ingressInsecureHostNames []string
	initialNumOfPools        = 0
	initialNumOfVSes         = 0
	initialNumOfFQDN         = 0
	ingressType              string
	numOfIng                 int
	clusterName              string
	timeout                  string
	dnsVSUUID                string
	testCaseTimeOut          = 1800
	testPollInterval         = "15s"
	mutex                    sync.Mutex
	REBOOTAKO                = false
	REBOOTCONTROLLER         = false
	REBOOTNODE               = false
)

func Setup() {
	var testbedParams lib.TestbedFields
	timeout = os.Args[4]
	testbedFileName = os.Args[7]
	testbed, err := os.Open(testbedFileName)
	if err != nil {
		fmt.Println("ERROR : Error opening testbed file ", testbedFileName)
		os.Exit(0)
	}
	defer testbed.Close()
	byteValue, _ := ioutil.ReadAll(testbed)
	json.Unmarshal(byteValue, &testbedParams)
	numGoRoutines, err = strconv.Atoi(os.Args[6])
	if err != nil {
		numGoRoutines = 5
	}
	if numGoRoutines <= 0 {
		fmt.Println("ERROR : Number of Go Routines cannot be zero or negative.")
		os.Exit(0)
	}
	NumOfIng, err = strconv.Atoi(os.Args[5])
	namespace = testbedParams.TestParams.Namespace
	appName = testbedParams.TestParams.AppName
	serviceNamePrefix = testbedParams.TestParams.ServiceNamePrefix
	ingressNamePrefix = testbedParams.TestParams.IngressNamePrefix
	clusterName = testbedParams.AkoParam.Clusters[0].ClusterName
	dnsVSUUID = testbedParams.TestParams.DnsVSUUID
	akoPodName = testbedParams.TestParams.AkoPodName
	os.Setenv("CTRL_USERNAME", testbedParams.Controller[0].UserName)
	os.Setenv("CTRL_PASSWORD", testbedParams.Controller[0].Password)
	os.Setenv("CTRL_IPADDRESS", testbedParams.Controller[0].Ip)
	lib.KubeInit(testbedParams.AkoParam.Clusters[0].KubeConfigFilePath)
	AviClients, err = lib.SharedAVIClients(2)
	if err != nil {
		fmt.Println("ERROR : Creating Avi Client : ", err)
		os.Exit(0)
	}
	err = lib.CreateApp(appName, namespace)
	if err != nil {
		fmt.Println("ERROR : Creation of Deployment "+appName+" failed due to the error : ", err)
		os.Exit(0)
	}
	listOfServicesCreated, err = lib.CreateService(serviceNamePrefix, appName, namespace, 2)
	if err != nil {
		fmt.Println("ERROR : Creation of Services failed due to the error : ", err)
		os.Exit(0)
	}
}

func Cleanup() {
	err := lib.DeleteService(listOfServicesCreated, namespace)
	if err != nil {
		fmt.Println("ERROR : Cleanup of Services ", listOfServicesCreated, " failed due to the error : ", err)
	}
	err = lib.DeleteApp(appName, namespace)
	if err != nil {
		fmt.Println("ERROR : Cleanup of Deployment "+appName+" failed due to the error : ", err)
	}
}

func SetupForTesting(t *testing.T) {
	pools := lib.FetchPools(t, AviClients[0])
	initialNumOfPools = len(pools)
	VSes := lib.FetchVirtualServices(t, AviClients[0])
	initialNumOfVSes = len(VSes)
	FQDNList := lib.FetchDNSARecordsFQDN(t, dnsVSUUID, AviClients[0])
	initialNumOfFQDN = len(FQDNList)
}

func Reboot(t *testing.T, nodeType string, controllerIP string, username string, password string) {
	t.Logf("Rebooting %s ... ", nodeType)
	loginID := username + "@" + controllerIP
	cmd := exec.Command("sshpass", "-p", password, "ssh", "-t", loginID, " `echo ", password, " |  sudo -S shutdown --reboot 0 && exit `")
	_, err := cmd.Output()
	if err != nil {
		fmt.Println(err.Error())
	}
	t.Logf("%s Rebooted", nodeType)
}

func ParallelReboot(t *testing.T, nodeType string, controllerIP string, username string, password string) {
	go Reboot(t, nodeType, controllerIP, username, password)
}

func RebootAko(t *testing.T) {
	akoNamespace := "avi-system"
	t.Logf("Rebooting AKO pod %s of namespace %s ...", akoPodName, akoNamespace)
	err := lib.DeletePod(akoPodName, akoNamespace)
	if err != nil {
		t.Fatalf("Cannot reboot Ako pod as : %v", err)
	}
	t.Logf("Ako rebooted.")
}

func ParallelAkoReboot(t *testing.T) {
	go RebootAko(t)
}

func DiffOfLists(list1 []string, list2 []string) []string {
	diffMap := map[string]int{}
	var diffString []string
	for _, l1 := range list1 {
		diffMap[l1] = 1
	}
	for _, l2 := range list2 {
		diffMap[l2] = diffMap[l2] + 1
	}
	var diffNum int
	for key, val := range diffMap {
		if val == 1 {
			diffNum = diffNum + 1
			diffString = append(diffString, key)
		}
	}
	return diffString
}

func PoolVerification(t *testing.T) bool {
	t.Logf("Verifying pools...")
	pools := lib.FetchPools(t, AviClients[0])
	if ingressType == MULTIHOST && (len(pools) < ((len(ingressesCreated) * 2) + initialNumOfPools)) {
		return false
	} else if len(pools) < len(ingressesCreated)+initialNumOfPools {
		return false
	}
	var ingressPoolList []string
	var poolList []string
	if ingressType == INSECURE {
		for i := 0; i < len(ingressHostNames); i++ {
			ingressPoolName := clusterName + "--" + ingressHostNames[i] + "-" + namespace + "-" + ingressesCreated[i]
			ingressPoolList = append(ingressPoolList, ingressPoolName)
		}
	} else if ingressType == SECURE {
		for i := 0; i < len(ingressHostNames); i++ {
			ingressPoolName := clusterName + "--" + namespace + "-" + ingressHostNames[i] + "-" + ingressesCreated[i]
			ingressPoolList = append(ingressPoolList, ingressPoolName)
		}
	} else if ingressType == MULTIHOST {
		for i := 0; i < len(ingressSecureHostNames); i++ {
			ingressPoolName := clusterName + "--" + namespace + "-" + ingressSecureHostNames[i] + "-" + ingressesCreated[i]
			ingressPoolList = append(ingressPoolList, ingressPoolName)
			ingressPoolName = clusterName + "--" + ingressInsecureHostNames[i] + "-" + namespace + "-" + ingressesCreated[i]
			ingressPoolList = append(ingressPoolList, ingressPoolName)
		}
	}
	for i := 0; i < len(pools); i++ {
		poolList = append(poolList, *pools[i].Name)
	}
	diffNum := len(DiffOfLists(ingressPoolList, poolList))
	if diffNum == initialNumOfPools {
		return true
	}
	return false
}

func DNSARecordsVerification(t *testing.T, hostNames []string) bool {
	t.Logf("Verifying DNS A Records...")
	FQDNList := lib.FetchDNSARecordsFQDN(t, dnsVSUUID, AviClients[0])
	diffString := DiffOfLists(FQDNList, hostNames)
	if len(diffString) == initialNumOfFQDN {
		return true
	}
	return false
}

func VSVerification(t *testing.T) bool {
	t.Logf("Verifying VSes...")
	VSes := lib.FetchVirtualServices(t, AviClients[0])
	var ingressVSList []string
	var VSList []string
	for i := 0; i < len(ingressesCreated); i++ {
		if ingressType != MULTIHOST {
			ingressVSName := clusterName + "--" + ingressesCreated[i] + ".avi.internal"
			ingressVSList = append(ingressVSList, ingressVSName)
		} else {
			ingressVSName := clusterName + "--" + ingressesCreated[i] + "-secure.avi.internal"
			ingressVSList = append(ingressVSList, ingressVSName)
		}
	}
	for i := 0; i < len(VSes); i++ {
		VSList = append(VSList, *VSes[i].Name)
	}
	diffNum := len(DiffOfLists(ingressVSList, VSList))
	if diffNum == initialNumOfVSes {
		return true
	}
	return false
}

func Verify(t *testing.T) bool {
	if ingressType == SECURE {
		if PoolVerification(t) == true && VSVerification(t) == true && DNSARecordsVerification(t, ingressHostNames) == true {
			t.Logf("Pools, VSes and DNS A Records verified")
			return true
		}
	} else if ingressType == MULTIHOST {
		hostName := append(ingressSecureHostNames, ingressInsecureHostNames...)
		if PoolVerification(t) == true && VSVerification(t) == true && DNSARecordsVerification(t, hostName) == true {
			t.Logf("Pools, VSes and DNS A Records verified")
			return true
		}
	} else if ingressType == INSECURE {
		if PoolVerification(t) == true && DNSARecordsVerification(t, ingressHostNames) == true {
			t.Logf("Pools and DNS A Records verified")
			return true
		}
	}
	return false
}

func parallelInsecureIngressCreation(t *testing.T, wg *sync.WaitGroup, serviceName string, namespace string, numOfIng int, startIndex int) {
	defer wg.Done()
	ingresses, hostNames, err := lib.CreateInsecureIngress(ingressNamePrefix, serviceName, namespace, numOfIng, startIndex)
	if err != nil {
		t.Fatalf("Failed to create %s ingresses as : %v", ingressType, err)
	}
	ingressesCreated = append(ingressesCreated, ingresses...)
	ingressHostNames = append(ingressHostNames, hostNames...)
}

func parallelSecureIngressCreation(t *testing.T, wg *sync.WaitGroup, serviceName string, namespace string, numOfIng int, startIndex int) {
	defer wg.Done()
	ingresses, hostNames, err := lib.CreateSecureIngress(ingressNamePrefix, serviceName, namespace, numOfIng, startIndex)
	if err != nil {
		t.Fatalf("Failed to create %s ingresses as : %v", ingressType, err)
	}
	ingressesCreated = append(ingressesCreated, ingresses...)
	ingressHostNames = append(ingressHostNames, hostNames...)
}

func parallelMultiHostIngressCreation(t *testing.T, wg *sync.WaitGroup, serviceName []string, namespace string, numOfIng int, startIndex int) {
	defer wg.Done()
	ingresses, secureHostNames, insecureHostNames, err := lib.CreateMultiHostIngress(ingressNamePrefix, serviceName, namespace, numOfIng, startIndex)
	if err != nil {
		t.Fatalf("Failed to create %s ingresses as : %v", ingressType, err)
	}
	ingressesCreated = append(ingressesCreated, ingresses...)
	ingressSecureHostNames = append(ingressSecureHostNames, secureHostNames...)
	ingressInsecureHostNames = append(ingressInsecureHostNames, insecureHostNames...)
}

func parallelIngressDeletion(t *testing.T, wg *sync.WaitGroup, namespace string, listOfIngressToDelete []string) {
	defer wg.Done()
	ingresses, err := lib.DeleteIngress(namespace, listOfIngressToDelete)
	if err != nil {
		t.Fatalf("Failed to delete ingresses as : %v", err)
	}
	ingressesDeleted = append(ingressesDeleted, ingresses...)
}

func CreateIngressesParallel(t *testing.T, numOfIng int, initialNumOfPools int) {
	ingressesCreated = []string{}
	var blockSize = numOfIng / numGoRoutines
	var remIng = numOfIng % numGoRoutines
	g := gomega.NewGomegaWithT(t)
	var wg sync.WaitGroup
	nextStartInd := 0
	switch {
	case ingressType == INSECURE:
		t.Logf("Creating %d %s Ingresses Parallely...", numOfIng, ingressType)
		if REBOOTAKO == true {
			wg.Add(1)
			go ParallelAkoReboot(t)
		}
		if REBOOTCONTROLLER == true {
			wg.Add(1)
			go ParallelReboot(t, CONTROLLER, os.Getenv("CTRL_IPADDRESS"), os.Getenv("CTRL_USERNAME"), os.Getenv("CTRL_PASSWORD"))
		}
		if REBOOTNODE == true {
			wg.Add(1)
			go ParallelReboot(t, KUBENODE, os.Getenv("CTRL_IPADDRESS"), os.Getenv("CTRL_USERNAME"), os.Getenv("CTRL_PASSWORD"))
		}
		for i := 0; i < numGoRoutines; i++ {
			wg.Add(1)
			if i+1 <= remIng {
				go parallelInsecureIngressCreation(t, &wg, listOfServicesCreated[0], namespace, blockSize+1, nextStartInd)
				nextStartInd = nextStartInd + blockSize + 1
			} else {
				go parallelInsecureIngressCreation(t, &wg, listOfServicesCreated[0], namespace, blockSize, nextStartInd)
				nextStartInd = nextStartInd + blockSize
			}
		}
	case ingressType == SECURE:
		t.Logf("Creating %d %s Ingresses Parallely...", numOfIng, ingressType)
		for i := 0; i < numGoRoutines; i++ {
			wg.Add(1)
			if i+1 <= remIng {
				go parallelSecureIngressCreation(t, &wg, listOfServicesCreated[0], namespace, blockSize+1, nextStartInd)
				nextStartInd = nextStartInd + blockSize + 1
			} else {
				go parallelSecureIngressCreation(t, &wg, listOfServicesCreated[0], namespace, blockSize, nextStartInd)
				nextStartInd = nextStartInd + blockSize
			}
		}
	case ingressType == MULTIHOST:
		t.Logf("Creating %d %s Ingresses Parallely...", numOfIng, ingressType)
		for i := 0; i < numGoRoutines; i++ {
			wg.Add(1)
			if (i + 1) <= remIng {
				go parallelMultiHostIngressCreation(t, &wg, listOfServicesCreated, namespace, blockSize+1, nextStartInd)
				nextStartInd = nextStartInd + blockSize + 1
			} else {
				go parallelMultiHostIngressCreation(t, &wg, listOfServicesCreated, namespace, blockSize, nextStartInd)
				nextStartInd = nextStartInd + blockSize
			}
		}
	}
	wg.Wait()
	g.Expect(ingressesCreated).To(gomega.HaveLen(numOfIng))
	t.Logf("Created %d %s Ingresses Parallely", numOfIng, ingressType)
	t.Logf("Verifiying Avi objects ...")
	pollInterval, _ := time.ParseDuration(testPollInterval)
	waitTimeIncr, _ := strconv.Atoi(testPollInterval[:len(testPollInterval)-1])
	for waitTime := 0; waitTime < testCaseTimeOut; {
		if Verify(t) == true {
			return
		}
		time.Sleep(pollInterval)
		waitTime = waitTime + waitTimeIncr
	}
	t.Fatalf("Error : Verification failed\n")
}

func DeleteIngressesParallel(t *testing.T, numOfIng int, initialNumOfPools int, AviClient *clients.AviClient) {
	var blockSize = numOfIng / numGoRoutines
	var remIng = numOfIng % numGoRoutines
	g := gomega.NewGomegaWithT(t)
	var wg sync.WaitGroup
	ingressesDeleted = []string{}
	t.Logf("Deleting %d %s Ingresses...", numOfIng, ingressType)
	nextStartInd := 0
	for i := 0; i < numGoRoutines; i++ {
		wg.Add(1)
		if (i + 1) <= remIng {
			go parallelIngressDeletion(t, &wg, namespace, ingressesCreated[nextStartInd:nextStartInd+blockSize+1])
			nextStartInd = nextStartInd + blockSize + 1
		} else {
			go parallelIngressDeletion(t, &wg, namespace, ingressesCreated[nextStartInd:nextStartInd+blockSize])
			nextStartInd = nextStartInd + blockSize
		}
	}
	wg.Wait()
	g.Expect(ingressesDeleted).To(gomega.HaveLen(numOfIng))
	t.Logf("Deleted %d %s Ingresses", numOfIng, ingressType)
	t.Logf("Verifiying Avi objects ...")
	g.Eventually(func() int {
		pools := lib.FetchPools(t, AviClient)
		return len(pools)
	}, testCaseTimeOut, testPollInterval).Should(gomega.Equal(initialNumOfPools))
	t.Logf("Deleted %d Pools", numOfIng)
}

func CreateIngressesSerial(t *testing.T, numOfIng int, initialNumOfPools int) {
	g := gomega.NewGomegaWithT(t)
	var err error
	switch {
	case ingressType == INSECURE:
		t.Logf("Creating %d %s Ingresses Serially...", numOfIng, ingressType)
		ingressesCreated, ingressHostNames, err = lib.CreateInsecureIngress(ingressNamePrefix, listOfServicesCreated[0], namespace, numOfIng)
		if err != nil {
			t.Fatalf("Failed to create %s ingresses as : %v", ingressType, err)
		}
	case ingressType == SECURE:
		t.Logf("Creating %d %s Ingresses Serially...", numOfIng, ingressType)
		ingressesCreated, ingressHostNames, err = lib.CreateSecureIngress(ingressNamePrefix, listOfServicesCreated[0], namespace, numOfIng)
		if err != nil {
			t.Fatalf("Failed to create %s ingresses as : %v", ingressType, err)
		}
	case ingressType == MULTIHOST:
		t.Logf("Creating %d %s Ingresses Serially...", numOfIng, ingressType)
		ingressesCreated, ingressSecureHostNames, ingressInsecureHostNames, err = lib.CreateMultiHostIngress(ingressNamePrefix, listOfServicesCreated, namespace, numOfIng)
		if err != nil {
			t.Fatalf("Failed to create %s ingresses as : %v", ingressType, err)
		}
	}
	g.Expect(ingressesCreated).To(gomega.HaveLen(numOfIng))
	t.Logf("Created %d %s Ingresses Serially", numOfIng, ingressType)
	t.Logf("Verifiying Avi objects ...")
	pollInterval, _ := time.ParseDuration(testPollInterval)
	waitTimeIncr, _ := strconv.Atoi(testPollInterval[:len(testPollInterval)-1])
	for waitTime := 0; waitTime < testCaseTimeOut; {
		if Verify(t) == true {
			return
		}
		time.Sleep(pollInterval)
		waitTime = waitTime + waitTimeIncr
	}
	t.Fatalf("Error : Verification failed\n")

}

func DeleteIngressesSerial(t *testing.T, numOfIng int, initialNumOfPools int, AviClient *clients.AviClient) {
	g := gomega.NewGomegaWithT(t)
	t.Logf("Deleting %d %s Ingresses Serially...", numOfIng, ingressType)
	ingressesDeleted, err := lib.DeleteIngress(namespace, ingressesCreated)
	if err != nil {
		t.Fatalf("Failed to delete ingresses as : %v", err)
	}
	g.Expect(ingressesDeleted).To(gomega.HaveLen(numOfIng))
	t.Logf("Deleted %d %s Ingresses Serially", numOfIng, ingressType)
	t.Logf("Verifiying Avi objects ...")
	g.Eventually(func() int {
		pools := lib.FetchPools(t, AviClient)
		return len(pools)
	}, testCaseTimeOut, testPollInterval).Should(gomega.Equal(initialNumOfPools))
	t.Logf("Deleted %d Pools", numOfIng)
}

func HybridCreation(t *testing.T, wg *sync.WaitGroup, numOfIng int, startIndex int) {
	mutex.Lock()
	ingresses, _, _ := lib.CreateSecureIngress(ingressNamePrefix, listOfServicesCreated[0], namespace, numOfIng, startIndex)
	t.Logf("Created ingresses %s", ingresses)
	ingressesCreated = append(ingressesCreated, ingresses...)
	mutex.Unlock()
	defer wg.Done()
}

func HybridDeletion(t *testing.T, wg *sync.WaitGroup, numOfIng int) {
	mutex.Lock()
	wg.Add(1)
	IngressList, _ := lib.ListIngress(t, namespace)
	var ingresses []string
	if len(IngressList) > 0 {
		ingresses = append(ingresses, IngressList...)
		deletedIngresses, err := lib.DeleteIngress(namespace, ingresses)
		if err != nil {
			fmt.Println("Error deleting ingresses -> ", err)
		}
		t.Logf("Deleted ingresses %s", deletedIngresses)
		ingressesDeleted = append(ingressesDeleted, deletedIngresses...)
	}
	mutex.Unlock()
	defer wg.Done()
}

func HybridExecution(t *testing.T, numOfIng int, deletionStartPoint int) {
	g := gomega.NewGomegaWithT(t)
	var wg sync.WaitGroup
	var err error
	ingressesCreated, _, err = lib.CreateInsecureIngress(ingressNamePrefix, listOfServicesCreated[0], namespace, deletionStartPoint)
	if err != nil {
		t.Fatalf("Failed to create %s ingresses as : %v", ingressType, err)
	}
	fmt.Println("Created ingresses in the start ", deletionStartPoint)
	for i := deletionStartPoint; i < numOfIng; i++ {
		wg.Add(1)
		go HybridCreation(t, &wg, 1, i)
	}
	for len(ingressesDeleted) < numOfIng {
		go HybridDeletion(t, &wg, numOfIng)
		time.Sleep(1 * time.Second)
	}
	HybridDeletion(t, &wg, numOfIng)
	wg.Wait()
	g.Expect(ingressesCreated).To(gomega.HaveLen(numOfIng))
	g.Expect(ingressesDeleted).To(gomega.HaveLen(numOfIng))
}

func CreateIngressParallelWithAkoReboot(t *testing.T) {
	SetupForTesting(t)
	REBOOTAKO = true
	CreateIngressesParallel(t, numOfIng, initialNumOfPools)
	REBOOTAKO = false
}

func CreateIngressParallelWithControllerReboot(t *testing.T) {
	SetupForTesting(t)
	CreateIngressesParallel(t, numOfIng, initialNumOfPools)
	ParallelAkoReboot(t)
}

func TestMain(t *testing.M) {
	Setup()
	t.Run()
	Cleanup()
}

// func TestReboot(t *testing.T) {
// 	RebootController(t, "10.79.169.144", "admin", "Aviuser123")
// 	lib.DeletePod("static-web", "default")
// }

// func TestUpdateIngress(t *testing.T) {
// 	ingressesCreated, _, err := lib.CreateInsecureIngress(ingressNamePrefix, listOfServicesCreated[0], namespace, 5)
// 	if err != nil {
// 		t.Fatalf("Failed to create %s ingresses as : %v", ingressType, err)
// 	}
// 	t.Logf("waiting")
// 	time.Sleep(10 * time.Second)
// 	_, _ = lib.UpdateIngress(namespace, ingressesCreated[0:1])
// }

// func TestMixedExecution(t *testing.T) {
// 	SetupForTesting(t)
// 	HybridExecution(t, 10, 3)
// }

// func TestParallel200InsecureIngresses(t *testing.T) {
// 	ingressType = INSECURE
// 	ParallelIngressHelper(t, 200)
// }

// func TestParallel200SecureIngresses(t *testing.T) {
// 	ingressType = SECURE
// 	ParallelIngressHelper(t, 200)
// }

// func TestParallel200MultiHostIngresses(t *testing.T) {
// 	ingressType = MULTIHOST
// 	SerialIngressHelper(t, 1)
// var err error
// ingressesCreated, ingressSecureHostNames, ingressInsecureHostNames, err = lib.CreateMultiHostIngress(ingressNamePrefix, listOfServicesCreated, namespace, 5, 0)
// if err != nil {
// 	t.Fatalf("Failed to create %s ingresses as : %v", ingressType, err)
// }
// for i := 0; i < len(ingressesCreated); i++ {
// 	t.Logf("%s\t%s\t%s\t", ingressesCreated[i], ingressSecureHostNames[i], ingressInsecureHostNames[i])
// }
// }

// func TestSerial200InsecureIngresses(t *testing.T) {
// 	ingressType = INSECURE
// 	SerialIngressHelper(t, 200)
// }

// func TestSerial200SecureIngresses(t *testing.T) {
// 	ingressType = SECURE
// 	SerialIngressHelper(t, 200)
// }

// func TestSerial200MultiHostIngresses(t *testing.T) {
// 	ingressType = MULTIHOST
// 	SerialIngressHelper(t, 200)
// }

// func TestParallel500InsecureIngresses(t *testing.T) {
// 	ingressType = INSECURE
// 	ParallelIngressHelper(t, 500)
// }

// func TestParallel500SecureIngresses(t *testing.T) {
// 	ingressType = SECURE
// 	ParallelIngressHelper(t, 500)
// }

// func TestParallel500MultiHostIngresses(t *testing.T) {
// 	ingressType = MULTIHOST
// 	ParallelIngressHelper(t, 500)
// }

// func TestSerial500InsecureIngresses(t *testing.T) {
// 	ingressType = INSECURE
// 	SerialIngressHelper(t, 500)
// }

// func TestSerial500SecureIngresses(t *testing.T) {
// 	ingressType = SECURE
// 	SerialIngressHelper(t, 500)
// }

// func TestSerial500MultiHostIngresses(t *testing.T) {
// 	ingressType = MULTIHOST
// 	SerialIngressHelper(t, 500)
// }

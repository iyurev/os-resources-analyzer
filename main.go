package main

import (
	"fmt"
	"github.com/fatih/color"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"log"
	"os"
	"path/filepath"
	"text/tabwriter"
)

const defaultConfName = "config"

type maxResourceVal struct {
	PodName   string
	Value     int64
	Namespace string
}

type clusterQuotaReport struct {
	SumCpuRequests int64
	SumCpuLimits   int64
	SumMemRequests int64
	SumMemLimits   int64
}

func (cqr *clusterQuotaReport) ToHumanReadableVal() {
	cqr.SumCpuRequests = MilCoreToCore(cqr.SumCpuRequests)
	cqr.SumCpuLimits = MilCoreToCore(cqr.SumCpuLimits)
	cqr.SumMemRequests = BytesToGi(cqr.SumMemRequests)
	cqr.SumMemLimits = BytesToGi(cqr.SumMemLimits)
}

func (cqr *clusterQuotaReport) PrettyPrint() {
	cqr.ToHumanReadableVal()
	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', tabwriter.AlignRight)
	title := fmt.Sprintf("%s\t%s\t%s\t%s\t", "CLUSTER CPU REQUESTS:", "CLUSTER CPU LIMITS:", "CLUSTER MEMORY REQUESTS:", "CLUSTER MEMORY LIMITS:")
	reportVal := fmt.Sprintf("%d core\t%d core\t%d Mi\t%d Mi\t", cqr.SumCpuRequests, cqr.SumCpuLimits, cqr.SumMemRequests, cqr.SumMemLimits)
	colorTitle := color.New(color.FgRed, color.Bold)
	magColorLine := color.New(color.FgMagenta, color.Bold)

	_, err := colorTitle.Fprintln(tw, title)
	if err != nil {
		log.Fatal(err)
	}
	_, err = magColorLine.Fprintln(tw, reportVal)
	if err != nil {
		log.Fatal(err)
	}
	err = tw.Flush()
	if err != nil {
		log.Fatal(err)
	}

}

type nodeReport struct {
	NodeName        string
	MaxCpuRequest   maxResourceVal
	MaxMemRequest   maxResourceVal
	MaxCpuLimit     maxResourceVal
	MaxMemLimit     maxResourceVal
	SumCputRequests int64
	SumMemRequests  int64
	SumCpuLimits    int64
	SumMemLimits    int64
}

func (nr *nodeReport) PrettyPrint() {
	nr.ToHumanReadableVal()
	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', tabwriter.AlignRight)

	MaxCpuRequestTitle := fmt.Sprintf("%s\t%s\t%s\t", "Max CPU request:", "Namespace:", "Pod name:")
	MaxCpuLimittTitle := fmt.Sprintf("%s\t%s\t%s\t", "Max CPU limit:", "Namespace:", "Pod name:")
	MaxMemRequestTitle := fmt.Sprintf("%s\t%s\t%s\t", "Max Memory request:", "Namespace:", "Pod name:")
	MaxMemLimitTitle := fmt.Sprintf("%s\t%s\t%s\t", "Max Memory limit:", "Namespace:", "Pod name:")

	SumCpuRequestReport := fmt.Sprintf("All requested CPU:\t %d\t", nr.SumCputRequests)
	SumMemRequestReport := fmt.Sprintf("All requested MEMORY:\t %d\t", nr.SumMemRequests)

	NodeNameTitle := fmt.Sprintf("Node name: %s\n", nr.NodeName)
	MaxCpuRequestReport := fmt.Sprintf("%d Cpu\t%s\t%s\t", nr.MaxCpuRequest.Value, nr.MaxCpuRequest.Namespace, nr.MaxCpuRequest.PodName)
	MaxCpuLimitReport := fmt.Sprintf("%d Cpu\t%s\t%s\t", nr.MaxCpuLimit.Value, nr.MaxCpuLimit.Namespace, nr.MaxCpuLimit.PodName)
	MaxMemRequestReport := fmt.Sprintf("%d Gi\t%s\t%s\t", nr.MaxMemRequest.Value, nr.MaxMemRequest.Namespace, nr.MaxMemRequest.PodName)
	MaxMemLimitReport := fmt.Sprintf("%d Gi\t%s\t%s\t", nr.MaxMemLimit.Value, nr.MaxMemLimit.Namespace, nr.MaxMemLimit.PodName)

	colorTitle := color.New(color.FgBlue, color.Bold)
	magColorLine := color.New(color.FgMagenta, color.Bold)
	greeColorLine := color.New(color.FgGreen, color.Bold)
	greeColorLine.Fprintln(tw, NodeNameTitle)

	colorTitle.Fprintln(tw, MaxCpuRequestTitle)
	magColorLine.Fprintln(tw, MaxCpuRequestReport)

	colorTitle.Fprintln(tw, MaxCpuLimittTitle)
	magColorLine.Fprintln(tw, MaxCpuLimitReport)

	colorTitle.Fprintln(tw, MaxMemRequestTitle)
	magColorLine.Fprintln(tw, MaxMemRequestReport)

	colorTitle.Fprintln(tw, MaxMemLimitTitle)
	magColorLine.Fprintln(tw, MaxMemLimitReport)

	magColorLine.Fprintln(tw, SumCpuRequestReport)
	magColorLine.Fprintln(tw, SumMemRequestReport)

	tw.Flush()

}

func (nr *nodeReport) ToHumanReadableVal() {
	nr.MaxCpuRequest.Value = MilCoreToCore(nr.MaxCpuRequest.Value)
	nr.MaxMemRequest.Value = BytesToGi(nr.MaxMemRequest.Value)
	nr.MaxCpuLimit.Value = MilCoreToCore(nr.MaxCpuLimit.Value)
	nr.MaxMemLimit.Value = BytesToGi(nr.MaxMemLimit.Value)
}

func (mrv *maxResourceVal) assignContext(obj *v1.ObjectMeta) {
	mrv.PodName = obj.Name
	mrv.Namespace = obj.Namespace
}

func RunningPodFromNodeOpt(nodeName string) v1.ListOptions {
	podState := "Running"
	fieldSelector := fmt.Sprintf("spec.nodeName=%s,status.phase=%s", nodeName, podState)
	return v1.ListOptions{FieldSelector: fieldSelector}

}

func MilCoreToCore(mc int64) int64 {
	return (mc / 1000)
}

func BytesToGi(bts int64) int64 {
	return bts / 1024 / 1024 / 1024 / 1024
}

func ClusterQuotaReport(clientset *kubernetes.Clientset) (*clusterQuotaReport, error) {
	clusterReport := &clusterQuotaReport{}
	allQuotas, err := clientset.CoreV1().ResourceQuotas("").List(v1.ListOptions{})
	if err != nil {
		return clusterReport, err
	}
	if len(allQuotas.Items) != 0 {
		for _, quota := range allQuotas.Items {
			cpuRequests := quota.Spec.Hard["requests.cpu"]
			cpuLimits := quota.Spec.Hard["limits.cpu"]
			memRequests := quota.Spec.Hard["requests.memory"]
			memLimits := quota.Spec.Hard["limits.memory"]
			clusterReport.SumCpuRequests += cpuRequests.MilliValue()
			clusterReport.SumCpuLimits += cpuLimits.MilliValue()
			clusterReport.SumMemRequests += memRequests.MilliValue()
			clusterReport.SumMemLimits += memLimits.MilliValue()
		}

		return clusterReport, nil
	}
	return clusterReport, fmt.Errorf("Empty quota list!!\n")
}

func CreteNodeReport(clientset *kubernetes.Clientset, nodeName string) (nodeReport, error) {
	podList, err := clientset.CoreV1().Pods("").List(RunningPodFromNodeOpt(nodeName))
	if err != nil {
		return nodeReport{}, err
	}
	NodeReport := nodeReport{}
	NodeReport.NodeName = nodeName

	for _, pod := range podList.Items {

		for _, container := range pod.Spec.Containers {
			cpuRequest := container.Resources.Requests.Cpu().MilliValue()
			cpuLimit := container.Resources.Limits.Cpu().MilliValue()
			memRequest := container.Resources.Requests.Memory().MilliValue()
			memLimit := container.Resources.Limits.Memory().MilliValue()
			if cpuRequest > NodeReport.MaxCpuRequest.Value {
				NodeReport.MaxCpuRequest.Value = cpuRequest
				NodeReport.MaxCpuRequest.PodName = pod.Name
				NodeReport.MaxCpuRequest.Namespace = pod.Namespace
			}
			if memRequest > NodeReport.MaxMemRequest.Value {
				NodeReport.MaxMemRequest.Value = memRequest
				NodeReport.MaxMemRequest.assignContext(&pod.ObjectMeta)
			}
			if cpuLimit > NodeReport.MaxCpuLimit.Value {
				NodeReport.MaxCpuLimit.Value = cpuLimit
				NodeReport.MaxCpuLimit.assignContext(&pod.ObjectMeta)
			}
			if memLimit > NodeReport.MaxMemLimit.Value {
				NodeReport.MaxMemLimit.Value = memLimit
				NodeReport.MaxMemLimit.assignContext(&pod.ObjectMeta)
			}
			NodeReport.SumCputRequests += cpuRequest
			NodeReport.SumCpuLimits += cpuLimit
			NodeReport.SumMemRequests += memRequest
			NodeReport.SumMemLimits += memLimit

		}
	}
	return NodeReport, nil
}

func main() {

	homePath := homedir.HomeDir()
	kubeConfPath := filepath.Join(homePath, ".kube", defaultConfName)

	kubeConfig, err := clientcmd.BuildConfigFromFlags("", kubeConfPath)
	if err != nil {
		log.Fatal(err)
	}
	client, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		log.Fatal(err)
	}
	nodeReport, err := CreteNodeReport(client, "os-node-05")
	if err != nil {
		log.Fatal(err)
	}
	nodeReport.PrettyPrint()
	//quotaReport, err := ClusterQuotaReport(client)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//quotaReport.PrettyPrint()

}

//fmt.Printf("Sum all CPU requests: %d\n", MilCoreToCore(sumCpuRequests))
//fmt.Printf("Sum all CPU limits: %d\n", MilCoreToCore(sumCpuLimits))
//fmt.Printf("Sum all MEMORY requests:  %d\n",  BytesToGi(sumMemRequests))
//fmt.Printf("Sum all MEMORY limits:  %d\n", BytesToGi(sumMemLimits))
//
//fmt.Printf("Max CPU request is: %d\n", maxCpuRequest.Value)
//fmt.Printf("Max CPU request is: %d\n", maxCpuLimit.Value)
//
//fmt.Printf("Max MEMORY request is: %d\n", BytesToGi(maxMemRequest.Value))
//fmt.Printf("Max MEMORY limit is: %d, Namespace: %s, Pod name: %s\n", BytesToGi(MaxMemLimit.Value), MaxMemLimit.Namespace, MaxMemLimit.PodName)

package main

import (
	"flag"
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
	SumCpuRequests     int64
	SumCpuLimits       int64
	SumMemRequests     int64
	SumMemLimits       int64
	SumUsedCpuRequets  int64
	SumUsedCpuLimits   int64
	SumUsedMemRequests int64
	SumUsedMemLimits   int64
}

func (cqr *clusterQuotaReport) ToHumanReadableVal() {
	cqr.SumCpuRequests = MilCoreToCore(cqr.SumCpuRequests)
	cqr.SumCpuLimits = MilCoreToCore(cqr.SumCpuLimits)
	cqr.SumMemRequests = BytesToGi(cqr.SumMemRequests)
	cqr.SumMemLimits = BytesToGi(cqr.SumMemLimits)

	cqr.SumUsedCpuRequets = MilCoreToCore(cqr.SumUsedCpuRequets)
	cqr.SumUsedMemRequests = BytesToGi(cqr.SumUsedMemRequests)
	cqr.SumUsedCpuLimits = MilCoreToCore(cqr.SumUsedCpuLimits)
	cqr.SumUsedMemLimits = BytesToGi(cqr.SumUsedMemLimits)
}

func (cqr *clusterQuotaReport) PrettyPrint() {
	cqr.ToHumanReadableVal()
	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', tabwriter.AlignRight)
	titleHardSpec := fmt.Sprintf("%s\t%s\t%s\t%s\t", "ALLOCATED CPU REQUESTS:", "ALLOCATED CPU LIMITS:", "ALLOCATED MEMORY REQUESTS:", "ALLOCATED MEMORY LIMITS:")
	titleStatusUsed := fmt.Sprintf("%s\t%s\t%s\t%s\t", "USED CPU REQUESTS:", "USED CPU LIMITS:", "USED MEMORY REQUESTS:", "USED MEMORY LIMITS:")
	reportAllocatedQuota := fmt.Sprintf("%d core\t%d core\t%d Gi\t%d Gi\t", cqr.SumCpuRequests, cqr.SumCpuLimits, cqr.SumMemRequests, cqr.SumMemLimits)
	reportUsedQuota := fmt.Sprintf("%d core\t%d core\t%d Gi\t%d Gi\t", cqr.SumUsedCpuRequets, cqr.SumUsedCpuLimits, cqr.SumUsedMemRequests, cqr.SumUsedMemLimits)
	colorTitle := color.New(color.FgRed, color.Bold)
	magColorLine := color.New(color.FgMagenta, color.Bold)

	_, err := colorTitle.Fprintln(tw, titleHardSpec)
	if err != nil {
		log.Fatal(err)
	}
	_, err = magColorLine.Fprintln(tw, reportAllocatedQuota)
	if err != nil {
		log.Fatal(err)
	}

	_, err = colorTitle.Fprintln(tw, titleStatusUsed)
	if err != nil {
		log.Fatal(err)
	}
	_, err = magColorLine.Fprintln(tw, reportUsedQuota)
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

	SummoryNodeResourcesTitle := fmt.Sprintf("%s\t%s\t%s\t%s\t", "All requested CPU:", "All requested MEMORY", "All CPU Limits:", "All MEMORY Limits")
	SummoryNodeResourceReport := fmt.Sprintf("%d Core\t%d Gi\t%d Core\t%d Gi\t", nr.SumCputRequests, nr.SumMemRequests, nr.SumCpuLimits, nr.SumMemLimits)

	NodeNameTitle := fmt.Sprintf("Node name: %s\n", nr.NodeName)
	MaxCpuRequestReport := fmt.Sprintf("%d Cpu\t%s\t%s\t", nr.MaxCpuRequest.Value, nr.MaxCpuRequest.Namespace, nr.MaxCpuRequest.PodName)
	MaxCpuLimitReport := fmt.Sprintf("%d Cpu\t%s\t%s\t", nr.MaxCpuLimit.Value, nr.MaxCpuLimit.Namespace, nr.MaxCpuLimit.PodName)
	MaxMemRequestReport := fmt.Sprintf("%d Gi\t%s\t%s\t", nr.MaxMemRequest.Value, nr.MaxMemRequest.Namespace, nr.MaxMemRequest.PodName)
	MaxMemLimitReport := fmt.Sprintf("%d Gi\t%s\t%s\t", nr.MaxMemLimit.Value, nr.MaxMemLimit.Namespace, nr.MaxMemLimit.PodName)

	redTitle := color.New(color.FgRed, color.Bold)
	yellowColorLine := color.New(color.FgYellow)

	colorTitle := color.New(color.FgBlue, color.Bold)
	magColorLine := color.New(color.FgMagenta, color.Bold)
	greeColorLine := color.New(color.FgGreen, color.Bold)
	_, err := greeColorLine.Fprintln(tw, NodeNameTitle)
	if err != nil {
		log.Fatal(err)
	}
	_, err = colorTitle.Fprintln(tw, MaxCpuRequestTitle)
	if err != nil {
		log.Fatal(err)
	}
	_, err = magColorLine.Fprintln(tw, MaxCpuRequestReport)
	if err != nil {
		log.Fatal(err)
	}
	_, err = colorTitle.Fprintln(tw, MaxCpuLimittTitle)
	if err != nil {
		log.Fatal(err)
	}
	_, err = magColorLine.Fprintln(tw, MaxCpuLimitReport)
	if err != nil {
		log.Fatal(err)
	}

	_, err = colorTitle.Fprintln(tw, MaxMemRequestTitle)
	if err != nil {
		log.Fatal(err)
	}
	_, err = magColorLine.Fprintln(tw, MaxMemRequestReport)
	if err != nil {
		log.Fatal(err)
	}
	_, err = colorTitle.Fprintln(tw, MaxMemLimitTitle)
	if err != nil {
		log.Fatal(err)
	}
	_, err = magColorLine.Fprintln(tw, MaxMemLimitReport)
	if err != nil {
		log.Fatal(err)
	}

	_, err = redTitle.Fprintln(tw, SummoryNodeResourcesTitle)
	if err != nil {
		log.Fatal(err)
	}
	_, err = yellowColorLine.Fprintln(tw, SummoryNodeResourceReport)
	if err != nil {
		log.Fatal(err)
	}
	err = tw.Flush()
	if err != nil {
		log.Fatal(err)
	}

}

func (nr *nodeReport) ToHumanReadableVal() {
	nr.MaxCpuRequest.Value = MilCoreToCore(nr.MaxCpuRequest.Value)
	nr.MaxMemRequest.Value = BytesToGi(nr.MaxMemRequest.Value)
	nr.MaxCpuLimit.Value = MilCoreToCore(nr.MaxCpuLimit.Value)
	nr.MaxMemLimit.Value = BytesToGi(nr.MaxMemLimit.Value)
	nr.SumCputRequests = MilCoreToCore(nr.SumCputRequests)
	nr.SumCpuLimits = MilCoreToCore(nr.SumCpuLimits)
	nr.SumMemRequests = BytesToGi(nr.SumMemRequests)
	nr.SumMemLimits = BytesToGi(nr.SumMemLimits)
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

func PendingPodFromNodeOpt(nodeName string) v1.ListOptions {
	podState := "Pending"
	fieldSelector := fmt.Sprintf("spec.nodeName=%s,status.phase=%s", nodeName, podState)
	return v1.ListOptions{FieldSelector: fieldSelector}

}

func MilCoreToCore(mc int64) int64 {
	return (mc / 1000)
}

func BytesToGi(bts int64) int64 {
	return bts / 1024 / 1024 / 1024
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

			usedCpuRequest := quota.Status.Used["requests.cpu"]
			usedMemRequest := quota.Status.Used["requests.memory"]
			usedCpuLimits := quota.Status.Used["limits.cpu"]
			UsedMemLimits := quota.Status.Used["limits.memory"]
			clusterReport.SumCpuRequests += cpuRequests.MilliValue()
			clusterReport.SumCpuLimits += cpuLimits.MilliValue()
			clusterReport.SumMemRequests += memRequests.Value()
			clusterReport.SumMemLimits += memLimits.Value()
			clusterReport.SumUsedCpuRequets += usedCpuRequest.MilliValue()
			clusterReport.SumUsedMemRequests += usedMemRequest.Value()
			clusterReport.SumUsedCpuLimits += usedCpuLimits.MilliValue()
			clusterReport.SumUsedMemLimits += UsedMemLimits.Value()
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
	pendingPodsList, err := clientset.CoreV1().Pods("").List(PendingPodFromNodeOpt(nodeName))
	if err != nil {
		return nodeReport{}, err
	}
	allPods := append(podList.Items, pendingPodsList.Items...)
	if len(allPods) == 0 {
		log.Fatal("Empty Pods list!!")
	}
	NodeReport := nodeReport{}
	NodeReport.NodeName = nodeName

	for _, pod := range allPods {

		for _, container := range pod.Spec.Containers {
			cpuRequest := container.Resources.Requests.Cpu().MilliValue()
			cpuLimit := container.Resources.Limits.Cpu().MilliValue()
			memRequest := container.Resources.Requests.Memory().Value()
			memLimit := container.Resources.Limits.Memory().Value()
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

var nodeName string
var clusterReport bool

func init() {
	flag.StringVar(&nodeName, "node-name", "", "OpenShift node name.")
	flag.BoolVar(&clusterReport, "cluster-report", false, "Get cluster resource report.")
	flag.Parse()
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

	if nodeName != "" {
		nodeReport, err := CreteNodeReport(client, nodeName)
		if err != nil {
			log.Fatal(err)
		}
		nodeReport.PrettyPrint()
	}
	if clusterReport {
		quotaReport, err := ClusterQuotaReport(client)
		if err != nil {
			log.Fatal(err)
		}
		quotaReport.PrettyPrint()
	}
}

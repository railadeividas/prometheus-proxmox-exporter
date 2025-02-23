package main

// https://pve.proxmox.com/pve-docs/api-viewer/
// https://pbs.proxmox.com/docs/command-syntax.html#proxmox-backup-manager

import (
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"time"
	"net/http"
	"sort"
	"strings"
	"flag"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/client_golang/prometheus/collectors"
)

type VM struct {
	Cpu				float64 `json:"cpu"`
	Disk			int64		`json:"disk"`
	DiskRead	int64		`json:"diskread"`
	DiskWrite	int64		`json:"diskwrite"`
	ID				string	`json:"id"`
	MaxCpu		int64		`json:"maxcpu"`
	MaxDisk		int64		`json:"maxdisk"`
	MaxMem		int64		`json:"maxmem"`
	Mem				int64		`json:"mem"`
	Name			string	`json:"name"`
	NetIn			int64		`json:"netin"`
	NetOut		int64		`json:"netout"`
	Node			string	`json:"node"`
	Status		string	`json:"status"`
	Template	int64	  `json:"template"`
	Type			string	`json:"type"`
	Uptime		int64		`json:"uptime"`
	VMID			int64		`json:"vmid"`
}

type NODE struct {
	CgroupMode	int64   `json:"cgroup-mode"`
	Cpu					float64	`json:"cpu"`
	Disk				int64		`json:"disk"`
	ID					string	`json:"id"`
	Level				string	`json:"level"`
	MaxCpu			int64		`json:"maxcpu"`
	MaxDisk			int64		`json:"maxdisk"`
	MaxMem			int64		`json:"maxmem"`
	Mem					int64		`json:"mem"`
	Node				string	`json:"node"`
	Status			string	`json:"status"`
	Type				string	`json:"type"`
	Uptime			int64  	`json:"uptime"`
}

type STORAGE struct {
	Content     string `json:"content"`
	Disk        int64  `json:"disk"`
	ID			    string `json:"id"`
	MaxDisk     int64  `json:"maxdisk"`
	Node	      string `json:"node"`
	PluginType	string `json:"plugintype"`
	Shared	    int64	 `json:"shared"`
	Status			string `json:"status"`
	Storage			string `json:"storage"`
	Type			  string `json:"type"`
}

var (
	pveResourceCpu = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "pve_resource_cpu",
			Help: "CPU utilization",
		},
		[]string{"id", "name", "node", "template", "type", "vmid"},
	)

	pveResourceDisk = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "pve_resource_disk",
			Help: "Used disk space in bytes",
		},
		[]string{"id", "name", "node", "template", "type", "vmid"},
	)

	pveResourceMaxDisk = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "pve_resource_maxdisk",
			Help: "Storage size in bytes",
		},
		[]string{"id", "name", "node", "template", "type", "vmid"},
	)

	pveResourceDiskRead = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "pve_resource_diskread",
			Help: "The amount of bytes the guest read from its block devices since the guest was started",
		},
		[]string{"id", "name", "node", "template", "type", "vmid"},
	)

	pveResourceDiskWrite = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "pve_resource_diskwrite",
			Help: "The amount of bytes the guest wrote to its block devices since the guest was started",
		},
		[]string{"id", "name", "node", "template", "type", "vmid"},
	)

	pveResourceMaxCpu = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "pve_resource_maxcpu",
			Help: "Number of available CPUs",
		},
		[]string{"id", "name", "node", "template", "type", "vmid"},
	)

	pveResourceMaxMem = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "pve_resource_maxmem",
			Help: "Number of available memory in bytes",
		},
		[]string{"id", "name", "node", "template", "type", "vmid"},
	)

	pveResourceMem = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "pve_resource_mem",
			Help: "Used memory in bytes",
		},
		[]string{"id", "name", "node", "template", "type", "vmid"},
	)

	pveResourceNetIn = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "pve_resource_netin",
			Help: "The amount of traffic in bytes that was sent to the guest over the network since it was started",
		},
		[]string{"id", "name", "node", "template", "type", "vmid"},
	)

	pveResourceNetOut = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "pve_resource_netout",
			Help: "The amount of traffic in bytes that was sent from the guest over the network since it was started",
		},
		[]string{"id", "name", "node", "template", "type", "vmid"},
	)

	pveResourceUptime = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "pve_resource_uptime",
			Help: "Uptime of node or virtual guest in seconds",
		},
		[]string{"id", "name", "node", "template", "type", "vmid"},
	)

	pveResourceStatus = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "pve_resource_status",
			Help: "Resource type dependent status",
		},
		[]string{"id", "name", "node", "template", "type", "vmid"},
	)

	pveStorageDisk = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "pve_storage_disk",
			Help: "Used disk space in bytes",
		},
		[]string{"id", "name", "node", "type", "plugintype", "shared", "content"},
	)

	pveStorageMaxDisk = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "pve_storage_maxdisk",
			Help: "Storage size in bytes",
		},
		[]string{"id", "name", "node", "type", "plugintype", "shared", "content"},
	)

	pveStorageStatus = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "pve_storage_status",
			Help: "Storage size in bytes",
		},
		[]string{"id", "name", "node", "type", "plugintype", "shared", "content"},
	)

	web_listen_address string
	runtime_scrape_interval int
	buildTime = "1970-01-01T00:00:00Z"
	version = "0.0.0"

)

func init() {
	// Unregister default collectors
	prometheus.Unregister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
	prometheus.Unregister(collectors.NewGoCollector())

	// Register Prometheus metrics
	prometheus.MustRegister(pveResourceCpu)
	prometheus.MustRegister(pveResourceDisk)
	prometheus.MustRegister(pveResourceDiskRead)
	prometheus.MustRegister(pveResourceDiskWrite)
	prometheus.MustRegister(pveResourceMaxCpu)
	prometheus.MustRegister(pveResourceMaxDisk)
	prometheus.MustRegister(pveResourceMaxMem)
	prometheus.MustRegister(pveResourceMem)
	prometheus.MustRegister(pveResourceNetIn)
	prometheus.MustRegister(pveResourceNetOut)
	prometheus.MustRegister(pveResourceUptime)
	prometheus.MustRegister(pveResourceStatus)

	prometheus.MustRegister(pveStorageDisk)
	prometheus.MustRegister(pveStorageMaxDisk)
	prometheus.MustRegister(pveStorageStatus)

}

func fetchAndUpdateVMpveResourceMetrics() error {
	cmd := exec.Command("pvesh", "get", "/cluster/resources", "--type", "vm", "--output-format", "json")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("error executing pvesh command: %w", err)
	}

	var vms []VM
	if err := json.Unmarshal(output, &vms); err != nil {
		return fmt.Errorf("error parsing JSON: %w", err)
	}

	for _, vm := range vms {
		labelID := fmt.Sprintf("node/%s/%s", vm.Node, vm.ID)
		labelVMID := fmt.Sprintf("%d", vm.VMID)
		labelNode := vm.Node
		labelTemplate := fmt.Sprintf("%d", vm.Template)
		labelType := vm.Type

		statusValue := 0
		if vm.Status == "running" {
			statusValue = 1
		}

		// Update Prometheus metrics
		pveResourceCpu.WithLabelValues(labelID, vm.Name, labelNode, labelTemplate, labelType, labelVMID).Set(vm.Cpu)
		pveResourceDisk.WithLabelValues(labelID, vm.Name, labelNode, labelTemplate, labelType, labelVMID).Set(float64(vm.Disk))
		pveResourceDiskRead.WithLabelValues(labelID, vm.Name, labelNode, labelTemplate, labelType, labelVMID).Set(float64(vm.DiskRead))
		pveResourceDiskWrite.WithLabelValues(labelID, vm.Name, labelNode, labelTemplate, labelType, labelVMID).Set(float64(vm.DiskWrite))
		pveResourceMaxCpu.WithLabelValues(labelID, vm.Name, labelNode, labelTemplate, labelType, labelVMID).Set(float64(vm.MaxCpu))
		pveResourceMaxDisk.WithLabelValues(labelID, vm.Name, labelNode, labelTemplate, labelType, labelVMID).Set(float64(vm.MaxDisk))
		pveResourceMaxMem.WithLabelValues(labelID, vm.Name, labelNode, labelTemplate, labelType, labelVMID).Set(float64(vm.MaxMem))
		pveResourceMem.WithLabelValues(labelID, vm.Name, labelNode, labelTemplate, labelType, labelVMID).Set(float64(vm.Mem))
		pveResourceNetIn.WithLabelValues(labelID, vm.Name, labelNode, labelTemplate, labelType, labelVMID).Set(float64(vm.NetIn))
		pveResourceNetOut.WithLabelValues(labelID, vm.Name, labelNode, labelTemplate, labelType, labelVMID).Set(float64(vm.NetOut))
		pveResourceUptime.WithLabelValues(labelID, vm.Name, labelNode, labelTemplate, labelType, labelVMID).Set(float64(vm.Uptime))
		pveResourceStatus.WithLabelValues(labelID, vm.Name, labelNode, labelTemplate, labelType, labelVMID).Set(float64(statusValue))
	}

	return nil
}


func fetchAndUpdateNodepveResourceMetrics() error {
	cmd := exec.Command("pvesh", "get", "/cluster/resources", "--type", "node", "--output-format", "json")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("error executing pvesh command: %w", err)
	}

	var nodes []NODE
	if err := json.Unmarshal(output, &nodes); err != nil {
		return fmt.Errorf("error parsing JSON: %w", err)
	}

	for _, node := range nodes {
		labelID := node.ID
		labelNode := node.Node
		labelType := node.Type
		labelName := node.Node

		statusValue := 0
		if node.Status == "online" {
			statusValue = 1
		}

		// Update Prometheus metrics
		pveResourceCpu.WithLabelValues(labelID, labelName, labelNode, "0", labelType, "0").Set(node.Cpu)
		pveResourceDisk.WithLabelValues(labelID, labelName, labelNode, "0", labelType, "0").Set(float64(node.Disk))
		pveResourceMaxCpu.WithLabelValues(labelID, labelName, labelNode, "0", labelType, "0").Set(float64(node.MaxCpu))
		pveResourceMaxDisk.WithLabelValues(labelID, labelName, labelNode, "0", labelType, "0").Set(float64(node.MaxDisk))
		pveResourceMaxMem.WithLabelValues(labelID, labelName, labelNode, "0", labelType, "0").Set(float64(node.MaxMem))
		pveResourceMem.WithLabelValues(labelID, labelName, labelNode, "0", labelType, "0").Set(float64(node.Mem))
		pveResourceUptime.WithLabelValues(labelID, labelName, labelNode, "0", labelType, "0").Set(float64(node.Uptime))
		pveResourceStatus.WithLabelValues(labelID, labelName, labelNode, "0", labelType, "0").Set(float64(statusValue))
	}

	return nil
}


func fetchAndUpdateStoragepveResourceMetrics() error {
	cmd := exec.Command("pvesh", "get", "/cluster/resources", "--type", "storage", "--output-format", "json")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("error executing pvesh command: %w", err)
	}

	var storages []STORAGE
	if err := json.Unmarshal(output, &storages); err != nil {
		return fmt.Errorf("error parsing JSON: %w", err)
	}

	for _, storage := range storages {
		labelID := storage.ID
		labelName := storage.Storage
		labelNode := storage.Node
		labelType := storage.Type
		labelPluginType := storage.PluginType
		labelShared := fmt.Sprintf("%d", storage.Shared)

		// Sort and format the content list
		contentList := strings.Split(storage.Content, ",")
		sort.Strings(contentList)
		labelContent := strings.Join(contentList, ",")

		statusValue := 0
		if storage.Status == "available" {
			statusValue = 1
		}

		// Update Prometheus metrics
		pveStorageDisk.WithLabelValues(labelID, labelName, labelNode, labelType, labelPluginType, labelShared, labelContent).Set(float64(storage.Disk))
		pveStorageMaxDisk.WithLabelValues(labelID, labelName, labelNode, labelType, labelPluginType, labelShared, labelContent).Set(float64(storage.MaxDisk))
		pveStorageStatus.WithLabelValues(labelID, labelName, labelNode, labelType, labelPluginType, labelShared, labelContent).Set(float64(statusValue))
	}

	return nil
}


func main() {
	showVersion := flag.Bool("version", false, "Print version and build time")
	flag.StringVar(&web_listen_address, "web.listen-address", "[::]:9221", "Address on which to expose metrics and web server")
	flag.IntVar(&runtime_scrape_interval, "runtime.scrape_interval", 10, "Interval in seconds to scrape proxmox metrics")
	flag.Parse()

	if *showVersion {
		fmt.Printf("Prometheus Proxmox Exporter Version: %s\n", version)
		fmt.Printf("Build Time: %s\n", buildTime)
		return
	}

	http.Handle("/metrics", promhttp.Handler())

	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Metrics fetcher crashed: %v", r)
			}
		}()

		ticker := time.NewTicker(time.Duration(runtime_scrape_interval) * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			if err := fetchAndUpdateVMpveResourceMetrics(); err != nil {
				log.Printf("Error fetching VM metrics: %v", err)
			}
			if err := fetchAndUpdateNodepveResourceMetrics(); err != nil {
				log.Printf("Error fetching Node metrics: %v", err)
			}
			if err := fetchAndUpdateStoragepveResourceMetrics(); err != nil {
				log.Printf("Error fetching Storage metrics: %v", err)
			}
		}
	}()

	log.Println("Prometheus proxmox exporter server started on", web_listen_address)
	log.Fatal(http.ListenAndServe(web_listen_address, nil))
}

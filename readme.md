# Prometheus Proxmox Exporter

Prometheus exporter for collecting metrics from a Proxmox:

- [x] Proxmox Virtual Environment (PVE) cluster
- [x] Proxmox Virtual Environment (PVE) node
- [ ] Proxmox Backup Server
- [ ] Proxmox Mail Gateway

## Features

- Collects resource usage metrics from Proxmox nodes, VMs, and containers
- Compatible with Prometheus and Grafana for monitoring and visualization

## Requirements

- Proxmox VE 8.0 or later

## Build from source

To build from source, you need Golang 1.22.0 or later.

1. Clone the repository:
   ```sh
   git clone https://github.com/railadeividas/prometheus-proxmox-exporter.git
   cd prometheus-proxmox-exporter
   ```

2. Build the exporter:
   ```sh
   go build -ldflags "-X main.buildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ)" -o prometheus-proxmox-exporter main.go
   ```

3. Run the exporter:
   ```sh
   chmod +x prometheus-proxmox-exporter
   ./prometheus-proxmox-exporter
   ```

## Download and Run Release

1. Download the latest release from GitHub:
   ```sh
   wget https://github.com/railadeividas/prometheus-proxmox-exporter/releases/latest/download/prometheus-proxmox-exporter
   ```

2. Make it executable:
   ```sh
   chmod +x prometheus-proxmox-exporter
   ```

3. Run the exporter:
   ```sh
   ./prometheus-proxmox-exporter --web.listen-address="[::]:9221" --runtime.scrape_interval=10
   ```

## Running as a Systemd Service

To run the exporter as a systemd service, follow these steps:

1. Create a systemd service file:
   ```sh
   sudo nano /etc/systemd/system/prometheus-proxmox-exporter.service
   ```

2. Add the following content:
   ```ini
   [Unit]
   Description=Prometheus Proxmox Exporter
   After=network.target

   [Service]
   User=root
   Group=root
   ExecStart=/usr/local/bin/prometheus-proxmox-exporter --web.listen-address="[::]:9221" --runtime.scrape_interval=10
   Restart=always
   RestartSec=10s

   [Install]
   WantedBy=multi-user.target
   ```

3. Move the binary to `/usr/local/bin/`:
   ```sh
   sudo mv prometheus-proxmox-exporter /usr/local/bin/
   ```

4. Reload systemd and enable the service:
   ```sh
   sudo systemctl daemon-reload
   sudo systemctl enable prometheus-proxmox-exporter
   sudo systemctl start prometheus-proxmox-exporter
   ```

5. Check the service status:
   ```sh
   sudo systemctl status prometheus-proxmox-exporter
   ```

## Configuration

The exporter supports environment variables for configuration:

| Variable                  | Description                      | Default     |
|---------------------------|----------------------------------|-------------|
| `web.listen-address`      | Address to listen on for metrics | `[::]:9221` |
| `runtime.scrape_interval` | Interval for scraping metrics    | `10`        |

Example usage:
```sh
./prometheus-proxmox-exporter --web.listen-address="[::]:9221" --runtime.scrape_interval=10
```

## Prometheus Configuration

Add the exporter as a scrape target in your Prometheus configuration:

```yaml
scrape_configs:
  - job_name: 'proxmox'
    static_configs:
      - targets: ['localhost:9221']
```

## Exported Metrics

[Raw example can be found here](https://raw.githubusercontent.com/railadeividas/prometheus-proxmox-exporter/refs/heads/master/docs/metrics_example.ini)

| Metric                   | Description                                                                                       |
|--------------------------|---------------------------------------------------------------------------------------------------|
| `pve_resource_cpu`       | CPU utilization                                                                                   |
| `pve_resource_disk`      | Used disk space in bytes                                                                          |
| `pve_resource_diskread`  | The amount of bytes the guest read from its block devices since the guest was started             |
| `pve_resource_diskwrite` | The amount of bytes the guest wrote to its block devices since the guest was started              |
| `pve_resource_maxcpu`    | Number of available CPUs                                                                          |
| `pve_resource_maxdisk`   | Storage size in bytes                                                                             |
| `pve_resource_maxmem`    | Number of available memory in bytes                                                               |
| `pve_resource_mem`       | Used memory in bytes                                                                              |
| `pve_resource_netin`     | The amount of traffic in bytes that was sent to the guest over the network since it was started   |
| `pve_resource_netout`    | The amount of traffic in bytes that was sent from the guest over the network since it was started |
| `pve_resource_status`    | Resource type dependent status                                                                    |
| `pve_resource_uptime`    | Uptime of node or virtual guest in seconds                                                        |
| `pve_storage_disk`       | Used disk space in bytes                                                                          |
| `pve_storage_maxdisk`    | Storage size in bytes                                                                             |
| `pve_storage_status`     | Storage size in bytes                                                                             |

## Grafana Dashboard

You can visualize the collected metrics in Grafana by importing a pre-built dashboard or creating your own.

- [Grafana dashboard JSON code](https://raw.githubusercontent.com/railadeividas/prometheus-proxmox-exporter/refs/heads/master/docs/grafana-dashboard-code.json)
- [Sreenshot of Proxmox dashboard](https://raw.githubusercontent.com/railadeividas/prometheus-proxmox-exporter/refs/heads/master/docs/grafana-dashboard-image.png)

<p>
<a href="https://raw.githubusercontent.com/railadeividas/prometheus-proxmox-exporter/refs/heads/master/docs/grafana-dashboard-image.png">
  <img
    width="600"
    alt="Grafana dashboard"
    src="https://raw.githubusercontent.com/railadeividas/prometheus-proxmox-exporter/refs/heads/master/docs/grafana-dashboard-image.png">
</a>
</p>

## License

This project is licensed under the MIT License.

## Contributions

Contributions are welcome! Feel free to open an issue or submit a pull request.

## Contact

For questions or support, open an issue in this repository.

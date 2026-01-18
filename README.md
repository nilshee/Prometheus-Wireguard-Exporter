# Prometheus Wireguard Exporter
A simple minimalistic wireguard connection stats exporter for Prometheus.
 
## Usage
```wireguard_exporter -p 9011 -l 127.0.0.1 -i=wg1,wg2,wg3```

### Command-line Options
| Flag | Descriptions  |  Required                    |
| :-------- | :------- | :-------------------------------- |
| `-p` | exporter listening port| No (defaults to 9011)|
| `-l` | address to listen on (e.g., 127.0.0.1, 192.168.1.10) | No (defaults to all interfaces)|
| `-i` | list of comma separated interface names to monitor  | No (monitors all if not specified)| 
| `-auth-user` | basic auth username | No (authentication disabled if not provided)|
| `-auth-pass` | basic auth password | No (authentication disabled if not provided)|
| `-verbose` | enable verbose logging (logs each request with source IP) | No (defaults to false)|

### Basic Authentication
To enable basic authentication, provide both `-auth-user` and `-auth-pass` flags:
```bash
wireguard_exporter -p 9011 -auth-user prometheus -auth-pass secretpassword
```

If only one of the authentication flags is provided, authentication will be disabled. Both username and password must be specified to enable basic auth.

When basic auth is enabled, Prometheus needs to be configured with the credentials:
```yaml
scrape_configs:
  - job_name: 'wireguard'
    static_configs:
      - targets: ['localhost:9011']
    basic_auth:
      username: 'prometheus'
      password: 'secretpassword'
```

### Verbose Logging
Enable verbose logging to see detailed information about each request including the source IP address:
```bash
wireguard_exporter -p 9011 -verbose
```

With verbose logging enabled, each request will be logged with:
- HTTP method and path
- HTTP status code
- Request duration
- Source IP address (respects `X-Forwarded-For` and `X-Real-IP` headers)

Example log output:
```
[GET] /metrics HTTP/1.1 - Status: 200 - Duration: 2.5ms - Source IP: 192.168.1.10:54321
```

# Exported metrics
- Latest Handshake 
- Bytes Received
- Bytes Transmitted

<img width="1508" alt="Screenshot 2024-02-19 at 6 01 37 PM" src="https://github.com/sathiraumesh/Prometheus-Wireguard-Exporter/assets/28914919/83327a18-ff5b-426a-bce8-bcbdb6750606">

## Deployment
Currently, there are no binaries. To build from the source run the following command in the project repository. Just so you know, this build is not the static binary.

```bash
  make
```

### Systemd Service Setup

To run the Wireguard exporter as a systemd service, follow these steps:

#### 1. Build and Install the Binary

```bash
# Build the binary
cd src
go build -o wireguard_exporter ./cmd/

# Install to system directory
mkdir -p /opt/wireguard_exporter
sudo cp wireguard_exporter /opt/wireguard_exporter/
sudo chmod +x /opt/wireguard_exporter/wireguard_exporter
```

#### 2. Create a Systemd Service File

Create a service file at `/etc/systemd/system/wireguard_exporter.service`:

```bash
sudo nano /etc/systemd/system/wireguard_exporter.service
```

Add the following content:

```ini
[Unit]
Description=Prometheus Wireguard Exporter
Documentation=https://github.com/sathiraumesh/Prometheus-Wireguard-Exporter
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=root
Group=root

# Adjust command-line flags as needed
ExecStart=/opt/wireguard_exporter/wireguard_exporter -p 9011

# Security settings
CapabilityBoundingSet=CAP_NET_ADMIN
AmbientCapabilities=CAP_NET_ADMIN
NoNewPrivileges=true
ProtectSystem=strict
ProtectHome=true
PrivateTmp=true

# Restart policy
Restart=on-failure
RestartSec=5s

# Logging
StandardOutput=journal
StandardError=journal
SyslogIdentifier=wireguard_exporter

[Install]
WantedBy=multi-user.target
```

**Configuration Options:**

To enable basic authentication:
```ini
ExecStart=/opt/wireguard_exporter/wireguard_exporter -p 9011 -auth-user prometheus -auth-pass secretpassword
```

To enable verbose logging:
```ini
ExecStart=/opt/wireguard_exporter/wireguard_exporter -p 9011 -verbose
```

To monitor specific interfaces:
```ini
ExecStart=/opt/wireguard_exporter/wireguard_exporter -p 9011 -i wg0,wg1
```

## Test
```bash
  make test
```

## Run Locally
This small setup was created to simulate and show the exporter in action. I have created an environment with multiple containers communicating via wireguard VPN. The setup includes Prometheus and Grafana configured to showcase the metrics. To start setup clone the project and go to the project directory.


Make sure docker, docker-compose, and make utility are installed. Run the following command to create a setup 


Run the project in a local setup
```bash
  make run
```

Monitor the metrics using Grafana Dashboard using the default password and username 
```admin, admin```

Import the dashboard from path 
```setup/grafana-provisioning```
```bash
http://localhost:3000/dashboards
```

provider "google" {
  project = "go-saga-microservices"
  region  = "us-central1"
  zone    = "us-central1-a"
}

resource "google_compute_instance" "small_vm" {
  name         = "small-20250816-103930"
  machine_type = "e2-small"
  zone         = "us-central1-a"

  can_ip_forward       = false
  deletion_protection  = false

  tags = ["http-server", "https-server"]

  labels = {
    goog-ops-agent-policy = "v2-x86-template-1-4-0"
  }

  boot_disk {
    auto_delete = true
    initialize_params {
      image = "debian-cloud/debian-12"
      size  = 10
    }
  }

  network_interface {
    network    = "default"
    subnetwork = "default"

    // TODO: check this
    access_config {
      // Ephemeral external IP (Terraform will allocate automatically)
      // If you want to force the static IP 146.148.87.133:
      nat_ip = "146.148.87.133"
    }
  }

  service_account {
    email  = "109803801845-compute@developer.gserviceaccount.com"
    scopes = [
      "https://www.googleapis.com/auth/devstorage.read_only",
      "https://www.googleapis.com/auth/logging.write",
      "https://www.googleapis.com/auth/monitoring.write",
      "https://www.googleapis.com/auth/service.management.readonly",
      "https://www.googleapis.com/auth/servicecontrol",
      "https://www.googleapis.com/auth/trace.append"
    ]
  }

  scheduling {
    preemptible       = true
    automatic_restart = false
    on_host_maintenance = "TERMINATE"
  }

  metadata = {
    enable-osconfig = "TRUE"
    ssh-keys = <<EOT
afmoroz:ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAID63agPJPBVFFr/4hvbSesTj085M7RPNmbFSlaFQ+0Ts afmoroz
EOT

    startup-script = <<EOT
#! /bin/bash
set -eux

mkdir app

sudo apt-get update -y
sudo apt-get install -y rsync make curl

curl -fsSL https://get.docker.com | sh

sudo usermod -aG docker $USER

sudo systemctl enable docker
sudo systemctl start docker

docker --version
docker-compose version
rsync --version
make --version

echo "Startup script completed successfully!"
EOT
  }
}

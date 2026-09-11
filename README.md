# sif.io

## Service deployment

The checked-in service configuration supports two deployment targets. The
systemd units and Compose files describe a Linux host deployment; the
`config/{smtp,www}/production.yml` manifests describe Kubernetes deployments.
They are independent deployment methods.

### Linux host

The host deployment keeps the Git checkout at
`/opt/sif.io/github_buckelij_sif.io` and installs runtime files under
`/opt/sif.io`. Native services run as the `sifio` user:

| Service | systemd unit | Executable | Environment file |
| --- | --- | --- | --- |
| SMTP and webmail | `sifiosmtp.service` | `/opt/sif.io/smtp/sif.io/main` | `/opt/sif.io/smtp/etc/env` |
| Website | `sifiowww.service` | `/opt/sif.io/www/sif.io/main` | `/opt/sif.io/www/etc/env` |

Install the unit files from `config/services` into `/etc/systemd/system`, then
reload systemd and enable the required services:

```sh
sudo install -m 0644 config/services/sifiosmtp.service /etc/systemd/system/
sudo install -m 0644 config/services/sifiowww.service /etc/systemd/system/
sudo install -m 0644 config/services/sifiosmtp-socat.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable --now sifiosmtp.service sifiowww.service sifiosmtp-socat.service
```

`sifiosmtp-socat.service` publishes TCP port 25 and forwards it to the SMTP
service on `127.0.0.1:1025`.

Deploy application changes from the canonical checkout by building the
relevant Go executable for the host architecture. The output paths below are
runtime targets, not Git working trees; do not run Git commands from those
directories:

```sh
cd /opt/sif.io/github_buckelij_sif.io
git fetch

go build -o /opt/sif.io/smtp/sif.io/main ./cmd/smtp
sudo systemctl restart sifiosmtp.service

go build -o /opt/sif.io/www/sif.io/main ./cmd/www
sudo systemctl restart sifiowww.service
```

The runtime directories may contain older source copies from previous
deployments, but systemd runs only their `main` executables. A binary built in
one of those copies appears as an untracked `main` file in that copy's Git
status.

When a unit file changes, reinstall it, run `sudo systemctl daemon-reload`,
then restart the affected unit. Keep `/opt/sif.io/{smtp,www}/etc/env` outside
source control and restrict it to the service user because it contains runtime
configuration and secrets.

### Docker services

Caddy and the Bluesky PDS run through systemd-managed Docker Compose projects.
Both require Docker to be running. Install their units as above, then enable
them:

```sh
sudo install -m 0644 config/services/caddy.service /etc/systemd/system/
sudo install -m 0644 config/services/pds.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable --now caddy.service pds.service
```

Caddy reads its Compose configuration from `/opt/sif.io/caddy/compose.yaml`,
mounts configuration from `/opt/sif.io/caddy/etc`, and stores persistent state
at `/opt/sif.io/caddy/data`. PDS persists data in `/pds` and reads its
environment from `/pds/pds.env`.

On the host, use `pdsadmin` to update the managed PDS/Caddy container stack. It
downloads the current PDS Compose file, pulls the PDS, Caddy, and Watchtower
images, and restarts PDS:

```sh
sudo pdsadmin update
```

Check the update using the systemd and container logs:

```sh
sudo journalctl --unit pds
sudo docker logs pds
```

The PDS Compose project also runs Watchtower, configured to check for updated
images nightly and remove old images after updates. Use the owning systemd
units to restart a service after changing its unit or mounted configuration.

`minecraft.service` is a standalone unit for an existing Bedrock installation
at `/opt/minecraft`; it runs as the `minecraft` user. Install it in the same
way and restart it after replacing the Bedrock server files:

```sh
sudo install -m 0644 config/services/minecraft.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable --now minecraft.service
sudo systemctl restart minecraft.service
```

### Kubernetes

The SMTP and website Dockerfiles build ARM64 images and run their test suites
during the image build. The Kubernetes manifests reference immutable image tags
through the `GITHUB.SHA` placeholder. A deployment pipeline must replace that
placeholder and the `INJECTED_*` values before applying the manifest; do not
commit live credentials to these files.

After publishing the image to `sifiocr.azurecr.io`, apply the rendered manifest
and monitor the rollout:

```sh
kubectl apply -f rendered-smtp.yml
kubectl rollout status deployment/sifio-smtp

kubectl apply -f rendered-www.yml
kubectl rollout status deployment/sifio-www
```
Wait for each rollout command to complete successfully before considering its update deployed.

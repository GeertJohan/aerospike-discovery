## CoreOS with flannel and cloud-config

This example shows how to setup an aerospike node on CoreOS by using a cloud-config.

To setup the node, there are several tasks/steps that need to be done.
 - dd the device to zero's on first boot
 - start aerospike-server
 - start aerospike-discovery

This configuration assumes:
 - An existing CoreOS cluster is already running providing services (etcd).
 - A volume/ssd is connected to the CoreOS instance.
 - A single aerospike instance is started on one machine
 - There is a flannel network in the CoreOS cluster.

Problems (TODO) with this setup:
 - restarting the aerospike container gives the node a new ip in the flannel network, and thereby a new identity in the cluster (will be added as new node in cluster).
 - check how reliable the `/home/core` folder is for the `.aerospike-clean-disk-once` file.

#cloud-config
```yaml
#cloud-config
coreos:
  etcd2:
    proxy: on
    discovery: <discovery url>
    advertise-client-urls: http://$private_ipv4:2379
    listen-client-urls: http://localhost:2379,http://$private_ipv4:2379
  fleet:
    etcd_servers: http://127.0.0.1:2379
    metadata: purpose=aerospike
  flannel:
    interface: $private_ipv4
    etcd_endpoints: http://127.0.0.1:2379
  units:
    - name: etcd2.service
      command: start
    - name: fleet.service
      command: start
    - name: flanneld.service
      command: start
      drop-ins:
        - name: 50-network-config.conf
          content: |
            [Service]
            ExecStartPre=/usr/bin/etcdctl set /coreos.com/network/config '{ "Network": "20.0.0.0/16" }'

    - name: clean-aerospike-data.service
      command: start
      content: |
        [Unit]
        Description=Clean (zero) the aerospike-data disk

        [Service]
        Type=oneshot
        RemainAfterExit=yes
        ExecStart=/opt/clean-aerospike-data.sh /dev/by-id/yourSSD

    - name: aerospike.service
      command: start
      content: |
        [Unit]
        Description=Aerospike
        After=flanneld.service
        Requires=flanneld.service
        After=clean-aerospike-data.service
        Requires=clean-aerospike-data.service

        [Service]
        Restart=always
        ExecStartPre=-/usr/bin/docker kill aerospike
        ExecStartPre=-/usr/bin/docker rm aerospike
        ExecStartPre=/bin/mkdir -p /opt/aerospike-workdir/smd
        ExecStartPre=/usr/bin/docker pull aerospike/aerospike-server
        ExecStart=/opt/aerospike-server.sh
        ExecStop=/usr/bin/docker stop -t 20 aerospike

    - name: aerospike-discovery.service
      command: start
      content: |
        [Unit]
        Description=Aerospike Discovery
        After=aerospike.service
        Requires=aerospike.service
        BindsTo=aerospike.service

        [Service]
        Restart=always
        ExecStartPre=-/usr/bin/docker kill aerospike-discovery
        ExecStartPre=-/usr/bin/docker rm aerospike-discovery
        ExecStartPre=/usr/bin/docker pull geertjohan/aerospike-discovery
        ExecStart=/opt/aerospike-discovery.sh
        ExecStop=/usr/bin/docker stop -t 5 aerospike-discovery

write_files:
  - path: /opt/aerospike/etc/aerospike.conf
    permissions: 0755
    owner: root
    content: |
      network {
        service {
          address any
          port 3000
        }
        fabric {
          address any
          port 3001
        }
        heartbeat {
          mode mesh
          port 3002

          interval 150
          timeout 20
        }
        info {
          address any
          port 3003
        }
      }

      service {
        work-directory /opt/aerospike-workdir
      }

      logging {
        file /dev/stdout {
          context any info
        }
      }

      namespace stuff {
        memory-size 4G
        replication-factor 2

        storage-engine device {
          device /dev/aerospike-data
          write-block-size 128k
        }

      }
  - path: /opt/clean-aerospike-data.sh
    permissions: 0755
    owner: root
    content: |
      #!/bin/bash
      cleanDevice="$1"
      cleanOnceFile="/home/core/.aerospike-clean-disk-once"
      if [ -f $cleanOnceFile ]; then
        echo "$cleanDevice is not being zero-ed again"
        exit 0
      fi
      while [ ! -b $cleanDevice ]
      do
        echo "waiting for $cleanDevice"
        sleep 1
      done
      /bin/dd if=/dev/zero of=$cleanDevice bs=1M conv=nocreat
      touch $cleanOnceFile
      exit 0
  - path: /opt/aerospike-server.sh
    permissions: 0755
    owner: root
    content: |
      #!/bin/bash
      devicepath=$(realpath /dev/by-id/yourSSD)
      /usr/bin/docker run --name aerospike \
        -v /opt/aerospike/etc:/opt/aerospike/etc \
        -v /opt/aerospike-workdir:/opt/aerospike-workdir \
        --device $devicepath:/dev/aerospike-data \
        aerospike/aerospike-server /usr/bin/asd --foreground --config-file /opt/aerospike/etc/aerospike.conf
  - path: /opt/aerospike-discovery.sh
    permissions: 0755
    owner: root
    content: |
      #!/bin/bash
      aerospike_id=$(docker inspect --format '{{ .Id }}' aerospike)
      aerospike_ipaddr=$(docker inspect --format '{{ .NetworkSettings.IPAddress }}' aerospike)
      local_etcd_ip=$(ip addr | grep 'eth0:' -A2 | tail -n1 | awk '{print $2}' | cut -f1 -d'/')
      /usr/bin/docker run --name aerospike-discovery geertjohan/aerospike-discovery \
        /usr/local/bin/aerospike-discovery \
        --local-aerospike-name="$aerospike_id" \
        --local-aerospike-ip="$aerospike_ipaddr" \
        --etcd-address="http://$local_etcd_ip:2379" \
        --logging-disable-timestamp
```
